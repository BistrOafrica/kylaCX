package main

import (
	"context"
	"fmt"
	"kyla-be/cmd/clients"
	"kyla-be/config"
	"kyla-be/internal/agentops"
	"kyla-be/internal/apps"
	"kyla-be/internal/audit"
	"kyla-be/internal/auth"
	"kyla-be/internal/ai"
	"kyla-be/internal/automation"
	"kyla-be/internal/automation/activities"
	"kyla-be/internal/campaigns"
	"kyla-be/internal/telephony"
	"kyla-be/internal/telephony/ivr"
	"kyla-be/internal/telephony/queues"
	"kyla-be/internal/telephony/xmlcurl"
	"kyla-be/internal/branch"
	casbinsvc "kyla-be/internal/casbin"
	"kyla-be/internal/communication"
	"kyla-be/internal/contact"
	"kyla-be/internal/crm"
	"kyla-be/internal/department"
	"kyla-be/internal/forms"
	"kyla-be/internal/invitation"
	"kyla-be/internal/knowledge"
	"kyla-be/internal/label"
	"kyla-be/internal/leave"
	"kyla-be/internal/middleware"
	"kyla-be/internal/objectcore"
	"kyla-be/internal/onboarding"
	"kyla-be/internal/organisation"
	"kyla-be/internal/projects"
	"kyla-be/internal/rbac"
	"kyla-be/internal/sharing"
	"kyla-be/internal/shift"
	"kyla-be/internal/tag"
	"kyla-be/internal/team"
	"kyla-be/internal/ticketing"
	"kyla-be/internal/user"
	"kyla-be/internal/workspace"
	"kyla-be/pkg/db"
	"kyla-be/pkg/pb"
	"kyla-be/pkg/service"
	"kyla-be/pkg/utils"
	"kyla-be/shared/events"
	"log"
	"net"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

const (
	secretKey            = "3ncRyP71on!"
	tokenDuration        = 10 * 24 * 60 * time.Minute // 10 days
	refreshTokenDuration = 10 * 24 * 60 * time.Minute // 10 days
)

// userGatewayAdapter bridges *user.UserStore to the apps.UserGateway interface.
// apps.UserGateway.FindByID returns *apps.UserProfile; user.UserStore.FindByID returns *user.User.
type userGatewayAdapter struct{ s *user.UserStore }

func (a *userGatewayAdapter) FindByID(id *uuid.UUID) (*apps.UserProfile, error) {
	u, err := a.s.FindByID(id)
	if err != nil {
		return nil, err
	}
	return &apps.UserProfile{Email: u.Email, FirstName: u.FirstName}, nil
}

func GinServer(
	onboardingHandler *onboarding.OnboardingHandler,
	webhookHandler *apps.WebhookHandler,
	waHandler *communication.WhatsAppHandler,
	smsWebhookHandler *communication.SMSWebhookHandler,
	webChatAdapter *communication.WebChatAdapter,
	xmlCurlHandler *xmlcurl.Handler,
) error {
	r := gin.Default()
	r.UseH2C = true // Enable H2C support for HTTP/2
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Onboarding REST API routes
	ob := r.Group("/api/v1/onboarding")
	{
		ob.POST("", onboardingHandler.CreateOnboarding)
		ob.GET("/:id", onboardingHandler.GetOnboarding)
		ob.PUT("/:id", onboardingHandler.UpdateOnboarding)
		ob.DELETE("/:id", onboardingHandler.DeleteOnboarding)
		ob.GET("", onboardingHandler.ListOnboardings)
	}

	// Webhook REST API routes — authenticated via API app Bearer token
	wh := r.Group("/api/v1/webhooks")
	{
		wh.POST("", webhookHandler.RegisterWebhook)
		wh.GET("", webhookHandler.ListWebhooks)
		wh.GET("/:id", webhookHandler.GetWebhook)
		wh.PUT("/:id", webhookHandler.UpdateWebhook)
		wh.DELETE("/:id", webhookHandler.DeleteWebhook)
	}

	// WhatsApp Business API webhook routes
	r.GET("/webhooks/whatsapp", waHandler.Verify)
	r.POST("/webhooks/whatsapp", waHandler.Receive)

	// SMS webhook routes (Twilio + Africa's Talking)
	if smsWebhookHandler != nil {
		r.POST("/webhooks/sms/twilio", smsWebhookHandler.ReceiveTwilio)
		r.POST("/webhooks/sms/africastalking", smsWebhookHandler.ReceiveAT)
	}

	// WebChat WebSocket endpoint
	r.GET("/ws/chat/:workspace_id", webChatAdapter.HandleWebSocket)

	// FreeSWITCH mod_xml_curl endpoint. Handler enforces RFC1918 source +
	// optional X-Kyla-XML-Token shared secret — no gRPC auth interceptor
	// here because FS is the caller, not a logged-in user.
	if xmlCurlHandler != nil {
		r.POST("/freeswitch/xml", xmlCurlHandler.Serve)
	}

	return r.Run(":8085")
}

func LoadCertificate(environment string) credentials.TransportCredentials {
	switch environment {
	case "production":
		creds, err := credentials.NewServerTLSFromFile("certs/prod/cert.pem", "certs/prod/key.pem")
		if err != nil {
			log.Fatalf("Failed to generate credentials %v", err)
		}
		log.Printf("Loaded production certificates %v", creds)
		return creds
	case "staging":
		creds, err := credentials.NewServerTLSFromFile("certs/staging/cert.pem", "certs/staging/key.pem")
		if err != nil {
			log.Fatalf("Failed to generate credentials %v", err)
		}
		log.Printf("Loaded staging certificates %v", creds)
		return creds
	case "development":
		creds, err := credentials.NewServerTLSFromFile("certs/dev/cert.pem", "certs/dev/key.pem")
		if err != nil {
			log.Fatalf("Failed to generate dev credentials %v", err)
		}
		log.Printf("Loaded dev certificates %v", creds)
		return creds
	case "local":
		return insecure.NewCredentials()
	default:
		log.Printf("Loaded insecure certificates %v", insecure.NewCredentials())
		return insecure.NewCredentials()
	}
}

func main() {
	configs, err := config.LoadConfig()
	if err != nil {
		log.Fatalln("Failed at config", err)
	}

	db.InitDB(&configs.PostgresConfig)
	if migrateErr := db.RunSQLMigrations(db.DB, "pkg/db/migrations"); migrateErr != nil {
		log.Printf("sql migration warning: %v", migrateErr)
	}

	// AutoMigrate workspace + audit + webhook + objectcore tables (idempotent — safe to run on every start).
	if migrateErr := db.DB.AutoMigrate(
		&workspace.Workspace{},
		&workspace.WorkspaceMember{},
		&audit.AuditLog{},
		&apps.Webhook{},
		&objectcore.ObjectType{},
		&objectcore.Object{},
		&objectcore.ObjectRelation{},
		&objectcore.ObjectEvent{},
		&objectcore.SavedView{},
		&communication.Conversation{},
		&communication.Message{},
		&communication.RoutingRule{},
		&communication.SLAPolicy{},
		&communication.SLARecord{},
		// Phase 4: CRM, Ticketing, Knowledge Base, Forms
		&crm.Pipeline{},
		&crm.PipelineStage{},
		&ticketing.TicketRoom{},
		&ticketing.TicketRoomMessage{},
		&ticketing.Macro{},
		&knowledge.KBCategory{},
		&knowledge.KBArticle{},
		&forms.Form{},
		&forms.FormSubmission{},
		&projects.Project{},
		// Phase 6: Automation workflow definitions + Temporal run projections
		&automation.Workflow{},
		&automation.WorkflowRun{},
		// Phase 6: Campaigns + per-recipient state + WhatsApp template mirror
		&campaigns.Campaign{},
		&campaigns.CampaignRecipient{},
		&campaigns.WhatsAppTemplate{},
		// Phase 5: Telephony — self-hosted SIP via FreeSWITCH
		&telephony.Call{},
		&telephony.CallEvent{},
		&telephony.SipDomain{},
		&telephony.SipExtension{},
		&telephony.SipTrunk{},
		// Phase 5c: IVR flows + DID mappings + run history
		&ivr.Flow{},
		&ivr.DIDMapping{},
		&ivr.Run{},
		// Phase 5d: call queues + agent membership + live entries
		&queues.Queue{},
		&queues.Membership{},
		&queues.Entry{},
		// Phase 5e: recording upload + transcription pipeline
		&telephony.CallRecording{},
	); migrateErr != nil {
		log.Printf("migration warning: %v", migrateErr)
	}

	// ── Casbin enforcer ──────────────────────────────────────────────────────
	// Graceful degradation: all protected routes will be denied if Casbin fails
	// to initialise rather than accidentally left open.
	var casbinEnforcer *casbinsvc.Enforcer
	if ce, ceErr := casbinsvc.NewEnforcer(db.DB); ceErr != nil {
		log.Printf("casbin enforcer init warning (RBAC disabled): %v", ceErr)
	} else {
		casbinEnforcer = ce
		log.Println("casbin enforcer initialised")
	}

	log.Printf("Starting server on port %v\n", configs.EnvConfigs.Port)

	// ── Event bus (NATS JetStream) ────────────────────────────────────────────
	// Graceful degradation: server starts with NoopBus if NATS is not available.
	var eventBus events.Bus
	natsBus, natsErr := events.NewNatsBus(configs.EnvConfigs.NatsURL)
	if natsErr != nil {
		log.Printf("NATS unavailable, using NoopBus: %v", natsErr)
		eventBus = &events.NoopBus{}
	} else {
		eventBus = natsBus
		defer natsBus.Close()
	}

	// ── Temporal client (automation engine backbone) ─────────────────────────
	// Graceful degradation: server starts without Temporal if TEMPORAL_HOST_PORT is unset.
	// When unset, the automation gRPC service can still persist workflow definitions but
	// no executions will run.
	temporalClient, tErr := automation.NewTemporalClient(automation.ClientConfig{
		HostPort:  configs.EnvConfigs.TemporalHostPort,
		Namespace: configs.EnvConfigs.TemporalNamespace,
		TaskQueue: configs.EnvConfigs.TemporalTaskQueue,
	})
	if tErr != nil {
		log.Printf("temporal unavailable, automation execution disabled: %v", tErr)
	}
	if temporalClient != nil {
		defer temporalClient.Close()
	}

	grpcClients, err := clients.NewClients(configs.Clients.LeadsAddress)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// ── Domain stores ────────────────────────────────────────────────────────
	domRbacStore := rbac.NewRbacStore(db.DB)
	domUserStore := user.NewUserStore(db.DB)
	domOrgStore := organisation.NewOrganisationStore(db.DB)
	domBranchStore := branch.NewBranchStore(db.DB)
	domDeptStore := department.NewDepartmentStore(db.DB)
	domTeamStore := team.NewTeamStore(db.DB)
	domAgentStore := agentops.NewAgentStatusStore(db.DB)
	domAppStore := apps.NewAPIAppStore(db.DB)
	domInvStore := invitation.NewInvitationStore(db.DB)
	domObStore := onboarding.NewOnboardingStore(db.DB)
	wsStore := workspace.NewWorkspaceStore(db.DB)
	ocStore := objectcore.NewObjectCoreStore(db.DB)
	viewStore := objectcore.NewViewStore(db.DB)
	convStore := communication.NewConversationStore(db.DB)
	msgStore := communication.NewMessageStore(db.DB)
	routingStore := communication.NewRoutingStore(db.DB)
	slaStore := communication.NewSLAStore(db.DB)

	contactStore := contact.NewContactStore(db.DB)
	contactGroupStore := contact.NewContactGroupStore(db.DB)
	domSharingStore := sharing.NewSharingStore(db.DB)
	tagStore := tag.NewTagStore(db.DB)
	labelStore := label.NewLabelStore(db.DB)
	leaveStore := leave.NewLeaveStore(db.DB)
	shiftStore := shift.NewShiftStore(db.DB)
	scheduleStore := shift.NewShiftScheduleStore(db.DB)
	breakStore := shift.NewBreakStore(db.DB)

	// Phase 4 stores
	crmStore := crm.NewCRMStore(db.DB)
	ticketingStore := ticketing.NewTicketingStore(db.DB)
	knowledgeStore := knowledge.NewKnowledgeStore(db.DB)
	formsStore := forms.NewFormsStore(db.DB)
	projectStore := projects.NewStore(db.DB)

	// ── Service-layer stores for auth infrastructure + department bridge ────
	// service.NewAuthStore requires the concrete service-layer types.
	// department.NewDepartmentServer still uses *service.BranchStore and
	// *service.UserStore; contact.BranchStore and contact.SharingStore are
	// type aliases for the service-layer types, so these are reused there too.
	svcRbacStore := service.NewRbacStore(db.DB)
	svcOrgStore := service.NewOrganisationStore(db.DB)
	svcUserStore := service.NewUserStore(db.DB)
	svcAppStore := service.NewAPIAppStore(db.DB)
	svcBranchStore := service.NewBranchStore(db.DB)   // contact.BranchStore = service.BranchStore
	svcSharingStore := service.NewSharingStore(db.DB) // contact.SharingStore = service.SharingStore

	// ── Auth infrastructure ──────────────────────────────────────────────────
	resendService := utils.NewResendService(
		configs.RsConfig.ResendApiKey,
		configs.RsConfig.ResendFromEmail,
		configs.RsConfig.ResendSupportEmail,
		configs.RsConfig.ResendBaseURL,
	)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     configs.RedisConfig.Addr,
		Password: configs.RedisConfig.Password,
		DB:       configs.RedisConfig.DB,
	})

	// jwtManager type is auth.JWTManager (= service.JWTManager via alias).
	jwtManager := service.NewJWTManager(secretKey, tokenDuration, refreshTokenDuration, svcRbacStore)
	svcAuthStore := service.NewAuthStore(svcRbacStore, svcOrgStore, jwtManager, svcAppStore, svcBranchStore, svcUserStore, redisClient)
	authAdaptor := auth.NewAuthStore(svcAuthStore, domRbacStore)
	// dbAuthStore type is auth.DBAuthStore (= service.DBAuthStore via alias).
	dbAuthStore := service.NewDBAuthStore(db.DB)

	firebaseAuth, err := service.NewFirebaseAuthService(&configs.FirebaseConfig, svcUserStore)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase Auth Service: %v", err)
	}

	webAuthn, _ := service.NewWebAuthn(&service.WebAuthnConfig{
		RPID:          configs.WebAuthnConfig.RPID,
		RPOrigin:      configs.WebAuthnConfig.RPOrigin,
		RPDisplayName: configs.WebAuthnConfig.RPDisplayName,
	})

	// ── gRPC servers ─────────────────────────────────────────────────────────
	// authServer stays in pkg/service — it owns JWT, Firebase, WebAuthn, passkey logic.
	authServer := service.NewAuthServer(svcUserStore, jwtManager, svcAuthStore, svcRbacStore, resendService, firebaseAuth, webAuthn, dbAuthStore)

	orgServer := organisation.NewOrganisationServer(domOrgStore, authAdaptor, db.DB, domRbacStore, domUserStore, domAgentStore, resendService, wsStore, casbinEnforcer)
	userServer := user.NewUserServer(domUserStore, authAdaptor, domRbacStore, domAgentStore, resendService)
	rbacServer := rbac.NewRbacServer(domRbacStore, authAdaptor, domUserStore, db.DB, casbinEnforcer)
	// svcRbacStore satisfies branch.RoleSaver (SaveRole takes *service.Role).
	branchServer := branch.NewBranchServer(domBranchStore, authAdaptor, svcRbacStore)
	// department still uses service-layer BranchStore and UserStore internally.
	deptServer := department.NewDepartmentServer(domDeptStore, authAdaptor, svcBranchStore, svcUserStore)
	teamServer := team.NewTeamServer(domTeamStore, authAdaptor)
	agentServer := agentops.NewAgentStatusServer(authAdaptor, domAgentStore)
	// appsServer uses the adapter so FindByID returns *apps.UserProfile.
	appsServer := apps.NewAppServer(authAdaptor, domAppStore, &userGatewayAdapter{domUserStore}, resendService)
	wsServer := workspace.NewWorkspaceServer(wsStore, authAdaptor, eventBus, casbinEnforcer, ocStore)
	ocServer := objectcore.NewObjectCoreServer(ocStore, authAdaptor, eventBus)
	viewServer := objectcore.NewViewServer(viewStore, authAdaptor)

	// ── Routing & SLA Engines ────────────────────────────────────────────────
	agentLB := communication.NewAgentLoadBalancer(redisClient)
	router := communication.NewRouter(routingStore, convStore, agentLB)
	slaEngine := communication.NewSLAEngine(slaStore, convStore, eventBus, configs.EnvConfigs.SLAScanIntervalSecs)

	// ── Channel Adapters ─────────────────────────────────────────────────────
	var smsProvider communication.SMSProvider
	switch configs.EnvConfigs.SMSProvider {
	case "twilio":
		smsProvider = communication.NewTwilioProvider(
			configs.EnvConfigs.TwilioAccountSID,
			configs.EnvConfigs.TwilioAuthToken,
			configs.EnvConfigs.TwilioFrom,
		)
	case "africastalking", "at":
		smsProvider = communication.NewAfricasTalkingProvider(
			configs.EnvConfigs.AfricasTalkingAPIKey,
			configs.EnvConfigs.AfricasTalkingUsername,
			configs.EnvConfigs.AfricasTalkingFrom,
		)
	default:
		log.Printf("No SMS provider configured, SMS adapter disabled")
	}

	// Create tenant resolver for webhook handlers (Recommendation #1: Tenant extraction)
	// TODO: Add contact store implementation for proper tenant resolution
	tenantResolver := communication.NewTenantResolver(
		nil, // contactStore - needs to be implemented
		configs.EnvConfigs.WhatsAppWebhookSecret,
		configs.EnvConfigs.TwilioAuthToken, // TODO: Use dedicated TWILIO_WEBHOOK_SECRET
	)

	var smsAdapter *communication.SMSAdapter
	if smsProvider != nil {
		smsAdapter = communication.NewSMSAdapter(smsProvider)
	}

	emailAdapter := communication.NewEmailAdapter(
		configs.EnvConfigs.SMTPHost, configs.EnvConfigs.SMTPPort,
		configs.EnvConfigs.SMTPUser, configs.EnvConfigs.SMTPPassword,
		configs.EnvConfigs.IMAPHost, configs.EnvConfigs.IMAPPort,
		configs.EnvConfigs.IMAPUser, configs.EnvConfigs.IMAPPassword,
		configs.EnvConfigs.IMAPPollIntervalSecs,
		convStore, msgStore, ocStore, eventBus, router, tenantResolver,
	)

	webChatAdapter := communication.NewWebChatAdapter(
		convStore, msgStore, ocStore, eventBus, router,
	)

	whatsAppAdapter := communication.NewWhatsAppAdapter(
		configs.EnvConfigs.WhatsAppAccessToken,
		configs.EnvConfigs.WhatsAppPhoneNumberID,
	)

	voiceBridge := communication.NewVoiceCallBridge(
		convStore, msgStore, ocStore, eventBus, eventBus,
	)

	// Build adapter registry
	var adapters []communication.ChannelAdapter
	if smsAdapter != nil {
		adapters = append(adapters, smsAdapter)
	}
	adapters = append(adapters, emailAdapter, webChatAdapter, whatsAppAdapter)

	adapterRegistry := communication.NewAdapterRegistry(adapters...)

	// Create streaming server (requires StreamBus interface)
	var streamBus events.StreamBus = natsBus
	if natsBus == nil {
		streamBus = &events.NoopBus{}
	}
	streamingServer := communication.NewStreamingServer(streamBus, convStore, msgStore, eventBus)

	// Create conversation server with all dependencies
	convServer := communication.NewConversationServer(
		convStore, msgStore, ocStore, authAdaptor, eventBus, adapterRegistry,
		streamingServer, slaEngine, slaStore, router, routingStore,
	)

	// SMS webhook handler
	var smsWebhookHandler *communication.SMSWebhookHandler
	if smsAdapter != nil {
		smsWebhookHandler = communication.NewSMSWebhookHandler(
			convStore, msgStore, ocStore, eventBus, router, tenantResolver,
		)
	}

	waHandler := communication.NewWhatsAppHandler(
		convStore, msgStore, ocStore, eventBus,
		configs.EnvConfigs.WhatsAppWebhookSecret,
		configs.EnvConfigs.WhatsAppVerifyToken,
	)

	domWebhookStore := apps.NewWebhookStore(db.DB)
	webhookHandler := apps.NewWebhookHandler(domWebhookStore, domAppStore)

	contactServer := contact.NewContactServer(contactStore, authAdaptor, svcBranchStore, contactGroupStore, grpcClients.LeadsClient, svcSharingStore)
	cgServer := contact.NewContactGroupServer(contactGroupStore, authAdaptor, contactStore)
	sharingServer := sharing.NewSharingServer(domSharingStore, authAdaptor)
	tagServer := tag.NewTagService(tagStore, authAdaptor)
	labelServer := label.NewLabelService(labelStore, authAdaptor)
	leaveServer := leave.NewLeaveServer(leaveStore, authAdaptor)
	shiftServer := shift.NewShiftServer(shiftStore, authAdaptor)
	schedServer := shift.NewScheduleServer(scheduleStore, authAdaptor)
	breakServer := shift.NewBreakServer(breakStore, authAdaptor)
	invServer := invitation.NewInvitationServer(
		domInvStore, resendService,
		domOrgStore, domBranchStore, domDeptStore, domTeamStore,
		domRbacStore, domUserStore,
		authAdaptor, domAgentStore,
		"app.kyla.cx",
	)
	obHandler := onboarding.NewOnboardingHandler(domObStore)
	obServer := onboarding.NewOnboardingServiceServer(db.DB)

	// Phase 4 servers
	crmServer := crm.NewCRMServer(crmStore, authAdaptor, eventBus)
	ticketingServer := ticketing.NewTicketingServer(ticketingStore, authAdaptor, eventBus)
	knowledgeServer := knowledge.NewKnowledgeServer(knowledgeStore, authAdaptor, eventBus)
	formsServer := forms.NewFormsServer(formsStore, authAdaptor, eventBus)
	projectServer := projects.NewProjectServer(projectStore, authAdaptor, eventBus)

	// Phase 6: AI engine — minimal LLMProvider used by both the gRPC
	// AIService and the automation `run_ai_skill` activity. Provider is
	// selected by LLM_PROVIDER env var; falls back to noop when keys are
	// missing so the binary always boots.
	aiProvider := ai.NewProvider(ai.ProviderConfig{
		Provider:        configs.EnvConfigs.LLMProvider,
		OpenAIAPIKey:    configs.EnvConfigs.OpenAIAPIKey,
		OpenAIModel:     configs.EnvConfigs.OpenAIModel,
		OpenAIBaseURL:   configs.EnvConfigs.OpenAIBaseURL,
		AnthropicAPIKey: configs.EnvConfigs.AnthropicAPIKey,
		AnthropicModel: configs.EnvConfigs.AnthropicModel,
	})
	aiServer := ai.NewServer(aiProvider)

	// Phase 6: Automation workflow service.
	// Workflow definitions live here; Temporal owns execution state.
	automationStore := automation.NewStore(db.DB)

	// Executor wraps the Temporal client for workflow starts. Nil-safe — when
	// Temporal is unavailable the executor reports !Enabled() and the gRPC
	// server returns FailedPrecondition for TestRunWorkflow.
	automationTaskQueue := configs.EnvConfigs.TemporalTaskQueue
	if automationTaskQueue == "" {
		automationTaskQueue = "kyla-automation"
	}
	automationExecutor := automation.NewExecutor(temporalClient, automationStore, automationTaskQueue)
	automationServer := automation.NewServer(automationStore, authAdaptor, automationExecutor)

	// Temporal worker — runs in the same binary per the architectural decision.
	// Registers AutomationWorkflow + all activities on the task queue.
	// Notifier and AI deps are wired in step 7 (and when notification ships);
	// activities gracefully degrade when these are nil.
	automationWorker, awErr := automation.StartWorker(temporalClient, automationTaskQueue, activities.Deps{
		ObjectStore:       ocStore,
		AdapterRegistry:   adapterRegistry,
		ConversationStore: convStore,
		MessageStore:      msgStore,
		SLATimer:          slaEngine,
		AI:                &ai.ActivityAdapter{Provider: aiProvider},
	})
	if awErr != nil {
		log.Printf("automation worker start failed: %v", awErr)
	}
	if automationWorker != nil {
		defer automationWorker.Stop()
	}

	// NATS consumer — fans matching domain events into the Executor so
	// Temporal workflows fire automatically. Uses a queue group so multiple
	// server instances share the load instead of duplicating starts.
	automationConsumer := automation.NewConsumer(automationStore, automationExecutor, eventBus)
	if acErr := automationConsumer.Start(context.Background()); acErr != nil {
		log.Printf("automation consumer start failed: %v", acErr)
	}
	defer automationConsumer.Stop()

	// Phase 6: Campaigns — channel-agnostic broadcast engine.
	// Shares the kyla-automation task queue with the automation worker but
	// runs in its own Temporal worker so a slow audience resolution can't
	// starve workflow execution.
	campaignsStore := campaigns.NewStore(db.DB)
	campaignsExecutor := campaigns.NewExecutor(temporalClient, automationTaskQueue)
	campaignsServer := campaigns.NewServer(campaignsStore, authAdaptor, campaignsExecutor)
	campaignsWorker, cwErr := campaigns.StartWorker(temporalClient, automationTaskQueue, campaigns.ActivityDeps{
		Store:             campaignsStore,
		ObjectStore:       ocStore,
		ConversationStore: convStore,
		MessageStore:      msgStore,
		AdapterRegistry:   adapterRegistry,
	})
	if cwErr != nil {
		log.Printf("campaigns worker start failed: %v", cwErr)
	}
	if campaignsWorker != nil {
		defer campaignsWorker.Stop()
	}

	// Phase 5: Telephony — self-hosted SIP via FreeSWITCH.
	// The PBX controller dials ESL on startup; if FS isn't reachable the
	// controller logs and stays disabled, the binary still boots, and gRPC
	// calls return FailedPrecondition (NoopPBX behaviour).
	telephonyStore := telephony.NewStore(db.DB)
	telephonyEventStream := telephony.NewCallEventStream(0)
	var telephonyPBX telephony.PBXController = telephony.NoopPBX{}
	if configs.EnvConfigs.FSEslHost != "" {
		fs := telephony.NewFreeSWITCHController(telephony.FreeSWITCHConfig{
			Host:     configs.EnvConfigs.FSEslHost,
			Port:     configs.EnvConfigs.FSEslPort,
			Password: configs.EnvConfigs.FSEslPassword,
		}, telephonyEventStream)
		if startErr := fs.Start(context.Background()); startErr != nil {
			log.Printf("freeswitch controller start failed: %v", startErr)
		}
		telephonyPBX = fs
		defer fs.Stop()
	}
	telephonyBridge := telephony.NewEventBridge(telephonyStore, eventBus, telephonyEventStream)

	// Phase 5c: IVR engine — node-based flow executor driven by PBX events.
	// The executor satisfies telephony.IVRHook; attaching it to the bridge
	// routes inbound DIDs into IVR flows and forwards playback/DTMF events.
	ivrStore := ivr.NewStore(db.DB)
	ivrExecutor := ivr.NewExecutor(ivrStore, telephonyPBX)
	telephonyBridge.AttachIVR(&ivrBridgeAdapter{store: ivrStore, exec: ivrExecutor})
	ivrServer := ivr.NewServer(ivrStore, authAdaptor)

	// Phase 5d: call queues — agent routing engine + wallboard data plane.
	// The router subscribes to bridge events (CHANNEL_ANSWER on agent legs,
	// CHANNEL_HANGUP on either leg) and originates per-agent legs via the
	// shared PBX controller.
	queuesStore := queues.NewStore(db.DB)
	queuesRouter := queues.NewRouter(queuesStore, telephonyPBX, telephonyStore, configs.EnvConfigs.FSDefaultTrunk)
	telephonyBridge.AttachQueues(queuesRouter)
	queuesServer := queues.NewServer(queuesStore, authAdaptor)

	// Phase 5e: recording upload + transcription pipeline.
	// Nil-safe: when RECORDINGS_BUCKET is empty NewRecordingPipeline returns
	// (nil, nil) and AttachRecordings keeps the bridge in legacy mode (file
	// path stamped on the call row, no S3, no transcript).
	recordingsPipeline, recErr := telephony.NewRecordingPipeline(
		context.Background(),
		telephony.RecordingPipelineConfig{
			AWSRegion:    configs.AwsCredentialsConfig.AwsRegion,
			AWSAccessKey: configs.AwsCredentialsConfig.AwsAccessKey,
			AWSSecretKey: configs.AwsCredentialsConfig.AwsSecretKey,
			Bucket:       configs.EnvConfigs.RecordingsBucket,
			KeyPrefix:    configs.EnvConfigs.RecordingsKeyPrefix,
			S3Endpoint:   configs.EnvConfigs.RecordingsS3Endpoint,
		},
		telephonyStore,
		&ai.TranscriberAdapter{Provider: aiProvider},
	)
	if recErr != nil {
		log.Printf("recording pipeline init failed: %v", recErr)
	}
	if recordingsPipeline != nil {
		telephonyBridge.AttachRecordings(recordingsPipeline)
	}

	// mod_xml_curl handler — FreeSWITCH POSTs directory/dialplan/configuration
	// lookups to /freeswitch/xml; the handler serves XML from the Postgres
	// control plane (sip_extensions + ivr_did_mappings + sip_trunks).
	xmlCurlHandler := xmlcurl.NewHandler(
		telephonyStore,
		ivrStore,
		configs.EnvConfigs.FSSipRealm,
		configs.EnvConfigs.FSWssURL,
		configs.EnvConfigs.FSXmlCurlToken,
	)

	go telephonyBridge.Start(context.Background())

	telephonyIssuer := telephony.NewJWTTokenIssuer(configs.EnvConfigs.JwtSecret, "kyla-be")
	telephonyServer := telephony.NewServer(
		telephonyStore, authAdaptor, telephonyPBX, eventBus, telephonyIssuer,
		telephony.ServerConfig{
			WssURL:       configs.EnvConfigs.FSWssURL,
			SipRealm:     configs.EnvConfigs.FSSipRealm,
			TurnURL:      configs.EnvConfigs.TurnURL,
			TurnUsername: configs.EnvConfigs.TurnUsername,
			TurnPassword: configs.EnvConfigs.TurnPassword,
		},
	)

	// ── Interceptors ─────────────────────────────────────────────────────────
	interceptor := middleware.NewAuthInterceptor(jwtManager, authAdaptor, casbinEnforcer)
	devicesInterceptor := middleware.NewSessionDevicesInterceptors(dbAuthStore)

	sqsClient := utils.NewSQSClient(
		configs.AwsCredentialsConfig.AwsRegion,
		configs.AwsCredentialsConfig.AwsAccessKey,
		configs.AwsCredentialsConfig.AwsSecretKey,
	)
	sqsAction := utils.NewSQSActions(sqsClient)
	responseInterceptor := middleware.NewResponseInterceptor(sqsAction)

	auditStore := audit.NewStore(db.DB)
	auditInterceptor := audit.NewInterceptor(auditStore)

	// ── gRPC server ───────────────────────────────────────────────────────────
	certificates := LoadCertificate(configs.EnvConfigs.Environment)
	excludedMethods := []string{
		"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo",
	}

	grpcServer := grpc.NewServer(
		grpc.Creds(certificates),
		grpc.ChainUnaryInterceptor(
			interceptor.Unary(),
			responseInterceptor.WithExcludedMethods(excludedMethods...).Unary(),
			devicesInterceptor.Unary(),
			auditInterceptor.Unary(),
		),
		grpc.ChainStreamInterceptor(interceptor.Stream(), responseInterceptor.Stream(), auditInterceptor.Stream()),
	)

	reflection.Register(grpcServer)

	pb.RegisterAuthServiceServer(grpcServer, authServer)
	pb.RegisterRoleServiceServer(grpcServer, rbacServer)
	pb.RegisterPermissionServiceServer(grpcServer, rbacServer)
	pb.RegisterOrganisationServiceServer(grpcServer, orgServer)
	pb.RegisterUserServiceServer(grpcServer, userServer)
	pb.RegisterBranchServiceServer(grpcServer, branchServer)
	pb.RegisterContactServiceServer(grpcServer, contactServer)
	pb.RegisterGroupServiceServer(grpcServer, cgServer)
	pb.RegisterTagServiceServer(grpcServer, tagServer)
	pb.RegisterLabelServiceServer(grpcServer, labelServer)
	pb.RegisterDepartmentServiceServer(grpcServer, deptServer)
	pb.RegisterAgentStatusServiceServer(grpcServer, agentServer)
	pb.RegisterAppServiceServer(grpcServer, appsServer)
	pb.RegisterLeaveServiceServer(grpcServer, leaveServer)
	pb.RegisterResourceSharingServer(grpcServer, sharingServer)
	pb.RegisterShiftServiceServer(grpcServer, shiftServer)
	pb.RegisterShiftScheduleServiceServer(grpcServer, schedServer)
	pb.RegisterBreakServiceServer(grpcServer, breakServer)
	pb.RegisterInvitationServiceServer(grpcServer, invServer)
	pb.RegisterTeamServiceServer(grpcServer, teamServer)
	pb.RegisterOnboardingServiceServer(grpcServer, obServer)
	pb.RegisterWorkspaceServiceServer(grpcServer, wsServer)
	pb.RegisterObjectCoreServiceServer(grpcServer, ocServer)
	pb.RegisterViewServiceServer(grpcServer, viewServer)
	pb.RegisterConversationServiceServer(grpcServer, convServer)
	pb.RegisterCRMServiceServer(grpcServer, crmServer)
	pb.RegisterTicketingServiceServer(grpcServer, ticketingServer)
	pb.RegisterKnowledgeServiceServer(grpcServer, knowledgeServer)
	pb.RegisterFormsServiceServer(grpcServer, formsServer)
	pb.RegisterProjectServiceServer(grpcServer, projectServer)
	pb.RegisterWorkflowServiceServer(grpcServer, automationServer)
	pb.RegisterAIServiceServer(grpcServer, aiServer)
	pb.RegisterCampaignServiceServer(grpcServer, campaignsServer)
	pb.RegisterTelephonyServiceServer(grpcServer, telephonyServer)
	pb.RegisterIVRServiceServer(grpcServer, ivrServer)
	pb.RegisterQueueServiceServer(grpcServer, queuesServer)

	// ── Run gRPC + HTTP servers + Background Services ────────────────────────
	// Create cancellation context for graceful shutdown of background services
	bgCtx := context.Background()

	var wg sync.WaitGroup
	wg.Add(5) // gRPC + HTTP + SLA scanner + Voice bridge + Email IMAP poller

	go func() {
		defer wg.Done()
		lis, err := net.Listen("tcp", fmt.Sprintf(":%v", configs.EnvConfigs.Port))
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		log.Printf("gRPC server started on port %v", configs.EnvConfigs.Port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		log.Println("Starting health check and REST API server on port 8085")
		if err := GinServer(obHandler, webhookHandler, waHandler, smsWebhookHandler, webChatAdapter, xmlCurlHandler); err != nil {
			log.Fatalf("failed to start health check server: %v", err)
		}
	}()

	// Start SLA breach scanner
	go func() {
		defer wg.Done()
		log.Println("Starting SLA breach scanner")
		slaEngine.Start(bgCtx)
	}()

	// Start Voice call bridge (NATS consumer)
	go func() {
		defer wg.Done()
		log.Println("Starting Voice call bridge")
		if err := voiceBridge.Start(bgCtx); err != nil {
			log.Printf("Voice bridge error: %v", err)
		}
	}()

	// Start Email IMAP poller
	go func() {
		defer wg.Done()
		log.Println("Starting Email IMAP poller")
		emailAdapter.StartIMAPPoller(bgCtx)
	}()

	wg.Wait()
}

// ivrBridgeAdapter satisfies telephony.IVRHook by composing the IVR store's
// DID lookup with the executor's StartForCall/Advance methods. Defined here
// (rather than inside the ivr package) so the executor stays free of any
// dependency on the telephony package, avoiding a circular import: telephony
// defines IVRHook; ivr provides the Executor; main.go glues them together.
type ivrBridgeAdapter struct {
	store *ivr.Store
	exec  *ivr.Executor
}

func (a *ivrBridgeAdapter) LookupFlowForDID(did string) (string, string, string, error) {
	return a.store.FindFlowIDForDID(did)
}

func (a *ivrBridgeAdapter) StartForCall(ctx context.Context, callUUID, flowID, orgID, workspaceID string) (string, error) {
	return a.exec.StartForCall(ctx, callUUID, flowID, orgID, workspaceID)
}

func (a *ivrBridgeAdapter) Advance(ctx context.Context, callUUID string, eventType telephony.PBXEventType, input string) {
	a.exec.Advance(ctx, callUUID, eventType, input)
}

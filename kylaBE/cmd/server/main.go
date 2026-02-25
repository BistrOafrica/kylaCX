package main

import (
	"fmt"
	"kyla-be/cmd/clients"
	"kyla-be/config"
	"kyla-be/pkg/db"
	"kyla-be/pkg/pb"
	"kyla-be/pkg/service"
	"kyla-be/pkg/utils"
	"log"
	"net"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
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

func GinServer(onboardingHandler *service.OnboardingHandler) error {
	r := gin.Default()
	r.UseH2C = true // Enable H2C support for HTTP/2
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Onboarding REST API routes
	onboarding := r.Group("/api/v1/onboarding")
	{
		onboarding.POST("", onboardingHandler.CreateOnboarding)
		onboarding.GET("/:id", onboardingHandler.GetOnboarding)
		onboarding.PUT("/:id", onboardingHandler.UpdateOnboarding)
		onboarding.DELETE("/:id", onboardingHandler.DeleteOnboarding)
		onboarding.GET("", onboardingHandler.ListOnboardings)
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

	log.Printf("Starting server on port %v\n", configs.EnvConfigs.Port)

	clients, err := clients.NewClients(configs.Clients.LeadsAddress)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	userStore := service.NewUserStore(db.DB)
	teamStore := service.NewTeamStore(db.DB)
	orgStore := service.NewOrganisationStore(db.DB)
	contactStore := service.NewContactStore(db.DB)
	branchStore := service.NewBranchStore(db.DB)
	departmentStore := service.NewDepartmentStore(db.DB)
	rbacStore := service.NewRbacStore(db.DB)
	appStore := service.NewAPIAppStore(db.DB)
	sharingStore := service.NewSharingStore(db.DB)
	jwtManager := service.NewJWTManager(secretKey, tokenDuration, refreshTokenDuration, rbacStore)
	resendService := utils.NewResendService(
		configs.RsConfig.ResendApiKey,
		configs.RsConfig.ResendFromEmail,
		configs.RsConfig.ResendSupportEmail,
		configs.RsConfig.ResendBaseURL,
	)

	// Shift related stores
	shiftStore := service.NewShiftStore(db.DB)
	scheduleStore := service.NewShiftScheduleStore(db.DB)
	breakStore := service.NewBreakStore(db.DB)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     configs.RedisConfig.Addr,
		Password: configs.RedisConfig.Password,
		DB:       configs.RedisConfig.DB,
	})

	authStore := service.NewAuthStore(rbacStore, orgStore, jwtManager, appStore, branchStore, userStore, redisClient)
	contactGroupStore := service.NewContactGroupStore(db.DB)
	tagStore := service.NewTagStore(db.DB)
	labelStore := service.NewLabelStore(db.DB)
	agentStatusStore := service.NewAgentStatusStore(db.DB)
	leaveStore := service.NewLeaveStore(db.DB)
	invitationStore := service.NewInvitationStore(db.DB)
	dbAuthStore := service.NewDBAuthStore(db.DB)
	orgServer := service.NewOrganisationServer(
		orgStore,
		authStore,
		branchStore,
		rbacStore,
		userStore,
		agentStatusStore,
		resendService,
	)
	userServer := service.NewUserServer(userStore, authStore, agentStatusStore, resendService)
	contactServer := service.NewContactServer(
		contactStore,
		authStore,
		branchStore,
		contactGroupStore,
		clients.LeadsClient,
		sharingStore,
	)
	contactGroupServer := service.NewContactGroupServer(contactGroupStore, authStore, contactStore)

	// Initialize Firebase Auth Service
	firebaseAuth, err := service.NewFirebaseAuthService(&configs.FirebaseConfig, userStore)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase Auth Service: %v", err)
	}

	webAuthn, _ := service.NewWebAuthn(&service.WebAuthnConfig{
		RPID:          configs.WebAuthnConfig.RPID,
		RPOrigin:      configs.WebAuthnConfig.RPOrigin,
		RPDisplayName: configs.WebAuthnConfig.RPDisplayName,
	})

	rbacServer := service.NewRbacServer(rbacStore, authStore, userStore, branchStore, departmentStore, teamStore)
	tagServer := service.NewTagService(tagStore, authStore)
	labelServer := service.NewLabelService(labelStore, authStore)
	authServer := service.NewAuthServer(userStore, jwtManager, authStore, rbacStore, resendService, firebaseAuth, webAuthn, dbAuthStore)
	branchServer := service.NewBranchServer(branchStore, authStore)
	departmentServer := service.NewDepartmentServer(departmentStore, authStore, branchStore, userStore)
	agentStatusServer := service.NewAgentStatusServer(authStore, agentStatusStore)
	appsServer := service.NewAppServer(authStore, appStore, userStore, resendService)
	leaveServer := service.NewLeaveServer(leaveStore, authStore)
	sharingServer := service.NewSharingServer(sharingStore, authStore)
	shiftServer := service.NewShiftServer(shiftStore, authStore)
	scheduleServer := service.NewScheduleServer(scheduleStore, authStore)
	breakServer := service.NewBreakServer(breakStore, authStore)
	teamServer := service.NewTeamServer(teamStore, authStore)
	invitationServer := service.NewInvitationServer(
		invitationStore,
		resendService,
		orgStore,
		branchStore,
		departmentStore,
		teamStore,
		rbacStore,
		userStore,
		authStore,
		agentStatusStore,
		"app.kyla.cx",
	)
	onboardingStore := service.NewOnboardingStore(db.DB)
	onboardingHandler := service.NewOnboardingHandler(onboardingStore)
	onboardingServer := service.NewOnboardingServiceServer(db.DB)
	interceptor := service.NewAuthInterceptor(jwtManager, authStore)
	sqsClient := utils.NewSQSClient(
		configs.AwsCredentialsConfig.AwsRegion,
		configs.AwsCredentialsConfig.AwsAccessKey,
		configs.AwsCredentialsConfig.AwsSecretKey,
	)

	sqsAction := utils.NewSQSActions(sqsClient)
	responseInterceptor := service.NewResponseInterceptor(sqsAction)
	devicesInterceptor := service.NewSessionDevicesInterceptors(dbAuthStore)

	certificates := LoadCertificate(configs.EnvConfigs.Environment)
	var excludedMethods = []string{
		"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo",
	}

	grpcServer := grpc.NewServer(
		grpc.Creds(certificates),
		grpc.ChainUnaryInterceptor(
			interceptor.Unary(), responseInterceptor.WithExcludedMethods(excludedMethods...).Unary(),
			devicesInterceptor.Unary(),
		),
		grpc.ChainStreamInterceptor(interceptor.Stream(), responseInterceptor.Stream()),
	)

	reflection.Register(grpcServer)

	pb.RegisterAuthServiceServer(grpcServer, authServer)
	pb.RegisterRoleServiceServer(grpcServer, rbacServer)
	pb.RegisterPermissionServiceServer(grpcServer, rbacServer)
	pb.RegisterOrganisationServiceServer(grpcServer, orgServer)
	pb.RegisterUserServiceServer(grpcServer, userServer)
	pb.RegisterBranchServiceServer(grpcServer, branchServer) // Register BranchService
	pb.RegisterContactServiceServer(grpcServer, contactServer)
	pb.RegisterGroupServiceServer(grpcServer, contactGroupServer)
	pb.RegisterTagServiceServer(grpcServer, tagServer)
	pb.RegisterLabelServiceServer(grpcServer, labelServer)
	pb.RegisterDepartmentServiceServer(grpcServer, departmentServer)
	pb.RegisterAgentStatusServiceServer(grpcServer, agentStatusServer)
	pb.RegisterAppServiceServer(grpcServer, appsServer)
	pb.RegisterLeaveServiceServer(grpcServer, leaveServer)
	pb.RegisterResourceSharingServer(grpcServer, sharingServer)
	pb.RegisterShiftServiceServer(grpcServer, shiftServer)
	pb.RegisterShiftScheduleServiceServer(grpcServer, scheduleServer)
	pb.RegisterBreakServiceServer(grpcServer, breakServer)
	pb.RegisterInvitationServiceServer(grpcServer, invitationServer)
	pb.RegisterTeamServiceServer(grpcServer, teamServer)
	pb.RegisterOnboardingServiceServer(grpcServer, onboardingServer)

	// Use wait groups to manage multiple servers
	var wg sync.WaitGroup

	// Start gRPC server in a goroutine
	wg.Add(2)
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

	// Start HTTP server for health checks and REST API in a goroutine
	go func() {
		defer wg.Done()
		log.Println("Starting health check and REST API server on port 8085")
		if err := GinServer(onboardingHandler); err != nil {
			log.Fatalf("failed to start health check server: %v", err)
		}
	}()

	// Wait for both servers to complete (they typically won't unless there's an error)
	wg.Wait()
}

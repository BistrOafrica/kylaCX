package apps

import (
	"net/http"
	"strings"

	"kyla-be/internal/authctx"
	"kyla-be/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// WebhookHandler provides Gin REST handlers for webhook registration.
// Authentication is performed via Bearer token (the API app's token).
type WebhookHandler struct {
	webhookStore *WebhookStore
	appStore     *AppStore
}

// NewWebhookHandler constructs a WebhookHandler.
func NewWebhookHandler(webhookStore *WebhookStore, appStore *AppStore) *WebhookHandler {
	return &WebhookHandler{webhookStore: webhookStore, appStore: appStore}
}

// lookupApp validates the Authorization: Bearer <token> header and returns the
// owning App. Returns false and writes the appropriate error response on failure.
func (h *WebhookHandler) lookupApp(c *gin.Context) (*App, bool) {
	header := c.GetHeader("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid Authorization header"})
		return nil, false
	}
	token := strings.TrimPrefix(header, "Bearer ")
	app, err := h.appStore.FindByToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return nil, false
	}
	if app.Status != "ACTIVE" {
		c.JSON(http.StatusForbidden, gin.H{"error": "app is not active"})
		return nil, false
	}
	return app, true
}

// orgIDFromApp extracts the organisation UUID from the app's owner.
// For org-scoped apps (OwnerType == ORGANISATIONS), OwnerId is the org UUID.
func orgIDFromApp(app *App) string {
	if app.OwnerType == authctx.ORGANISATIONS {
		return app.OwnerId
	}
	return app.OwnerId
}

// RegisterWebhook POST /api/v1/webhooks
func (h *WebhookHandler) RegisterWebhook(c *gin.Context) {
	app, ok := h.lookupApp(c)
	if !ok {
		return
	}

	var req RegisterWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	secret, err := utils.GENERATE_RANDOM_KEY(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate signing secret"})
		return
	}

	w := &Webhook{
		ID:        uuid.New(),
		AppID:     app.ID,
		OrgID:     orgIDFromApp(app),
		URL:       req.URL,
		Events:    pq.StringArray(req.Events),
		Secret:    secret,
		IsActive:  true,
		CreatedBy: app.ID.String(),
	}
	if req.WorkspaceID != "" {
		w.WorkspaceID = &req.WorkspaceID
	} else if app.WorkspaceID != nil {
		w.WorkspaceID = app.WorkspaceID
	}

	created, err := h.webhookStore.Create(w)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register webhook"})
		return
	}

	c.JSON(http.StatusCreated, webhookToResponse(created, true))
}

// ListWebhooks GET /api/v1/webhooks
// Returns all webhooks owned by the authenticated app's organisation,
// optionally filtered by ?workspace_id=<uuid>.
func (h *WebhookHandler) ListWebhooks(c *gin.Context) {
	app, ok := h.lookupApp(c)
	if !ok {
		return
	}

	var wsID *string
	if q := c.Query("workspace_id"); q != "" {
		wsID = &q
	}

	webhooks, err := h.webhookStore.FindByOrg(orgIDFromApp(app), wsID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list webhooks"})
		return
	}

	resp := make([]WebhookResponse, len(webhooks))
	for i, wh := range webhooks {
		resp[i] = webhookToResponse(wh, false)
	}
	c.JSON(http.StatusOK, gin.H{"webhooks": resp})
}

// GetWebhook GET /api/v1/webhooks/:id
func (h *WebhookHandler) GetWebhook(c *gin.Context) {
	app, ok := h.lookupApp(c)
	if !ok {
		return
	}

	wh, err := h.webhookStore.FindByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "webhook not found"})
		return
	}
	if wh.OrgID != orgIDFromApp(app) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	c.JSON(http.StatusOK, webhookToResponse(wh, false))
}

// UpdateWebhook PUT /api/v1/webhooks/:id
func (h *WebhookHandler) UpdateWebhook(c *gin.Context) {
	app, ok := h.lookupApp(c)
	if !ok {
		return
	}

	wh, err := h.webhookStore.FindByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "webhook not found"})
		return
	}
	if wh.OrgID != orgIDFromApp(app) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req UpdateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.URL != "" {
		wh.URL = req.URL
	}
	if len(req.Events) > 0 {
		wh.Events = pq.StringArray(req.Events)
	}
	if req.IsActive != nil {
		wh.IsActive = *req.IsActive
	}

	updated, err := h.webhookStore.Update(wh)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update webhook"})
		return
	}
	c.JSON(http.StatusOK, webhookToResponse(updated, false))
}

// DeleteWebhook DELETE /api/v1/webhooks/:id
func (h *WebhookHandler) DeleteWebhook(c *gin.Context) {
	app, ok := h.lookupApp(c)
	if !ok {
		return
	}

	wh, err := h.webhookStore.FindByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "webhook not found"})
		return
	}
	if wh.OrgID != orgIDFromApp(app) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.webhookStore.Delete(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete webhook"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

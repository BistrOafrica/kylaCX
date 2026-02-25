package service

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OnboardingHandler struct {
	onboardingStore *OnboardingStore
}

func NewOnboardingHandler(onboardingStore *OnboardingStore) *OnboardingHandler {
	return &OnboardingHandler{
		onboardingStore: onboardingStore,
	}
}

// GIN HTTP handlers for teh onboarding methods
// CreateOnboarding
func (h *OnboardingHandler) CreateOnboarding(ctx *gin.Context) {
	data := &Onboarding{}
	if err := ctx.ShouldBindJSON(data); err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}
	createdOnboarding, err := h.onboardingStore.CreateOnboarding(data)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(201, gin.H{"data": createdOnboarding})
}

// GetOnboarding
func (h *OnboardingHandler) GetOnboarding(ctx *gin.Context) {
	id := ctx.Param("id")
	onboarding, err := h.onboardingStore.GetOnboardingByID(id)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, gin.H{"data": onboarding})
}

// UpdateOnboarding
func (h *OnboardingHandler) UpdateOnboarding(ctx *gin.Context) {
	id := ctx.Param("id")
	data := &Onboarding{}
	if err := ctx.ShouldBindJSON(data); err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}
	uuidID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "invalid UUID"})
		return
	}
	data.ID = uuidID
	updatedOnboarding, err := h.onboardingStore.UpdateOnboarding(data)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, gin.H{"data": updatedOnboarding})
}

// DeleteOnboarding
func (h *OnboardingHandler) DeleteOnboarding(ctx *gin.Context) {
	id := ctx.Param("id")
	err := h.onboardingStore.DeleteOnboarding(id)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, gin.H{"data": "onboarding deleted successfully"})
}

// ListOnboardings
func (h *OnboardingHandler) ListOnboardings(ctx *gin.Context) {
	var params ListOnboardingsParams
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}
	onboardings, total, err := h.onboardingStore.ListOnboardings(params)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, gin.H{"data": onboardings, "total": total})
}

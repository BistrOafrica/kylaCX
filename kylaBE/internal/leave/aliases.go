// Package leave provides access to the leave domain types.
// Phase 0 bridge: re-exports from pkg/service with forwarding constructors.
package leave

import (
	"kyla-be/internal/auth"
	"kyla-be/pkg/service"

	"gorm.io/gorm"
)

// Type aliases — exact re-exports of the service layer types.
type LeaveType = service.LeaveType
type LeaveBalance = service.LeaveBalance
type LeaveRequest = service.LeaveRequest
type LeaveRequestAttachment = service.LeaveRequestAttachment
type LeaveRequestEvent = service.LeaveRequestEvent
type EarnedLeaveCondition = service.EarnedLeaveCondition
type LeaveRequestsMetrics = service.LeaveRequestsMetrics
type LeaveStoreDB = service.LeaveStoreDB
type LeaveServer = service.LeaveServer

// NewLeaveStore creates a new LeaveStoreDB.
func NewLeaveStore(db *gorm.DB) *LeaveStoreDB {
	return service.NewLeaveStore(db)
}

// NewLeaveServer creates a new LeaveServer.
func NewLeaveServer(leaveStore *LeaveStoreDB, authStore *auth.AuthStore) *LeaveServer {
	return service.NewLeaveServer(leaveStore, authStore.Inner())
}

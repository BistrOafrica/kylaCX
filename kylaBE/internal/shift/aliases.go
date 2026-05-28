// Package shift provides access to shift, schedule, and break domain types.
// Phase 0 bridge: re-exports from pkg/service with forwarding constructors.
package shift

import (
	"kyla-be/internal/auth"
	"kyla-be/pkg/service"

	"gorm.io/gorm"
)

// Type aliases — exact re-exports of the service layer types.
type Shift = service.Shift
type ShiftStore = service.ShiftStore
type ShiftServer = service.ShiftServer
type ShiftScheduleStore = service.ShiftScheduleStore
type ShiftScheduleServer = service.ShiftScheduleServer
type BreakStoreDB = service.BreakStoreDB
type BreakServer = service.BreakServer

// NewShiftStore creates a new ShiftStore.
func NewShiftStore(db *gorm.DB) *ShiftStore {
	return service.NewShiftStore(db)
}

// NewShiftScheduleStore creates a new ShiftScheduleStore.
func NewShiftScheduleStore(db *gorm.DB) *ShiftScheduleStore {
	return service.NewShiftScheduleStore(db)
}

// NewBreakStore creates a new BreakStoreDB.
func NewBreakStore(db *gorm.DB) *BreakStoreDB {
	return service.NewBreakStore(db)
}

// NewShiftServer creates a new ShiftServer.
func NewShiftServer(shiftStore *ShiftStore, authStore *auth.AuthStore) *ShiftServer {
	return service.NewShiftServer(shiftStore, authStore.Inner())
}

// NewScheduleServer creates a new ShiftScheduleServer.
func NewScheduleServer(scheduleStore *ShiftScheduleStore, authStore *auth.AuthStore) *ShiftScheduleServer {
	return service.NewScheduleServer(scheduleStore, authStore.Inner())
}

// NewBreakServer creates a new BreakServer.
func NewBreakServer(breakStore *BreakStoreDB, authStore *auth.AuthStore) *BreakServer {
	return service.NewBreakServer(breakStore, authStore.Inner())
}

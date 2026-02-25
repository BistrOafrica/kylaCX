package main

import (
	"fmt"
	"kyla-be/config"
	"kyla-be/pkg/db"
	"kyla-be/pkg/service"
	"log"
	"os"

	"gorm.io/gorm"
)

func handleMigrateError(modelName string, err error) {
	if err != nil {
		log.Fatalf("Failed to auto migrate %s database: %v", modelName, err)
	}
	fmt.Printf("%s Database connected successfully...\n", modelName)
}

func migratePhoneNormalization(tx *gorm.DB) error {
	// Read and execute phone normalization SQL
	phoneNormalizationSqlBytes, err := os.ReadFile("pkg/db/migrations/phone_normalization.sql")
	if err != nil {
		return fmt.Errorf("failed to read phone normalization SQL file: %v", err)
	}

	if err := tx.Exec(string(phoneNormalizationSqlBytes)).Error; err != nil {
		return fmt.Errorf("failed to execute phone normalization SQL: %v", err)
	}

	log.Println("Phone number normalization completed successfully")
	return nil
}

func main() {
	configs, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configurations %v", err)
	}
	db.InitDB(&configs.PostgresConfig)
	DB := db.DB
	// Auto migrate models with Transactions
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	tx.Debug()
	// // Migrate the schema
	// handleMigrateError("\n:::===> :Organisation", DB.AutoMigrate(&service.Organisation{}))
	// handleMigrateError(":::===> :User", DB.AutoMigrate(&service.User{}))
	// // handleMigrateError(":::===> :App", DB.AutoMigrate(&service.App{}))
	// handleMigrateError(":::===> :Branch", DB.AutoMigrate(&service.Branch{}))
	// handleMigrateError(":::===> :Department", DB.AutoMigrate(&service.Department{}))
	// handleMigrateError(":::===> :Team", DB.AutoMigrate(&service.Team{}))
	// BreakRecord
	handleMigrateError(":::===> :BreakRecord", DB.AutoMigrate(&service.BreakRecord{}))
	// UserShiftRecord
	handleMigrateError(":::===> :UserShiftRecord", DB.AutoMigrate(&service.UserShiftRecord{}))
	// userSessions
	handleMigrateError(":::===> :UserSession", DB.AutoMigrate(&service.UserSession{}))
	// userDeviceInfo
	handleMigrateError(":::===> :UserDeviceInfo", DB.AutoMigrate(&service.UserDeviceInfo{}))
	// handleMigrateError(":::===> :ContactGroup", DB.AutoMigrate(&service.ContactGroup{}))
	// handleMigrateError(":::===> :Contact tag", DB.AutoMigrate(&service.Tag{}))
	// handleMigrateError(":::===> :SocialProfile", DB.AutoMigrate(&service.SocialProfile{}))
	// handleMigrateError(":::===> :Contact", DB.AutoMigrate(&service.Contact{}))
	// handleMigrateError(":::===> :Label", DB.AutoMigrate(&service.Label{}))
	// handleMigrateError(":::===> :Roles Table", DB.AutoMigrate(&service.Role{}))
	// handleMigrateError(":::===> :AgentStatus", DB.AutoMigrate(&service.AgentStatus{}))
	// handleMigrateError(":::===> :StatusChange", DB.AutoMigrate(&service.StatusChange{}))
	// handleMigrateError(":::===> :CustomFieldDefinition Table", DB.AutoMigrate(&service.CustomFieldDefinition{}))
	// handleMigrateError(":::===> :CustomFieldValue Table\n", DB.AutoMigrate(&service.CustomFieldValue{}))
	// handleMigrateError(":::===> :Shift Table", DB.AutoMigrate(&service.Shift{}))
	// handleMigrateError(":::===> :ShiftSchedule Table", DB.AutoMigrate(&service.ShiftSchedule{}))
	// handleMigrateError(":::===> :ShiftBreak Table", DB.AutoMigrate(&service.ScheduleBreak{}))
	// handleMigrateError(":::===> :LeaveType", DB.AutoMigrate(&service.LeaveType{}))
	// handleMigrateError(":::===> :EarnedLeaveCondition", DB.AutoMigrate(&service.EarnedLeaveCondition{}))
	// handleMigrateError(":::===> :LeaveBalance", DB.AutoMigrate(&service.LeaveBalance{}))

	// handleMigrateError(":::===> :LeaveRequest", DB.AutoMigrate(&service.LeaveRequest{}))
	// handleMigrateError(":::===> :LeaveRequestEvent", DB.AutoMigrate(&service.LeaveRequestEvent{}))
	// handleMigrateError(":::===> :LeaveRequestAttachment", DB.AutoMigrate(&service.LeaveRequestAttachment{}))
	// handleMigrateError(":::===> :ContactMergeHistory", DB.AutoMigrate(&service.ContactMergeHistory{}))
	// invitations
	// 	handleMigrateError(":::===> :Invitation", DB.AutoMigrate(&service.Invitation{}))
	// 	DB.AutoMigrate(&service.Passkey{}, &service.AuthenticatorInfo{})
	// 	if err := tx.Exec(`
	//     DO $$
	//     BEGIN
	//         IF EXISTS (SELECT 1 FROM information_schema.columns
	//                   WHERE table_name = 'leave_balances' AND column_name = 'remaining') THEN
	//             -- Check if the column is NOT generated (PostgreSQL-specific)
	//             IF NOT EXISTS (
	//                 SELECT 1 FROM information_schema.columns
	//                 WHERE table_name = 'leave_balances'
	//                 AND column_name = 'remaining'
	//                 AND is_generated = 'ALWAYS'
	//             ) THEN
	//                 ALTER TABLE leave_balances DROP COLUMN remaining;
	//                 ALTER TABLE leave_balances
	//                     ADD COLUMN remaining INTEGER GENERATED ALWAYS AS (total_eligible - used) STORED;
	//             END IF;
	//         ELSE
	//             -- Column doesn't exist; create it fresh
	//             ALTER TABLE leave_balances
	//                 ADD COLUMN remaining INTEGER GENERATED ALWAYS AS (total_eligible - used) STORED;
	//         END IF;
	//     END $$;
	// `).Error; err != nil {
	// 		log.Fatalf("Failed to update 'remaining' column: %v", err)
	// 	}

	// 	DB.AutoMigrate(&service.Entity{}, &service.EntityLink{}, &service.AccessRequest{})
	// 	DB.AutoMigrate(&service.Shift{}, &service.ShiftSchedule{}, &service.ScheduleBreak{})

	// 	// Phone number normalization migration
	// 	if err := migratePhoneNormalization(tx); err != nil {
	// 		log.Fatalf("Phone number normalization migration failed: %v", err)
	// 	}

	tx.Commit()
}

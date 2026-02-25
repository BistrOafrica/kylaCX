package service

import (
	"log"
	"reflect"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// func ProcessNodes(db *gorm.DB) error {
// 	return db.Transaction(func(tx *gorm.DB) error {
// 		// Define a slice of entity types and their corresponding node types
// 		entities := []struct {
// 			data     interface{}
// 			nodeType string
// 		}{
// 			{&[]Organisation{}, "org"},
// 			{&[]User{}, "user"},
// 			{&[]Branch{}, "branch"},
// 			{&[]Department{}, "department"},
// 			{&[]Team{}, "team"},
// 			{&[]Role{}, "role"},
// 			{&[]Contact{}, "contact"},
// 			{&[]ContactGroup{}, "contact_group"},
// 			{&[]App{}, "app"},
// 		}

// 		// Iterate over each entity type
// 		for _, entity := range entities {
// 			// Fetch all records for the current entity type
// 			if err := tx.Find(entity.data).Error; err != nil {
// 				return err
// 			}

// 			// Reflect over the slice of entities
// 			val := reflect.ValueOf(entity.data).Elem()
// 			for i := 0; i < val.Len(); i++ {
// 				item := val.Index(i).Interface()

// 				// Create a node for the current entity
// 				node := &Node{
// 					ID:              getID(item),
// 					Type:            entity.nodeType,
// 					OwnershipEntity: isOwnershipEntity(entity.nodeType),
// 				}
// 				if err := tx.FirstOrCreate(node).Error; err != nil {
// 					return err
// 				}

// 				// Check if the entity has an owner and create an edge if it does
// 				ownerID, ownerType := getOwner(item)
// 				if ownerID != "" && ownerType != "" {
// 					edge := &Edge{
// 						FromID: ownerID,
// 						ToID:   node.ID,
// 						Type:   OWNS,
// 					}
// 					if err := tx.FirstOrCreate(edge).Error; err != nil {
// 						return err
// 					}
// 				}
// 			}
// 		}
// 		return nil
// 	})
// }

// // Helper function to determine if the node type is an ownership entity
// func isOwnershipEntity(nodeType string) bool {
// 	ownershipEntities := map[string]bool{
// 		"org":           true,
// 		"user":          true,
// 		"branch":        true,
// 		"department":    true,
// 		"team":          true,
// 		"role":          true,
// 		"contact_group": true,
// 		"app":           true,
// 		// Add other types as needed
// 	}
// 	return ownershipEntities[nodeType]
// }

// // Helper function to extract the ID from an entity
// func getID(entity interface{}) string {
// 	val := reflect.ValueOf(entity)
// 	idField := val.FieldByName("ID")
// 	if idField.IsValid() {
// 		return idField.String()
// 	}
// 	return ""
// }

// // Helper function to extract the owner ID and owner type from an entity
// func getOwner(entity interface{}) (string, string) {
// 	val := reflect.ValueOf(entity)
// 	ownerIDField := val.FieldByName("OwnerID")
// 	ownerTypeField := val.FieldByName("OwnerType")
// 	if ownerIDField.IsValid() && ownerTypeField.IsValid() {
// 		return ownerIDField.String(), ownerTypeField.String()
// 	}
// 	return "", ""
// }

func ProcessNodes(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		entities := []struct {
			data     interface{}
			nodeType string
		}{
			{&[]Organisation{}, "organisation"},
			{&[]User{}, "user"},
			{&[]Branch{}, "branch"},
			{&[]Department{}, "department"},
			{&[]Team{}, "team"},
			{&[]Role{}, "role"},
			{&[]Contact{}, "contact"},
			{&[]ContactGroup{}, "contact_group"},
			{&[]App{}, "app"},
		}

		for _, entity := range entities {
			if err := tx.Find(entity.data).Error; err != nil {
				return err
			}

			val := reflect.ValueOf(entity.data).Elem()
			for i := 0; i < val.Len(); i++ {
				item := val.Index(i).Interface()
				node := &Entity{
					ID:              getID(item),
					Type:            entity.nodeType,
					OwnershipEntity: isOwnershipEntity(entity.nodeType),
				}
				if err := tx.FirstOrCreate(node).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func ProcessEdges(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		entities := []struct {
			data     interface{}
			nodeType string
		}{
			{&[]Organisation{}, "organisation"},
			{&[]User{}, "user"},
			{&[]Branch{}, "branch"},
			{&[]Department{}, "department"},
			{&[]Team{}, "team"},
			{&[]Role{}, "role"},
			{&[]Contact{}, "contact"},
			{&[]ContactGroup{}, "contact_group"},
			{&[]App{}, "app"},
		}

		for _, entity := range entities {
			if err := tx.Find(entity.data).Error; err != nil {
				return err
			}

			val := reflect.ValueOf(entity.data).Elem()
			for i := 0; i < val.Len(); i++ {
				item := val.Index(i).Interface()
				ownerID, ownerType := getOwner(item)
				toID := getID(item)
				if ownerID != uuid.Nil && ownerType != "" {
					log.Printf("Edges : %s - %s", ownerType, ownerID)
					edge := &EntityLink{
						ID:          uuid.New(),
						FromID:      ownerID,
						FromType:    ownerType,
						ToID:        toID,
						ToType:      entity.nodeType,
						Type:        OWNS,
						Roles:       "ADMIN",
						Permissions: "CREATE_READ_UPDATE_DELETE_SHARE",
						Conditions:  "is_owner",
					}
					if err := tx.FirstOrCreate(edge, "from_id = ? AND to_id = ?", ownerID, toID).Error; err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
}

func isOwnershipEntity(nodeType string) bool {
	ownershipEntities := map[string]bool{
		"organisation":  true,
		"user":          true,
		"branch":        true,
		"department":    true,
		"team":          true,
		"contact_group": true,
		"app":           true,
	}
	return ownershipEntities[nodeType]
}

func getID(entity interface{}) uuid.UUID {
	val := reflect.ValueOf(entity)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	idField := val.FieldByName("ID")
	if idField.IsValid() && idField.CanInterface() {
		if id, ok := idField.Interface().(uuid.UUID); ok {
			return id
		}
	}
	return uuid.Nil
}
func getOwner(entity interface{}) (uuid.UUID, string) {
	val := reflect.ValueOf(entity)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	ownerIDField := val.FieldByName("OwnerID")
	ownerTypeField := val.FieldByName("OwnerType")

	var ownerID uuid.UUID
	var ownerType string

	if ownerIDField.IsValid() && ownerIDField.CanInterface() {
		if id, ok := ownerIDField.Interface().(uuid.UUID); ok {
			ownerID = id
		}
	}
	if ownerTypeField.IsValid() && ownerTypeField.CanInterface() {
		if t, ok := ownerTypeField.Interface().(OwnerType); ok {
			log.Printf("OwnerTtype: %v ", t)
			ownerType = string(t)
		}
	}
	return ownerID, ownerType
}

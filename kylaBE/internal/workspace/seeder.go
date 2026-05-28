package workspace

import "log"

// SystemTypeSeedable is implemented by objectcore.ObjectCoreStore.
// Defined here (not in objectcore) to avoid circular imports.
type SystemTypeSeedable interface {
	SeedSystemObjectTypes(orgID, workspaceID, template string) error
}

// SeedWorkspace is called immediately after a new Workspace is created.
// When an ocSeeder is provided it creates the system object types for the
// workspace's domain template; otherwise it falls back to a log stub.
func SeedWorkspace(w *Workspace, ocSeeder SystemTypeSeedable) error {
	if ocSeeder != nil {
		if err := ocSeeder.SeedSystemObjectTypes(w.OrgID, w.ID, string(w.DomainTemplate)); err != nil {
			log.Printf("[workspace/seed] object type seed error for workspace %s: %v", w.ID, err)
			return err
		}
		log.Printf("[workspace/seed] seeded system object types for workspace %s (template=%s)", w.ID, w.DomainTemplate)
		return nil
	}

	// Phase 1 stub — logs pending seeds when no seeder is injected.
	log.Printf("[workspace/seed] workspace %s (%s) created with template %q — no seeder injected", w.ID, w.Name, w.DomainTemplate)
	switch w.DomainTemplate {
	case DomainTemplateSupport:
		log.Printf("[workspace/seed] support seeds pending: Ticket, Conversation, SLA, Knowledge Article")
	case DomainTemplateSales:
		log.Printf("[workspace/seed] sales seeds pending: Lead, Deal, Contact, Company, Activity")
	case DomainTemplateMarketing:
		log.Printf("[workspace/seed] marketing seeds pending: Campaign, Audience, Form, Journey")
	case DomainTemplateOperations:
		log.Printf("[workspace/seed] operations seeds pending: Task, Project, Process")
	default:
		log.Printf("[workspace/seed] custom workspace — no default seeds")
	}
	return nil
}

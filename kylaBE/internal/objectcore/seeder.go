package objectcore

// systemObjectTypes returns the predefined system ObjectTypes for the given
// workspace domain template.  These are created once when a workspace is
// first provisioned and marked is_system=true so they cannot be deleted.
func systemObjectTypes(orgID, workspaceID, template string) []ObjectType {
	switch template {
	case "support":
		return supportTypes(orgID, workspaceID)
	case "sales":
		return salesTypes(orgID, workspaceID)
	case "marketing":
		return marketingTypes(orgID, workspaceID)
	case "operations":
		return operationsTypes(orgID, workspaceID)
	default:
		return nil
	}
}

func base(orgID, workspaceID, slug, name, plural, icon, color string, fields []FieldDefinition) ObjectType {
	return ObjectType{
		OrgID:       orgID,
		WorkspaceID: workspaceID,
		Slug:        slug,
		Name:        name,
		PluralName:  plural,
		Icon:        icon,
		Color:       color,
		IsSystem:    true,
		Schema:      ObjectSchema{Fields: fields},
	}
}

func sel(opts ...string) []SelectOption {
	out := make([]SelectOption, len(opts))
	for i, o := range opts {
		out[i] = SelectOption{Value: o, Label: o}
	}
	return out
}

// ── Support ───────────────────────────────────────────────────────────────────

func supportTypes(org, ws string) []ObjectType {
	return []ObjectType{
		base(org, ws, "ticket", "Ticket", "Tickets", "ticket", "#6366f1", []FieldDefinition{
			{Key: "title", Label: "Title", Type: FieldTypeText, Required: true, Searchable: true},
			{Key: "description", Label: "Description", Type: FieldTypeText, Searchable: true},
			{Key: "status", Label: "Status", Type: FieldTypeSelect, Options: sel("open", "pending", "resolved", "closed")},
			{Key: "priority", Label: "Priority", Type: FieldTypeSelect, Options: sel("low", "normal", "high", "urgent")},
			{Key: "assignee", Label: "Assignee", Type: FieldTypeUser},
			{Key: "contact", Label: "Contact", Type: FieldTypeRelation, RelatesTo: "customer"},
			{Key: "due_date", Label: "Due Date", Type: FieldTypeDate},
			{Key: "tags", Label: "Tags", Type: FieldTypeMulti},
		}),
		base(org, ws, "conversation", "Conversation", "Conversations", "messages", "#06b6d4", []FieldDefinition{
			{Key: "channel", Label: "Channel", Type: FieldTypeSelect, Options: sel("whatsapp", "email", "sms", "voice", "webchat")},
			{Key: "status", Label: "Status", Type: FieldTypeSelect, Options: sel("open", "pending", "resolved", "snoozed")},
			{Key: "assignee", Label: "Assignee", Type: FieldTypeUser},
			{Key: "contact", Label: "Contact", Type: FieldTypeRelation, RelatesTo: "customer"},
			{Key: "subject", Label: "Subject", Type: FieldTypeText, Searchable: true},
		}),
		base(org, ws, "customer", "Customer", "Customers", "user", "#10b981", []FieldDefinition{
			{Key: "first_name", Label: "First Name", Type: FieldTypeText, Required: true, Searchable: true},
			{Key: "last_name", Label: "Last Name", Type: FieldTypeText, Searchable: true},
			{Key: "email", Label: "Email", Type: FieldTypeEmail, Unique: true, Searchable: true},
			{Key: "phone", Label: "Phone", Type: FieldTypePhone, Searchable: true},
			{Key: "company", Label: "Company", Type: FieldTypeText, Searchable: true},
		}),
		base(org, ws, "sla_policy", "SLA Policy", "SLA Policies", "clock", "#f59e0b", []FieldDefinition{
			{Key: "name", Label: "Name", Type: FieldTypeText, Required: true, Searchable: true},
			{Key: "response_hours", Label: "First Response (hrs)", Type: FieldTypeNumber},
			{Key: "resolution_hours", Label: "Resolution (hrs)", Type: FieldTypeNumber},
			{Key: "priority", Label: "Priority", Type: FieldTypeSelect, Options: sel("low", "normal", "high", "urgent")},
		}),
		base(org, ws, "knowledge_article", "Knowledge Article", "Knowledge Articles", "book", "#8b5cf6", []FieldDefinition{
			{Key: "title", Label: "Title", Type: FieldTypeText, Required: true, Searchable: true},
			{Key: "content", Label: "Content", Type: FieldTypeText, Searchable: true},
			{Key: "category", Label: "Category", Type: FieldTypeSelect, Options: sel("general", "product", "billing", "technical")},
			{Key: "status", Label: "Status", Type: FieldTypeSelect, Options: sel("draft", "published", "archived")},
			{Key: "author", Label: "Author", Type: FieldTypeUser},
		}),
	}
}

// ── Sales ─────────────────────────────────────────────────────────────────────

func salesTypes(org, ws string) []ObjectType {
	return []ObjectType{
		base(org, ws, "lead", "Lead", "Leads", "users", "#f59e0b", []FieldDefinition{
			{Key: "name", Label: "Name", Type: FieldTypeText, Required: true, Searchable: true},
			{Key: "email", Label: "Email", Type: FieldTypeEmail, Searchable: true},
			{Key: "phone", Label: "Phone", Type: FieldTypePhone, Searchable: true},
			{Key: "company", Label: "Company", Type: FieldTypeText, Searchable: true},
			{Key: "status", Label: "Status", Type: FieldTypeSelect, Options: sel("new", "contacted", "qualified", "converted", "dead")},
			{Key: "source", Label: "Source", Type: FieldTypeSelect, Options: sel("organic", "paid", "referral", "event", "cold")},
			{Key: "assignee", Label: "Assignee", Type: FieldTypeUser},
		}),
		base(org, ws, "deal", "Deal", "Deals", "currency-dollar", "#10b981", []FieldDefinition{
			{Key: "name", Label: "Name", Type: FieldTypeText, Required: true, Searchable: true},
			{Key: "value", Label: "Value", Type: FieldTypeCurrency},
			{Key: "stage", Label: "Stage", Type: FieldTypeSelect, Options: sel("prospect", "qualified", "proposal", "negotiation", "closed_won", "closed_lost")},
			{Key: "probability", Label: "Probability (%)", Type: FieldTypeNumber},
			{Key: "close_date", Label: "Close Date", Type: FieldTypeDate},
			{Key: "contact", Label: "Contact", Type: FieldTypeRelation, RelatesTo: "contact"},
			{Key: "assignee", Label: "Assignee", Type: FieldTypeUser},
		}),
		base(org, ws, "contact", "Contact", "Contacts", "user", "#6366f1", []FieldDefinition{
			{Key: "first_name", Label: "First Name", Type: FieldTypeText, Required: true, Searchable: true},
			{Key: "last_name", Label: "Last Name", Type: FieldTypeText, Searchable: true},
			{Key: "email", Label: "Email", Type: FieldTypeEmail, Unique: true, Searchable: true},
			{Key: "phone", Label: "Phone", Type: FieldTypePhone, Searchable: true},
			{Key: "company", Label: "Company", Type: FieldTypeRelation, RelatesTo: "company"},
			{Key: "assignee", Label: "Assignee", Type: FieldTypeUser},
		}),
		base(org, ws, "company", "Company", "Companies", "building", "#0ea5e9", []FieldDefinition{
			{Key: "name", Label: "Name", Type: FieldTypeText, Required: true, Searchable: true},
			{Key: "industry", Label: "Industry", Type: FieldTypeSelect, Options: sel("technology", "finance", "healthcare", "retail", "manufacturing", "other")},
			{Key: "website", Label: "Website", Type: FieldTypeURL},
			{Key: "size", Label: "Size", Type: FieldTypeSelect, Options: sel("1-10", "11-50", "51-200", "201-1000", "1001+")},
			{Key: "phone", Label: "Phone", Type: FieldTypePhone, Searchable: true},
		}),
		base(org, ws, "activity", "Activity", "Activities", "activity", "#ec4899", []FieldDefinition{
			{Key: "title", Label: "Title", Type: FieldTypeText, Required: true, Searchable: true},
			{Key: "type", Label: "Type", Type: FieldTypeSelect, Options: sel("call", "email", "meeting", "note", "task")},
			{Key: "due_date", Label: "Due Date", Type: FieldTypeDateTime},
			{Key: "assignee", Label: "Assignee", Type: FieldTypeUser},
			{Key: "related_to", Label: "Related To", Type: FieldTypeRelation},
			{Key: "notes", Label: "Notes", Type: FieldTypeText, Searchable: true},
		}),
	}
}

// ── Marketing ─────────────────────────────────────────────────────────────────

func marketingTypes(org, ws string) []ObjectType {
	return []ObjectType{
		base(org, ws, "campaign", "Campaign", "Campaigns", "megaphone", "#f97316", []FieldDefinition{
			{Key: "name", Label: "Name", Type: FieldTypeText, Required: true, Searchable: true},
			{Key: "channel", Label: "Channel", Type: FieldTypeSelect, Options: sel("whatsapp", "email", "sms")},
			{Key: "status", Label: "Status", Type: FieldTypeSelect, Options: sel("draft", "scheduled", "running", "paused", "completed")},
			{Key: "audience_size", Label: "Audience Size", Type: FieldTypeNumber},
			{Key: "scheduled_at", Label: "Scheduled At", Type: FieldTypeDateTime},
			{Key: "owner", Label: "Owner", Type: FieldTypeUser},
		}),
		base(org, ws, "audience", "Audience", "Audiences", "users-group", "#8b5cf6", []FieldDefinition{
			{Key: "name", Label: "Name", Type: FieldTypeText, Required: true, Searchable: true},
			{Key: "description", Label: "Description", Type: FieldTypeText},
			{Key: "size", Label: "Estimated Size", Type: FieldTypeNumber},
			{Key: "filters", Label: "Filter Config", Type: FieldTypeText},
		}),
		base(org, ws, "form", "Form", "Forms", "forms", "#06b6d4", []FieldDefinition{
			{Key: "name", Label: "Name", Type: FieldTypeText, Required: true, Searchable: true},
			{Key: "description", Label: "Description", Type: FieldTypeText},
			{Key: "status", Label: "Status", Type: FieldTypeSelect, Options: sel("draft", "active", "closed")},
			{Key: "submissions", Label: "Submission Count", Type: FieldTypeNumber},
		}),
		base(org, ws, "journey", "Journey", "Journeys", "route", "#10b981", []FieldDefinition{
			{Key: "name", Label: "Name", Type: FieldTypeText, Required: true, Searchable: true},
			{Key: "trigger", Label: "Trigger", Type: FieldTypeSelect, Options: sel("form_submit", "contact_created", "contact_updated", "tag_added")},
			{Key: "status", Label: "Status", Type: FieldTypeSelect, Options: sel("draft", "active", "paused", "archived")},
		}),
	}
}

// ── Operations ────────────────────────────────────────────────────────────────

func operationsTypes(org, ws string) []ObjectType {
	return []ObjectType{
		base(org, ws, "task", "Task", "Tasks", "checkbox", "#6366f1", []FieldDefinition{
			{Key: "title", Label: "Title", Type: FieldTypeText, Required: true, Searchable: true},
			{Key: "description", Label: "Description", Type: FieldTypeText, Searchable: true},
			{Key: "status", Label: "Status", Type: FieldTypeSelect, Options: sel("todo", "in_progress", "blocked", "done")},
			{Key: "priority", Label: "Priority", Type: FieldTypeSelect, Options: sel("low", "normal", "high", "urgent")},
			{Key: "assignee", Label: "Assignee", Type: FieldTypeUser},
			{Key: "due_date", Label: "Due Date", Type: FieldTypeDate},
			{Key: "project", Label: "Project", Type: FieldTypeRelation, RelatesTo: "project"},
		}),
		base(org, ws, "project", "Project", "Projects", "folder", "#0ea5e9", []FieldDefinition{
			{Key: "name", Label: "Name", Type: FieldTypeText, Required: true, Searchable: true},
			{Key: "description", Label: "Description", Type: FieldTypeText},
			{Key: "status", Label: "Status", Type: FieldTypeSelect, Options: sel("planning", "active", "on_hold", "completed", "cancelled")},
			{Key: "start_date", Label: "Start Date", Type: FieldTypeDate},
			{Key: "end_date", Label: "End Date", Type: FieldTypeDate},
			{Key: "owner", Label: "Owner", Type: FieldTypeUser},
		}),
		base(org, ws, "process", "Process", "Processes", "flow", "#10b981", []FieldDefinition{
			{Key: "name", Label: "Name", Type: FieldTypeText, Required: true, Searchable: true},
			{Key: "description", Label: "Description", Type: FieldTypeText},
			{Key: "status", Label: "Status", Type: FieldTypeSelect, Options: sel("draft", "active", "inactive")},
			{Key: "owner", Label: "Owner", Type: FieldTypeUser},
		}),
	}
}

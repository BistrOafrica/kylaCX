import { ObjectList } from "@/features/crm/components/ObjectList"

/**
 * TicketsList — schema-aware list bound to the `ticket` Object Core
 * type. Tickets are first-class Object Core records (the system seeds
 * the type on workspace creation), so we reuse the CRM ObjectList
 * primitive — saved views and field rendering come along for free.
 */
export function TicketsList() {
  return (
    <ObjectList
      typeSlug="ticket"
      title="Tickets"
      description="Customer support issues"
      basePath="/tickets"
    />
  )
}

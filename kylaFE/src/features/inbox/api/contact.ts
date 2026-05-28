import { services, unary } from "@/lib/rpc"
import { ReadContactRequest, type Contact } from "@/pb/contact"

/**
 * Resolve a single contact for the side panel. The backend has many
 * other contact methods (search, bulk import, dedup) — the inbox only
 * needs a read.
 */
export async function readContact(id: string): Promise<Contact | null> {
  const res = await unary(
    services.contact.readContact(
      ReadContactRequest.create({ contactId: id }) as ReadContactRequest,
    ),
  )
  return res.contact ?? null
}

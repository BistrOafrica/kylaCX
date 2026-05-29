import { services } from "@/lib/rpc/services"
import type {
  SipExtension,
  SipTrunk,
  SipDomain,
} from "@/pb/telephony"

/**
 * SIP admin helpers — extensions, trunks, domains.
 *
 * All RPCs come from the unified Phase 5 TelephonyService. The TelephonyService
 * server strips trunk passwords on read paths, so listings here are safe to
 * surface in the admin UI without leaking credentials.
 */

// ── Extensions ──────────────────────────────────────────────────────────────

export async function listSipExtensions(workspaceId: string): Promise<SipExtension[]> {
  const res = await services.telephony.listSipExtensions({ workspaceId })
  return res.response.extensions
}

export async function createSipExtension(extension: Partial<SipExtension>): Promise<SipExtension> {
  const res = await services.telephony.createSipExtension({
    extension: emptyExtension(extension),
  })
  return res.response
}

export async function deleteSipExtension(id: string): Promise<void> {
  await services.telephony.deleteSipExtension({ id })
}

// ── Trunks ──────────────────────────────────────────────────────────────────

export async function listSipTrunks(orgId: string): Promise<SipTrunk[]> {
  const res = await services.telephony.listSipTrunks({ orgId })
  return res.response.trunks
}

export async function createSipTrunk(trunk: Partial<SipTrunk>): Promise<SipTrunk> {
  const res = await services.telephony.createSipTrunk({
    trunk: emptyTrunk(trunk),
  })
  return res.response
}

export async function updateSipTrunk(trunk: SipTrunk): Promise<SipTrunk> {
  const res = await services.telephony.updateSipTrunk({ trunk })
  return res.response
}

export async function deleteSipTrunk(id: string): Promise<void> {
  await services.telephony.deleteSipTrunk({ id })
}

// ── Domains ─────────────────────────────────────────────────────────────────

export async function listSipDomains(orgId: string): Promise<SipDomain[]> {
  const res = await services.telephony.listSipDomains({ orgId })
  return res.response.domains
}

export async function createSipDomain(domain: Partial<SipDomain>): Promise<SipDomain> {
  const res = await services.telephony.createSipDomain({
    domain: emptyDomain(domain),
  })
  return res.response
}

export async function deleteSipDomain(id: string): Promise<void> {
  await services.telephony.deleteSipDomain({ id })
}

// ── Empty-message builders ──────────────────────────────────────────────────

function emptyExtension(seed: Partial<SipExtension>): SipExtension {
  return {
    id: "",
    orgId: "",
    workspaceId: "",
    userId: "",
    extension: "",
    displayName: "",
    status: "",
    ...seed,
  } as SipExtension
}

function emptyTrunk(seed: Partial<SipTrunk>): SipTrunk {
  return {
    id: "",
    orgId: "",
    name: "",
    gatewayName: "",
    provider: "custom",
    sipServer: "",
    username: "",
    password: "",
    fromUri: "",
    isActive: true,
    ...seed,
  } as SipTrunk
}

function emptyDomain(seed: Partial<SipDomain>): SipDomain {
  return {
    id: "",
    orgId: "",
    domain: "",
    isDefault: false,
    ...seed,
  } as SipDomain
}

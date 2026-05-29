import { services } from "@/lib/rpc/services"
import type { SoftphoneToken } from "@/pb/telephony"

/**
 * Fetches a short-lived softphone bootstrap from the backend.
 *
 * The token is bound to (org, user, extension) and signed with the platform
 * JWT_SECRET_KEY — the FreeSWITCH WSS profile validates it before allowing
 * SIP REGISTER. Pair the token with the supplied ws_url + ice_servers to
 * stand up a working SIP.js UserAgent.
 *
 * agentId is optional; when omitted the backend resolves the calling user
 * from the gRPC auth context.
 */
export async function fetchSoftphoneToken(agentId?: string): Promise<SoftphoneToken> {
  const res = await services.telephony.issueSoftphoneToken({
    agentId: agentId ?? "",
  })
  const token = res.response.token
  if (!token) {
    throw new Error("issueSoftphoneToken returned an empty token")
  }
  return token
}

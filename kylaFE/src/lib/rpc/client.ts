import { GrpcWebFetchTransport } from "@protobuf-ts/grpcweb-transport"
import type { RpcOptions, RpcMetadata } from "@protobuf-ts/runtime-rpc"
import { env } from "./env"

/**
 * Shared gRPC-web transport.
 *
 * One transport instance reused by every service client. Auth metadata
 * is injected per-call via withAuth() rather than at transport
 * construction, so changes in token / workspace / locale flow through
 * without re-instantiating clients.
 */
export const transport = new GrpcWebFetchTransport({
  baseUrl: env.apiUrl,
  format: "binary",
})

type MetaProvider = () => RpcMetadata

const metaProviders: MetaProvider[] = []

/**
 * Register a function that contributes gRPC metadata on every call.
 * Used by auth, workspace, and i18n modules to inject their headers
 * without component code knowing about them.
 *
 *   registerMetaProvider(() => ({ authorization: store.accessToken }))
 *
 * Returns an unregister function so providers can be torn down.
 */
export function registerMetaProvider(provider: MetaProvider): () => void {
  metaProviders.push(provider)
  return () => {
    const i = metaProviders.indexOf(provider)
    if (i >= 0) metaProviders.splice(i, 1)
  }
}

/** Build the merged metadata for a call. */
export function buildMeta(extra?: RpcMetadata): RpcMetadata {
  const meta: RpcMetadata = {}
  for (const provider of metaProviders) {
    Object.assign(meta, provider())
  }
  if (extra) Object.assign(meta, extra)
  return meta
}

/**
 * Build RpcOptions with metadata + timeout sensible defaults.
 *
 *   client.login(req, rpcOptions())
 */
export function rpcOptions(extra?: Partial<RpcOptions>): RpcOptions {
  return {
    timeout: 30_000,
    meta: buildMeta(extra?.meta),
    ...extra,
  }
}

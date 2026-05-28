import type { UnaryCall, ServerStreamingCall } from "@protobuf-ts/runtime-rpc"
import { rpcOptions } from "./client"
import { RpcError } from "./errors"

/**
 * Higher-level call helpers around the @protobuf-ts call objects.
 *
 *   const { token } = await unary(services.auth.login(req))
 *
 * unary() awaits the response and normalizes errors to RpcError. It
 * accepts either a bare UnaryCall (caller built options) or a
 * caller-builder pattern: `unary((opts) => services.auth.login(req, opts))`
 * which is the recommended form so the helper can inject metadata.
 */

type CallBuilder<I extends object, O extends object> = (
  opts: ReturnType<typeof rpcOptions>,
) => UnaryCall<I, O>

export async function unary<I extends object, O extends object>(
  callOrBuilder: UnaryCall<I, O> | CallBuilder<I, O>,
): Promise<O> {
  const call =
    typeof callOrBuilder === "function" ? callOrBuilder(rpcOptions()) : callOrBuilder
  try {
    const result = await call.response
    return result
  } catch (err) {
    throw RpcError.from(err)
  }
}

/**
 * Subscribe to a server-streaming RPC.
 *
 *   const sub = stream(
 *     (opts) => services.communication.streamConversationUpdates(req, opts),
 *     {
 *       onMessage: (msg) => …,
 *       onError: (err)  => …,
 *       onDone:  ()     => …,
 *     }
 *   )
 *   // later
 *   sub.cancel()
 */
export interface StreamHandlers<O> {
  onMessage: (msg: O) => void
  onError?: (err: RpcError) => void
  onDone?: () => void
}

export interface Subscription {
  cancel(): void
}

type StreamBuilder<I extends object, O extends object> = (
  opts: ReturnType<typeof rpcOptions>,
) => ServerStreamingCall<I, O>

export function stream<I extends object, O extends object>(
  builder: StreamBuilder<I, O>,
  handlers: StreamHandlers<O>,
): Subscription {
  const controller = new AbortController()
  const opts = rpcOptions({ abort: controller.signal })
  const call = builder(opts)

  ;(async () => {
    try {
      for await (const msg of call.responses) {
        handlers.onMessage(msg)
      }
      handlers.onDone?.()
    } catch (err) {
      if (controller.signal.aborted) return
      handlers.onError?.(RpcError.from(err))
    }
  })()

  return {
    cancel: () => controller.abort(),
  }
}

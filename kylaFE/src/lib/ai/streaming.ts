/**
 * AI streaming bridge — placeholder.
 *
 * F1 will wire this to the AIService (ClassifyText, SummarizeText,
 * GenerateReply) via gRPC server-streaming. For F0 the function below
 * exists so consumers can be written against the contract; it currently
 * fakes a token-by-token stream so the UI can be exercised.
 */

export interface StreamOptions {
  onToken: (token: string) => void
  onDone?: () => void
  onError?: (err: Error) => void
  signal?: AbortSignal
}

/** Mock streamer — replace with real AIService binding in F1. */
export async function streamMockResponse(text: string, opts: StreamOptions) {
  const words = text.split(/(\s+)/)
  try {
    for (const word of words) {
      if (opts.signal?.aborted) return
      await new Promise((r) => setTimeout(r, 28))
      opts.onToken(word)
    }
    opts.onDone?.()
  } catch (err) {
    opts.onError?.(err as Error)
  }
}

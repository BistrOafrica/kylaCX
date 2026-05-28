/**
 * Typed application error wrapping a gRPC failure.
 *
 * The gRPC-web runtime throws values with `code` and `message` fields
 * but the type information is lossy. This module normalizes errors so
 * the rest of the app can switch on a stable enum.
 */

export type RpcCode =
  | "UNAUTHENTICATED"
  | "PERMISSION_DENIED"
  | "NOT_FOUND"
  | "ALREADY_EXISTS"
  | "INVALID_ARGUMENT"
  | "FAILED_PRECONDITION"
  | "RESOURCE_EXHAUSTED"
  | "UNAVAILABLE"
  | "DEADLINE_EXCEEDED"
  | "INTERNAL"
  | "UNKNOWN"

const GRPC_CODE_NAMES: Record<number, RpcCode> = {
  1: "UNKNOWN",
  3: "INVALID_ARGUMENT",
  4: "DEADLINE_EXCEEDED",
  5: "NOT_FOUND",
  6: "ALREADY_EXISTS",
  7: "PERMISSION_DENIED",
  8: "RESOURCE_EXHAUSTED",
  9: "FAILED_PRECONDITION",
  13: "INTERNAL",
  14: "UNAVAILABLE",
  16: "UNAUTHENTICATED",
}

export class RpcError extends Error {
  readonly code: RpcCode
  readonly status: number
  readonly meta: Record<string, string>

  constructor(message: string, code: RpcCode, status = 0, meta: Record<string, string> = {}) {
    super(message)
    this.name = "RpcError"
    this.code = code
    this.status = status
    this.meta = meta
  }

  static from(err: unknown): RpcError {
    if (err instanceof RpcError) return err

    const e = err as { code?: number | string; message?: string; meta?: Record<string, string> }
    const codeName: RpcCode =
      typeof e.code === "number"
        ? GRPC_CODE_NAMES[e.code] ?? "UNKNOWN"
        : (e.code as RpcCode) ?? "UNKNOWN"

    return new RpcError(
      e.message ?? "Request failed",
      codeName,
      typeof e.code === "number" ? e.code : 0,
      e.meta ?? {},
    )
  }

  /** True if the error indicates the caller is not (or no longer) authenticated. */
  get isUnauthenticated() {
    return this.code === "UNAUTHENTICATED"
  }

  /** True for errors a UI should show as a toast rather than a destructive failure. */
  get isUserFacing() {
    return (
      this.code === "INVALID_ARGUMENT" ||
      this.code === "ALREADY_EXISTS" ||
      this.code === "FAILED_PRECONDITION" ||
      this.code === "NOT_FOUND" ||
      this.code === "PERMISSION_DENIED"
    )
  }
}

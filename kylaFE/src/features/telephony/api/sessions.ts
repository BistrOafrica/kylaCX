import { services, unary } from "@/lib/rpc"
import {
  StartCallSessionRequest,
  EndCallSessionRequest,
  ReadCallSessionRequest,
  PlaceOnHoldRequest,
  RemoveFromHoldRequest,
  AddNoteRequest,
  AddDestinationRequest,
  type CallSession,
  Destination,
  CallDirection,
} from "@/pb/call_session"
import { useAuthStore } from "@/lib/auth"

/**
 * Call session API — active-call control surface.
 *
 * This is what the softphone widget invokes to start / end / hold /
 * note an ongoing call. Live state is read via the monitoring streams
 * (see api/calls.ts).
 */

function selfId() {
  return useAuthStore.getState().identity?.userId ?? ""
}

export async function startSession(input: {
  destinationNumber: string
  direction?: CallDirection
}): Promise<{ sessionId: string }> {
  // StartCallSession only takes agentId + direction — the actual
  // destination is added via AddDestination after the session is
  // started.
  const res = await unary(
    services.callSession.startSession(
      StartCallSessionRequest.create({
        agentId: selfId(),
        direction: input.direction ?? CallDirection.OUTBOUND,
      }) as StartCallSessionRequest,
    ),
  )
  const sessionId =
    (res as { sessionId?: string; id?: string }).sessionId ??
    (res as { id?: string }).id ??
    ""

  if (sessionId && input.destinationNumber) {
    await addDestination({
      sessionId,
      destinationType: Destination.EXTENSION,
      destinationId: "",
      destinationNumber: input.destinationNumber,
    })
  }
  return { sessionId }
}

export async function endSession(sessionId: string): Promise<void> {
  await unary(
    services.callSession.endSession(
      EndCallSessionRequest.create({ sessionId }) as EndCallSessionRequest,
    ),
  )
}

export async function getSession(id: string): Promise<CallSession> {
  return unary(
    services.callSession.readSession(
      ReadCallSessionRequest.create({ id }) as ReadCallSessionRequest,
    ),
  )
}

export async function placeOnHold(sessionId: string): Promise<void> {
  await unary(
    services.callSession.placeOnHold(
      PlaceOnHoldRequest.create({
        sessionId,
        agentId: selfId(),
      }) as PlaceOnHoldRequest,
    ),
  )
}

export async function removeFromHold(sessionId: string): Promise<void> {
  await unary(
    services.callSession.removeFromHold(
      RemoveFromHoldRequest.create({
        sessionId,
      }) as RemoveFromHoldRequest,
    ),
  )
}

export async function addNote(input: {
  sessionId: string
  name: string
  content: string
}): Promise<void> {
  await unary(
    services.callSession.addNote(
      AddNoteRequest.create({
        sessionId: input.sessionId,
        name: input.name,
        content: input.content,
        createdBy: selfId(),
      }) as AddNoteRequest,
    ),
  )
}

export async function addDestination(input: {
  sessionId: string
  destinationType: Destination
  destinationId: string
  destinationNumber: string
}): Promise<void> {
  await unary(
    services.callSession.addDestination(
      AddDestinationRequest.create({
        sessionId: input.sessionId,
        destinationType: input.destinationType,
        destinationId: input.destinationId,
        destinationNumber: input.destinationNumber,
      }) as AddDestinationRequest,
    ),
  )
}

export { Destination }
export type { CallSession }

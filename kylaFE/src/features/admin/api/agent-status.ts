import { services, unary } from "@/lib/rpc"
import {
  StatusChangeRequest,
  AgentStatusRequest,
  AgentAvailabilityRequest,
  StatusType,
  StatusChange,
  type AgentStatus,
} from "@/pb/agents"
import { useAuthStore } from "@/lib/auth"
import { OwnerType } from "@/pb/owner_type"
import { useWorkspaceStore } from "@/lib/workspace"

function selfId() {
  return useAuthStore.getState().identity?.userId ?? ""
}

export async function getLatestStatus(agentId?: string): Promise<StatusChange | null> {
  const res = await unary(
    services.agentStatus.readLatestStatusChange(
      AgentStatusRequest.create({
        agentId: agentId ?? selfId(),
      }) as AgentStatusRequest,
    ),
  )
  return (res as { statusChange?: StatusChange }).statusChange ?? null
}

export async function getStatusHistory(agentId?: string): Promise<AgentStatus | null> {
  const res = await unary(
    services.agentStatus.readUserStatusHistory(
      AgentStatusRequest.create({
        agentId: agentId ?? selfId(),
      }) as AgentStatusRequest,
    ),
  )
  return (res as { agentStatus?: AgentStatus }).agentStatus ?? null
}

export async function setStatus(input: {
  statusType: StatusType
  description?: string
}): Promise<StatusChange | null> {
  const userId = selfId()
  const change = StatusChange.create({
    statusType: input.statusType,
    description: input.description ?? "",
    startTime: new Date().toISOString(),
    ownerId: userId,
    ownerType: OwnerType.USERS,
  }) as StatusChange
  const res = await unary(
    services.agentStatus.createStatusChange(
      StatusChangeRequest.create({
        statusChange: change,
      }) as StatusChangeRequest,
    ),
  )
  return (res as { statusChange?: StatusChange }).statusChange ?? null
}

export async function getAvailability(agentId?: string) {
  return unary(
    services.agentStatus.getAgentAvailability(
      AgentAvailabilityRequest.create({
        agentId: agentId ?? selfId(),
      }) as AgentAvailabilityRequest,
    ),
  )
}

void useWorkspaceStore   // keep import — surface may take a workspace scope later
export { StatusType }
export type { StatusChange, AgentStatus }

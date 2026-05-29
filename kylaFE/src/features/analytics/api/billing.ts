import { services, unary } from "@/lib/rpc"
import {
  ReadAllWalletsRequest,
  ReadWalletTransactionsRequest,
  type Wallet,
  type WalletTransaction,
} from "@/pb/billing_wallets"
import { useWorkspaceStore } from "@/lib/workspace"
import { OwnerType } from "@/pb/owner_type"

/**
 * Billing — currently surfaced via Wallets + Subscription protos.
 *
 * The Subscription proto only exposes `subscribeToPlan`,
 * `updateSubscription`, and `cancelSubscription` — there's no
 * "read my subscription" RPC, so the UI shows wallet + recent
 * transactions today and defers the plan view to F6.x once the
 * proto gains a read method.
 */

function ownerScope() {
  const orgId = useWorkspaceStore.getState().organisation?.id ?? ""
  return { ownerType: OwnerType.ORGANISATIONS, ownerId: orgId }
}

export async function listWallets(): Promise<Wallet[]> {
  const { ownerType, ownerId } = ownerScope()
  const res = await unary(
    services.wallets.readAllWallets(
      ReadAllWalletsRequest.create({
        ownerType,
        ownerId,
      }) as ReadAllWalletsRequest,
    ),
  )
  return res.wallets ?? []
}

export async function listWalletTransactions(
  walletId: string,
): Promise<WalletTransaction[]> {
  const res = await unary(
    services.wallets.readWalletTransactions(
      ReadWalletTransactionsRequest.create({
        walletId,
      }) as ReadWalletTransactionsRequest,
    ),
  )
  return res.walletTransactions ?? []
}

export function formatMoney(amount: bigint | number, currency = "USD"): string {
  const n = typeof amount === "bigint" ? Number(amount) : amount
  return new Intl.NumberFormat(undefined, {
    style: "currency",
    currency,
    maximumFractionDigits: 0,
  }).format(n / 100)
}

export type { Wallet, WalletTransaction }

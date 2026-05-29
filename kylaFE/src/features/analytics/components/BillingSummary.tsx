import {
  IconWallet,
  IconArrowDown,
  IconArrowUp,
  IconCreditCard,
} from "@tabler/icons-react"
import {
  PageHeader,
  Surface,
  EmptyState,
  ErrorState,
  CardSkeleton,
} from "@/design-system"
import { cn } from "@/lib/utils"
import { useWallets, useWalletTransactions } from "../hooks/queries"
import { formatMoney, type Wallet, type WalletTransaction } from "../api/billing"

/**
 * BillingSummary — wallets + recent transactions.
 *
 * The Subscription proto doesn't expose a "read my plan" RPC yet, so
 * F6 ships the wallet view (balance + usage + recent transactions)
 * and leaves the plan + payment-methods sections for F6.x once the
 * backend regenerates a richer Subscription service.
 */
export function BillingSummary() {
  const wallets = useWallets()
  const mainWallet = wallets.data?.find((w) => w.isMain) ?? wallets.data?.[0]
  const txns = useWalletTransactions(mainWallet?.id ?? null)

  return (
    <div className="flex flex-col h-full overflow-y-auto bg-canvas">
      <PageHeader
        title="Billing"
        description="Wallets, usage and recent transactions"
      />

      <div className="p-4 space-y-4 max-w-4xl">
        {wallets.isPending ? (
          <CardSkeleton lines={3} />
        ) : wallets.isError ? (
          <ErrorState
            title="Couldn't load wallets"
            description={(wallets.error as Error).message}
            onRetry={() => void wallets.refetch()}
          />
        ) : (wallets.data?.length ?? 0) === 0 ? (
          <EmptyState
            icon={<IconWallet className="size-5" />}
            title="No wallets yet"
            description="Wallets get created when you set up billing or top up credit."
          />
        ) : (
          <>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
              {wallets.data!.map((w) => (
                <WalletCard key={w.id} wallet={w} primary={w.id === mainWallet?.id} />
              ))}
            </div>

            <Surface level={1} radius="md" className="flex flex-col">
              <header className="flex items-center gap-2 px-4 py-3 border-b border-border">
                <IconCreditCard className="size-3.5 text-fg-muted" aria-hidden />
                <span className="text-md font-medium text-fg flex-1">
                  Recent transactions
                </span>
                {mainWallet && (
                  <span className="font-mono text-xs text-fg-muted truncate">
                    {mainWallet.id.slice(0, 12)}…
                  </span>
                )}
              </header>
              {txns.isPending ? (
                <div className="p-3">
                  <CardSkeleton lines={3} />
                </div>
              ) : txns.isError ? (
                <ErrorState
                  title="Couldn't load transactions"
                  description={(txns.error as Error).message}
                  onRetry={() => void txns.refetch()}
                />
              ) : (txns.data?.length ?? 0) === 0 ? (
                <EmptyState
                  title="No transactions yet"
                  description="Top-ups and charges will appear here."
                  size="sm"
                />
              ) : (
                <ul className="divide-y divide-border-subtle" role="list">
                  {txns.data!.map((t) => (
                    <TransactionRow
                      key={t.id}
                      transaction={t}
                      currency={mainWallet?.currency}
                    />
                  ))}
                </ul>
              )}
            </Surface>
          </>
        )}
      </div>
    </div>
  )
}

function WalletCard({ wallet, primary }: { wallet: Wallet; primary?: boolean }) {
  return (
    <Surface
      level={1}
      radius="md"
      className={cn("p-4 space-y-2", primary && "border-accent")}
    >
      <div className="flex items-center gap-2">
        <IconWallet className="size-4 text-fg-muted" aria-hidden />
        <span className="text-md font-medium text-fg flex-1 truncate">
          {primary ? "Primary wallet" : "Wallet"}
        </span>
        {primary && (
          <span className="inline-flex items-center h-4 px-1 rounded-xs text-[10px] font-mono uppercase bg-accent-subtle text-accent">
            main
          </span>
        )}
      </div>
      <div className="text-2xl font-semibold tabular-nums text-fg">
        {formatMoney(wallet.balance, wallet.currency || "USD")}
      </div>
      <dl className="grid grid-cols-2 gap-x-3 gap-y-1 text-sm">
        <dt className="text-fg-muted">Purchased</dt>
        <dd className="font-mono text-fg-secondary text-end">
          {formatMoney(wallet.lifetimePurchased, wallet.currency || "USD")}
        </dd>
        <dt className="text-fg-muted">Spent</dt>
        <dd className="font-mono text-fg-secondary text-end">
          {formatMoney(wallet.lifetimeSpent, wallet.currency || "USD")}
        </dd>
        <dt className="text-fg-muted">Status</dt>
        <dd className="font-mono text-fg-secondary text-end uppercase tracking-wider text-xs">
          {wallet.status || "—"}
        </dd>
      </dl>
    </Surface>
  )
}

function TransactionRow({
  transaction,
  currency,
}: {
  transaction: WalletTransaction
  currency?: string
}) {
  const isCredit =
    transaction.transactionType === "credit" ||
    transaction.transactionType === "TOP_UP"
  return (
    <li className="flex items-center gap-3 px-4 py-2.5">
      <div
        className={cn(
          "size-7 rounded-sm flex items-center justify-center shrink-0",
          isCredit
            ? "bg-success-subtle text-success"
            : "bg-warn-subtle text-warn",
        )}
        aria-hidden
      >
        {isCredit ? (
          <IconArrowDown className="size-3.5" />
        ) : (
          <IconArrowUp className="size-3.5" />
        )}
      </div>
      <div className="min-w-0 flex-1">
        <div className="text-md text-fg truncate">
          {transaction.transactionType || "Transaction"}
        </div>
        <div className="text-xs font-mono text-fg-muted truncate">
          {transaction.id.slice(0, 12)}
        </div>
      </div>
      <span
        className={cn(
          "font-mono tabular-nums text-md",
          isCredit ? "text-success" : "text-fg",
        )}
      >
        {isCredit ? "+" : "−"}
        {formatMoney(transaction.amount, currency)}
      </span>
    </li>
  )
}

import { Link, useNavigate } from "react-router-dom"
import { IconPlus, IconTree } from "@tabler/icons-react"
import { toast } from "sonner"
import {
  PageHeader,
  EmptyState,
  ErrorState,
  ListRowSkeleton,
} from "@/design-system"
import { cn } from "@/lib/utils"
import { useIvrFlows } from "../hooks/queries"
import { createIvrFlow } from "../api/ivr"

export function IvrList() {
  const list = useIvrFlows()
  const navigate = useNavigate()

  const onNew = async () => {
    try {
      await createIvrFlow({ name: "Untitled IVR" })
      void list.refetch()
      toast.success("IVR flow created")
    } catch (err) {
      toast.error((err as Error).message)
    }
    // The proto's CreateIvrFlow response shape doesn't carry the new
    // id — refetch the list and let the user click in.
    navigate("/ivr")
  }

  return (
    <div className="flex flex-col h-full bg-canvas">
      <PageHeader
        title="IVR flows"
        description="Interactive voice response menus and triggers"
        actions={
          <button
            type="button"
            onClick={() => void onNew()}
            className={cn(
              "inline-flex items-center gap-1.5 h-8 px-3 rounded-sm text-md font-medium",
              "bg-accent text-accent-fg hover:bg-accent-hover",
            )}
          >
            <IconPlus className="size-3.5" />
            New flow
          </button>
        }
      />

      {list.isPending ? (
        <div className="p-2">
          <ListRowSkeleton count={4} />
        </div>
      ) : list.isError ? (
        <ErrorState
          title="Couldn't load IVR flows"
          description={(list.error as Error).message}
          onRetry={() => void list.refetch()}
        />
      ) : (list.data?.items.length ?? 0) === 0 ? (
        <EmptyState
          icon={<IconTree className="size-5" />}
          title="No IVR flows yet"
          description="Build menu trees that route inbound callers to the right queue."
        />
      ) : (
        <ul className="flex-1 overflow-y-auto p-1" role="list">
          {list.data!.items.map((flow) => (
            <li key={flow.id}>
              <Link
                to={`/ivr/${flow.id}`}
                className="flex items-center gap-3 px-3 py-2 rounded-sm hover:bg-subtle transition-colors"
              >
                <div
                  className="size-7 rounded-sm bg-accent-subtle text-fg-secondary flex items-center justify-center shrink-0"
                  aria-hidden
                >
                  <IconTree className="size-3.5" />
                </div>
                <div className="min-w-0 flex-1">
                  <div className="text-md font-medium text-fg truncate">
                    {flow.name || "(unnamed)"}
                  </div>
                  <div className="text-sm text-fg-muted truncate">
                    {flow.description || "No description"}
                  </div>
                </div>
                <span className="font-mono text-xs text-fg-muted">
                  {flow.nodes?.length ?? 0} nodes
                </span>
              </Link>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}

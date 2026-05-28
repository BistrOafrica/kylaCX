import { useMemo, useState } from "react"
import { Link } from "react-router-dom"
import { IconSearch, IconPlus, IconBook, IconWorld } from "@tabler/icons-react"
import {
  PageHeader,
  EmptyState,
  ErrorState,
  ListRowSkeleton,
} from "@/design-system"
import { Input } from "@/components/ui/input"
import { cn } from "@/lib/utils"
import { useKnowledgeBases, useKnowledgeSearch } from "../hooks/queries"
import { KnowledgeBaseType, type KnowledgeBase } from "../api/knowledge"
import { relativeShort } from "@/features/crm/utils/relative"

/**
 * KnowledgeList — browse knowledge sources.
 *
 * The backend currently exposes the LLM-knowledge-source proto (text /
 * website / FAQ), so this lists those rather than the categorised
 * article model. When the article+category proto regenerates, we add
 * a second tab without breaking this surface.
 */
export function KnowledgeList() {
  const all = useKnowledgeBases()
  const [search, setSearch] = useState("")
  const searchResults = useKnowledgeSearch(search)

  const source = search.trim().length >= 2 ? searchResults : all
  const items = useMemo<KnowledgeBase[]>(() => {
    if (search.trim().length >= 2) {
      return searchResults.data?.items ?? []
    }
    return all.data?.items ?? []
  }, [search, all.data, searchResults.data])

  return (
    <div className="flex flex-col h-full bg-canvas">
      <PageHeader
        title="Knowledge"
        description="Sources Kyla searches when answering"
        actions={
          <button
            type="button"
            className="inline-flex items-center gap-1.5 h-8 px-3 rounded-sm text-md font-medium bg-accent text-accent-fg hover:bg-accent-hover"
          >
            <IconPlus className="size-3.5" />
            New source
          </button>
        }
      />

      <div className="flex items-center gap-2 px-3 h-9 border-b border-border">
        <div className="relative flex-1 max-w-md">
          <IconSearch
            className="absolute start-2 top-1/2 -translate-y-1/2 size-3.5 text-fg-muted"
            aria-hidden
          />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search knowledge…"
            className="h-7 ps-7 text-base"
            aria-label="Search knowledge"
          />
        </div>
        <span className="ms-auto font-mono text-xs text-fg-muted">
          {items.length}
        </span>
      </div>

      {source.isPending ? (
        <div className="flex-1 p-2">
          <ListRowSkeleton count={8} />
        </div>
      ) : source.isError ? (
        <div className="flex-1 flex items-center justify-center">
          <ErrorState
            title="Couldn't load knowledge"
            description={(source.error as Error).message}
            onRetry={() => void source.refetch()}
          />
        </div>
      ) : items.length === 0 ? (
        <div className="flex-1 flex items-center justify-center">
          <EmptyState
            icon={<IconBook className="size-5" />}
            title={search ? "No matches" : "No knowledge sources yet"}
            description={
              search
                ? "Try different search terms or clear the filter."
                : "Add text articles or crawl public websites for Kyla to draw on."
            }
          />
        </div>
      ) : (
        <ul className="flex-1 overflow-y-auto p-1" role="list">
          {items.map((item) => (
            <li key={item.id}>
              <KnowledgeRow item={item} />
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}

function KnowledgeRow({ item }: { item: KnowledgeBase }) {
  const Icon = item.type === KnowledgeBaseType.WEBSITE ? IconWorld : IconBook
  return (
    <Link
      to={`/knowledge/${item.id}`}
      className={cn(
        "flex items-center gap-3 px-3 py-2 rounded-sm",
        "hover:bg-subtle transition-colors",
      )}
    >
      <div
        className={cn(
          "size-7 rounded-sm flex items-center justify-center shrink-0",
          "bg-accent-subtle text-fg-secondary",
        )}
        aria-hidden
      >
        <Icon className="size-3.5" />
      </div>
      <div className="min-w-0 flex-1">
        <div className="text-md font-medium text-fg truncate">{item.name}</div>
        {item.description && (
          <div className="text-sm text-fg-muted truncate">{item.description}</div>
        )}
      </div>
      <KbTypeBadge type={item.type} />
      <span className="w-16 text-end font-mono text-xs text-fg-muted">
        {relativeShort(item.updatedAt)}
      </span>
    </Link>
  )
}

function KbTypeBadge({ type }: { type: KnowledgeBaseType }) {
  const meta =
    type === KnowledgeBaseType.WEBSITE
      ? { label: "Website", cls: "bg-info-subtle text-info" }
      : type === KnowledgeBaseType.DA_FAQ
        ? { label: "FAQ", cls: "bg-warn-subtle text-warn" }
        : { label: "Text", cls: "bg-subtle text-fg-muted" }
  return (
    <span
      className={cn(
        "inline-flex items-center h-5 px-1.5 rounded-xs text-xs font-medium",
        meta.cls,
      )}
    >
      {meta.label}
    </span>
  )
}

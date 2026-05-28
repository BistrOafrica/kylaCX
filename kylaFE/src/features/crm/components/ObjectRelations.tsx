import { useMemo } from "react"
import { IconLink } from "@tabler/icons-react"
import { CardSkeleton, EmptyState } from "@/design-system"
import { useObjectRelations } from "../hooks/queries"
import type { ObjectRelation } from "@/pb/object_core"

/**
 * ObjectRelations — list of links to / from the current object.
 *
 * Groups by `relation` name (e.g. `belongs_to_company`, `has_deal`)
 * so the user gets a clean grouped view. Linking to actual related
 * objects awaits an Object resolver (we have the IDs but not the
 * type slug per relation); F2.x can swap the ID display for resolved
 * names once a batch resolver lands.
 */
export function ObjectRelations({ objectId }: { objectId: string }) {
  const relations = useObjectRelations(objectId)

  const grouped = useMemo(() => {
    const map = new Map<string, ObjectRelation[]>()
    for (const r of relations.data ?? []) {
      const arr = map.get(r.relation) ?? []
      arr.push(r)
      map.set(r.relation, arr)
    }
    return [...map.entries()]
  }, [relations.data])

  if (relations.isPending) {
    return <CardSkeleton lines={2} />
  }

  if (grouped.length === 0) {
    return (
      <EmptyState
        icon={<IconLink className="size-4" />}
        title="No related objects"
        description="Link contacts, deals, or tickets to see them here."
        size="sm"
      />
    )
  }

  return (
    <div className="space-y-3">
      {grouped.map(([relation, items]) => (
        <div key={relation} className="space-y-1.5">
          <div className="text-xs font-mono uppercase tracking-wider text-fg-muted">
            {relation.replaceAll("_", " ")}
          </div>
          <ul className="space-y-1">
            {items.map((rel) => (
              <li
                key={rel.id}
                className="flex items-center gap-2 px-2 h-7 rounded-sm border border-border bg-surface text-base"
              >
                <IconLink className="size-3 text-fg-muted shrink-0" aria-hidden />
                <span className="font-mono text-sm text-fg-secondary truncate flex-1">
                  {rel.fromId === objectId ? rel.toId : rel.fromId}
                </span>
              </li>
            ))}
          </ul>
        </div>
      ))}
    </div>
  )
}

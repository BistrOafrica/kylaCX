import { useEffect, useMemo } from "react"
import { useNavigate, useParams } from "react-router-dom"
import { useHotkeys } from "react-hotkeys-hook"
import { useInbox } from "./queries"
import type { InboxFilters } from "../api/conversations"

/**
 * j / k navigation across the active inbox list.
 *
 * Reads the same query the list is rendering, so navigation always
 * stays in sync with what the agent sees. Wrapping at the ends is
 * intentional — Linear-style — so a single press never gets stuck.
 */
export function useInboxNavigation(filters: InboxFilters) {
  const navigate = useNavigate()
  const { id } = useParams()
  const inbox = useInbox(filters)

  const flat = useMemo(
    () => inbox.data?.pages.flatMap((p) => p.conversations) ?? [],
    [inbox.data],
  )

  const currentIndex = useMemo(
    () => (id ? flat.findIndex((c) => c.id === id) : -1),
    [flat, id],
  )

  // Auto-select the first row when entering /inbox with no selection
  // and rows exist. Disabled while inbox is still loading to avoid
  // flashing.
  useEffect(() => {
    if (!id && flat.length > 0 && !inbox.isPending) {
      navigate(`/inbox/${flat[0]!.id}`, { replace: true })
    }
  }, [id, flat, navigate, inbox.isPending])

  useHotkeys(
    "j",
    () => {
      if (!flat.length) return
      const next = currentIndex >= 0 ? (currentIndex + 1) % flat.length : 0
      navigate(`/inbox/${flat[next]!.id}`)
    },
    { enableOnFormTags: false },
    [flat, currentIndex],
  )

  useHotkeys(
    "k",
    () => {
      if (!flat.length) return
      const prev =
        currentIndex > 0
          ? currentIndex - 1
          : currentIndex < 0
            ? 0
            : flat.length - 1
      navigate(`/inbox/${flat[prev]!.id}`)
    },
    { enableOnFormTags: false },
    [flat, currentIndex],
  )
}

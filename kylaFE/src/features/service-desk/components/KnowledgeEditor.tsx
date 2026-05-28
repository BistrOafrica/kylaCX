import { useState } from "react"
import { Link } from "react-router-dom"
import { IconArrowLeft, IconLoader2, IconDeviceFloppy } from "@tabler/icons-react"
import { toast } from "sonner"
import {
  CardSkeleton,
  ErrorState,
  RichEditor,
  Surface,
} from "@/design-system"
import { Input } from "@/components/ui/input"
import { cn } from "@/lib/utils"
import {
  useKnowledgeBase,
  useUpdateKnowledgeBase,
} from "../hooks/queries"
import { KnowledgeBaseType, KnowledgeBase } from "../api/knowledge"

/**
 * KnowledgeEditor — full-page editor for a single knowledge source.
 *
 * For TEXT-type sources we use the Tiptap RichEditor; for WEBSITE-type
 * we expose the crawl URL. Save button persists via UpdateKnowledgeBase.
 *
 * The outer component just loads; the inner form is re-keyed by id so
 * its initial state always captures the freshly-loaded values without
 * a setState-in-effect dance.
 */
export function KnowledgeEditor({ id }: { id: string }) {
  const kb = useKnowledgeBase(id)

  if (kb.isPending) {
    return (
      <div className="p-6 space-y-3">
        <CardSkeleton lines={3} />
      </div>
    )
  }

  if (kb.isError || !kb.data) {
    return (
      <div className="p-12">
        <ErrorState
          title="Couldn't load source"
          description={(kb.error as Error | undefined)?.message ?? "Not found"}
          onRetry={() => void kb.refetch()}
        />
      </div>
    )
  }

  return <EditorForm key={kb.data.id} data={kb.data} />
}

function EditorForm({ data }: { data: KnowledgeBase }) {
  const update = useUpdateKnowledgeBase()
  const [name, setName] = useState(data.name)
  const [description, setDescription] = useState(data.description)
  const [text, setText] = useState(data.text)
  const [url, setUrl] = useState(data.url)
  const [dirty, setDirty] = useState(false)

  const isWebsite = data.type === KnowledgeBaseType.WEBSITE

  const onSave = async () => {
    const next = KnowledgeBase.create({
      ...data,
      name,
      description,
      text: isWebsite ? "" : text,
      url: isWebsite ? url : "",
    }) as KnowledgeBase
    try {
      await update.mutateAsync(next)
      setDirty(false)
      toast.success("Saved")
    } catch {
      /* mutationCache toasts errors */
    }
  }

  return (
    <div className="h-full overflow-y-auto bg-canvas">
      <header className="sticky top-0 z-10 flex items-center gap-3 px-6 py-3 border-b border-border bg-canvas">
        <Link
          to="/knowledge"
          className={cn(
            "inline-flex items-center justify-center size-7 rounded-sm",
            "text-fg-secondary hover:text-fg hover:bg-subtle transition-colors",
          )}
          aria-label="Back to knowledge"
        >
          <IconArrowLeft className="size-4" />
        </Link>
        <div className="min-w-0 flex-1">
          <Input
            value={name}
            onChange={(e) => {
              setName(e.target.value)
              setDirty(true)
            }}
            placeholder="Source name"
            className="h-8 text-lg font-semibold border-0 bg-transparent shadow-none px-0"
          />
        </div>
        <button
          type="button"
          onClick={() => void onSave()}
          disabled={!dirty || update.isPending}
          className={cn(
            "inline-flex items-center gap-1.5 h-8 px-3 rounded-sm text-md font-medium",
            "bg-accent text-accent-fg hover:bg-accent-hover",
            "disabled:opacity-40 disabled:pointer-events-none transition-colors",
          )}
        >
          {update.isPending ? (
            <IconLoader2 className="size-3.5 animate-spin" />
          ) : (
            <IconDeviceFloppy className="size-3.5" />
          )}
          {dirty ? "Save" : "Saved"}
        </button>
      </header>

      <div className="p-6 max-w-3xl space-y-4">
        <Surface level={0} radius="sm" className="p-3 space-y-2">
          <label className="block">
            <span className="text-xs font-mono uppercase tracking-wider text-fg-muted">
              Description
            </span>
            <Input
              value={description}
              onChange={(e) => {
                setDescription(e.target.value)
                setDirty(true)
              }}
              placeholder="Short summary of what this source contains"
              className="h-8 mt-1"
            />
          </label>
        </Surface>

        {isWebsite ? (
          <Surface level={0} radius="sm" className="p-3 space-y-2">
            <label className="block">
              <span className="text-xs font-mono uppercase tracking-wider text-fg-muted">
                URL to crawl
              </span>
              <Input
                value={url}
                onChange={(e) => {
                  setUrl(e.target.value)
                  setDirty(true)
                }}
                placeholder="https://docs.example.com"
                className="h-8 mt-1 font-mono"
              />
            </label>
          </Surface>
        ) : (
          <div className="space-y-2">
            <div className="text-xs font-mono uppercase tracking-wider text-fg-muted">
              Content
            </div>
            <RichEditor
              value={text}
              onChange={(html) => {
                setText(html)
                setDirty(true)
              }}
              placeholder="Write the article body…"
              minHeight="24rem"
            />
          </div>
        )}
      </div>
    </div>
  )
}

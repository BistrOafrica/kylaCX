import { useEditor, EditorContent, type Editor } from "@tiptap/react"
import StarterKit from "@tiptap/starter-kit"
import Placeholder from "@tiptap/extension-placeholder"
import Link from "@tiptap/extension-link"
import { useEffect } from "react"
import {
  IconBold,
  IconItalic,
  IconStrikethrough,
  IconList,
  IconListNumbers,
  IconQuote,
  IconLink,
  IconCode,
  IconH1,
  IconH2,
} from "@tabler/icons-react"
import { cn } from "@/lib/utils"

/**
 * RichEditor — Tiptap-powered editor used by the KB editor, ticket
 * replies (F3.x), and any future surface that needs rich text. Exposes
 * an HTML string via onChange so consumers can persist it to the
 * backend's text fields.
 */
export interface RichEditorProps {
  value: string
  onChange: (html: string) => void
  placeholder?: string
  editable?: boolean
  className?: string
  minHeight?: string
}

export function RichEditor({
  value,
  onChange,
  placeholder = "Write something…",
  editable = true,
  className,
  minHeight = "12rem",
}: RichEditorProps) {
  const editor = useEditor({
    extensions: [
      StarterKit,
      Placeholder.configure({ placeholder }),
      Link.configure({
        openOnClick: false,
        autolink: true,
        HTMLAttributes: { rel: "noreferrer nofollow", target: "_blank" },
      }),
    ],
    content: value,
    editable,
    onUpdate: ({ editor }) => onChange(editor.getHTML()),
    editorProps: {
      attributes: {
        class: cn(
          "tiptap prose prose-sm max-w-none focus:outline-none",
          "text-base text-fg",
        ),
      },
    },
  })

  // Sync external value changes (e.g. when switching docs) into the editor.
  useEffect(() => {
    if (!editor) return
    if (editor.getHTML() === value) return
    editor.commands.setContent(value, { emitUpdate: false })
  }, [value, editor])

  if (!editor) return null

  return (
    <div
      className={cn(
        "rounded-md border border-border bg-surface overflow-hidden",
        className,
      )}
    >
      {editable && <Toolbar editor={editor} />}
      <div
        className="px-3 py-3 overflow-y-auto"
        style={{ minHeight }}
      >
        <EditorContent editor={editor} />
      </div>
    </div>
  )
}

function Toolbar({ editor }: { editor: Editor }) {
  const btn = (active: boolean) =>
    cn(
      "inline-flex items-center justify-center size-7 rounded-xs transition-colors",
      "text-fg-secondary hover:text-fg hover:bg-subtle",
      active && "bg-accent-subtle text-fg",
    )

  return (
    <div className="flex items-center gap-px px-2 h-9 border-b border-border bg-subtle">
      <button
        type="button"
        onClick={() => editor.chain().focus().toggleHeading({ level: 1 }).run()}
        className={btn(editor.isActive("heading", { level: 1 }))}
        aria-label="Heading 1"
        title="H1"
      >
        <IconH1 className="size-3.5" />
      </button>
      <button
        type="button"
        onClick={() => editor.chain().focus().toggleHeading({ level: 2 }).run()}
        className={btn(editor.isActive("heading", { level: 2 }))}
        aria-label="Heading 2"
        title="H2"
      >
        <IconH2 className="size-3.5" />
      </button>
      <span className="w-px h-4 bg-border-strong mx-1" aria-hidden />
      <button
        type="button"
        onClick={() => editor.chain().focus().toggleBold().run()}
        className={btn(editor.isActive("bold"))}
        aria-label="Bold"
        title="Bold"
      >
        <IconBold className="size-3.5" />
      </button>
      <button
        type="button"
        onClick={() => editor.chain().focus().toggleItalic().run()}
        className={btn(editor.isActive("italic"))}
        aria-label="Italic"
        title="Italic"
      >
        <IconItalic className="size-3.5" />
      </button>
      <button
        type="button"
        onClick={() => editor.chain().focus().toggleStrike().run()}
        className={btn(editor.isActive("strike"))}
        aria-label="Strikethrough"
        title="Strikethrough"
      >
        <IconStrikethrough className="size-3.5" />
      </button>
      <span className="w-px h-4 bg-border-strong mx-1" aria-hidden />
      <button
        type="button"
        onClick={() => editor.chain().focus().toggleBulletList().run()}
        className={btn(editor.isActive("bulletList"))}
        aria-label="Bullet list"
        title="Bullets"
      >
        <IconList className="size-3.5" />
      </button>
      <button
        type="button"
        onClick={() => editor.chain().focus().toggleOrderedList().run()}
        className={btn(editor.isActive("orderedList"))}
        aria-label="Numbered list"
        title="Numbered"
      >
        <IconListNumbers className="size-3.5" />
      </button>
      <button
        type="button"
        onClick={() => editor.chain().focus().toggleBlockquote().run()}
        className={btn(editor.isActive("blockquote"))}
        aria-label="Blockquote"
        title="Quote"
      >
        <IconQuote className="size-3.5" />
      </button>
      <button
        type="button"
        onClick={() => editor.chain().focus().toggleCodeBlock().run()}
        className={btn(editor.isActive("codeBlock"))}
        aria-label="Code block"
        title="Code"
      >
        <IconCode className="size-3.5" />
      </button>
      <button
        type="button"
        onClick={() => {
          const prev = editor.getAttributes("link").href as string | undefined
          const url = window.prompt("Link URL", prev ?? "https://")
          if (url === null) return
          if (url === "") {
            editor.chain().focus().unsetLink().run()
          } else {
            editor.chain().focus().setLink({ href: url }).run()
          }
        }}
        className={btn(editor.isActive("link"))}
        aria-label="Link"
        title="Link"
      >
        <IconLink className="size-3.5" />
      </button>
    </div>
  )
}

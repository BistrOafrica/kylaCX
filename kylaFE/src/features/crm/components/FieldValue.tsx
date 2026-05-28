import { FieldType, type FieldDefinition } from "@/pb/object_core"
import { cn } from "@/lib/utils"

/**
 * FieldValue — schema-aware cell renderer.
 *
 * The default `formatFieldValue` returns a string for table cells; this
 * component returns richer JSX (color chips for SELECT, mailto/tel
 * links for EMAIL/PHONE, etc.) and is used in the detail view + the
 * Linear-dense table.
 */
export function FieldValue({
  field,
  value,
}: {
  field: FieldDefinition
  value: unknown
}) {
  if (value === undefined || value === null || value === "") {
    return <span className="text-fg-muted">—</span>
  }

  switch (field.type) {
    case FieldType.EMAIL:
      return (
        <a
          href={`mailto:${value}`}
          className="text-fg-link hover:underline truncate"
        >
          {String(value)}
        </a>
      )

    case FieldType.PHONE:
      return (
        <a
          href={`tel:${value}`}
          className="text-fg-link hover:underline truncate font-mono text-sm"
        >
          {String(value)}
        </a>
      )

    case FieldType.URL: {
      const href = String(value)
      const display = href.replace(/^https?:\/\//, "")
      return (
        <a
          href={href}
          target="_blank"
          rel="noreferrer"
          className="text-fg-link hover:underline truncate"
        >
          {display}
        </a>
      )
    }

    case FieldType.CURRENCY: {
      const n = Number(value)
      const formatted = Number.isNaN(n)
        ? String(value)
        : new Intl.NumberFormat(undefined, {
            style: "currency",
            currency: "USD",
            maximumFractionDigits: 0,
          }).format(n)
      return <span className="font-mono tabular-nums">{formatted}</span>
    }

    case FieldType.NUMBER:
      return (
        <span className="font-mono tabular-nums">
          {Number(value).toLocaleString()}
        </span>
      )

    case FieldType.DATE:
    case FieldType.DATETIME: {
      const d = new Date(String(value))
      if (Number.isNaN(d.getTime())) return <span>{String(value)}</span>
      return (
        <span className="font-mono text-sm">
          {field.type === FieldType.DATETIME
            ? d.toLocaleString()
            : d.toLocaleDateString()}
        </span>
      )
    }

    case FieldType.BOOLEAN:
      return value ? (
        <span className="inline-flex items-center h-4 px-1.5 rounded-xs text-[10px] font-medium bg-success-subtle text-success">
          Yes
        </span>
      ) : (
        <span className="inline-flex items-center h-4 px-1.5 rounded-xs text-[10px] font-medium bg-subtle text-fg-muted">
          No
        </span>
      )

    case FieldType.SELECT: {
      const opt = field.options?.find((o) => o.value === value)
      if (!opt) return <span>{String(value)}</span>
      return (
        <span
          className={cn(
            "inline-flex items-center h-5 px-1.5 rounded-xs text-xs font-medium",
            "border border-border",
          )}
          style={
            opt.color
              ? { background: `${opt.color}1F`, color: opt.color }
              : undefined
          }
        >
          {opt.label}
        </span>
      )
    }

    case FieldType.MULTI: {
      if (!Array.isArray(value)) return <span>{String(value)}</span>
      return (
        <div className="flex flex-wrap gap-1">
          {value.map((v) => {
            const opt = field.options?.find((o) => o.value === v)
            return (
              <span
                key={String(v)}
                className="inline-flex items-center h-4 px-1.5 rounded-xs text-[10px] font-medium border border-border"
                style={
                  opt?.color
                    ? { background: `${opt.color}1F`, color: opt.color }
                    : undefined
                }
              >
                {opt?.label ?? String(v)}
              </span>
            )
          })}
        </div>
      )
    }

    case FieldType.USER:
    case FieldType.RELATION:
      return (
        <span className="font-mono text-sm text-fg-secondary truncate">
          {String(value).slice(0, 12)}
          {String(value).length > 12 && "…"}
        </span>
      )

    default:
      return <span className="truncate">{String(value)}</span>
  }
}

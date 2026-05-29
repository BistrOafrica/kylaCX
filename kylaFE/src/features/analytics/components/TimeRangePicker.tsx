import { IconChevronDown, IconCalendar } from "@tabler/icons-react"
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
} from "@/components/ui/dropdown-menu"
import { cn } from "@/lib/utils"
import { PRESETS, type TimeRangePreset } from "../utils/time-range"

export function TimeRangePicker({
  value,
  onChange,
}: {
  value: TimeRangePreset
  onChange: (v: TimeRangePreset) => void
}) {
  const current = PRESETS.find((p) => p.id === value) ?? PRESETS[0]!

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        className={cn(
          "inline-flex items-center gap-1.5 h-7 px-2.5 rounded-sm text-md",
          "border border-border bg-surface hover:bg-subtle transition-colors",
        )}
        aria-label="Time range"
      >
        <IconCalendar className="size-3.5 text-fg-muted" aria-hidden />
        {current.label}
        <IconChevronDown className="size-3 text-fg-muted" aria-hidden />
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuRadioGroup
          value={value}
          onValueChange={(v) => onChange(v as TimeRangePreset)}
        >
          {PRESETS.map((p) => (
            <DropdownMenuRadioItem key={p.id} value={p.id}>
              {p.label}
            </DropdownMenuRadioItem>
          ))}
        </DropdownMenuRadioGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

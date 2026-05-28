/**
 * Kyla design system — single import surface.
 *
 *   import { Kbd, EmptyState, tokens } from "@/design-system"
 *
 * Tokens are imported via the side-effect CSS in src/index.css, so this
 * module exports only the typed accessor + the new primitives and patterns.
 * The legacy shadcn primitives in src/components/ui/ keep their own
 * import paths until they are migrated piece by piece.
 */

// Tokens (typed accessor; CSS is loaded at app boot)
export { tokens } from "./tokens/tokens"
export type { Channel, StatusKind } from "./tokens/tokens"

// Primitives — new, Linear-dense
export { Kbd } from "./primitives/Kbd"
export type { KbdProps } from "./primitives/Kbd"
export { StatusDot } from "./primitives/StatusDot"
export type { StatusDotProps, StatusTone } from "./primitives/StatusDot"
export { Surface } from "./primitives/Surface"
export type { SurfaceProps } from "./primitives/Surface"
export { ChannelBadge } from "./primitives/ChannelBadge"
export type { ChannelBadgeProps } from "./primitives/ChannelBadge"

// Patterns — composed primitives
export { EmptyState } from "./patterns/EmptyState"
export type { EmptyStateProps } from "./patterns/EmptyState"
export { ErrorState } from "./patterns/ErrorState"
export type { ErrorStateProps } from "./patterns/ErrorState"
export {
  Skeleton,
  ListRowSkeleton,
  CardSkeleton,
} from "./patterns/LoadingSkeleton"
export type { SkeletonProps } from "./patterns/LoadingSkeleton"
export { PageHeader } from "./patterns/PageHeader"
export type { PageHeaderProps } from "./patterns/PageHeader"
export { StreamingText } from "./patterns/StreamingText"
export type { StreamingTextProps } from "./patterns/StreamingText"
export { AISuggestionCard } from "./patterns/AISuggestion"
export type { AISuggestionCardProps } from "./patterns/AISuggestion"
export { RichEditor } from "./patterns/RichEditor"
export type { RichEditorProps } from "./patterns/RichEditor"

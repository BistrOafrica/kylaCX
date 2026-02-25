import { type Icon } from "@tabler/icons-react"
import { type ReactNode } from "react"

import { Badge } from "@/components/ui/badge"
import {
  Card,
  CardAction,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

export interface StatCardProps {
  title: string
  value: string | number
  description?: string
  trend?: {
    value: string
    icon: Icon
    direction?: "up" | "down"
  }
  footer?: {
    label: string | ReactNode
    sublabel?: string
  }
  icon?: Icon
}

export function StatCard({
  title,
  value,
  description,
  trend,
  footer,
  icon: Icon,
}: StatCardProps) {
  return (
    <Card className="@container/card">
      <CardHeader>
        <CardDescription>{title}</CardDescription>
        <CardTitle className="text-2xl font-semibold tabular-nums @[250px]/card:text-3xl">
          {value}
        </CardTitle>
        {trend && (
          <CardAction>
            <Badge variant="outline">
              <trend.icon />
              {trend.value}
            </Badge>
          </CardAction>
        )}
        {Icon && !trend && (
          <CardAction>
            <Icon className="text-muted-foreground size-4" />
          </CardAction>
        )}
      </CardHeader>
      {footer && (
        <CardFooter className="flex-col items-start gap-1.5 text-sm">
          <div className="line-clamp-1 flex gap-2 font-medium">
            {footer.label}
            {trend?.icon && <trend.icon className="size-4" />}
          </div>
          {footer.sublabel && (
            <div className="text-muted-foreground">{footer.sublabel}</div>
          )}
          {description && !footer.sublabel && (
            <div className="text-muted-foreground">{description}</div>
          )}
        </CardFooter>
      )}
      {!footer && description && (
        <CardFooter className="flex-col items-start gap-1.5 text-sm">
          <div className="text-muted-foreground">{description}</div>
        </CardFooter>
      )}
    </Card>
  )
}

export function StatCardGrid({ children }: { children: ReactNode }) {
  return (
    <div className="*:data-[slot=card]:from-primary/5 *:data-[slot=card]:to-card dark:*:data-[slot=card]:bg-card grid grid-cols-1 gap-4 px-4 *:data-[slot=card]:bg-gradient-to-t *:data-[slot=card]:shadow-xs lg:px-0 @xl/main:grid-cols-2 @5xl/main:grid-cols-4">
      {children}
    </div>
  )
}

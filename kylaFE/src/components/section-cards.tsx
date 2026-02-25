import { IconTrendingDown, IconTrendingUp } from "@tabler/icons-react"

import { StatCard, StatCardGrid } from "@/components/stat-card"

export function SectionCards() {
  return (
    <StatCardGrid>
      <StatCard
        title="Total Revenue"
        value="$1,250.00"
        trend={{
          value: "+12.5%",
          icon: IconTrendingUp,
        }}
        footer={{
          label: "Trending up this month",
          sublabel: "Visitors for the last 6 months",
        }}
      />
      <StatCard
        title="New Customers"
        value="1,234"
        trend={{
          value: "-20%",
          icon: IconTrendingDown,
        }}
        footer={{
          label: "Down 20% this period",
          sublabel: "Acquisition needs attention",
        }}
      />
      <StatCard
        title="Active Accounts"
        value="45,678"
        trend={{
          value: "+12.5%",
          icon: IconTrendingUp,
        }}
        footer={{
          label: "Strong user retention",
          sublabel: "Engagement exceed targets",
        }}
      />
      <StatCard
        title="Growth Rate"
        value="4.5%"
        trend={{
          value: "+4.5%",
          icon: IconTrendingUp,
        }}
        footer={{
          label: "Steady performance increase",
          sublabel: "Meets growth projections",
        }}
      />
    </StatCardGrid>
  )
}

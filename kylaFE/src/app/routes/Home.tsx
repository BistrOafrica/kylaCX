import { useTranslation } from "react-i18next"
import { IconSparkles } from "@tabler/icons-react"
import { EmptyState, PageHeader } from "@/design-system"

/**
 * Home — the post-login landing page in F0.
 *
 * F1 will replace this with the omnichannel inbox. Until then it
 * shows a friendly empty state explaining the workbench is being
 * built out, with the AI rail and ⌘K already functional.
 */
export function Home() {
  const { t } = useTranslation()

  return (
    <div className="h-full flex flex-col">
      <PageHeader
        title={t("common.appName")}
        description={t("common.tagline")}
      />
      <div className="flex-1 flex items-center justify-center p-6">
        <EmptyState
          icon={<IconSparkles className="size-5" />}
          title="Workbench ready"
          description={
            "Press ⌘K to explore commands or ⌘J to open the Kyla copilot. " +
            "The inbox, CRM, tickets and automation surfaces are coming in F1–F4."
          }
          size="lg"
        />
      </div>
    </div>
  )
}

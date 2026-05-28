import { Link } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { EmptyState } from "@/design-system"

export function NotFound() {
  const { t } = useTranslation()
  return (
    <div className="min-h-dvh flex items-center justify-center bg-canvas p-6">
      <EmptyState
        title="404"
        description="That page is somewhere — but not here."
        action={
          <Link
            to="/"
            className="inline-flex items-center h-8 px-3 rounded-sm border border-border hover:bg-subtle text-md"
          >
            {t("common.back")}
          </Link>
        }
        size="lg"
      />
    </div>
  )
}

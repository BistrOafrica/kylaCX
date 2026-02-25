import { SettingsDialog as SettingsDialogComponent } from "@/components/settings-dialog"

export { SettingsDialogComponent as SettingsDialog }

export default function Page() {
  return (
    <div className="flex h-svh items-center justify-center">
      <SettingsDialogComponent />
    </div>
  )
}

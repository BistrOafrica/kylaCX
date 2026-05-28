import * as React from "react"
import { SupportConversation } from "./support/SupportConversation"
import { SupportSidebar } from "./support/SupportSidebar"
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar"

export default function SupportPage() {
  return (
    <SidebarProvider
      style={
        {
          "--sidebar-width": "350px",
        } as React.CSSProperties
      }
    >
      <SupportSidebar />
      <SidebarInset className="flex flex-col overflow-hidden">
        <SupportConversation />
      </SidebarInset>
    </SidebarProvider>
  )
}

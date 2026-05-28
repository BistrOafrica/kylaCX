"use client"

import * as React from "react"
import { useLocation, useNavigate } from "react-router-dom"

import { Label } from "@/components/ui/label"
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarInput,
} from "@/components/ui/sidebar"
import { Switch } from "@/components/ui/switch"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"

// Support ticket data organized by labels
const ticketsByLabel = {
  all: [
    {
      id: 1,
      from: "John Doe",
      email: "john@example.com",
      subject: "Payment Issue - Urgent",
      preview: "I'm having trouble processing my payment...",
      time: "2 hours ago",
      unread: true,
      priority: "high",
      label: "urgent",
    },
    {
      id: 2,
      from: "Sarah Smith",
      email: "sarah@example.com",
      subject: "Feature Request",
      preview: "Would love to see a dark mode option...",
      time: "5 hours ago",
      unread: true,
      priority: "medium",
      label: "feature",
    },
    {
      id: 3,
      from: "Mike Johnson",
      email: "mike@example.com",
      subject: "Thank you!",
      preview: "Great service, very helpful support team...",
      time: "1 day ago",
      unread: false,
      priority: "low",
      label: "general",
    },
    {
      id: 4,
      from: "Emily Davis",
      email: "emily@example.com",
      subject: "Bug Report - Login Issues",
      preview: "Unable to login after password reset...",
      time: "3 hours ago",
      unread: true,
      priority: "high",
      label: "bug",
    },
    {
      id: 5,
      from: "Tom Wilson",
      email: "tom@example.com",
      subject: "Billing Question",
      preview: "Can you explain the charges on my account?",
      time: "Yesterday",
      unread: true,
      priority: "medium",
      label: "billing",
    },
  ],
}

const labelFilters = [
  { title: "All", value: "all", color: "#6b7280" },
  { title: "Urgent", value: "urgent", color: "#ef4444" },
  { title: "Bug", value: "bug", color: "#f97316" },
  { title: "Feature", value: "feature", color: "#3b82f6" },
  { title: "Billing", value: "billing", color: "#8b5cf6" },
  { title: "General", value: "general", color: "#6b7280" },
]

export function SupportSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const location = useLocation()
  const navigate = useNavigate()
  const [searchQuery, setSearchQuery] = React.useState("")
  const [showUnreadsOnly, setShowUnreadsOnly] = React.useState(false)

  // Get current label from URL params
  const searchParams = new URLSearchParams(location.search)
  const currentLabel = searchParams.get("label") || "all"
//   const currentFilter = searchParams.get("filter") || "inbox"

  // Filter tickets based on current label and filter
  const getFilteredTickets = () => {
    let tickets = ticketsByLabel.all

    // Filter by label if not "all"
    if (currentLabel !== "all") {
      tickets = tickets.filter((t) => t.label === currentLabel)
    }

    // Filter by folder (additional filtering logic can be added when archived/deleted fields are added)
    // if (currentFilter === "inbox") {
    //   tickets = tickets.filter((t) => !t.archived && !t.deleted)
    // }

    // Filter by search
    if (searchQuery) {
      tickets = tickets.filter(
        (t) =>
          t.subject.toLowerCase().includes(searchQuery.toLowerCase()) ||
          t.from.toLowerCase().includes(searchQuery.toLowerCase()) ||
          t.preview.toLowerCase().includes(searchQuery.toLowerCase())
      )
    }

    // Filter by unread
    if (showUnreadsOnly) {
      tickets = tickets.filter((t) => t.unread)
    }

    return tickets
  }

  const filteredTickets = getFilteredTickets()

  const handleTicketClick = (ticketId: number) => {
    const params = new URLSearchParams(location.search)
    params.set("ticket", ticketId.toString())
    navigate(`/support?${params.toString()}`)
  }

  const handleLabelClick = (label: string) => {
    const params = new URLSearchParams(location.search)
    params.set("label", label)
    params.delete("ticket") // Clear ticket selection when changing label
    navigate(`/support?${params.toString()}`)
  }

  return (
    <Sidebar collapsible="none" className="hidden min-w-[350px] md:flex border-r" {...props}>
      <SidebarHeader className="gap-3.5 border-b p-4">
        {/* Label filters */}
        <div className="flex flex-wrap gap-1">
          {labelFilters.map((label) => (
            <Button
              key={label.value}
              variant={currentLabel === label.value ? "default" : "outline"}
              size="sm"
              onClick={() => handleLabelClick(label.value)}
              className="h-7 text-xs"
            >
              <div
                className="size-2 rounded-full mr-1.5"
                style={{ backgroundColor: label.color }}
              />
              {label.title}
            </Button>
          ))}
        </div>

        <div className="flex w-full items-center justify-between">
          <Label className="flex items-center gap-2 text-sm">
            <span>Unreads only</span>
            <Switch
              className="shadow-none"
              checked={showUnreadsOnly}
              onCheckedChange={setShowUnreadsOnly}
            />
          </Label>
        </div>
        <SidebarInput
          placeholder="Type to search..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
        />
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup className="px-0">
          <SidebarGroupContent>
            {filteredTickets.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
                <p className="text-sm">No tickets found</p>
              </div>
            ) : (
              filteredTickets.map((ticket) => (
                <button
                  key={ticket.id}
                  onClick={() => handleTicketClick(ticket.id)}
                  className="hover:bg-sidebar-accent hover:text-sidebar-accent-foreground flex flex-col items-start gap-2 border-b p-4 text-sm leading-tight whitespace-nowrap last:border-b-0 text-left w-full transition-colors"
                >
                  <div className="flex w-full items-center gap-2">
                    <span className={ticket.unread ? "font-semibold" : ""}>
                      {ticket.from}
                    </span>
                    <span className="ml-auto text-xs text-muted-foreground">
                      {ticket.time}
                    </span>
                  </div>
                  <div className="flex items-center gap-2 w-full">
                    <span className={`font-medium flex-1 ${ticket.unread ? "" : "text-muted-foreground"}`}>
                      {ticket.subject}
                    </span>
                    {ticket.unread && (
                      <Badge variant="default" className="h-5 px-1.5 text-xs">
                        New
                      </Badge>
                    )}
                  </div>
                  <span className="line-clamp-2 w-[300px] text-xs whitespace-break-spaces text-muted-foreground">
                    {ticket.preview}
                  </span>
                  <div className="flex items-center gap-1">
                    <div
                      className="size-2 rounded-full"
                      style={{
                        backgroundColor:
                          ticket.label === "urgent"
                            ? "#ef4444"
                            : ticket.label === "bug"
                              ? "#f97316"
                              : ticket.label === "feature"
                                ? "#3b82f6"
                                : ticket.label === "billing"
                                  ? "#8b5cf6"
                                  : "#6b7280",
                      }}
                    />
                    <span className="text-xs text-muted-foreground capitalize">
                      {ticket.label}
                    </span>
                  </div>
                </button>
              ))
            )}
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
    </Sidebar>
  )
}

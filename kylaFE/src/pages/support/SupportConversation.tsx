"use client"

import * as React from "react"
import { useLocation } from "react-router-dom"
import {
  IconSend,
  IconPaperclip,
  IconDotsVertical,
  IconStar,
  IconTrash,
  IconArchive,
  IconTag,
} from "@tabler/icons-react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Separator } from "@/components/ui/separator"
import { SidebarTrigger } from "@/components/ui/sidebar"
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Badge } from "@/components/ui/badge"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"

const ticketData = {
  1: {
    id: 1,
    from: "John Doe",
    email: "john@example.com",
    subject: "Payment Issue - Urgent",
    label: "urgent",
    priority: "high",
    messages: [
      {
        id: 1,
        sender: "John Doe",
        email: "john@example.com",
        content:
          "Hi, I'm having trouble processing my payment. When I try to submit the payment form, I get an error message saying 'Payment gateway unavailable'. I've tried multiple times with different cards but the same issue persists.",
        timestamp: "2 hours ago",
        isAgent: false,
      },
      {
        id: 2,
        sender: "Support Agent",
        email: "support@kyla.cx",
        content:
          "Thank you for reaching out. I apologize for the inconvenience. Let me look into this issue for you. Can you please provide the error code or take a screenshot of the error message?",
        timestamp: "1 hour ago",
        isAgent: true,
      },
    ],
  },
  2: {
    id: 2,
    from: "Sarah Smith",
    email: "sarah@example.com",
    subject: "Feature Request",
    label: "feature",
    priority: "medium",
    messages: [
      {
        id: 1,
        sender: "Sarah Smith",
        email: "sarah@example.com",
        content:
          "Would love to see a dark mode option in the application. It would make using the app at night much more comfortable for the eyes.",
        timestamp: "5 hours ago",
        isAgent: false,
      },
    ],
  },
}

export function SupportConversation() {
  const location = useLocation()
  const searchParams = new URLSearchParams(location.search)
  const ticketId = searchParams.get("ticket")
  const [replyMessage, setReplyMessage] = React.useState("")

  const ticketIdNum = ticketId ? parseInt(ticketId, 10) : null
  const ticket = ticketIdNum && ticketIdNum in ticketData 
    ? ticketData[ticketIdNum as keyof typeof ticketData] 
    : null

  // Reset reply message when ticket changes
  React.useEffect(() => {
    setReplyMessage("")
  }, [ticketId])

  const handleSendReply = () => {
    if (replyMessage.trim()) {
      // Handle sending reply
      console.log("Sending reply:", replyMessage)
      setReplyMessage("")
    }
  }

  if (!ticket) {
    return (
      <div className="flex flex-1 flex-col">
        <header className="bg-background sticky top-0 flex shrink-0 items-center gap-2 border-b p-4">
          <SidebarTrigger className="-ml-1" />
          <Separator
            orientation="vertical"
            className="mr-2 data-[orientation=vertical]:h-4"
          />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem className="hidden md:block">
                <BreadcrumbLink href="/support">Support</BreadcrumbLink>
              </BreadcrumbItem>
              <BreadcrumbSeparator className="hidden md:block" />
              <BreadcrumbItem>
                <BreadcrumbPage>Inbox</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div className="flex flex-1 flex-col items-center justify-center p-8 text-center">
          <h3 className="text-xl font-semibold mb-2">No ticket selected</h3>
          <p className="text-muted-foreground">
            Select a ticket from the sidebar to view the conversation
          </p>
        </div>
      </div>
    )
  }

  const getLabelColor = (label: string) => {
    switch (label) {
      case "urgent":
        return "#ef4444"
      case "bug":
        return "#f97316"
      case "feature":
        return "#3b82f6"
      case "billing":
        return "#8b5cf6"
      default:
        return "#6b7280"
    }
  }

  return (
    <div className="flex flex-1 flex-col h-[calc(100vh- var(--sidebar-header-height))]">
      {/* Header */}
      <header className="bg-background sticky top-0 flex shrink-0 items-center gap-2 border-b p-4 z-10">
        <SidebarTrigger className="-ml-1" />
        <Separator
          orientation="vertical"
          className="mr-2 data-[orientation=vertical]:h-4"
        />
        <div className="flex flex-1 items-center justify-between">
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem className="hidden md:block">
                <BreadcrumbLink href="/support">Support</BreadcrumbLink>
              </BreadcrumbItem>
              <BreadcrumbSeparator className="hidden md:block" />
              <BreadcrumbItem>
                <BreadcrumbPage>{ticket.subject}</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
          <div className="flex items-center gap-2">
            <Button variant="ghost" size="icon">
              <IconStar className="size-4" />
            </Button>
            <Button variant="ghost" size="icon">
              <IconArchive className="size-4" />
            </Button>
            <Button variant="ghost" size="icon">
              <IconTrash className="size-4" />
            </Button>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="icon">
                  <IconDotsVertical className="size-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem>
                  <IconTag className="size-4 mr-2" />
                  Change Label
                </DropdownMenuItem>
                <DropdownMenuItem>Mark as Spam</DropdownMenuItem>
                <DropdownMenuItem>Print</DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>
      </header>

      {/* Ticket Header */}
      <div className="border-b p-6">
        <div className="flex items-start justify-between mb-3">
          <h2 className="text-2xl font-semibold">{ticket.subject}</h2>
          <div className="flex items-center gap-2">
            <div
              className="size-2 rounded-full"
              style={{ backgroundColor: getLabelColor(ticket.label) }}
            />
            <Badge variant="outline" className="capitalize">
              {ticket.label}
            </Badge>
            <Badge
              variant={
                ticket.priority === "high"
                  ? "destructive"
                  : ticket.priority === "medium"
                    ? "default"
                    : "secondary"
              }
            >
              {ticket.priority}
            </Badge>
          </div>
        </div>
        <div className="flex items-center gap-3 text-sm text-muted-foreground">
          <Avatar className="size-8">
            <AvatarFallback>{ticket.from.charAt(0)}</AvatarFallback>
          </Avatar>
          <div>
            <div className="font-medium text-foreground">{ticket.from}</div>
            <div>{ticket.email}</div>
          </div>
        </div>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-6 space-y-6">
        {ticket.messages.map((message) => (
          <div
            key={message.id}
            className={`flex gap-3 ${message.isAgent ? "flex-row-reverse" : ""}`}
          >
            <Avatar className="size-8">
              <AvatarFallback>
                {message.sender.charAt(0)}
              </AvatarFallback>
            </Avatar>
            <div className={`flex-1 ${message.isAgent ? "items-end" : ""}`}>
              <div className="flex items-baseline gap-2 mb-1">
                <span className="font-medium text-sm">{message.sender}</span>
                <span className="text-xs text-muted-foreground">
                  {message.timestamp}
                </span>
              </div>
              <div
                className={`rounded-lg p-4 ${
                  message.isAgent
                    ? "bg-primary text-primary-foreground"
                    : "bg-muted"
                }`}
              >
                <p className="text-sm whitespace-pre-wrap">{message.content}</p>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Reply Input */}
      <div className="border-t p-4">
        <div className="flex items-end gap-2">
          <div className="flex-1">
            <Input
              placeholder="Type your reply..."
              value={replyMessage}
              onChange={(e) => setReplyMessage(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter" && !e.shiftKey) {
                  e.preventDefault()
                  handleSendReply()
                }
              }}
              className="min-h-[80px]"
            />
          </div>
          <div className="flex gap-2">
            <Button variant="outline" size="icon">
              <IconPaperclip className="size-4" />
            </Button>
            <Button onClick={handleSendReply}>
              <IconSend className="size-4 mr-2" />
              Send
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}

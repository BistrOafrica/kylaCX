import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { useNavigate } from "react-router-dom";
import {
  IconSearch,
  IconMail,
  IconUsers,
  IconTarget,
  IconCurrencyDollar,
  IconChecklist,
  IconSend,
  IconRobot,
  IconSettings,
  IconHome,
} from "@tabler/icons-react";

type QuickAction = {
  id: string;
  title: string;
  description: string;
  icon: React.ElementType;
  action: () => void;
  keywords: string[];
};

export function QuickCreateDialog({
  open,
  onOpenChange,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const navigate = useNavigate();
  const [search, setSearch] = useState("");

  const actions: QuickAction[] = [
    {
      id: "new-ticket",
      title: "New Support Ticket",
      description: "Create a new customer support ticket",
      icon: IconMail,
      action: () => {
        navigate("/support");
        onOpenChange(false);
      },
      keywords: ["support", "ticket", "email", "help"],
    },
    {
      id: "new-contact",
      title: "Add Contact",
      description: "Create a new contact in CRM",
      icon: IconUsers,
      action: () => {
        navigate("/relations");
        onOpenChange(false);
      },
      keywords: ["contact", "crm", "person", "customer"],
    },
    {
      id: "new-lead",
      title: "Add Lead",
      description: "Create a new sales lead",
      icon: IconTarget,
      action: () => {
        navigate("/relations");
        onOpenChange(false);
      },
      keywords: ["lead", "prospect", "sales"],
    },
    {
      id: "new-deal",
      title: "Add Deal",
      description: "Create a new sales deal",
      icon: IconCurrencyDollar,
      action: () => {
        navigate("/relations");
        onOpenChange(false);
      },
      keywords: ["deal", "sales", "opportunity"],
    },
    {
      id: "new-task",
      title: "Add Task",
      description: "Create a new task",
      icon: IconChecklist,
      action: () => {
        navigate("/relations");
        onOpenChange(false);
      },
      keywords: ["task", "todo", "reminder"],
    },
    {
      id: "new-campaign",
      title: "New Campaign",
      description: "Create email, SMS, or WhatsApp campaign",
      icon: IconSend,
      action: () => {
        navigate("/blasts");
        onOpenChange(false);
      },
      keywords: ["campaign", "email", "sms", "whatsapp", "marketing"],
    },
    {
      id: "new-workflow",
      title: "New Workflow",
      description: "Create an automation workflow",
      icon: IconSettings,
      action: () => {
        navigate("/flows");
        onOpenChange(false);
      },
      keywords: ["workflow", "automation", "flow"],
    },
    {
      id: "ask-ai",
      title: "Ask Kyla AI",
      description: "Get help from AI assistant",
      icon: IconRobot,
      action: () => {
        navigate("/kyla-ai");
        onOpenChange(false);
      },
      keywords: ["ai", "assistant", "help", "copilot"],
    },
    {
      id: "dashboard",
      title: "Go to Dashboard",
      description: "Navigate to main dashboard",
      icon: IconHome,
      action: () => {
        navigate("/");
        onOpenChange(false);
      },
      keywords: ["dashboard", "home", "overview"],
    },
  ];

  const filteredActions = actions.filter(
    (action) =>
      action.title.toLowerCase().includes(search.toLowerCase()) ||
      action.description.toLowerCase().includes(search.toLowerCase()) ||
      action.keywords.some((keyword) =>
        keyword.toLowerCase().includes(search.toLowerCase())
      )
  );

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Quick Actions</DialogTitle>
          <DialogDescription>
            Quickly navigate or create items in the system
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div className="relative">
            <IconSearch className="text-muted-foreground absolute left-3 top-1/2 size-4 -translate-y-1/2" />
            <Input
              placeholder="Search actions..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="pl-9"
              autoFocus
            />
          </div>

          <div className="max-h-[400px] space-y-2 overflow-y-auto">
            {filteredActions.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-8">
                <p className="text-muted-foreground text-sm">No actions found</p>
              </div>
            ) : (
              filteredActions.map((action) => (
                <Button
                  key={action.id}
                  variant="ghost"
                  className="h-auto w-full justify-start gap-3 p-4 text-left"
                  onClick={action.action}
                >
                  <action.icon className="text-primary size-5 shrink-0" />
                  <div className="flex-1">
                    <p className="font-medium">{action.title}</p>
                    <p className="text-muted-foreground text-xs">
                      {action.description}
                    </p>
                  </div>
                </Button>
              ))
            )}
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}

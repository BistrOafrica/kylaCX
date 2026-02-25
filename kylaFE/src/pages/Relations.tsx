import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  IconUsers,
  IconChecklist,
  IconTarget,
  IconCurrencyDollar,
  IconFolder,
  IconSearch,
  IconPlus,
  IconMail,
  IconPhone,
} from "@tabler/icons-react";
import PageContainer from "@/components/page-container";

const contacts = [
  {
    id: 1,
    name: "Alice Johnson",
    email: "alice@company.com",
    phone: "+1 234 567 8900",
    company: "Tech Corp",
    status: "active",
  },
  {
    id: 2,
    name: "Bob Smith",
    email: "bob@startup.io",
    phone: "+1 234 567 8901",
    company: "Startup Inc",
    status: "active",
  },
];

const leads = [
  {
    id: 1,
    name: "Enterprise Deal",
    source: "Website",
    value: "$50,000",
    status: "qualified",
  },
  {
    id: 2,
    name: "SMB Prospect",
    source: "Referral",
    value: "$10,000",
    status: "new",
  },
];

const deals = [
  {
    id: 1,
    name: "Q1 Enterprise",
    value: "$100,000",
    stage: "negotiation",
    probability: "75%",
  },
  {
    id: 2,
    name: "Mid-Market",
    value: "$45,000",
    stage: "proposal",
    probability: "50%",
  },
];

const tasks = [
  {
    id: 1,
    title: "Follow up with Alice",
    due: "Today",
    priority: "high",
    completed: false,
  },
  {
    id: 2,
    title: "Send proposal to Bob",
    due: "Tomorrow",
    priority: "medium",
    completed: false,
  },
];

export default function RelationsPage() {
  return (
    <PageContainer>
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Relations (CRM)</h1>
          <p className="text-muted-foreground">
            Manage contacts, leads, deals, and tasks
          </p>
        </div>
      </div>

      <Tabs defaultValue="contacts" className="w-full">
        <TabsList>
          <TabsTrigger value="contacts" className="gap-2">
            <IconUsers className="size-4" />
            Contacts
          </TabsTrigger>
          <TabsTrigger value="leads" className="gap-2">
            <IconTarget className="size-4" />
            Leads
          </TabsTrigger>
          <TabsTrigger value="deals" className="gap-2">
            <IconCurrencyDollar className="size-4" />
            Deals
          </TabsTrigger>
          <TabsTrigger value="tasks" className="gap-2">
            <IconChecklist className="size-4" />
            Tasks
          </TabsTrigger>
          <TabsTrigger value="files" className="gap-2">
            <IconFolder className="size-4" />
            Files
          </TabsTrigger>
        </TabsList>

        <TabsContent value="contacts" className="mt-4 space-y-4">
          <div className="flex items-center gap-2">
            <div className="relative flex-1">
              <IconSearch className="text-muted-foreground absolute left-3 top-1/2 size-4 -translate-y-1/2" />
              <Input placeholder="Search contacts..." className="pl-9" />
            </div>
            <Button className="gap-2">
              <IconPlus className="size-4" />
              Add Contact
            </Button>
          </div>

          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {contacts.map((contact) => (
              <Card key={contact.id}>
                <CardHeader>
                  <CardTitle>{contact.name}</CardTitle>
                  <CardDescription>{contact.company}</CardDescription>
                </CardHeader>
                <CardContent className="space-y-2">
                  <div className="flex items-center gap-2 text-sm">
                    <IconMail className="text-muted-foreground size-4" />
                    <span>{contact.email}</span>
                  </div>
                  <div className="flex items-center gap-2 text-sm">
                    <IconPhone className="text-muted-foreground size-4" />
                    <span>{contact.phone}</span>
                  </div>
                  <Badge variant="secondary">{contact.status}</Badge>
                </CardContent>
              </Card>
            ))}
          </div>
        </TabsContent>

        <TabsContent value="leads" className="mt-4 space-y-4">
          <div className="flex items-center gap-2">
            <div className="relative flex-1">
              <IconSearch className="text-muted-foreground absolute left-3 top-1/2 size-4 -translate-y-1/2" />
              <Input placeholder="Search leads..." className="pl-9" />
            </div>
            <Button className="gap-2">
              <IconPlus className="size-4" />
              Add Lead
            </Button>
          </div>

          <div className="grid gap-4">
            {leads.map((lead) => (
              <Card key={lead.id}>
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <CardTitle>{lead.name}</CardTitle>
                    <Badge
                      variant={
                        lead.status === "qualified" ? "default" : "secondary"
                      }
                    >
                      {lead.status}
                    </Badge>
                  </div>
                  <CardDescription>
                    Source: {lead.source} • Value: {lead.value}
                  </CardDescription>
                </CardHeader>
              </Card>
            ))}
          </div>
        </TabsContent>

        <TabsContent value="deals" className="mt-4 space-y-4">
          <div className="flex items-center gap-2">
            <div className="relative flex-1">
              <IconSearch className="text-muted-foreground absolute left-3 top-1/2 size-4 -translate-y-1/2" />
              <Input placeholder="Search deals..." className="pl-9" />
            </div>
            <Button className="gap-2">
              <IconPlus className="size-4" />
              Add Deal
            </Button>
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            {deals.map((deal) => (
              <Card key={deal.id}>
                <CardHeader>
                  <CardTitle className="flex items-center justify-between">
                    <span>{deal.name}</span>
                    <span className="text-primary">{deal.value}</span>
                  </CardTitle>
                  <CardDescription>
                    Stage: {deal.stage} • Probability: {deal.probability}
                  </CardDescription>
                </CardHeader>
              </Card>
            ))}
          </div>
        </TabsContent>

        <TabsContent value="tasks" className="mt-4 space-y-4">
          <div className="flex items-center gap-2">
            <Button className="gap-2">
              <IconPlus className="size-4" />
              Add Task
            </Button>
          </div>

          <div className="grid gap-2">
            {tasks.map((task) => (
              <Card key={task.id}>
                <CardContent className="flex items-center justify-between py-4">
                  <div className="flex items-center gap-3">
                    <input
                      type="checkbox"
                      checked={task.completed}
                      className="size-4"
                    />
                    <div>
                      <p className="font-medium">{task.title}</p>
                      <p className="text-muted-foreground text-sm">
                        Due: {task.due}
                      </p>
                    </div>
                  </div>
                  <Badge
                    variant={
                      task.priority === "high"
                        ? "destructive"
                        : task.priority === "medium"
                          ? "default"
                          : "secondary"
                    }
                  >
                    {task.priority}
                  </Badge>
                </CardContent>
              </Card>
            ))}
          </div>
        </TabsContent>

        <TabsContent value="files" className="mt-4">
          <Card>
            <CardContent className="flex flex-col items-center justify-center py-12">
              <IconFolder className="text-muted-foreground mb-4 size-12" />
              <p className="text-muted-foreground mb-4">No files yet</p>
              <Button className="gap-2">
                <IconPlus className="size-4" />
                Upload File
              </Button>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </PageContainer>
  );
}

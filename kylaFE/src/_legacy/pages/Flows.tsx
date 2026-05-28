import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  IconPlus,
  IconSearch,
  IconPlayerPlay,
  IconPlayerPause,
  IconSettings,
} from "@tabler/icons-react";
import PageContainer from "@/components/page-container";
import { StatCard, StatCardGrid } from "@/components/stat-card";

const workflows = [
  {
    id: 1,
    name: "Welcome Email Sequence",
    description: "Automated welcome emails for new subscribers",
    status: "active",
    triggers: 450,
    lastRun: "2 hours ago",
  },
  {
    id: 2,
    name: "Abandoned Cart Recovery",
    description: "Send reminders for abandoned shopping carts",
    status: "active",
    triggers: 78,
    lastRun: "30 minutes ago",
  },
  {
    id: 3,
    name: "Lead Scoring",
    description: "Automatically score and route leads",
    status: "paused",
    triggers: 0,
    lastRun: "3 days ago",
  },
];

export default function FlowsPage() {
  return (
    <PageContainer>
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Automation Flows</h1>
          <p className="text-muted-foreground">
            Create and manage automated workflows
          </p>
        </div>
        <Button className="gap-2">
          <IconPlus className="size-4" />
          New Workflow
        </Button>
      </div>

      {/* Stats Cards */}
      <StatCardGrid>
        <StatCard
          title="Active Workflows"
          value="12"
          description="2 paused"
          icon={IconPlayerPlay}
        />
        <StatCard
          title="Total Triggers"
          value="1,234"
          description="This month"
          icon={IconSettings}
        />
        <StatCard
          title="Success Rate"
          value="98.5%"
          description="+2% from last month"
          icon={IconPlayerPlay}
        />
        <StatCard
          title="Failure Rate"
          value="1.5%"
          description="+0.5% from last month"
          icon={IconPlayerPause}
        />
      </StatCardGrid>

      <div className="space-y-4">
        <div className="flex items-center gap-2">
          <div className="relative flex-1">
            <IconSearch className="text-muted-foreground absolute left-3 top-1/2 size-4 -translate-y-1/2" />
            <Input placeholder="Search workflows..." className="pl-9" />
          </div>
        </div>

        <div className="grid gap-4">
          {workflows.map((workflow) => (
            <Card key={workflow.id}>
              <CardHeader>
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <CardTitle className="text-lg">{workflow.name}</CardTitle>
                      <Badge
                        variant={
                          workflow.status === "active" ? "default" : "secondary"
                        }
                      >
                        {workflow.status}
                      </Badge>
                    </div>
                    <CardDescription className="mt-1">
                      {workflow.description}
                    </CardDescription>
                  </div>
                  <div className="flex gap-2">
                    <Button variant="outline" size="sm" className="gap-2">
                      {workflow.status === "active" ? (
                        <>
                          <IconPlayerPause className="size-4" />
                          Pause
                        </>
                      ) : (
                        <>
                          <IconPlayerPlay className="size-4" />
                          Start
                        </>
                      )}
                    </Button>
                    <Button variant="outline" size="sm">
                      Edit
                    </Button>
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-2 gap-4 md:grid-cols-3">
                  <div>
                    <p className="text-muted-foreground text-sm">Triggers</p>
                    <p className="text-lg font-semibold">{workflow.triggers}</p>
                  </div>
                  <div>
                    <p className="text-muted-foreground text-sm">Last Run</p>
                    <p className="text-sm font-medium">{workflow.lastRun}</p>
                  </div>
                  <div>
                    <p className="text-muted-foreground text-sm">Status</p>
                    <p className="text-sm font-medium capitalize">
                      {workflow.status}
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>

        {/* Visual Workflow Builder Placeholder */}
        <Card className="border-dashed">
          <CardContent className="flex flex-col items-center justify-center py-12">
            <IconSettings className="text-muted-foreground mb-4 size-12" />
            <p className="text-muted-foreground mb-4 text-center">
              Visual workflow builder with drag-and-drop widgets
              <br />
              <span className="text-xs">
                Connect triggers, actions, and conditions to create automated
                flows
              </span>
            </p>
            <Button className="gap-2">
              <IconPlus className="size-4" />
              Create Your First Workflow
            </Button>
          </CardContent>
        </Card>
      </div>
    </PageContainer>
  );
}

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
  IconMail,
  IconMessage,
  IconBrandWhatsapp,
  IconPlus,
  IconSearch,
  IconSend,
  IconUsers,
} from "@tabler/icons-react";
import PageContainer from "@/components/page-container";
import { StatCard, StatCardGrid } from "@/components/stat-card";

const campaigns = [
  {
    id: 1,
    name: "Product Launch Email",
    type: "email",
    status: "sent",
    recipients: 5000,
    opens: 1250,
    clicks: 300,
    date: "2 days ago",
  },
  {
    id: 2,
    name: "Flash Sale SMS",
    type: "sms",
    status: "scheduled",
    recipients: 2000,
    date: "Tomorrow at 9:00 AM",
  },
  {
    id: 3,
    name: "Customer Feedback WhatsApp",
    type: "whatsapp",
    status: "draft",
    recipients: 0,
    date: "Not scheduled",
  },
];

export default function BlastsPage() {
  return (
    <PageContainer>
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Marketing Campaigns</h1>
          <p className="text-muted-foreground">
            Create and manage email, SMS, and WhatsApp campaigns
          </p>
        </div>
        <Button className="gap-2">
          <IconPlus className="size-4" />
          New Campaign
        </Button>
      </div>

      {/* Stats Cards */}
      <StatCardGrid>
        <StatCard
          title="Total Sent"
          value="24,589"
          description="+12% from last month"
          icon={IconSend}
        />
        <StatCard
          title="Open Rate"
          value="45.2%"
          description="+3% from last month"
          icon={IconMail}
        />
        <StatCard
          title="Click Rate"
          value="12.8%"
          description="+1.2% from last month"
          icon={IconUsers}
        />
      </StatCardGrid>

      <Tabs defaultValue="all" className="w-full">
        <TabsList>
          <TabsTrigger value="all">All Campaigns</TabsTrigger>
          <TabsTrigger value="email" className="gap-2">
            <IconMail className="size-4" />
            Email
          </TabsTrigger>
          <TabsTrigger value="sms" className="gap-2">
            <IconMessage className="size-4" />
            SMS
          </TabsTrigger>
          <TabsTrigger value="whatsapp" className="gap-2">
            <IconBrandWhatsapp className="size-4" />
            WhatsApp
          </TabsTrigger>
        </TabsList>

        <TabsContent value="all" className="mt-4 space-y-4">
          <div className="flex items-center gap-2">
            <div className="relative flex-1">
              <IconSearch className="text-muted-foreground absolute left-3 top-1/2 size-4 -translate-y-1/2" />
              <Input placeholder="Search campaigns..." className="pl-9" />
            </div>
          </div>

          <div className="grid gap-4">
            {campaigns.map((campaign) => (
              <Card key={campaign.id}>
                <CardHeader>
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-2">
                        <CardTitle className="text-lg">
                          {campaign.name}
                        </CardTitle>
                        <Badge
                          variant={
                            campaign.status === "sent"
                              ? "default"
                              : campaign.status === "scheduled"
                                ? "secondary"
                                : "outline"
                          }
                        >
                          {campaign.status}
                        </Badge>
                      </div>
                      <CardDescription className="mt-1">
                        {campaign.type === "email" && (
                          <span className="flex items-center gap-1">
                            <IconMail className="size-4" /> Email Campaign
                          </span>
                        )}
                        {campaign.type === "sms" && (
                          <span className="flex items-center gap-1">
                            <IconMessage className="size-4" /> SMS Campaign
                          </span>
                        )}
                        {campaign.type === "whatsapp" && (
                          <span className="flex items-center gap-1">
                            <IconBrandWhatsapp className="size-4" /> WhatsApp
                            Campaign
                          </span>
                        )}
                      </CardDescription>
                    </div>
                    <Button variant="outline" size="sm">
                      View Details
                    </Button>
                  </div>
                </CardHeader>
                <CardContent>
                  <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
                    <div>
                      <p className="text-muted-foreground text-sm">
                        Recipients
                      </p>
                      <p className="text-lg font-semibold">
                        {campaign.recipients.toLocaleString()}
                      </p>
                    </div>
                    {campaign.status === "sent" && (
                      <>
                        <div>
                          <p className="text-muted-foreground text-sm">Opens</p>
                          <p className="text-lg font-semibold">
                            {campaign.opens?.toLocaleString()}
                          </p>
                        </div>
                        <div>
                          <p className="text-muted-foreground text-sm">
                            Clicks
                          </p>
                          <p className="text-lg font-semibold">
                            {campaign.clicks?.toLocaleString()}
                          </p>
                        </div>
                      </>
                    )}
                    <div>
                      <p className="text-muted-foreground text-sm">Date</p>
                      <p className="text-sm font-medium">{campaign.date}</p>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </TabsContent>

        <TabsContent value="email" className="mt-4">
          <Card>
            <CardContent className="flex flex-col items-center justify-center py-12">
              <IconMail className="text-muted-foreground mb-4 size-12" />
              <p className="text-muted-foreground mb-4">
                No email campaigns found
              </p>
              <Button className="gap-2">
                <IconPlus className="size-4" />
                Create Email Campaign
              </Button>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="sms" className="mt-4">
          <Card>
            <CardContent className="flex flex-col items-center justify-center py-12">
              <IconMessage className="text-muted-foreground mb-4 size-12" />
              <p className="text-muted-foreground mb-4">
                No SMS campaigns found
              </p>
              <Button className="gap-2">
                <IconPlus className="size-4" />
                Create SMS Campaign
              </Button>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="whatsapp" className="mt-4">
          <Card>
            <CardContent className="flex flex-col items-center justify-center py-12">
              <IconBrandWhatsapp className="text-muted-foreground mb-4 size-12" />
              <p className="text-muted-foreground mb-4">
                No WhatsApp campaigns found
              </p>
              <Button className="gap-2">
                <IconPlus className="size-4" />
                Create WhatsApp Campaign
              </Button>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </PageContainer>
  );
}

import * as React from "react";
import {
  IconHome,
  IconUsers,
  IconSend,
  IconGitBranch,
  IconRobot,
  IconSettings,
  IconInnerShadowTop,
  IconHeadset,
  IconInbox,
  IconStar,
  IconArchive,
  IconTrash,
} from "@tabler/icons-react";

import { NavMain } from "@/components/nav-main";
import { NavSecondary } from "@/components/nav-secondary";
import { NavSupport } from "@/components/nav-support";
import { SidebarCalendar } from "@/components/sidebar-calendar";
import { SettingsDialog } from "@/pages/SettingsDialog";
import { QuickCreateDialog } from "@/components/quick-create-dialog";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarSeparator,
} from "@/components/ui/sidebar";

const data = {
  user: {
    name: "shadcn",
    email: "m@example.com",
    avatar: "/avatars/shadcn.jpg",
  },
  navMain: [
    {
      title: "Home",
      url: "/",
      icon: IconHome,
    },
    {
      title: "Relations",
      url: "/relations",
      icon: IconUsers,
    },
    {
      title: "Blasts",
      url: "/blasts",
      icon: IconSend,
    },
    {
      title: "Flows",
      url: "/flows",
      icon: IconGitBranch,
    },
    {
      title: "Kyla AI",
      url: "/kyla-ai",
      icon: IconRobot,
    },
  ],
  navSupport: [
    {
      title: "Support",
      url: "/support",
      icon: IconHeadset,
      items: [
        {
          title: "Inbox",
          url: "/support?filter=inbox",
          count: 5,
          icon: IconInbox,
        },
        {
          title: "Starred",
          url: "/support?filter=starred",
          count: 2,
          icon: IconStar,
        },
        {
          title: "Sent",
          url: "/support?filter=sent",
          count: 0,
          icon: IconSend,
        },
        {
          title: "Archive",
          url: "/support?filter=archive",
          count: 12,
          icon: IconArchive,
        },
        {
          title: "Trash",
          url: "/support?filter=trash",
          count: 3,
          icon: IconTrash,
        },
      ],
    },
  ],
  navSecondary: [],
};

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const [settingsOpen, setSettingsOpen] = React.useState(false);
  const [quickCreateOpen, setQuickCreateOpen] = React.useState(false);

  const handleSettingsClick = () => {
    setSettingsOpen(true);
  };

  const handleQuickCreateClick = () => {
    setQuickCreateOpen(true);
  };

  return (
    <>
      <Sidebar collapsible="icon" {...props}>
        <SidebarHeader className="h-[calc(var(--header-height))]">
          <SidebarMenu>
            <SidebarMenuItem className="flex flex-row">
              <SidebarMenuButton
                asChild
                className="data-[slot=sidebar-menu-button]:!p-1.5 "
              >
                <a href="/">
                  <IconInnerShadowTop className="!size-5" />
                  <span className="text-base font-semibold">Kyla</span>
                </a>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarHeader>
        <SidebarContent>
          <SidebarSeparator />
          <NavMain
            items={data.navMain}
            onQuickCreate={handleQuickCreateClick}
          />
          <SidebarSeparator className="mt-1" />
          <NavSupport items={data.navSupport} />
        </SidebarContent>
        <SidebarFooter>
          <SidebarSeparator className="mt-1" />
          <SidebarCalendar />
          <SidebarSeparator className="mt-1" />
          <NavSecondary
            items={[
              ...data.navSecondary,
              {
                title: "Settings",
                url: "#",
                icon: IconSettings,
                onClick: handleSettingsClick,
              },
            ]}
            className="mt-auto"
          />
        </SidebarFooter>
      </Sidebar>
      <SettingsDialog open={settingsOpen} onOpenChange={setSettingsOpen} />
      <QuickCreateDialog
        open={quickCreateOpen}
        onOpenChange={setQuickCreateOpen}
      />
    </>
  );
}

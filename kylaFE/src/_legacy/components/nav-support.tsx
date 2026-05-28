"use client";

import { IconTicket, type Icon } from "@tabler/icons-react";
import { Link, useLocation } from "react-router-dom";

import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
} from "@/components/ui/sidebar";
import { Badge } from "@/components/ui/badge";

export function NavSupport({
  items,
}: {
  items: {
    title: string;
    url: string;
    icon?: Icon;
    items?: {
      title: string;
      url: string;
      color?: string;
      count?: number;
      icon?: Icon;
    }[];
  }[];
}) {
  const location = useLocation();

  return (
    <>
      {items.map((item) => (
        <SidebarGroup key={item.title}>
          <SidebarGroupLabel asChild>
            <div className="flex items-center gap-2 group cursor-pointer">
              {item.title}
              <IconTicket className="ml-auto transition-transform duration-200" />
            </div>
          </SidebarGroupLabel>
          <SidebarMenu>
            {!item.items?.length ? (
              <SidebarMenuItem>
                <SidebarMenuButton
                  asChild
                  isActive={location.pathname === item.url}
                >
                  <Link to={item.url}>
                    <span>{item.title}</span>
                  </Link>
                </SidebarMenuButton>
              </SidebarMenuItem>
            ) : null}
            {item.items?.length ? (
              <SidebarMenuSub className="border-l-0 ml-0 px-0">
                {item.items.map((subItem) => {
                  const isActive = location.pathname + location.search === subItem.url;
                  return (
                    <SidebarMenuSubItem key={subItem.title}>
                      <SidebarMenuSubButton asChild isActive={isActive}>
                        <Link to={subItem.url} className="flex items-center gap-2">
                          {subItem.icon && <subItem.icon className="size-4" />}
                          {subItem.color && (
                            <div
                              className="size-2 rounded-full"
                              style={{ backgroundColor: subItem.color }}
                            />
                          )}
                          <span className="flex-1">{subItem.title}</span>
                          {subItem.count !== undefined && subItem.count > 0 && (
                            <Badge variant="secondary" className="h-5 px-1.5 text-xs ml-auto">
                              {subItem.count}
                            </Badge>
                          )}
                        </Link>
                      </SidebarMenuSubButton>
                    </SidebarMenuSubItem>
                  );
                })}
              </SidebarMenuSub>
            ) : null}
          </SidebarMenu>
        </SidebarGroup>
      ))}
    </>
  );
}

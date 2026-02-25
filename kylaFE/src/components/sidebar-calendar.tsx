import { useState } from "react";

import { Calendar } from "@/components/ui/calendar";
import { SidebarGroup, SidebarGroupContent } from "@/components/ui/sidebar";

// Sample data for tasks, events, and support tickets
// const todayItems = [
//   {
//     title: "Fix login bug",
//     type: "task",
//     from: new Date().toISOString(),
//     to: new Date().toISOString(),
//   },
//   {
//     title: "Client onboarding call",
//     type: "event",
//     from: new Date(Date.now() + 2 * 60 * 60 * 1000).toISOString(),
//     to: new Date(Date.now() + 3 * 60 * 60 * 1000).toISOString(),
//   },
//   {
//     title: "Support: Payment issue",
//     type: "ticket",
//     from: new Date().toISOString(),
//     to: new Date().toISOString(),
//   },
// ];

export function SidebarCalendar({
  ...props
}: React.ComponentPropsWithoutRef<typeof SidebarGroup>) {
  const [date, setDate] = useState<Date | undefined>(new Date());

  // Get items for today (limit to 3 as requested)
  // const displayItems = todayItems.slice(0, 3);

  return (
    <SidebarGroup {...props} className="group-data-[state=collapsed]:hidden">
      <SidebarGroupContent>
        <div className="w-full rounded-lg">
          <Calendar
            mode="single"
            selected={date}
            onSelect={setDate}
            className="w-full bg-transparent p-0"
            required
          />
          {/* <div className="mt-3 border-t pt-3">
            <div className="mb-2 flex items-center justify-between px-1">
              <div className="text-xs font-medium">Today's Items</div>
              <Button
                variant="ghost"
                size="icon"
                className="size-5"
                title="Add Item"
              >
                <PlusIcon className="size-3" />
                <span className="sr-only">Add Item</span>
              </Button>
            </div>
            <div className="flex flex-col gap-1.5">
              {displayItems.length > 0 ? (
                displayItems.map((item, index) => (
                  <div
                    key={index}
                    className={`relative rounded-md p-2 pl-5 text-xs after:absolute after:inset-y-2 after:left-2 after:w-1 after:rounded-full ${
                      item.type === "task"
                        ? "bg-blue-50 after:bg-blue-500 dark:bg-blue-950"
                        : item.type === "event"
                        ? "bg-emerald-50 after:bg-emerald-500 dark:bg-emerald-950"
                        : "bg-orange-50 after:bg-orange-500 dark:bg-orange-950"
                    }`}
                  >
                    <div className="font-medium">{item.title}</div>
                    {item.from !== item.to && (
                      <div className="text-muted-foreground text-[10px]">
                        {formatDateRange(new Date(item.from), new Date(item.to))}
                      </div>
                    )}
                  </div>
                ))
              ) : (
                <div className="text-muted-foreground px-2 py-3 text-center text-xs">
                  No items for today
                </div>
              )}
            </div>
          </div> */}
        </div>
      </SidebarGroupContent>
    </SidebarGroup>
  );
}

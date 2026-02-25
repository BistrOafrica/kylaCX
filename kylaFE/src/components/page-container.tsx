import { NavUser } from "./nav-user";
import ThemeToggle from "./theme-toggle";
import { SidebarTrigger } from "./ui/sidebar";

export default function PageContainer({
  children,
}: {
  children: React.ReactNode;
}) {
  const data = {
    user: {
      name: "shadcn",
      email: "m@example.com",
      avatar: "/avatars/shadcn.jpg",
    },
  };
  return (
    <div className="flex flex-col gap-4 p-4 md:p-6">
      <div className="flex gap-4 px-2 md:px-0 justify-end items-center">
        <NavUser user={data.user} />
        <SidebarTrigger className="-ml-1" />
        <ThemeToggle />
      </div>
      {children}
    </div>
  );
}

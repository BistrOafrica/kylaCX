import { SidebarProvider, SidebarInset } from "../components/ui/sidebar";
import { Navigate, Outlet } from "react-router-dom";
import { AppSidebar } from "@/components/app-sidebar";
import { FloatingCallButton } from "@/components/floating-call-button";

// This is a simple auth check - replace with your actual auth logic
const useAuth = () => {
  // For now, let's check if there's a token in localStorage
  // You can replace this with your actual authentication logic
  const token = localStorage.getItem("authToken");
  return { isAuthenticated: !!token };
};

export const Layout = () => {
  const { isAuthenticated } = useAuth();

  if (!isAuthenticated) {
    // Redirect to login if not authenticated
    return <Navigate to="/login" replace />;
  }

  // Render child routes
  return (
    <SidebarProvider
      style={
        {
          "--sidebar-width": "calc(var(--spacing) * 72)",
          "--header-height": "calc(var(--spacing) * 12)",
        } as React.CSSProperties
      }
    >
      <AppSidebar variant="sidebar" />
      <SidebarInset>
        {/* <SiteHeader /> */}
        <div className="flex flex-1 flex-col overflow-hidden">
          <div className="@container/main flex flex-1 flex-col overflow-hidden">
            <Outlet />
          </div>
        </div>
      </SidebarInset>
      <FloatingCallButton />
    </SidebarProvider>
  );
};

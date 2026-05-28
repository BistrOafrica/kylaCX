import type { RouteObject } from "react-router-dom"
import { Navigate } from "react-router-dom"
import { RequireAuth, RequireGuest } from "@/lib/auth"
import { AppShell } from "@/app/shell/AppShell"
import { Login } from "./Login"
import { Home } from "./Home"
import { NotFound } from "./NotFound"
import { InboxRoute } from "@/features/inbox/InboxRoute"
import { ConversationRoute } from "@/features/inbox/ConversationRoute"
import { CrmRoute } from "@/features/crm/routes/CrmRoute"
import {
  ContactsListRoute,
  ContactDetailRoute,
} from "@/features/crm/routes/ContactsRoute"
import {
  CompaniesListRoute,
  CompanyDetailRoute,
} from "@/features/crm/routes/CompaniesRoute"
import {
  LeadsListRoute,
  LeadDetailRoute,
} from "@/features/crm/routes/LeadsRoute"
import {
  DealsBoardRoute,
  DealDetailRoute,
} from "@/features/crm/routes/DealsRoute"
import {
  TicketsListRoute,
  TicketDetailRoute,
} from "@/features/service-desk/routes/TicketsRoute"
import {
  KnowledgeListRoute,
  KnowledgeDetailRoute,
} from "@/features/service-desk/routes/KnowledgeRoute"
import {
  FormsListRoute,
  FormDetailRoute,
} from "@/features/service-desk/routes/FormsRoute"

/**
 * F0 route table.
 *
 * Public routes (login, signup, OTP, reset) live outside the shell and
 * are wrapped in RequireGuest so an authenticated user is bounced home.
 *
 * Everything else renders inside AppShell guarded by RequireAuth.
 * Feature surfaces (inbox, crm, tickets, …) get added as child routes
 * of the protected branch starting in F1.
 */

export const routeConfig: RouteObject[] = [
  {
    path: "/login",
    element: (
      <RequireGuest>
        <Login />
      </RequireGuest>
    ),
  },
  {
    path: "/signup",
    element: (
      <RequireGuest>
        <Navigate to="/login" replace />
      </RequireGuest>
    ),
  },
  {
    path: "/",
    element: (
      <RequireAuth>
        <AppShell />
      </RequireAuth>
    ),
    children: [
      { index: true, element: <Navigate to="/inbox" replace /> },
      {
        path: "inbox",
        element: <InboxRoute />,
        children: [
          { index: true, element: <ConversationRoute /> },
          { path: ":id", element: <ConversationRoute /> },
        ],
      },
      { path: "home", element: <Home /> },
      {
        path: "crm",
        element: <CrmRoute />,
        children: [
          { index: true, element: <Navigate to="/crm/contacts" replace /> },
          { path: "contacts",       element: <ContactsListRoute /> },
          { path: "contacts/:id",   element: <ContactDetailRoute /> },
          { path: "companies",      element: <CompaniesListRoute /> },
          { path: "companies/:id",  element: <CompanyDetailRoute /> },
          { path: "leads",          element: <LeadsListRoute /> },
          { path: "leads/:id",      element: <LeadDetailRoute /> },
          { path: "deals",          element: <DealsBoardRoute /> },
          { path: "deals/:id",      element: <DealDetailRoute /> },
        ],
      },
      { path: "tickets",        element: <TicketsListRoute /> },
      { path: "tickets/:id",    element: <TicketDetailRoute /> },
      { path: "knowledge",      element: <KnowledgeListRoute /> },
      { path: "knowledge/:id",  element: <KnowledgeDetailRoute /> },
      { path: "forms",          element: <FormsListRoute /> },
      { path: "forms/:id",      element: <FormDetailRoute /> },
      // Future feature surfaces:
      //   { path: "automation/*", element: <AutomationRouter /> },
      //   { path: "admin/*",      element: <AdminRouter /> },
    ],
  },
  {
    path: "*",
    element: <NotFound />,
  },
]

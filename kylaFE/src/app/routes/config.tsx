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
import {
  AutomationListRoute,
  WorkflowDetailRoute,
  AIStudioRoute,
} from "@/features/automation/routes/AutomationRoute"
import { AdminRoute } from "@/features/admin/routes/AdminRoute"
import { OrganisationTree } from "@/features/admin/components/OrganisationTree"
import { UsersList } from "@/features/admin/components/UsersList"
import { RolesMatrix } from "@/features/admin/components/RolesMatrix"
import { AppsList } from "@/features/admin/components/AppsList"
import { AgentProfile } from "@/features/admin/components/AgentProfile"
import { AnalyticsRoute } from "@/features/analytics/routes/AnalyticsRoute"
import { OverviewDashboard } from "@/features/analytics/components/OverviewDashboard"
import { CallsDashboard } from "@/features/analytics/components/CallsDashboard"
import { BillingSummary } from "@/features/analytics/components/BillingSummary"
import {
  CallsRoute,
  CallsHistoryRoute,
  CallDetailRoute,
  WallboardRoute,
  IvrListRoute,
  IvrFlowRoute,
  IvrV2FlowRoute,
  QueueWallboardRoute,
  SipAdminRoute,
} from "@/features/telephony/routes/CallsRoute"
import {
  CampaignsListRoute,
  WhatsappCampaignRoute,
  AutodialerCampaignRoute,
  TemplatesRoute,
} from "@/features/campaigns/routes/CampaignsRoute"

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
      { path: "automation",     element: <AutomationListRoute /> },
      { path: "automation/:id", element: <WorkflowDetailRoute /> },
      { path: "ai-studio",      element: <AIStudioRoute /> },
      {
        path: "analytics",
        element: <AnalyticsRoute />,
        children: [
          { index: true, element: <Navigate to="/analytics/overview" replace /> },
          { path: "overview", element: <OverviewDashboard /> },
          { path: "calls",    element: <CallsDashboard /> },
          { path: "billing",  element: <BillingSummary /> },
        ],
      },
      {
        path: "calls",
        element: <CallsRoute />,
        children: [
          { index: true, element: <Navigate to="/calls/history" replace /> },
          { path: "history",       element: <CallsHistoryRoute /> },
          { path: "wallboard",     element: <WallboardRoute /> },
          { path: "ivr",           element: <IvrListRoute /> },
        ],
      },
      { path: "calls/:id",   element: <CallDetailRoute /> },
      { path: "ivr",         element: <IvrListRoute /> },
      { path: "ivr/:id",     element: <IvrFlowRoute /> },
      // New canvas targeting the Phase 5c IVRService.
      { path: "ivr-v2/:id",  element: <IvrV2FlowRoute /> },
      // Phase 5d live wallboard for the new QueueService.
      { path: "queues-live", element: <QueueWallboardRoute /> },
      // Phase 5 SIP infrastructure admin (extensions/trunks/domains).
      { path: "sip-admin",   element: <SipAdminRoute /> },
      { path: "campaigns",                  element: <CampaignsListRoute /> },
      { path: "campaigns/templates",        element: <TemplatesRoute /> },
      { path: "campaigns/whatsapp/:id",     element: <WhatsappCampaignRoute /> },
      { path: "campaigns/voice/:id",        element: <AutodialerCampaignRoute /> },
      {
        path: "admin",
        element: <AdminRoute />,
        children: [
          { index: true, element: <Navigate to="/admin/organisation" replace /> },
          { path: "organisation", element: <OrganisationTree /> },
          { path: "users",        element: <UsersList /> },
          { path: "rbac",         element: <RolesMatrix /> },
          { path: "apps",         element: <AppsList /> },
          { path: "profile",      element: <AgentProfile /> },
        ],
      },
    ],
  },
  {
    path: "*",
    element: <NotFound />,
  },
]

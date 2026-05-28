import type { RouteObject } from "react-router-dom";
import LoginPage from "@/pages/Login";
import SignupPage from "@/pages/SignUp";
import OTPPage from "@/pages/Otp";
import Dashboard from "@/pages/Dashboard";
import SupportPage from "@/pages/Support";
import RelationsPage from "@/pages/Relations";
import BlastsPage from "@/pages/Blasts";
import FlowsPage from "@/pages/Flows";
import KylaAIPage from "@/pages/KylaAI";
import { Layout } from "./ProtectedRoute";

export const routeConfig: RouteObject[] = [
  // Unprotected routes
  {
    path: "/login",
    element: <LoginPage />,
  },
  {
    path: "/signup",
    element: <SignupPage />,
  },
  {
    path: "/otp",
    element: <OTPPage />,
  },
  // Protected routes
  {
    path: "/",
    element: <Layout />,
    children: [
      {
        index: true,
        element: <Dashboard />,
      },
      {
        path: "support",
        element: <SupportPage />,
      },
      {
        path: "relations",
        element: <RelationsPage />,
      },
      {
        path: "blasts",
        element: <BlastsPage />,
      },
      {
        path: "flows",
        element: <FlowsPage />,
      },
      {
        path: "kyla-ai",
        element: <KylaAIPage />,
      },
      {
        path: "settings",
        element: <Dashboard />, // Settings handled via modal
      },
    ],
  },
];

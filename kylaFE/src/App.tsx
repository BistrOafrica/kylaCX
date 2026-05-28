import { createBrowserRouter, RouterProvider } from "react-router-dom"
import { routeConfig } from "./app/routes/config"

const router = createBrowserRouter(routeConfig)

export function App() {
  return <RouterProvider router={router} />
}

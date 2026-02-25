import { createBrowserRouter, RouterProvider } from "react-router-dom";
import { routeConfig } from "./routes/config";
import { Toaster } from "./components/ui/sonner";
import "./App.css";

// Create router with the route configuration
const router = createBrowserRouter(routeConfig);

function App() {
  return (
    <>
      <RouterProvider router={router} />
      <Toaster />
    </>
  );
}

export default App;

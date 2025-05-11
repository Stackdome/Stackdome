import { Route, RouterProvider, createBrowserRouter, createRoutesFromElements } from "react-router-dom"
import Login from "@/pages/login"
import Signup from "@/pages/signup"
import StacksPage from "@/pages/stacks"
import StackCreatePage from "@/pages/stacks/components/create"
import StackDetailPage from "@/pages/stacks/components/detail"
import StackActivityPage from "@/pages/stacks/components/activity"
import StackSettingsPage from "@/pages/stacks/components/settings"
import { StackProvider } from "./pages/stacks/contexts/stack-context"
import { logoutAndRedirect } from "@/helpers/common"

const Logout = () => {
  logoutAndRedirect("/login");
  return null;
}

// Create router with routes
const router = createBrowserRouter(
  createRoutesFromElements(
    <>
      <Route path="/" element={<StacksPage />} />
      <Route path="/login" element={<Login />} />
      <Route path="/sign-up" element={<Signup />} />
      <Route path="/dashboard" element={<StacksPage />} />
      <Route path="/stacks" element={<StacksPage />} />
      <Route path="/stacks/create" element={<StackCreatePage />} />
      <Route path="/stacks/:id" element={<StackDetailPage />} />
      <Route path="/stacks/:id/activity" element={<StackActivityPage />} />
      <Route path="/stacks/:id/settings" element={<StackSettingsPage />} />
      <Route path="/logout" element={<Logout />} />
    </>
  )
)

function App() {
  return (
    <StackProvider>
      <RouterProvider router={router} />
    </StackProvider>
  )
}

export default App

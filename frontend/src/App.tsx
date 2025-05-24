import { Route, RouterProvider, createBrowserRouter, createRoutesFromElements } from "react-router-dom"
import Login from "@/pages/login"
import Signup from "@/pages/signup"
import StacksPage from "@/pages/stacks/components/list"
import StackCreatePage from "@/pages/stacks/components/create"
import StackDetailPage from "@/pages/stacks/components/detail"
import ClustersPage from "@/pages/clusters"
import ClusterCreatePage from "@/pages/clusters/components/create"
import ClusterDetailPage from "@/pages/clusters/components/detail"
import DomainsPage from "@/pages/domains"
import { StackProvider } from "@/pages/stacks/contexts/stack-context"
import { logoutAndRedirect } from "@/helpers/common"
import { AppLayout } from "@/components/app-layout"
import { Toaster } from "@/components/ui/toaster"

const Logout = () => {
  logoutAndRedirect("/login");
  return null;
}

// Create router with routes
const router = createBrowserRouter(
  createRoutesFromElements(
    <>
      <Route element={<AppLayout />}>
        <Route path="/" element={<StacksPage />} />
        <Route path="/dashboard" element={<StacksPage />} />
        <Route path="/stacks" element={<StacksPage />} />
        <Route path="/stacks/create" element={<StackCreatePage />} />
        <Route path="/stacks/:id" element={<StackDetailPage />} />
        <Route path="/clusters" element={<ClustersPage />} />
        <Route path="/clusters/create" element={<ClusterCreatePage />} />
        <Route path="/clusters/:id" element={<ClusterDetailPage />} />
        <Route path="/domains" element={<DomainsPage />} />
      </Route>
      <Route path="/login" element={<Login />} />
      <Route path="/sign-up" element={<Signup />} />
      <Route path="/logout" element={<Logout />} />
    </>
  )
)

function App() {
  return (
    <StackProvider>
      <RouterProvider router={router} />
      <Toaster />
    </StackProvider>
  )
}

export default App

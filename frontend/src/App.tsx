import { Route, RouterProvider, createBrowserRouter, createRoutesFromElements } from "react-router-dom"
import Login from "@/pages/login"
import Signup from "@/pages/signup"
import StacksPage from "@/pages/stacks/components/list"
import StackCreatePage from "@/pages/stacks/components/create"
import StackDetailPage from "@/pages/stacks/components/detail"
import ClustersPage from "@/pages/clusters"
import ClusterDetailPage from "@/pages/clusters/components/detail"
import SecretsPage from "@/pages/secrets"
import DomainsPage from "@/pages/domains"
import AddonsPage from "@/pages/addons"
import PostgresFormPage from "@/pages/addons/components/postgres-create-page"
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
        <Route path="/clusters/:id" element={<ClusterDetailPage />} />
        <Route path="/secrets" element={<SecretsPage />} />
        <Route path="/addons" element={<AddonsPage />} />
        <Route path="/addons/create/postgres" element={<PostgresFormPage />} />
        <Route path="/addons/postgres/:id/edit" element={<PostgresFormPage />} />
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

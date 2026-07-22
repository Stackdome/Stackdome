import { Navigate, Route, RouterProvider, createBrowserRouter, createRoutesFromElements } from "react-router-dom"
import Login from "@/pages/login"
import Signup from "@/pages/signup"
import GithubCallbackPage from "@/pages/auth/github-callback"
import StacksPage from "@/pages/stacks/components/list"
import CanvasEditorPage from "@/pages/stacks/components/editor"
import ClustersPage from "@/pages/clusters"
import ClusterDetailPage from "@/pages/clusters/components/detail"
import SecretsPage from "@/pages/secrets"
import DomainsPage from "@/pages/domains"
import AddonsPage from "@/pages/addons"
import PostgresFormPage from "@/pages/addons/components/postgres-create-page"
import PostgresDetailPage from "@/pages/addons/postgres-detail-page"
import ObjectStoresPage from "@/pages/object-stores"
import PreviewsPage from "@/pages/previews"
import PreviewConfigDetailPage from "@/pages/previews/config-detail"
import GitIntegrationsPage from "@/pages/git-integrations"
import ImageRegistriesPage from "@/pages/image-registries"
import NotFoundPage from "@/pages/not-found"
import { StackProvider } from "@/pages/stacks/contexts/stack-context"
import { isUserLoggedIn, logoutAndRedirect } from "@/helpers/common"
import { AppLayout } from "@/components/app-layout"
import { Toaster } from "@/components/ui/toaster"
import { ConfirmProvider } from "@/components/branded/confirm"
import { ThemeProvider } from "@/contexts/theme-provider"
import { CurrentUserProvider } from "@/contexts/current-user-context"
import { RequireAdmin } from "@/components/require-admin"
// Users + Projects (workspace collaboration) are shelved — pages remain in the
// repo (src/pages/users, src/pages/projects) but are unrouted; /settings/* below
// redirects home so a typed URL doesn't 404.

const Logout = () => {
  logoutAndRedirect("/sign-in");
  return null;
}

const RequireAuth = ({ children }: { children: React.ReactNode }) => {
  if (!isUserLoggedIn()) {
    return <Navigate to="/sign-in" replace />;
  }
  return <>{children}</>;
}

// Create router with routes
const router = createBrowserRouter(
  createRoutesFromElements(
    <>
      <Route element={<RequireAuth><AppLayout /></RequireAuth>}>
        <Route path="/" element={<StacksPage />} />
        <Route path="/dashboard" element={<StacksPage />} />
        <Route path="/stacks" element={<StacksPage />} />
        <Route path="/stacks/new" element={<CanvasEditorPage />} />
        <Route path="/stacks/create" element={<Navigate to="/stacks/new" replace />} />
        <Route path="/stacks/:id" element={<CanvasEditorPage />} />
        <Route path="/secrets" element={<SecretsPage />} />
        <Route path="/object-stores" element={<ObjectStoresPage />} />
        <Route path="/addons" element={<AddonsPage />} />
        <Route path="/addons/create/postgres" element={<PostgresFormPage />} />
        <Route path="/addons/postgres/:id/edit" element={<PostgresFormPage />} />
        <Route path="/addons/postgres/:id" element={<PostgresDetailPage />} />
        {/* Org-scoped, admin-only pages — members are redirected to "/" */}
        <Route element={<RequireAdmin />}>
          <Route path="/clusters" element={<ClustersPage />} />
          <Route path="/clusters/:id" element={<ClusterDetailPage />} />
          <Route path="/domains" element={<DomainsPage />} />
        </Route>
        <Route path="/previews" element={<PreviewsPage />} />
        <Route path="/previews/:configId" element={<PreviewConfigDetailPage />} />
        <Route path="/git-integrations" element={<GitIntegrationsPage />} />
        <Route path="/image-registries" element={<ImageRegistriesPage />} />
        {/* Workspace collaboration (Users + Projects) shelved — redirect home. */}
        <Route path="/settings/*" element={<Navigate to="/" replace />} />
      </Route>
      <Route path="/sign-in" element={<Login />} />
      <Route path="/sign-up" element={<Signup />} />
      <Route path="/auth/github/callback" element={<GithubCallbackPage />} />
      <Route path="/logout" element={<Logout />} />
      <Route path="*" element={<NotFoundPage />} />
    </>
  )
)

// Safety net for the Radix body pointer-events wedge: a modal layer unmounted
// mid-close (e.g. by navigation) can "restore" a stale pointer-events:none to
// <body> and dead-lock the page (radix-ui/primitives#1836 class). After any
// navigation, a body lock with no open body-locking layer (dialog/menu — not
// tooltips/accordions, which never lock) is stale by definition — clear it.
// Checked twice: once after the unmounts flush, once after close animations.
const OPEN_LOCKING_LAYER =
  "[data-state='open']:is([role='dialog'],[role='alertdialog'],[role='menu'])";
router.subscribe(() => {
  for (const delay of [0, 300]) {
    setTimeout(() => {
      if (!document.querySelector(OPEN_LOCKING_LAYER)) {
        document.body.style.pointerEvents = "";
      }
    }, delay);
  }
})

function App() {
  return (
    <ThemeProvider defaultTheme="system" storageKey="stackdome-ui-theme">
      <StackProvider>
        <CurrentUserProvider>
          <ConfirmProvider>
            <RouterProvider router={router} />
          </ConfirmProvider>
        </CurrentUserProvider>
        <Toaster />
      </StackProvider>
    </ThemeProvider>
  )
}

export default App

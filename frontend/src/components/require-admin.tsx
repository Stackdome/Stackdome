import { Navigate, Outlet } from "react-router-dom";
import { useCurrentUser } from "@/hooks/use-current-user";

export function RequireAdmin() {
  const { isOrgAdmin, loading } = useCurrentUser();
  if (loading) return null;
  if (!isOrgAdmin) return <Navigate to="/" replace />;
  return <Outlet />;
}

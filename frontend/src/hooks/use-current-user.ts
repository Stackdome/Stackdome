import { useContext } from "react";
import { CurrentUserContext } from "@/contexts/current-user-context";

export function useCurrentUser() {
  const ctx = useContext(CurrentUserContext);
  if (!ctx) throw new Error("useCurrentUser must be used within CurrentUserProvider");
  return ctx;
}

import { useCallback } from "react";
import { getCurrentOrganizationId } from "@/helpers/common";
import { getErrorMessage } from "@/api/client";
import { promoteAdmin, demoteAdmin } from "@/api/organizations";

type ActionResult = { ok: true } | { ok: false; error: string };

export function useUserActions() {
  const orgId = getCurrentOrganizationId();

  const promote = useCallback(
    async (userId: string): Promise<ActionResult> => {
      if (!orgId) return { ok: false, error: "no organisation" };
      try {
        await promoteAdmin(orgId, { user_id: userId });
        return { ok: true };
      } catch (e) {
        return { ok: false, error: getErrorMessage(e) };
      }
    },
    [orgId]
  );

  const demote = useCallback(
    async (
      userId: string,
      teamName: string,
      role: "Developer" | "Viewer"
    ): Promise<ActionResult> => {
      if (!orgId) return { ok: false, error: "no organisation" };
      try {
        await demoteAdmin(orgId, userId, { team_name: teamName, role });
        return { ok: true };
      } catch (e) {
        return { ok: false, error: getErrorMessage(e) };
      }
    },
    [orgId]
  );

  return { promote, demote };
}

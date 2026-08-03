import { useCallback, useState } from "react";
import { getCurrentOrganizationId } from "@/lib/common";
import { getErrorMessage } from "@/api/client";
import { createInvite, resendInvite, revokeInvite, type OrgInviteCreateResponse } from "@/api/invites";
import { toInviteCreateRequest, type InviteFormData } from "../schemas/invite-schema";

type CreateResult = "idle" | "sent" | "failed";

export function useInvites() {
  const orgId = getCurrentOrganizationId();
  const [submitting, setSubmitting] = useState(false);
  const [serverError, setServerError] = useState<string | null>(null);
  const [result, setResult] = useState<CreateResult>("idle");

  const create = useCallback(async (data: InviteFormData): Promise<{ invite: OrgInviteCreateResponse; token?: string }> => {
    if (!orgId) throw new Error("no organisation");
    setSubmitting(true);
    setServerError(null);
    try {
      const invite = await createInvite(orgId, toInviteCreateRequest(data));
      setResult(invite.email_sent ? "sent" : "failed");
      return { invite, token: invite.invite_token };
    } catch (e: unknown) {
      setServerError(getErrorMessage(e));
      throw e;
    } finally {
      setSubmitting(false);
    }
  }, [orgId]);

  const resend = useCallback(async (inviteId: string) => {
    if (!orgId) return;
    await resendInvite(orgId, inviteId);
  }, [orgId]);

  const revoke = useCallback(async (inviteId: string) => {
    if (!orgId) return;
    await revokeInvite(orgId, inviteId);
  }, [orgId]);

  const reset = useCallback(() => { setResult("idle"); setServerError(null); }, []);

  return { create, resend, revoke, reset, submitting, serverError, result };
}

import { useEffect, useState } from "react";
import { getPublicInviteInfo, type OrgInviteInfo } from "@/api/invites";
import { getErrorStatus, getErrorMessage } from "@/api/client";

export type InviteInfoState =
  | "loading" | "new-user" | "not-found" | "expired" | "revoked" | "already-used";

export function useInviteInfo(token: string | null) {
  const [state, setState] = useState<InviteInfoState>("loading");
  const [info, setInfo] = useState<OrgInviteInfo | null>(null);

  useEffect(() => {
    let cancelled = false;
    if (!token) { setState("not-found"); return; }
    (async () => {
      try {
        const data = await getPublicInviteInfo(token);
        if (cancelled) return;
        setInfo(data);
        setState("new-user");
      } catch (e) {
        if (cancelled) return;
        const status = getErrorStatus(e);
        if (status === 404) { setState("not-found"); return; }
        if (status === 410) {
          const msg = getErrorMessage(e).toLowerCase();
          if (msg.includes("revoked")) setState("revoked");
          else if (msg.includes("already")) setState("already-used");
          else setState("expired");
          return;
        }
        setState("not-found");
      }
    })();
    return () => { cancelled = true; };
  }, [token]);

  return { state, info };
}

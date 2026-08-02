import { useEffect, useState } from "react";
import { getOrganization } from "@/api/organizations";

/** FQDNs owned by the organisation — used to classify public ingress URLs. */
export function useOrgDomains(orgId: string | undefined): string[] {
  const [domains, setDomains] = useState<string[]>([]);
  useEffect(() => {
    if (!orgId) return;
    let cancelled = false;
    getOrganization(orgId)
      .then((org) => {
        if (cancelled) return;
        setDomains((org.domains ?? []).map((d) => d.fqdn).filter((f): f is string => !!f));
      })
      .catch(() => {
        // Classification degrades gracefully: with no org domains every URL
        // reads as "custom", which still renders correctly in the pill.
      });
    return () => {
      cancelled = true;
    };
  }, [orgId]);
  return domains;
}

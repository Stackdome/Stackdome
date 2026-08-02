import { useEffect, useState } from "react";
import { Input } from "@/components/ui/input";
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from "@/components/ui/select";
import { listRepositoryBranches } from "@/api/git-integrations";
import { getCurrentOrganizationId } from "@/lib/common";

interface BranchFieldProps {
  id: string;
  value: string;
  onChange: (branch: string) => void;
  /** null → no listing; free text only */
  integrationId: string | null;
  /** "owner/name"; listing needs both parts */
  repoFullName?: string;
  placeholder?: string;
}

/** Branch picker: a Select of listed branches when the integration can list
    them, otherwise a free-text Input. Falls back to free text on fetch
    failure too — some hosts/tokens can't list branches. */
export function BranchField({
  id, value, onChange, integrationId, repoFullName, placeholder = "main",
}: BranchFieldProps) {
  const [branches, setBranches] = useState<string[]>([]);

  useEffect(() => {
    setBranches([]);
    const orgId = getCurrentOrganizationId();
    const [owner, repo] = (repoFullName ?? "").split("/");
    if (!orgId || !integrationId || !owner || !repo) return;
    let cancelled = false;
    listRepositoryBranches(orgId, integrationId, owner, repo)
      .then((res) => {
        if (!cancelled) setBranches(res.items ?? []);
      })
      .catch(() => {
        // fall back to free-text branch input
      });
    return () => { cancelled = true; };
  }, [integrationId, repoFullName]);

  if (branches.length > 0) {
    return (
      <Select value={value} onValueChange={onChange}>
        <SelectTrigger id={id}>
          <SelectValue placeholder="Select branch" />
        </SelectTrigger>
        <SelectContent>
          {branches.map((b) => (
            <SelectItem key={b} value={b}>{b}</SelectItem>
          ))}
        </SelectContent>
      </Select>
    );
  }

  return (
    <Input
      id={id}
      value={value}
      placeholder={placeholder}
      onChange={(e) => onChange(e.target.value)}
    />
  );
}

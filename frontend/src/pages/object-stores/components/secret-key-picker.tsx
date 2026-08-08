import { useEffect, useMemo, useState, type ReactNode } from "react";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { getCurrentOrganizationId } from "@/lib/common";
import { getSecrets, type Secret } from "@/api/secrets";
import { getErrorMessage } from "@/api/client";

type Props = {
  label: ReactNode;
  helpText?: string;
  value: { secret_id: string; key: string };
  onChange: (next: { secret_id: string; key: string }) => void;
  /**
   * Optional name of the key we typically expect inside the chosen secret.
   * If the selected secret doesn't contain a matching key, the picker surfaces
   * an advisory hint. It does NOT filter the dropdown options.
   */
  expectedKeyHint?: string;
  error?: string;
};

export function SecretKeyPicker({
  label,
  helpText,
  value,
  onChange,
  expectedKeyHint,
  error,
}: Props) {
  const [secrets, setSecrets] = useState<Array<Secret & { id: string }>>([]);
  const [loading, setLoading] = useState(false);
  const [fetchError, setFetchError] = useState<string | null>(null);

  useEffect(() => {
    const orgId = getCurrentOrganizationId();
    if (!orgId) return;
    let cancelled = false;
    setLoading(true);
    getSecrets(orgId)
      .then((res) => {
        if (cancelled) return;
        setSecrets(
          (res.items ?? []).filter(
            (s): s is Secret & { id: string } => !!s.id && s.type === "Generic",
          ),
        );
      })
      .catch((e) => {
        if (cancelled) return;
        setFetchError(getErrorMessage(e));
      })
      .finally(() => {
        if (cancelled) return;
        setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const selectedSecret = useMemo(
    () => secrets.find((s) => s.id === value.secret_id),
    [secrets, value.secret_id],
  );

  const availableKeys = useMemo(
    () => (selectedSecret?.data ?? []).map((d) => d.key),
    [selectedSecret],
  );
  const keyMissing =
    !!value.key && availableKeys.length > 0 && !availableKeys.includes(value.key);

  return (
    <div className="flex flex-col gap-2">
      <Label>{label}</Label>
      {helpText && <p className="text-meta text-muted-foreground">{helpText}</p>}

      <div className="grid grid-cols-2 gap-2">
        <Select
          value={value.secret_id}
          onValueChange={(v) => onChange({ secret_id: v, key: "" })}
          disabled={loading}
        >
          <SelectTrigger>
            <SelectValue placeholder={loading ? "Loading secrets…" : "Select a Generic secret"} />
          </SelectTrigger>
          <SelectContent>
            {secrets.map((s) => (
              <SelectItem key={s.id} value={s.id}>
                {s.name}
              </SelectItem>
            ))}
            {!loading && secrets.length === 0 && (
              <div className="px-3 py-2 text-meta text-muted-foreground">
                No Generic secrets — create one on the Secrets page.
              </div>
            )}
          </SelectContent>
        </Select>

        <Select
          value={value.key}
          onValueChange={(v) => onChange({ secret_id: value.secret_id, key: v })}
          disabled={!selectedSecret}
        >
          <SelectTrigger>
            <SelectValue placeholder="Select a key" />
          </SelectTrigger>
          <SelectContent>
            {availableKeys.map((k) => (
              <SelectItem key={k} value={k}>
                {k}
              </SelectItem>
            ))}
            {selectedSecret && availableKeys.length === 0 && (
              <div className="px-3 py-2 text-meta text-muted-foreground">
                Selected secret has no keys.
              </div>
            )}
          </SelectContent>
        </Select>
      </div>

      {fetchError ? (
        <p className="text-meta text-danger">{fetchError}</p>
      ) : error ? (
        <p className="text-meta text-danger">{error}</p>
      ) : keyMissing ? (
        <p className="text-meta text-danger">
          The selected key is no longer present in the chosen secret.
        </p>
      ) : expectedKeyHint && selectedSecret && !availableKeys.includes(expectedKeyHint) ? (
        <p className="text-meta text-warn">
          Tip: typically named <code className="font-mono">{expectedKeyHint}</code>. The selected secret doesn't have that key.
        </p>
      ) : null}
    </div>
  );
}

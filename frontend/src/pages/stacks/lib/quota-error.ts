// Compute-quota limits count across the whole organisation, not per stack.
// The backend sends only the limit, no usage numbers, so the server reason is
// passed through verbatim.

// pkg/errors ErrorComputeQuotaExceeded, serialized as a string on the wire.
export const COMPUTE_QUOTA_EXCEEDED_CODE = "30";

export type QuotaMessage = {
  title: string;
  description: string;
};

const GENERIC_TITLE = "Plan limit reached";
const ORG_WIDE_NOTE = "That's counted across your whole organisation, not per stack.";

// Matched against the server-authored reason, most specific first: "volume
// size" must beat "volumes", and "stack resources" must beat "stacks".
const SCOPES: { match: string; title: string; remedy: string; orgWide: boolean }[] = [
  {
    match: "stack resources",
    title: "Stack resource limit reached",
    remedy: "Remove a resource here, or delete one from another stack.",
    orgWide: true,
  },
  {
    match: "stacks",
    title: "Stack limit reached",
    remedy: "Delete a stack you no longer need.",
    orgWide: true,
  },
  {
    match: "volume size",
    title: "Volume size limit reached",
    remedy: "Reduce the volume's size.",
    orgWide: false,
  },
  {
    match: "volumes",
    title: "Volume limit reached",
    remedy: "Remove a volume you no longer need.",
    orgWide: true,
  },
];

export function quotaMessage(reason: string): QuotaMessage {
  const text = reason.trim();
  const scope = SCOPES.find((s) => text.toLowerCase().includes(s.match));
  if (!scope) {
    return { title: GENERIC_TITLE, description: text };
  }
  const sentence = text.endsWith(".") ? text : `${text}.`;
  const framing = scope.orgWide ? `${ORG_WIDE_NOTE} ` : "";
  return {
    title: scope.title,
    description: `${sentence} ${framing}${scope.remedy}`,
  };
}

export interface ParsedOutput {
  group: "internal" | "public";
  label: string;
  port?: string;
  key: string;
}

// Longest bases first so "public_url" wins over "url" and "public_host" over "host".
const BASES: { base: string; group: "internal" | "public"; label: string }[] = [
  { base: "public_host", group: "public", label: "Public Host" },
  { base: "public_url", group: "public", label: "Public URL" },
  { base: "port", group: "internal", label: "Port" },
  { base: "url", group: "internal", label: "URL" },
  { base: "host", group: "internal", label: "Host" },
];

export function parseOutputKey(key: string): ParsedOutput {
  for (const { base, group, label } of BASES) {
    if (key === base) return { group, label, key };
    if (key.startsWith(base + ".")) {
      return { group, label, port: key.slice(base.length + 1), key };
    }
  }
  return { group: "internal", label: key, key };
}

export function groupOutputs(outputs: string[]): {
  internal: ParsedOutput[];
  public: ParsedOutput[];
} {
  const internal: ParsedOutput[] = [];
  const pub: ParsedOutput[] = [];
  for (const o of outputs) {
    const parsed = parseOutputKey(o);
    (parsed.group === "public" ? pub : internal).push(parsed);
  }
  return { internal, public: pub };
}

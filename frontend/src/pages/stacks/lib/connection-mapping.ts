const SIMPLE_KEY = /^[A-Za-z0-9_]+$/;

// Mirror of pkg/models/output_descriptor.go secretOutputAccessor: simple keys
// get a dot accessor; anything else is bracket-quoted with ' and \ escaped.
export function secretOutputAccessor(key: string): string {
  if (SIMPLE_KEY.test(key)) return `key.${key}`;
  const escaped = key.replace(/\\/g, "\\\\").replace(/'/g, "\\'");
  return `key['${escaped}']`;
}

// Reverse secretOutputAccessor. Returns the key, or null if the accessor is not
// a secret-key accessor.
export function parseSecretOutput(output: string): string | null {
  if (output.startsWith("key.") && output.length > 4) return output.slice(4);
  const m = output.match(/^key\['(.*)'\]$/s);
  if (!m) return null;
  return m[1].replace(/\\(['\\])/g, "$1");
}

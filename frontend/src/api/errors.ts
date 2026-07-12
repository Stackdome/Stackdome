// Parses the backend Error envelope into a structured shape the UI can consume:
// a top-level message plus per-field validation errors (details.errors[]) and
// the credential-required detail variant. Falls back to the flat reason string
// when no structured details are present.
import { getErrorMessage, getErrorStatus, isAxiosError } from "./client";

export type ParsedFieldError = {
  field: string;
  code: string;
  message: string;
};

export type CredentialErrorTarget = {
  kind: string;
  host: string;
  ref: string;
};

export type CredentialErrorInfo = {
  code: string;
  target: CredentialErrorTarget;
};

export type ParsedApiError = {
  status: number | undefined;
  topLevel: string;
  fieldErrors: ParsedFieldError[];
  credential?: CredentialErrorInfo;
};

const CREDENTIAL_CODES = new Set(["credentials_required", "credentials_invalid"]);

function asRecord(value: unknown): Record<string, unknown> | undefined {
  return value && typeof value === "object" ? (value as Record<string, unknown>) : undefined;
}

// The wire body is either an Error or an ErrorList; the structured details live
// on the primary Error (the list's first item when it's a list).
function primaryError(data: unknown): Record<string, unknown> | undefined {
  const record = asRecord(data);
  if (!record) return undefined;
  if (Array.isArray(record.items)) {
    return asRecord(record.items[0]);
  }
  return record;
}

function extractFieldErrors(details: Record<string, unknown> | undefined): ParsedFieldError[] {
  if (!details || !Array.isArray(details.errors)) return [];
  return details.errors.flatMap((entry): ParsedFieldError[] => {
    const fe = asRecord(entry);
    if (!fe || typeof fe.field !== "string" || typeof fe.message !== "string") return [];
    return [{ field: fe.field, code: typeof fe.code === "string" ? fe.code : "", message: fe.message }];
  });
}

function extractCredential(details: Record<string, unknown> | undefined): CredentialErrorInfo | undefined {
  if (!details || typeof details.code !== "string" || !CREDENTIAL_CODES.has(details.code)) return undefined;
  const target = asRecord(details.target);
  if (!target) return undefined;
  return {
    code: details.code,
    target: {
      kind: String(target.kind ?? ""),
      host: String(target.host ?? ""),
      ref: String(target.ref ?? ""),
    },
  };
}

export function parseApiError(error: unknown): ParsedApiError {
  const primary = isAxiosError(error) ? primaryError(error.response?.data) : undefined;
  const details = asRecord(primary?.details);
  // Derive topLevel from the primary error's reason so ErrorList bodies (whose
  // reason lives on items[0]) resolve correctly; getErrorMessage matches any
  // object body as a single Error first and would miss it.
  const reason = typeof primary?.reason === "string" ? primary.reason : undefined;
  return {
    status: getErrorStatus(error),
    topLevel: reason || getErrorMessage(error),
    fieldErrors: extractFieldErrors(details),
    credential: extractCredential(details),
  };
}

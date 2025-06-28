import type { Secret, SecretType, SecretData } from "@/api/secrets";

export type { Secret, SecretType, SecretData };

export interface SecretFormData {
  name: string;
  description?: string;
  type: SecretType;
  data: SecretData[];
}

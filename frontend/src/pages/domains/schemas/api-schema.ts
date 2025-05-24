import { z } from 'zod';
import type { components } from '@/api/types/openapi';

export type DomainName = components["schemas"]["DomainName"];

/**
 * Domain validation regex pattern from backend
 * This matches the pattern in pkg/services/stack_domain_service.go
 *
 * Pattern breakdown:
 * - ^(?i:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,}$
 * - Case insensitive validation for fully qualified domain names
 * - Must start and end with alphanumeric characters
 * - Can contain hyphens in the middle (up to 61 chars per label)
 * - Must have at least one dot
 * - Must end with at least 2 letter TLD
 */
const DOMAIN_REGEX = /^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,}$/i;

/**
 * Zod schema for domain validation that matches the OpenAPI DomainName schema
 */
export const domainSchema = z.object({
  id: z.string().optional(),
  fqdn: z
    .string()
    .min(1, 'Domain name is required')
    .max(253, 'Domain name cannot exceed 253 characters')
    .regex(
      DOMAIN_REGEX,
      'Please enter a valid domain name'
    )
    .transform((val) => val.toLowerCase().trim())
    .optional(),
}) satisfies z.ZodType<DomainName>;

type DomainFormData = z.infer<typeof domainSchema>;

export function createDomainFromForm(formData: DomainFormData): DomainName {
  return {
    fqdn: formData.fqdn,
  };
}

export function validateDomainName(domain: string): { isValid: boolean; error?: string } {
  try {
    domainSchema.parse({ fqdn: domain });
    return { isValid: true };
  } catch (error) {
    if (error instanceof z.ZodError) {
      return {
        isValid: false,
        error: error.errors[0]?.message || 'Invalid domain name'
      };
    }
    return { isValid: false, error: 'Invalid domain name' };
  }
}

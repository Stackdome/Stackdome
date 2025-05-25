import { z } from 'zod';

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
 * Zod schema for domain validation
 */
export const domainSchema = z.object({
  fqdn: z
    .string()
    .min(1, 'Domain name is required')
    .max(253, 'Domain name cannot exceed 253 characters')
    .regex(
      DOMAIN_REGEX,
      'Please enter a valid domain name (e.g., example.com, subdomain.example.org)'
    )
    .transform((val) => val.toLowerCase().trim()),
});

/**
 * Form data type for adding domains
 */
export type DomainFormData = z.infer<typeof domainSchema>;


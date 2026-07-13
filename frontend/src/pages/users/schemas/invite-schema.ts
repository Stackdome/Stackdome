import { z } from "zod";
import type { OrgInviteCreateRequest } from "@/api/invites";

export const inviteSchema = z.object({
  email: z.string().email("Enter a valid email"),
  project_name: z.string().min(1, "Pick a project"),
  role: z.enum(["Developer", "Viewer"]),
});

export type InviteFormData = z.infer<typeof inviteSchema>;

export function toInviteCreateRequest(data: InviteFormData): OrgInviteCreateRequest {
  return {
    email: data.email,
    project_name: data.project_name,
    role: data.role,
    expires_in_days: 1,
  };
}

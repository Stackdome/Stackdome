import { z } from 'zod';

export const loginSchema = z.object({
  email: z.string().email('Email is invalid'),
  password: z.string().min(1, 'Password is required'),
});

export type LoginFormData = z.infer<typeof loginSchema>;

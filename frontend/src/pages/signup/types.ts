import { z } from 'zod';

export const signupSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  email: z.string().email('Email is invalid'),
  password: z.string().min(8, 'Password must be at least 8 characters'),
  organisationName: z.string().min(1, 'Organization Name is required'),
  // confirmPassword is for UI only, not in OpenAPI, but needed for validation
  confirmPassword: z.string().min(1, 'Please confirm your password'),
}).refine((data) => data.password === data.confirmPassword, {
  message: 'Passwords do not match',
  path: ['confirmPassword'],
});

export type SignupFormData = z.infer<typeof signupSchema>;

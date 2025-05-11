import { useState } from "react"
import { Link, useNavigate } from "react-router-dom"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { useSignup } from '../hooks/use-signup';
import type { UserSignupRequest, UserSignupResponse } from '@/api/users';
import type { SignupFormData } from '../types';
import { signupSchema } from '../types';

export function SignupForm({
  className,
  ...props
}: React.ComponentProps<"div">) {
  const [formData, setFormData] = useState<SignupFormData>({ 
    email: "", 
    password: "",
    confirmPassword: "",
    name: "",
    organisationName: ""
  });
  const [errors, setErrors] = useState<Partial<SignupFormData>>({});
  const [isLoading, setIsLoading] = useState(false);
  const { signup } = useSignup();
  const navigate = useNavigate();

  const validateForm = (): boolean => {
    const result = signupSchema.safeParse(formData);
    if (!result.success) {
      const fieldErrors: Partial<SignupFormData> = {};
      result.error.errors.forEach(err => {
        const field = err.path[0] as keyof SignupFormData;
        fieldErrors[field] = err.message;
      });
      setErrors(fieldErrors);
      return false;
    }
    setErrors({});
    return true;
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
    if (errors[name as keyof SignupFormData]) {
      setErrors(prev => ({ ...prev, [name]: undefined }));
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateForm()) return;
    setIsLoading(true);
    setErrors({});
    try {
      const payload: UserSignupRequest = {
        name: formData.name,
        email: formData.email,
        password: formData.password,
        organisation: { name: formData.organisationName }
      };
      const response: UserSignupResponse = await signup(payload);
      if (response && response.jwt_token) {
        localStorage.setItem('authToken', response.jwt_token);
      }
      navigate("/dashboard");
    } catch (err) {
      setErrors(prev => ({ ...prev, email: (err as Error)?.message || 'Signup failed' }));
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className={cn("mx-auto flex w-full flex-col justify-center space-y-6 sm:w-[350px]", className)} {...props}>
      <div className="flex flex-col space-y-2 text-center">
        <h1 className="text-2xl font-semibold tracking-tight">Create an account</h1>
        <p className="text-sm text-muted-foreground">Enter your email below to create your account</p>
      </div>
      <div className="grid gap-6">
        <form onSubmit={handleSubmit}>
          <div className="grid gap-4">
            <div className="grid gap-2">
              <Label htmlFor="name">Name</Label>
              <Input
                id="name"
                name="name"
                type="text"
                value={formData.name || ''}
                onChange={handleChange}
                className={errors.name ? "border-red-500" : ""}
              />
              {errors.name && (
                <p className="text-red-500 text-sm">{errors.name}</p>
              )}
            </div>
            <div className="grid gap-2">
              <Label htmlFor="organisationName">Organization Name</Label>
              <Input
                id="organisationName"
                name="organisationName"
                type="text"
                value={formData.organisationName || ''}
                onChange={handleChange}
                className={errors.organisationName ? "border-red-500" : ""}
              />
              {errors.organisationName && (
                <p className="text-red-500 text-sm">{errors.organisationName}</p>
              )}
            </div>
            <div className="grid gap-2">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                name="email"
                type="email"
                autoCapitalize="none"
                autoComplete="email"
                autoCorrect="off"
                value={formData.email}
                onChange={handleChange}
                className={errors.email ? "border-red-500" : ""}
              />
              {errors.email && (
                <p className="text-red-500 text-sm">{errors.email}</p>
              )}
            </div>
            <div className="grid gap-2">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                name="password"
                type="password"
                autoComplete="new-password"
                value={formData.password}
                onChange={handleChange}
                className={errors.password ? "border-red-500" : ""}
              />
              {errors.password && (
                <p className="text-red-500 text-sm">{errors.password}</p>
              )}
            </div>
            <div className="grid gap-2">
              <Label htmlFor="confirmPassword">Confirm Password</Label>
              <Input
                id="confirmPassword"
                name="confirmPassword"
                type="password"
                autoComplete="new-password"
                value={formData.confirmPassword}
                onChange={handleChange}
                className={errors.confirmPassword ? "border-red-500" : ""}
              />
              {errors.confirmPassword && (
                <p className="text-red-500 text-sm">{errors.confirmPassword}</p>
              )}
            </div>
            <Button type="submit" className="w-full mt-2" disabled={isLoading}>
              {isLoading ? (
                <>
                  <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  Creating account...
                </>
              ) : (
                "Sign Up"
              )}
            </Button>
          </div>
        </form>
      </div>
      <div className="text-center">
        <p className="text-sm text-muted-foreground">
          Already have an account?{" "}
          <Link to="/login" className="underline underline-offset-4 hover:text-primary">
            Login
          </Link>
        </p>
      </div>
    </div>
  )
}

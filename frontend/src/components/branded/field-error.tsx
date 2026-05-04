import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

interface FieldErrorProps {
  children?: ReactNode;
  className?: string;
}

export function FieldError({ children, className }: FieldErrorProps) {
  if (!children) return null;
  return (
    <p className={cn("text-[11.5px] text-danger mt-1 leading-tight", className)}>
      {children}
    </p>
  );
}

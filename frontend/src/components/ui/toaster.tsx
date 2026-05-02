import {
  Toast,
  ToastClose,
  ToastDescription,
  ToastProvider,
  ToastTitle,
  ToastViewport,
} from "@/components/ui/toast"
import { useToast } from "@/components/ui/use-toast"
import { CheckCircle2, AlertTriangle, Info } from "lucide-react"
import { cn } from "@/lib/utils"

function variantChip(variant?: "default" | "destructive" | "success" | null) {
  switch (variant) {
    case "destructive":
      return {
        Icon: AlertTriangle,
        ring: "bg-danger-bg border-danger-border text-danger",
      };
    case "success":
      return {
        Icon: CheckCircle2,
        ring: "bg-success-bg border-success-border text-success",
      };
    default:
      return {
        Icon: Info,
        ring: "bg-brand-bg border-brand-border text-brand",
      };
  }
}

export function Toaster() {
  const { toasts } = useToast()

  return (
    <ToastProvider>
      {toasts.map(function ({ id, title, description, action, variant, ...props }) {
        const { Icon, ring } = variantChip(variant);
        return (
          <Toast key={id} variant={variant} {...props}>
            <span className={cn("flex h-7 w-7 shrink-0 items-center justify-center rounded-md border", ring)}>
              <Icon className="h-4 w-4" />
            </span>
            <div className="flex-1 min-w-0 grid gap-1">
              {title && <ToastTitle>{title}</ToastTitle>}
              {description && (
                <ToastDescription>{description}</ToastDescription>
              )}
            </div>
            {action}
            <ToastClose />
          </Toast>
        )
      })}
      <ToastViewport />
    </ToastProvider>
  )
}

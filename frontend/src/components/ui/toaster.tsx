import {
  Toast,
  ToastClose,
  ToastDescription,
  ToastProvider,
  ToastTitle,
  ToastViewport,
} from "@/components/ui/toast"
import { useToast } from "@/components/ui/use-toast"
import { CheckCircle2, XCircle, AlertTriangle, Info } from "lucide-react"

function variantChip(variant?: "default" | "destructive" | "success" | "warning" | "info" | null) {
  switch (variant) {
    case "destructive":
      return { Icon: XCircle, color: "text-danger" };
    case "success":
      return { Icon: CheckCircle2, color: "text-success" };
    case "warning":
      return { Icon: AlertTriangle, color: "text-warn" };
    case "info":
      return { Icon: Info, color: "text-info" };
    default:
      return { Icon: Info, color: "text-brand" };
  }
}

export function Toaster() {
  const { toasts } = useToast()

  return (
    <ToastProvider>
      {[...toasts].reverse().map(function ({ id, title, description, action, variant, ...props }) {
        const { Icon, color } = variantChip(variant);
        return (
          <Toast key={id} variant={variant} {...props}>
            <Icon className={`mt-0.5 h-4 w-4 shrink-0 ${color}`} />
            <div className="flex-1 min-w-0 grid gap-0.5 pr-4">
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

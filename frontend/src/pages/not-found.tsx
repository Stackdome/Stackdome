import { ArrowLeft, ArrowRight } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { EyebrowLabel } from "@/components/branded/eyebrow-label";
import { StackdomeWordmark } from "@/components/branded/stackdome-mark";
import { ThemeToggle } from "@/components/theme-toggle";
import { Button } from "@/components/ui/button";

export default function NotFoundPage() {
  const navigate = useNavigate();
  const canGoBack = typeof window !== "undefined" && window.history.length > 1;

  return (
    <div className="relative min-h-svh overflow-hidden bg-background text-foreground">
      <div
        aria-hidden="true"
        className="pointer-events-none fixed inset-0 z-0 opacity-[0.10] text-foreground [background-image:linear-gradient(to_right,currentColor_1px,transparent_1px),linear-gradient(to_bottom,currentColor_1px,transparent_1px)] [background-size:64px_64px] [mask-image:radial-gradient(ellipse_at_center,black_30%,transparent_80%)] [-webkit-mask-image:radial-gradient(ellipse_at_center,black_30%,transparent_80%)]"
      />

      <div className="absolute left-4 top-4 z-30 md:left-6 md:top-6">
        <StackdomeWordmark size={20} />
      </div>

      <div className="absolute right-4 top-4 z-30 md:right-6 md:top-6">
        <ThemeToggle />
      </div>

      <main className="relative z-10 flex min-h-svh items-center justify-center px-6">
        <div className="flex max-w-xl flex-col items-center gap-6 text-center">
          <EyebrowLabel tone="brand" className="font-semibold">
            Error
          </EyebrowLabel>

          <h1
            className="font-mono text-[120px] font-semibold leading-none tracking-tight text-brand sm:text-[160px] xl:text-[200px]"
            style={{
              WebkitTextStroke: "2px currentColor",
              WebkitTextFillColor: "transparent",
            }}
          >
            404
          </h1>

          <h2 className="text-head font-semibold leading-tight tracking-tight">
            Page not found
          </h2>

          <p className="max-w-md text-body leading-relaxed text-muted-foreground">
            The page you&apos;re looking for doesn&apos;t exist or has moved.
            Check the URL, or head back to your stacks.
          </p>

          <div className="mt-2 flex flex-wrap items-center justify-center gap-3">
            {canGoBack && (
              <Button variant="ghost" onClick={() => navigate(-1)}>
                <ArrowLeft />
                Go back
              </Button>
            )}
            <Button onClick={() => navigate("/stacks")}>
              Back to stacks
              <ArrowRight />
            </Button>
          </div>
        </div>
      </main>

      <div className="absolute inset-x-0 bottom-6 z-10 flex items-center justify-between px-6 font-mono text-label text-muted-foreground md:px-10">
        <span>status: 404</span>
        <span>route_not_found</span>
      </div>
    </div>
  );
}

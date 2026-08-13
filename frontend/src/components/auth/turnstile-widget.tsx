import {
  forwardRef,
  useEffect,
  useImperativeHandle,
  useRef,
} from "react";
import { useTheme } from "@/contexts/theme-provider";

const turnstileScriptID = "cloudflare-turnstile-script";
const turnstileScriptURL =
  "https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit";

export interface TurnstileRenderOptions {
  sitekey: string;
  action: string;
  /* Default is a fixed 300px, narrower than the form's fields. */
  size?: "normal" | "flexible" | "compact";
  theme?: "light" | "dark" | "auto";
  callback: (token: string) => void;
  "expired-callback": () => void;
  "error-callback": () => void;
}

export interface TurnstileAPI {
  render(container: HTMLElement, options: TurnstileRenderOptions): string;
  reset(widgetID: string): void;
  remove(widgetID: string): void;
}

declare global {
  interface Window {
    turnstile?: TurnstileAPI;
  }
}

export interface TurnstileWidgetHandle {
  reset(): void;
}

interface TurnstileWidgetProps {
  siteKey: string;
  action: string;
  onToken(token: string): void;
  onUnavailable(): void;
}

let scriptPromise: Promise<TurnstileAPI> | null = null;

function loadTurnstile(): Promise<TurnstileAPI> {
  if (window.turnstile) return Promise.resolve(window.turnstile);
  if (scriptPromise) return scriptPromise;

  scriptPromise = new Promise<TurnstileAPI>((resolve, reject) => {
    const existingScript = document.getElementById(turnstileScriptID) as HTMLScriptElement | null;
    const script = existingScript ?? document.createElement("script");

    const handleLoad = () => {
      if (window.turnstile) {
        resolve(window.turnstile);
        return;
      }
      reject(new Error("Turnstile loaded without a browser API"));
    };
    const handleError = () => reject(new Error("failed to load Turnstile"));

    script.addEventListener("load", handleLoad, { once: true });
    script.addEventListener("error", handleError, { once: true });

    if (!existingScript) {
      script.id = turnstileScriptID;
      script.src = turnstileScriptURL;
      script.async = true;
      script.defer = true;
      document.head.appendChild(script);
    }
  }).catch((error: unknown) => {
    scriptPromise = null;
    throw error;
  });

  return scriptPromise;
}

export const TurnstileWidget = forwardRef<TurnstileWidgetHandle, TurnstileWidgetProps>(
  function TurnstileWidget({ siteKey, action, onToken, onUnavailable }, ref) {
    const containerRef = useRef<HTMLDivElement>(null);
    const widgetIDRef = useRef<string | null>(null);
    /* ThemeProvider's "system" and Turnstile's "auto" both mean
       prefers-color-scheme, so the names map straight across. */
    const { theme } = useTheme();
    const widgetTheme = theme === "system" ? "auto" : theme;
    const onTokenRef = useRef(onToken);
    const onUnavailableRef = useRef(onUnavailable);

    onTokenRef.current = onToken;
    onUnavailableRef.current = onUnavailable;

    useImperativeHandle(ref, () => ({
      reset() {
        if (widgetIDRef.current && window.turnstile) {
          window.turnstile.reset(widgetIDRef.current);
        }
      },
    }));

    useEffect(() => {
      let active = true;
      let turnstile: TurnstileAPI | undefined;

      loadTurnstile()
        .then((api) => {
          if (!active || !containerRef.current) return;
          turnstile = api;
          widgetIDRef.current = api.render(containerRef.current, {
            sitekey: siteKey,
            action,
            size: "flexible",
            theme: widgetTheme,
            callback: (token) => onTokenRef.current(token),
            "expired-callback": () => onTokenRef.current(""),
            "error-callback": () => {
              onTokenRef.current("");
              onUnavailableRef.current();
            },
          });
        })
        .catch(() => {
          if (active) onUnavailableRef.current();
        });

      return () => {
        active = false;
        if (widgetIDRef.current && turnstile) {
          turnstile.remove(widgetIDRef.current);
          widgetIDRef.current = null;
        }
      };
      /* widgetTheme is a dep because Turnstile has no setter for it: repainting
         means remove and re-render, which costs the visitor their token. */
    }, [siteKey, action, widgetTheme]);

    return (
      /* min-h reserves the height so the submit button doesn't jump when the
         widget paints. Padding, not margin: the form's space-y-3 owns the
         margins, so py-2 stacks to 20px of air here against the fields' 12px.
         That air is the only separation available, since Cloudflare renders the
         plate itself out of reach of our stylesheets. */
      <div
        ref={containerRef}
        className="cf-turnstile min-h-[65px] py-2"
        data-sitekey={siteKey}
        data-action={action}
      />
    );
  },
);

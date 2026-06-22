import type { Tone } from "../derive";
import { toneDotClass } from "../derive";

export interface RailNodeProps { tone: Tone; big?: boolean; pulse?: boolean; isLast?: boolean; children: React.ReactNode; }

export function RailNode({ tone, big, pulse, isLast, children }: RailNodeProps) {
  return (
    <div className="flex items-stretch gap-3.5">
      <div className="flex w-[34px] flex-none flex-col items-center">
        <span
          data-testid="rail-dot"
          className={[
            "flex-none rounded-full",
            // Align the dot with the first line of content: the big card header
            // sits ~2px down; a single-line history row's text centers ~9px down.
            big ? "mt-0.5 h-[15px] w-[15px] border-2 border-current bg-background" : "mt-[9px] h-2.5 w-2.5",
            big ? toneDotClass(tone).replace("bg-", "text-") : toneDotClass(tone),
            pulse ? "animate-pulse" : "",
          ].join(" ")}
        />
        <span
          data-testid="rail-connector"
          className={["mt-1 w-px flex-1 bg-border", isLast ? "invisible" : "visible"].join(" ")}
          style={{ minHeight: isLast ? 0 : 16 }}
        />
      </div>
      <div className={["min-w-0 flex-1", isLast ? "pb-1" : "pb-5"].join(" ")}>{children}</div>
    </div>
  );
}

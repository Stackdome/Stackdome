import { Button } from "@/components/ui/button";

type Preset = { label: string; expression: string; help: string };

const PRESETS: Preset[] = [
  { label: "Hourly",  expression: "0 0 * * * *", help: "Top of every hour" },
  { label: "Daily",   expression: "0 0 3 * * *", help: "Every day at 03:00 UTC" },
  { label: "Weekly",  expression: "0 0 3 * * 0", help: "Sundays at 03:00 UTC" },
  { label: "Monthly", expression: "0 0 3 1 * *", help: "1st of every month at 03:00 UTC" },
];

type Props = {
  value: string;
  onChange: (expression: string) => void;
  disabled?: boolean;
};

export function CronPresets({ value, onChange, disabled }: Props) {
  return (
    <div className="flex flex-wrap items-center gap-2">
      <span className="text-xs text-muted-foreground">Presets:</span>
      {PRESETS.map((p) => (
        <Button
          key={p.label}
          type="button"
          variant={value === p.expression ? "default" : "outline"}
          size="sm"
          disabled={disabled}
          title={p.help}
          aria-label={`${p.label} — ${p.help}`}
          aria-pressed={value === p.expression}
          onClick={() => onChange(p.expression)}
        >
          {p.label}
        </Button>
      ))}
    </div>
  );
}

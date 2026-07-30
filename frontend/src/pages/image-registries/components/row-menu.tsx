import { Ellipsis, KeyRound, ShieldCheck, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

export function RowMenu({
  onVerify,
  onUpdateCredentials,
  onRemove,
}: {
  onVerify: () => void;
  onUpdateCredentials: () => void;
  onRemove: () => void;
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="outline"
          size="icon"
          aria-label="Open row menu"
          className="h-8 w-8 rounded-md border-border"
        >
          <Ellipsis className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-[200px]">
        {/* Items open dialogs; defer callbacks until the menu has fully closed.
            Radix's DropdownMenu→Dialog composition races the menu's close
            against the dialog's mount and can leave body pointer-events "none".
            See https://github.com/radix-ui/primitives/issues/1836 */}
        <DropdownMenuItem onSelect={() => setTimeout(() => onVerify(), 0)}>
          <ShieldCheck className="h-4 w-4" />
          Verify registry access
        </DropdownMenuItem>
        <DropdownMenuItem onSelect={() => setTimeout(() => onUpdateCredentials(), 0)}>
          <KeyRound className="h-4 w-4" />
          Update credentials
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem variant="destructive" onSelect={() => setTimeout(() => onRemove(), 0)}>
          <Trash2 className="h-4 w-4" />
          Remove registry
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

import { type DomainName } from "../schemas/api-schema";
import { Globe, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";

interface DomainListItemProps {
  domain: Partial<DomainName>;
  index: number;
  onRemove: (index: number) => void;
}

export default function DomainListItem({ domain, index, onRemove }: DomainListItemProps) {
  return (
    <TooltipProvider>
      <div className="flex items-center justify-between p-4 border-b last:border-b-0">
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-2">
            <Globe className="h-4 w-4 text-muted-foreground" />
            <span>{domain.fqdn || "No domain specified"}</span>
          </div>
        </div>
        <div>
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => onRemove(index)}
                className="h-8 w-8 p-0 text-red-600 hover:text-red-700 hover:bg-red-50"
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>
              <p>Remove domain</p>
            </TooltipContent>
          </Tooltip>
        </div>
      </div>
    </TooltipProvider>
  );
}

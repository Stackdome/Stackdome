import { useState } from "react";
import { AlertCircle, Info } from "lucide-react";
import { type DomainName, validateDomainName } from "../schemas/api-schema";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";

interface AddDomainDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onAddDomain: (domain: DomainName) => void;
  existingDomains?: Partial<DomainName>[];
  isLoading?: boolean;
}

export default function AddDomainDialog({
  open,
  onOpenChange,
  onAddDomain,
  existingDomains = [],
  isLoading = false,
}: AddDomainDialogProps) {
  const [domainFqdn, setDomainFqdn] = useState("");
  const [error, setError] = useState<string | null>(null);

  const handleDomainInputChange = (value: string) => {
    setDomainFqdn(value);
    if (error) {
      setError(null);
    }
  };

  const handleAddDomain = () => {
    if (!domainFqdn.trim()) {
      setError("Domain name is required");
      return;
    }

    const validation = validateDomainName(domainFqdn.trim());
    if (!validation.isValid) {
      setError(validation.error || "Invalid domain name");
      return;
    }

    // Check for duplicates
    const normalizedDomain = domainFqdn.trim().toLowerCase();
    const existingDomain = existingDomains.find(
      domain => domain.fqdn?.toLowerCase() === normalizedDomain
    );

    if (existingDomain) {
      setError("This domain already exists");
      return;
    }

    // Add the domain
    onAddDomain({ fqdn: normalizedDomain });

    // Reset state
    setDomainFqdn("");
    setError(null);
  };

  // Reset form when dialog closes
  const handleOpenChange = (open: boolean) => {
    if (!open) {
      setDomainFqdn("");
      setError(null);
    }
    onOpenChange(open);
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md mx-auto">
        <DialogHeader>
          <DialogTitle>Add Domain</DialogTitle>
          <DialogDescription>
            Configure the domain for your organization.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          <div>
            <div className="flex items-center gap-1 mb-2">
              <Label htmlFor="domain-fqdn" className="text-sm font-medium">Domain Name</Label>
              <TooltipProvider>
                <Tooltip delayDuration={300}>
                  <TooltipTrigger asChild>
                    <Info className="size-3.5 text-muted-foreground cursor-pointer" />
                  </TooltipTrigger>
                  <TooltipContent side="top">
                    Enter a valid domain name like example.com
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </div>
            <Input
              id="domain-fqdn"
              placeholder="example.com"
              value={domainFqdn}
              onChange={(e) => handleDomainInputChange(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter" && domainFqdn.trim() && !error) {
                  handleAddDomain();
                }
              }}
              className={error ? "border-red-500 focus:border-red-500" : ""}
            />
            {error && (
              <p className="text-sm text-red-600 mt-1 flex items-center gap-1">
                <AlertCircle className="h-4 w-4" />
                {error}
              </p>
            )}
            <p className="text-xs text-muted-foreground mt-1">
              Examples: example.com, subdomain.example.org, app.company.co.uk
            </p>
          </div>
        </div>
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="outline" type="button">Cancel</Button>
          </DialogClose>
          <Button
            onClick={handleAddDomain}
            disabled={!domainFqdn.trim() || !!error || isLoading}
          >
            {isLoading ? "Adding..." : "Add Domain"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

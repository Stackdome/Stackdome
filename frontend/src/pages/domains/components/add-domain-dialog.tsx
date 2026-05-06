import { useState } from "react";
import { Loader2 } from "lucide-react";
import { type DomainName, validateDomainName } from "../schemas/api-schema";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { FieldShell } from "@/components/branded";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

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
          <FieldShell
            label="Domain Name"
            htmlFor="domain-fqdn"
            required
            hint="Examples: example.com, subdomain.example.org, app.company.co.uk"
            error={error}
          >
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
              className={error ? "border-danger focus:border-danger" : ""}
            />
          </FieldShell>
        </div>
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="outline" type="button">Cancel</Button>
          </DialogClose>
          <Button
            onClick={handleAddDomain}
            disabled={!domainFqdn.trim() || !!error || isLoading}
          >
            {isLoading && <Loader2 className="h-4 w-4 animate-spin" />}
            Add Domain
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

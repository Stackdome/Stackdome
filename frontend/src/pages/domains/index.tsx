import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { AlertCircle, Globe, Plus } from "lucide-react";
import { getCurrentOrganizationId } from "@/helpers/common";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { TooltipProvider } from "@/components/ui/tooltip";
import { useBreadcrumb } from "@/hooks/use-breadcrumb";
import { useToast } from "@/components/ui/use-toast";
import { Separator } from "@/components/ui/separator";
import * as organizationApi from "@/api/organizations";
import type { Organization } from "@/api/organizations";
import { type DomainName, createDomainFromForm } from "./schemas/api-schema";
import DomainListItem from "./components/domain-list-item";
import AddDomainDialog from "./components/add-domain-dialog";

export default function DomainsPage() {
  const [organization, setOrganization] = useState<Organization | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showAddDialog, setShowAddDialog] = useState(false);

  const { setCustomLabel, setPathLoading } = useBreadcrumb();
  const { toast } = useToast();

  useEffect(() => {
    const loadOrganization = async () => {
      const orgId = getCurrentOrganizationId();
      if (!orgId) {
        setError("Organization ID not found.");
        setLoading(false);
        return;
      }

      const currentPath = `/domains`;
      setPathLoading(currentPath, true);
      setLoading(true);

      try {
        const data = await organizationApi.getOrganization(orgId);
        setOrganization(data);
        setCustomLabel(currentPath, "Domain configuration");
      } catch (err) {
        console.error("Failed to load organization:", err);
        setError("Failed to load domains. Please try again later.");
      } finally {
        setLoading(false);
        setPathLoading(currentPath, false);
      }
    };

    loadOrganization();
  }, [setCustomLabel, setPathLoading]);

  const handleDomainsChange = async (updatedDomains: Partial<DomainName>[]) => {
    if (!organization) return;

    const orgId = getCurrentOrganizationId();
    if (!orgId) return;

    setSaving(true);

    try {
      // Validate and convert domains
      const validatedDomains = updatedDomains
        .filter(domain => domain.fqdn && domain.fqdn.trim())
        .map(domain => createDomainFromForm({ fqdn: domain.fqdn }));

      const updatedOrg = await organizationApi.updateOrganization(orgId, {
        ...organization,
        domains: validatedDomains
      });

      setOrganization(updatedOrg);

      toast({
        title: "Domain configured",
        description: "Domain configuration has been saved successfully.",
      });
    } catch (err) {
      console.error("Failed to update domains:", err);
      toast({
        title: "Error",
        description: "Failed to save domain configuration. Please try again.",
        variant: "destructive",
      });
    } finally {
      setSaving(false);
    }
  };

  const handleAddDomain = (newDomain: DomainName) => {
    if (!organization) return;

    const updatedDomains = [...(organization.domains || []), newDomain];
    handleDomainsChange(updatedDomains);
    setShowAddDialog(false);
  };

  const handleRemoveDomain = (index: number) => {
    if (!organization) return;

    const updatedDomains = [...(organization.domains || [])];
    updatedDomains.splice(index, 1);
    handleDomainsChange(updatedDomains);

    toast({
      title: "Domain deleted",
      description: "The domain has been removed from your configuration.",
      variant: "destructive",
    });
  };

  if (loading) {
    return (
      <div className="flex flex-1 flex-col p-4 pt-0 h-full items-center justify-center">
        <svg className="animate-spin h-10 w-10 text-primary" viewBox="0 0 24 24">
          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8z" />
        </svg>
        <p className="mt-2 text-muted-foreground">Loading domains...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-8 text-center">
        <AlertCircle className="mx-auto h-12 w-12 text-destructive mb-4" />
        <h2 className="text-xl font-semibold mb-2">Error Loading Domains</h2>
        <p className="text-muted-foreground mb-4">{error}</p>
        <Button onClick={() => window.location.reload()}>
          Try Again
        </Button>
      </div>
    );
  }

  const domains = organization?.domains || [];

  return (
    <TooltipProvider>
      <div className="p-6">
        <header className="mb-6">
          <div className="flex justify-between items-center">
            <div>
              <div className="flex items-center gap-3 mb-1">
                <h1 className="text-2xl font-bold">Domain configuration</h1>
              </div>
            </div>
          </div>
          <Separator className="mt-4" />
        </header>

        <Card className="rounded-lg">
          <CardHeader className="pb-3">
            <CardTitle className="text-xl flex items-center gap-2">
              <Globe className="h-5 w-5" />
              Domains
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            {domains.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-20">
                <Globe className="h-12 w-12 mb-4 text-muted-foreground" />
                <h3 className="text-xl font-medium mb-2">No domain configured</h3>
                <p className="text-muted-foreground mb-6">Configure a domain for your organization.</p>
                <Button onClick={() => setShowAddDialog(true)}>
                  <Plus className="mr-2 h-4 w-4" />
                  Add Domain
                </Button>
              </div>
            ) : (
              <div className="space-y-0">
                <div className="border rounded-lg">
                  {domains.map((domain, index) => (
                    <DomainListItem
                      key={index}
                      domain={domain}
                      index={index}
                      onRemove={handleRemoveDomain}
                    />
                  ))}
                </div>
              </div>
            )}
          </CardContent>
        </Card>

        <AddDomainDialog
          open={showAddDialog}
          onOpenChange={setShowAddDialog}
          onAddDomain={handleAddDomain}
          existingDomains={domains}
          isLoading={saving}
        />
      </div>
    </TooltipProvider>
  );
}

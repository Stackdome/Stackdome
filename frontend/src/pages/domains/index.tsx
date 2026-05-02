import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { AlertCircle, Globe, PlusCircle, Loader2 } from "lucide-react";
import { getCurrentOrganizationId } from "@/helpers/common";
import { TooltipProvider } from "@/components/ui/tooltip";
import { useBreadcrumb } from "@/hooks/use-breadcrumb";
import { useToast } from "@/components/ui/use-toast";
import { PageHeader, Panel, EmptyState } from "@/components/branded";
import * as organizationApi from "@/api/organizations";
import type { Organization } from "@/api/organizations";
import { getErrorMessage } from "@/api/client";
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
        setCustomLabel(currentPath, "Domains");
      } catch (err) {
        console.error("Failed to load organization:", err);
        setError(getErrorMessage(err));
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


    } catch (err) {
      console.error("Failed to update domains:", err);
      toast({
        title: "Error",
        description: getErrorMessage(err),
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
    toast({
      title: "Domain added",
      description: "Domain configuration added successfully.",
    });
  };

  const handleRemoveDomain = (index: number) => {
    if (!organization) return;

    const updatedDomains = [...(organization.domains || [])];
    updatedDomains.splice(index, 1);
    handleDomainsChange(updatedDomains);

    toast({
      title: "Domain deleted",
      description: "The domain configuration removed from your configuration.",
      variant: "destructive",
    });
  };

  if (loading) {
    return (
      <div className="flex flex-1 flex-col items-center justify-center min-h-[calc(100vh-4rem)] p-4">
        <Loader2 className="h-10 w-10 animate-spin text-primary" />
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
      <div className="p-8 space-y-8">
        <PageHeader
          eyebrow="Platform"
          title="Domains"
          subtitle="Configure custom domains for your organization"
          actions={
            <Button onClick={() => setShowAddDialog(true)} className="bg-brand text-white hover:bg-brand-darker">
              <PlusCircle className="mr-2 h-4 w-4" />
              Add Domain
            </Button>
          }
        />

        <Panel
          title="All Domains"
          count={domains.length}
          bodyClassName={domains.length === 0 ? "p-5" : "p-0"}
        >
          {domains.length === 0 ? (
            <EmptyState
              icon={<Globe className="h-8 w-8" />}
              title="No domain configured"
              description="Configure a domain for your organization."
              action={
                <Button onClick={() => setShowAddDialog(true)} variant="outline">
                  <PlusCircle className="mr-2 h-4 w-4" />
                  Add Domain
                </Button>
              }
            />
          ) : (
            domains.map((domain, index) => (
              <DomainListItem
                key={index}
                domain={domain}
                index={index}
                onRemove={handleRemoveDomain}
              />
            ))
          )}
        </Panel>

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

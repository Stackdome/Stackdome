import { useCallback, useEffect, useState } from "react";
import { PlusCircle, AlertCircle, Loader2, KeyRound } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { TooltipProvider } from "@/components/ui/tooltip";
import { PageHeader, Panel, EmptyState } from "@/components/branded";
import { useConfirm } from "@/components/branded/confirm";
import { useToast } from "@/components/ui/use-toast";
import { listApiTokens, revokeApiToken, type APIToken } from "@/api/api-tokens";
import { getErrorMessage } from "@/api/client";
import { useBreadcrumb } from "@/hooks/use-breadcrumb";
import { CreateTokenDialog } from "./components/create-token-dialog";
import { cn } from "@/lib/utils";

function formatDate(value?: string): string {
  if (!value) return "Never";
  return new Date(value).toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" });
}

export default function ApiTokensPage() {
  const [tokens, setTokens] = useState<APIToken[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const { toast } = useToast();
  const confirm = useConfirm();
  const { setCustomLabel, setPathLoading } = useBreadcrumb();

  const fetchTokens = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await listApiTokens();
      setTokens(data.items || []);
    } catch (e) {
      console.error("Failed to fetch API tokens:", e);
      setError(getErrorMessage(e));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchTokens();
  }, [fetchTokens]);

  useEffect(() => {
    const currentPath = "/settings/api-tokens";
    setCustomLabel(currentPath, "API Tokens");
    setPathLoading(currentPath, loading);
  }, [setCustomLabel, setPathLoading, loading]);

  async function requestRevoke(token: APIToken) {
    if (!token.id) return;
    const ok = await confirm({
      title: "Revoke token?",
      description: `"${token.name}" will stop working immediately. This cannot be undone.`,
      confirmLabel: "Revoke",
      variant: "destructive",
    });
    if (!ok) return;
    try {
      await revokeApiToken(token.id);
      fetchTokens();
      toast({ title: "Token revoked", description: "The API token has been revoked.", variant: "success" });
    } catch (e) {
      console.error("Failed to revoke API token:", e);
      toast({ title: "Failed to revoke token", description: getErrorMessage(e), variant: "destructive" });
    }
  }

  if (loading) {
    return (
      <div className="flex flex-1 flex-col items-center justify-center min-h-[calc(100vh-4rem)] p-4">
        <Loader2 className="h-10 w-10 animate-spin text-primary" />
        <p className="mt-2 text-muted-foreground">Loading API tokens...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-8 text-center">
        <AlertCircle className="mx-auto h-12 w-12 text-destructive mb-4" />
        <h2 className="text-xl font-semibold mb-2">Error Loading API Tokens</h2>
        <p className="text-muted-foreground mb-4">{error}</p>
        <Button onClick={() => window.location.reload()}>Try Again</Button>
      </div>
    );
  }

  return (
    <TooltipProvider>
      <div className="p-8 space-y-8">
        <PageHeader
          eyebrow="Platform"
          title="API Tokens"
          subtitle="Issue tokens for scripts, CI, and agents to authenticate as you"
          actions={
            <Button onClick={() => setShowCreateDialog(true)}>
              <PlusCircle className="h-4 w-4" />
              Create token
            </Button>
          }
        />

        {tokens.length === 0 ? (
          <EmptyState
            icon={<KeyRound className="h-8 w-8" />}
            title="No API tokens yet"
            description="Create a token to let scripts, CI, or agents authenticate as you."
            action={
              <Button onClick={() => setShowCreateDialog(true)}>
                <PlusCircle className="h-4 w-4" />
                Create token
              </Button>
            }
          />
        ) : (
          <Panel title="API Tokens" count={tokens.length} bodyClassName="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Prefix</TableHead>
                  <TableHead>Scopes</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead>Expires</TableHead>
                  <TableHead>Last used</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {tokens.map((token) => {
                  const revoked = !!token.revoked_at;
                  return (
                    <TableRow key={token.id} className={cn(revoked && "opacity-50")}>
                      <TableCell className="font-medium">
                        <div className="flex items-center gap-2">
                          {token.name}
                          {revoked && <Badge variant="secondary">Revoked</Badge>}
                        </div>
                      </TableCell>
                      <TableCell className="font-mono text-xs text-muted-foreground">{token.token_prefix}</TableCell>
                      <TableCell className="max-w-[220px] truncate text-xs text-muted-foreground" title={token.scopes?.join(", ")}>
                        {token.scopes?.join(", ") || "—"}
                      </TableCell>
                      <TableCell className="text-muted-foreground">{formatDate(token.created_at)}</TableCell>
                      <TableCell className="text-muted-foreground">{formatDate(token.expires_at)}</TableCell>
                      <TableCell className="text-muted-foreground">{formatDate(token.last_used_at)}</TableCell>
                      <TableCell className="text-right">
                        {!revoked && (
                          <Button variant="outline" size="sm" onClick={() => requestRevoke(token)}>
                            Revoke
                          </Button>
                        )}
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          </Panel>
        )}

        <CreateTokenDialog
          open={showCreateDialog}
          onOpenChange={setShowCreateDialog}
          onCreated={fetchTokens}
        />
      </div>
    </TooltipProvider>
  );
}

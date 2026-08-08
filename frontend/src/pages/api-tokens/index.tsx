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
import { PageHeader, Panel, EmptyState } from "@/components/branded";
import { useConfirm } from "@/components/branded/confirm";
import { useToast } from "@/components/ui/use-toast";
import { listApiTokens, revokeApiToken, type APIToken } from "@/api/api-tokens";
import { getErrorMessage } from "@/api/client";
import { useBreadcrumb } from "@/hooks/use-breadcrumb";
import { CreateTokenDialog } from "./components/create-token-dialog";
import { accessLabel } from "./access";
import { cn } from "@/lib/utils";

const COLUMN_HEAD_CLASS = "text-[11px] uppercase tracking-widest text-muted-foreground font-mono";

// The breadcrumb key must match the router path exactly or the label never applies.
export const API_TOKENS_PATH = "/settings/api-tokens";

function formatDate(value: string | undefined, fallback: string): string {
  if (!value) return fallback;
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
      console.error("Failed to fetch API tokens:", getErrorMessage(e));
      setError(getErrorMessage(e));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchTokens();
  }, [fetchTokens]);

  useEffect(() => {
    setCustomLabel(API_TOKENS_PATH, "API Tokens");
    setPathLoading(API_TOKENS_PATH, loading);
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
      console.error("Failed to revoke API token:", getErrorMessage(e));
      toast({ title: "Failed to revoke token", description: getErrorMessage(e), variant: "destructive" });
    }
  }

  // Only the first load takes over the page — a refetch after revoke keeps
  // the table on screen instead of flashing back to a spinner.
  if (loading && tokens.length === 0 && !error) {
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
        <Button onClick={fetchTokens}>Try Again</Button>
      </div>
    );
  }

  return (
    <div className="p-8 space-y-8">
      <PageHeader
        eyebrow="Platform"
        title="API Tokens"
        subtitle="Issue tokens for scripts, CI, and agents"
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
          description="Create a token to let scripts, CI, or agents call the API."
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
              <TableRow className="hover:bg-transparent">
                <TableHead className={COLUMN_HEAD_CLASS}>Token</TableHead>
                <TableHead className={COLUMN_HEAD_CLASS}>Access</TableHead>
                <TableHead className={COLUMN_HEAD_CLASS}>Created</TableHead>
                <TableHead className={COLUMN_HEAD_CLASS}>Expires</TableHead>
                <TableHead className={COLUMN_HEAD_CLASS}>Last used</TableHead>
                <TableHead>
                  <span className="sr-only">Actions</span>
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {tokens.map((token) => {
                const expired = !!token.expires_at && new Date(token.expires_at) < new Date();
                return (
                  <TableRow
                    key={token.id}
                    className={cn("border-b border-border hover:bg-muted/50", expired && "opacity-50")}
                  >
                    <TableCell className="py-3.5">
                      <div className="flex items-center gap-2">
                        <div className="min-w-0">
                          <div className="truncate text-sm font-medium text-foreground">{token.name}</div>
                          <div className="font-mono text-xs text-muted-foreground">{token.token_prefix}</div>
                        </div>
                        {expired && <Badge variant="secondary">Expired</Badge>}
                      </div>
                    </TableCell>
                    <TableCell className="py-3.5">
                      <span
                        className="inline-flex items-center rounded border border-border bg-muted px-2 py-px font-mono text-[11px] text-muted-foreground"
                        title={token.scopes?.join(", ")}
                      >
                        {accessLabel(token.scopes)}
                      </span>
                    </TableCell>
                    <TableCell className="py-3.5 text-sm text-muted-foreground">{formatDate(token.created_at, "—")}</TableCell>
                    <TableCell className="py-3.5 text-sm text-muted-foreground">{formatDate(token.expires_at, "Never")}</TableCell>
                    <TableCell className="py-3.5 text-sm text-muted-foreground">{formatDate(token.last_used_at, "Never")}</TableCell>
                    <TableCell className="py-3.5 text-right">
                      <Button variant="outline" size="sm" onClick={() => requestRevoke(token)}>
                        Revoke
                      </Button>
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
  );
}

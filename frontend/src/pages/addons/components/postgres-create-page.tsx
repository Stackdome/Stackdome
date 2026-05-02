import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Loader2, X, Rocket, ExternalLink, AlertCircle, ChevronDown } from "lucide-react";
import { ZodError } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Separator } from "@/components/ui/separator";
import { cn } from "@/lib/utils";
import { getCurrentOrganizationId } from "@/helpers/common";
import { getErrorMessage } from "@/api/client";
import { useToast } from "@/components/ui/use-toast";
import { useBreadcrumb } from "@/hooks/use-breadcrumb";
import * as addonsApi from "@/api/addons";
import {
  PostgresAddonFormSchema,
  defaultFormValues,
  type PostgresAddonFormValues,
} from "../schemas/form-schema";
import { PLAN_PRESETS, type PlanId } from "../lib/plan-presets";
import {
  addonToFormValues,
  buildCreateInput,
  JsonAreaParseError,
} from "../lib/payload";

// Mirrors backend: pkg/worker/postgresaddon/image_catalog_reconciler.go (CloudNativePG images)
// and config/openapi/stackdome_api.yaml (PostgresVersion.major: 13..17).
// TODO: ideally fetched from a backend "supported versions" endpoint so the
// list can't drift from the operator's image catalog.
const POSTGRES_VERSIONS = [17, 16, 15, 14, 13];
const ADVANCED_DOCS_URL = "https://docs.stackdome.io/addons/postgres/advanced";

type FormErrors = Partial<Record<string, string>>;

export default function PostgresFormPage() {
  const navigate = useNavigate();
  const { id: editId } = useParams<{ id: string }>();
  const isEdit = Boolean(editId);
  const { toast } = useToast();
  const { setCustomLabel, setPathLoading, registerNonClickablePath } = useBreadcrumb();
  const [values, setValues] = useState<PostgresAddonFormValues>(() =>
    defaultFormValues(""),
  );
  const [errors, setErrors] = useState<FormErrors>({});
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [loadingAddon, setLoadingAddon] = useState(isEdit);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [originalValues, setOriginalValues] = useState<PostgresAddonFormValues | null>(null);
  const [confirmAdvancedOpen, setConfirmAdvancedOpen] = useState(false);

  // Static labels + non-clickable registrations (don't depend on values.name).
  useEffect(() => {
    if (isEdit && editId) {
      setCustomLabel("/addons/postgres", "Postgres");
      setCustomLabel(`/addons/postgres/${editId}/edit`, "Edit");
      const disposers = [
        registerNonClickablePath("/addons/postgres"),
        registerNonClickablePath(`/addons/postgres/${editId}`),
      ];
      return () => disposers.forEach((d) => d());
    }
    setCustomLabel("/addons/create", "Create");
    setCustomLabel("/addons/create/postgres", "Postgres");
    setPathLoading("/addons/create/postgres", false);
    const dispose = registerNonClickablePath("/addons/create");
    return () => dispose();
  }, [setCustomLabel, setPathLoading, registerNonClickablePath, isEdit, editId]);

  // Dynamic label that follows the typed name in edit mode.
  useEffect(() => {
    if (!isEdit || !editId) return;
    setCustomLabel(`/addons/postgres/${editId}`, values.name || "Postgres");
    setPathLoading(`/addons/postgres/${editId}/edit`, loadingAddon);
  }, [setCustomLabel, setPathLoading, isEdit, editId, values.name, loadingAddon]);

  // Hydrate form from existing addon when editing
  useEffect(() => {
    if (!isEdit || !editId) return;
    const orgId = getCurrentOrganizationId();
    if (!orgId) {
      setLoadError("No organization selected");
      setLoadingAddon(false);
      return;
    }
    let cancelled = false;
    setLoadingAddon(true);
    addonsApi
      .getPostgresAddon(orgId, editId)
      .then((addon) => {
        if (cancelled) return;
        const hydrated = addonToFormValues(addon);
        setValues(hydrated);
        setOriginalValues(hydrated);
        setLoadError(null);
      })
      .catch((e) => {
        if (cancelled) return;
        console.error("Failed to load addon:", e);
        setLoadError(getErrorMessage(e));
      })
      .finally(() => {
        if (!cancelled) setLoadingAddon(false);
      });
    return () => {
      cancelled = true;
    };
  }, [isEdit, editId]);

  const update = <K extends keyof PostgresAddonFormValues>(
    key: K,
    value: PostgresAddonFormValues[K],
  ) => {
    setValues((v) => ({ ...v, [key]: value }));
    if (errors[key as string]) {
      setErrors((e) => ({ ...e, [key as string]: undefined }));
    }
  };

  const validate = (): boolean => {
    try {
      PostgresAddonFormSchema.parse(values);
      if (values.plan === "custom") {
        const anyCustom =
          values.customCpuRequest ||
          values.customCpuLimit ||
          values.customMemoryRequest ||
          values.customMemoryLimit;
        if (!anyCustom) {
          setErrors({
            customCpuRequest: "Set at least one CPU or memory value for the Custom plan",
          });
          return false;
        }
      }
      setErrors({});
      return true;
    } catch (e) {
      if (e instanceof ZodError) {
        const next: FormErrors = {};
        for (const issue of e.issues) {
          next[issue.path.join(".")] = issue.message;
        }
        setErrors(next);
      }
      return false;
    }
  };

  const handleCancel = () => {
    if (window.history.length > 2 && window.history.state?.idx !== 0) {
      navigate(-1);
    } else {
      navigate("/addons", { replace: true });
    }
  };

  const persist = async () => {
    const orgId = getCurrentOrganizationId();
    if (!orgId) {
      setSubmitError("No organization selected");
      return;
    }
    setSubmitting(true);
    setSubmitError(null);
    try {
      const input = buildCreateInput(values);
      if (isEdit && editId) {
        await addonsApi.updatePostgresAddon(orgId, editId, input);
        toast({
          title: "Addon updated",
          description: "Changes have been applied.",
        });
      } else {
        await addonsApi.createPostgresAddon(orgId, input);
        toast({
          title: "Addon created",
          description: "Provisioning has started; status will update as it's ready.",
        });
      }
      setConfirmAdvancedOpen(false);
      navigate("/addons", { replace: true });
    } catch (e) {
      if (e instanceof JsonAreaParseError) {
        setErrors((errs) => ({ ...errs, advancedJson: e.message }));
        setConfirmAdvancedOpen(false);
      } else {
        console.error("Failed to save addon:", e);
        setSubmitError(getErrorMessage(e));
      }
    } finally {
      setSubmitting(false);
    }
  };

  const handleSubmit = () => {
    if (!validate()) return;
    // Only gate on advanced changes, since basic-form fields are visible to the
    // user as they edit. Advanced JSON is the one place where data can vanish
    // unnoticed, so confirm intent there.
    if (
      isEdit &&
      originalValues &&
      values.advancedJson !== originalValues.advancedJson
    ) {
      setConfirmAdvancedOpen(true);
      return;
    }
    void persist();
  };

  const handleResetAdvanced = () => {
    if (!originalValues) return;
    setValues((v) => ({ ...v, advancedJson: originalValues.advancedJson }));
    if (errors.advancedJson) {
      setErrors((e) => ({ ...e, advancedJson: undefined }));
    }
  };

  const showCustomCompute = values.plan === "custom";
  const advancedDirty =
    isEdit &&
    originalValues != null &&
    values.advancedJson !== originalValues.advancedJson;
  const advancedCleared =
    advancedDirty &&
    Boolean(originalValues?.advancedJson.trim()) &&
    !values.advancedJson.trim();

  if (isEdit && loadingAddon) {
    return (
      <div className="flex flex-1 flex-col items-center justify-center min-h-[calc(100vh-4rem)] p-4">
        <Loader2 className="h-10 w-10 animate-spin text-primary" />
        <p className="mt-2 text-muted-foreground">Loading addon...</p>
      </div>
    );
  }

  if (isEdit && loadError) {
    return (
      <div className="p-8 text-center">
        <AlertCircle className="mx-auto h-12 w-12 text-destructive mb-4" />
        <h2 className="text-xl font-semibold mb-2">Failed to load addon</h2>
        <p className="text-muted-foreground mb-4">{loadError}</p>
        <Button onClick={() => navigate("/addons")}>Back to Addons</Button>
      </div>
    );
  }

  const submitLabel = isEdit
    ? submitting ? "Saving..." : "Save changes"
    : submitting ? "Creating..." : "Create addon";

  return (
    <div className="p-6">
      <header className="mb-6">
        <div className="flex justify-between items-center">
          <div>
            <h1 className="text-2xl font-bold mb-1">
              {isEdit ? `Edit ${values.name}` : "Create Postgres addon"}
            </h1>
            <p className="text-muted-foreground text-sm">
              {isEdit
                ? "Update configuration for this Postgres addon."
                : "Provision a managed PostgreSQL database for your stacks."}
            </p>
          </div>
          <div className="flex gap-3">
            <Button variant="outline" size="lg" onClick={handleCancel} disabled={submitting}>
              <X className="h-4 w-4" />
              Cancel
            </Button>
            <Button size="lg" onClick={handleSubmit} disabled={submitting}>
              {submitting ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <Rocket className="h-4 w-4" />
              )}
              <span className="font-semibold">{submitLabel}</span>
            </Button>
          </div>
        </div>
        <Separator className="mt-4" />
      </header>

      {submitError && (
        <div className="text-sm text-destructive bg-destructive/10 p-3 rounded-md mb-6">
          {submitError}
        </div>
      )}

      <div className="flex flex-col">
        <Card className="mb-6 rounded-lg overflow-hidden">
          <CardHeader className="pb-3">
            <CardTitle className="text-xl">Addon Information</CardTitle>
          </CardHeader>
          <Separator />
          <CardContent className="pt-6">
            <div className="grid gap-6 max-w-5xl">
              <div>
                <Label htmlFor="addon-name" className="text-sm font-medium flex items-center gap-1 mb-2">
                  Name <span className="text-red-500">*</span>
                </Label>
                <Input
                  id="addon-name"
                  value={values.name}
                  onChange={(e) => update("name", e.target.value)}
                  placeholder="e.g. main-db"
                  disabled={isEdit}
                  className={cn("max-w-md", errors.name ? "border-destructive" : "")}
                />
                <p className="text-xs text-muted-foreground mt-2">
                  {isEdit
                    ? "Name cannot be changed after creation."
                    : "Lowercase letters, numbers, and hyphens. Must start and end with a letter or number."}
                </p>
                {errors.name && (
                  <p className="text-sm text-destructive mt-1">{errors.name}</p>
                )}
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="mb-6 rounded-lg overflow-hidden">
          <CardHeader className="pb-3">
            <CardTitle className="text-xl">Configuration</CardTitle>
            <CardDescription className="mt-1">
              Pick a plan, set storage, and choose a Postgres version.
            </CardDescription>
          </CardHeader>
          <Separator />
          <CardContent className="pt-6">
            <div className="rounded-md border overflow-hidden max-w-3xl">
              <table className="w-full text-sm">
                <thead className="bg-muted/30">
                  <tr>
                    <th className="text-left px-4 py-2.5 font-medium">Plan</th>
                    <th className="text-left px-4 py-2.5 font-medium">CPU</th>
                    <th className="text-left px-4 py-2.5 font-medium">Memory</th>
                  </tr>
                </thead>
                <tbody>
                  {PLAN_PRESETS.map((preset) => {
                    const selected = values.plan === preset.id;
                    return (
                      <tr
                        key={preset.id}
                        className={cn(
                          "border-t cursor-pointer transition-colors",
                          selected ? "bg-primary/5" : "hover:bg-muted/30",
                        )}
                        onClick={() => update("plan", preset.id as PlanId)}
                      >
                        <td className="px-4 py-2.5">
                          <label className="flex items-center gap-2 cursor-pointer">
                            <input
                              type="radio"
                              name="plan"
                              value={preset.id}
                              checked={selected}
                              onChange={() => update("plan", preset.id as PlanId)}
                              className="cursor-pointer"
                            />
                            <span className={selected ? "font-medium" : ""}>
                              {preset.label}
                            </span>
                          </label>
                        </td>
                        <td className="px-4 py-2.5 text-muted-foreground">{preset.cpu}</td>
                        <td className="px-4 py-2.5 text-muted-foreground">{preset.memory}</td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>

            {showCustomCompute && (
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 mt-4 max-w-2xl border rounded-md p-4 bg-muted/20">
                <p className="sm:col-span-2 text-sm font-medium">Custom resources</p>
                <div className="space-y-1">
                  <Label htmlFor="cpu-req" className="text-xs">CPU request</Label>
                  <Input
                    id="cpu-req"
                    placeholder="250m"
                    value={values.customCpuRequest ?? ""}
                    onChange={(e) => update("customCpuRequest", e.target.value)}
                  />
                </div>
                <div className="space-y-1">
                  <Label htmlFor="cpu-lim" className="text-xs">CPU limit</Label>
                  <Input
                    id="cpu-lim"
                    placeholder="500m"
                    value={values.customCpuLimit ?? ""}
                    onChange={(e) => update("customCpuLimit", e.target.value)}
                  />
                </div>
                <div className="space-y-1">
                  <Label htmlFor="mem-req" className="text-xs">Memory request</Label>
                  <Input
                    id="mem-req"
                    placeholder="512Mi"
                    value={values.customMemoryRequest ?? ""}
                    onChange={(e) => update("customMemoryRequest", e.target.value)}
                  />
                </div>
                <div className="space-y-1">
                  <Label htmlFor="mem-lim" className="text-xs">Memory limit</Label>
                  <Input
                    id="mem-lim"
                    placeholder="1Gi"
                    value={values.customMemoryLimit ?? ""}
                    onChange={(e) => update("customMemoryLimit", e.target.value)}
                  />
                </div>
                {errors.customCpuRequest && (
                  <p className="sm:col-span-2 text-sm text-destructive">
                    {errors.customCpuRequest}
                  </p>
                )}
              </div>
            )}

            <div className="grid grid-cols-1 sm:grid-cols-3 gap-6 mt-6 max-w-3xl">
              <div>
                <Label htmlFor="storage-size" className="text-sm font-medium mb-2 block">
                  Storage size
                </Label>
                <div className="flex items-center gap-2">
                  <Input
                    id="storage-size"
                    type="number"
                    min={1}
                    value={values.storageGB}
                    onChange={(e) => update("storageGB", Number(e.target.value) || 0)}
                    className={errors.storageGB ? "border-destructive" : ""}
                  />
                  <span className="text-sm text-muted-foreground">GB</span>
                </div>
                {errors.storageGB && (
                  <p className="text-sm text-destructive mt-1">{errors.storageGB}</p>
                )}
              </div>

              <div>
                <Label htmlFor="pg-version" className="text-sm font-medium mb-2 block">
                  Version
                </Label>
                <Select
                  value={String(values.versionMajor)}
                  onValueChange={(v) => update("versionMajor", Number(v))}
                >
                  <SelectTrigger id="pg-version">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {POSTGRES_VERSIONS.map((v) => (
                      <SelectItem key={v} value={String(v)}>
                        PG {v} {v === 17 ? "(default)" : ""}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div>
                <Label htmlFor="ha-toggle" className="text-sm font-medium mb-2 block">
                  High availability
                </Label>
                <div className="flex items-center justify-between h-10">
                  <span className="text-sm text-muted-foreground">
                    Replicated across 2 instances
                  </span>
                  <Switch
                    id="ha-toggle"
                    checked={values.highAvailability}
                    onCheckedChange={(c) => update("highAvailability", c)}
                  />
                </div>
              </div>
            </div>

            <Collapsible
              defaultOpen={isEdit && Boolean(values.advancedJson)}
              className="mt-8 -mx-6 -mb-6 border-t"
            >
              <div className="flex items-center gap-3 hover:bg-muted/30 transition-colors">
                <CollapsibleTrigger asChild>
                  <button
                    type="button"
                    className="flex-1 text-left group focus:outline-none px-6 py-4"
                  >
                    <div className="flex items-baseline gap-2 flex-wrap">
                      <span className="text-base font-semibold">Advanced</span>
                      <ChevronDown className="h-4 w-4 text-muted-foreground transition-transform group-data-[state=open]:rotate-180" />
                      {isEdit && advancedDirty && (
                        <span className="text-[10px] uppercase tracking-wide font-medium text-amber-700 dark:text-amber-400 bg-amber-100 dark:bg-amber-950/40 px-1.5 py-0.5 rounded">
                          Modified
                        </span>
                      )}
                    </div>
                  </button>
                </CollapsibleTrigger>
                <a
                  href={ADVANCED_DOCS_URL}
                  target="_blank"
                  rel="noreferrer noopener"
                  className="text-sm text-primary hover:underline inline-flex items-center gap-1 whitespace-nowrap pr-6"
                  onClick={(e) => e.stopPropagation()}
                >
                  Read the documentation
                  <ExternalLink className="h-3.5 w-3.5" />
                </a>
              </div>
              <CollapsibleContent>
                <div className="px-6 pb-6 pt-3 space-y-2">
                  {isEdit && advancedDirty && (
                    <div className="flex items-center justify-between text-xs text-muted-foreground max-w-3xl">
                      <span className="inline-flex items-center gap-1.5">
                        {advancedCleared && (
                          <>
                            <AlertCircle className="h-3.5 w-3.5 text-amber-600 dark:text-amber-400" />
                            <span>
                              These advanced fields will be removed on save.
                            </span>
                          </>
                        )}
                      </span>
                      <button
                        type="button"
                        onClick={handleResetAdvanced}
                        className="text-primary hover:underline font-medium"
                      >
                        Reset to current
                      </button>
                    </div>
                  )}
                  <Textarea
                    id="advanced-json"
                    rows={12}
                    value={values.advancedJson}
                    onChange={(e) => update("advancedJson", e.target.value)}
                    placeholder={'{\n  "configuration": {\n    "parameters": { "max_connections": "200" }\n  }\n}'}
                    className={cn(
                      "font-mono text-xs [field-sizing:fixed] max-w-3xl",
                      errors.advancedJson ? "border-destructive" : "",
                    )}
                    spellCheck={false}
                  />
                  {errors.advancedJson && (
                    <p className="text-sm text-destructive mt-2 whitespace-pre-wrap">
                      {errors.advancedJson}
                    </p>
                  )}
                </div>
              </CollapsibleContent>
            </Collapsible>
          </CardContent>
        </Card>
      </div>

      <Dialog
        open={confirmAdvancedOpen}
        onOpenChange={(o) => !o && !submitting && setConfirmAdvancedOpen(false)}
      >
        <DialogContent className="sm:max-w-[460px]">
          <DialogHeader>
            <DialogTitle>Save advanced changes?</DialogTitle>
            <DialogDescription>
              {advancedCleared
                ? "You're about to remove the advanced fields shown previously. The addon will redeploy without those overrides."
                : "You've edited the advanced configuration. These changes will be applied to the addon."}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => !submitting && setConfirmAdvancedOpen(false)}
              disabled={submitting}
            >
              Cancel
            </Button>
            <Button onClick={() => void persist()} disabled={submitting}>
              {submitting && <Loader2 className="h-4 w-4 animate-spin" />}
              Save changes
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

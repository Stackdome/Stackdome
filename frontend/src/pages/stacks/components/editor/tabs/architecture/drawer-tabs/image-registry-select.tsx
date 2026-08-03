import { useEffect, useState } from "react";
import { ChevronDown, Globe, Package } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import {
  Command,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from "@/components/ui/command";
import { cn } from "@/lib/utils";
import { listRegistryCredentials, type RegistryCredential } from "@/api/registry-credentials";
import { getCurrentOrganizationId } from "@/lib/common";
import { splitImageRef, joinImageRef, dockerHostsEqual } from "@/pages/stacks/lib/image-ref";

export interface ImageSourcePatch {
  ref: string;
  registry_credentials_id: string | undefined;
}

interface ImageRegistrySelectProps {
  id: string;
  /** Full source.image.ref (host included when present). */
  imageRef: string;
  registryCredentialsId?: string;
  onChange: (patch: ImageSourcePatch) => void;
}

const PUBLIC_LABEL = "Public / Docker Hub";

/** Registry picker for the image source: rewrites the ref's host segment and
    keeps registry_credentials_id in sync with the chosen credential. */
export function ImageRegistrySelect({ id, imageRef, registryCredentialsId, onChange }: ImageRegistrySelectProps) {
  const [open, setOpen] = useState(false);
  const [credentials, setCredentials] = useState<RegistryCredential[]>([]);
  const [query, setQuery] = useState("");

  const { host, remainder } = splitImageRef(imageRef);

  useEffect(() => {
    const orgId = getCurrentOrganizationId();
    if (!orgId) return;
    let cancelled = false;
    listRegistryCredentials(orgId)
      .then((res) => {
        if (!cancelled) setCredentials(res.items ?? []);
      })
      .catch(() => {
        // Credential list is an enhancement; free-text hosts still work.
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const matched =
    credentials.find((c) => c.id === registryCredentialsId) ??
    (host ? credentials.find((c) => dockerHostsEqual(c.host, host)) : undefined);

  const pickCredential = (cred: RegistryCredential) => {
    onChange({ ref: joinImageRef(cred.host ?? null, remainder), registry_credentials_id: cred.id });
    setOpen(false);
  };

  const pickPublic = () => {
    onChange({ ref: remainder, registry_credentials_id: undefined });
    setOpen(false);
  };

  const pickCustom = () => {
    onChange({ ref: joinImageRef(query.trim(), remainder), registry_credentials_id: undefined });
    setOpen(false);
  };

  const display = matched?.host ?? host ?? PUBLIC_LABEL;

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          id={id}
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className={cn(
            "h-9 w-full justify-between font-mono text-[12.5px] font-normal",
            !host && !matched && "text-muted-foreground",
          )}
        >
          <span className="truncate">{display}</span>
          <ChevronDown className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[var(--radix-popover-trigger-width)] p-0" align="start">
        <Command shouldFilter={false}>
          <CommandInput placeholder="Registry host…" value={query} onValueChange={setQuery} />
          <CommandList>
            <CommandGroup>
              <CommandItem value="public" onSelect={pickPublic}>
                <Globe className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="text-[12.5px]">{PUBLIC_LABEL}</span>
              </CommandItem>
              {credentials.map((cred) => (
                <CommandItem key={cred.id} value={cred.id!} onSelect={() => pickCredential(cred)}>
                  <Package className="h-3.5 w-3.5 text-muted-foreground" />
                  <span className="flex-1 truncate font-mono text-[12.5px]">{cred.host}</span>
                  <span className="text-[11px] text-muted-foreground">{cred.username}</span>
                </CommandItem>
              ))}
            </CommandGroup>
            {query.trim() && (
              <>
                <CommandSeparator />
                <CommandGroup>
                  <CommandItem value={`custom-${query}`} onSelect={pickCustom}>
                    <span className="truncate text-[12.5px]">Use &quot;{query.trim()}&quot; as registry host</span>
                  </CommandItem>
                </CommandGroup>
              </>
            )}
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}

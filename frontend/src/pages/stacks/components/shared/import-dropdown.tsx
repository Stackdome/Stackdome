import React from 'react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Button, buttonVariants } from '@/components/ui/button';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import type { VariantProps } from 'class-variance-authority';
import { ChevronDown, Import } from 'lucide-react';
import { buildImportOptions, type ImportOption } from './import-options';

interface ImportDropdownProps {
  variant?: VariantProps<typeof buttonVariants>['variant'];
  size?: VariantProps<typeof buttonVariants>['size'];
  children?: React.ReactNode;
  className?: string;
  disabled?: boolean;
  importOptions: ImportOption[];
}

export default function ImportDropdown({
  variant = 'outline',
  size = 'default',
  children,
  className,
  disabled = false,
  importOptions,
}: ImportDropdownProps) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant={variant}
          size={size}
          className={className}
          disabled={disabled}
        >
          {children || (
            <>
              <Import className="h-4 w-4" />
              Import
              <ChevronDown className="ml-2 h-4 w-4" />
            </>
          )}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent
        align="end"
        className="w-56"
        onCloseAutoFocus={(e) => e.preventDefault()}
      >
        <TooltipProvider delayDuration={200}>
          {importOptions.map((option) => {
            const item = (
              <DropdownMenuItem
                onClick={option.onClick}
                disabled={option.disabled}
                className="flex items-center gap-2 cursor-pointer"
              >
                {option.icon}
                <span>{option.label}</span>
              </DropdownMenuItem>
            );
            if (option.disabled) {
              return (
                <Tooltip key={option.id}>
                  <TooltipTrigger asChild>
                    <div>{item}</div>
                  </TooltipTrigger>
                  <TooltipContent side="left">Coming soon</TooltipContent>
                </Tooltip>
              );
            }
            return <React.Fragment key={option.id}>{item}</React.Fragment>;
          })}
        </TooltipProvider>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

// Pre-configured Docker Compose + Templates import dropdown
interface DockerComposeImportDropdownProps {
  onDockerComposeImport: () => void;
  onTemplates: () => void;
  variant?: VariantProps<typeof buttonVariants>['variant'];
  size?: VariantProps<typeof buttonVariants>['size'];
  children?: React.ReactNode;
  className?: string;
  disabled?: boolean;
}

export function DockerComposeImportDropdown({
  onDockerComposeImport,
  onTemplates,
  variant = 'outline',
  size = 'default',
  children,
  className,
  disabled = false,
}: DockerComposeImportDropdownProps) {
  const importOptions = buildImportOptions({
    onDockerCompose: onDockerComposeImport,
    onTemplates,
  });

  return (
    <ImportDropdown
      variant={variant}
      size={size}
      className={className}
      disabled={disabled}
      importOptions={importOptions}
    >
      {children}
    </ImportDropdown>
  );
}

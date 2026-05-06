import React from 'react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Button, buttonVariants } from '@/components/ui/button';
import type { VariantProps } from 'class-variance-authority';
import { ChevronDown, Upload } from 'lucide-react';
import dockerIconUrl from '@/assets/docker.svg';

export interface ImportOption {
  id: string;
  label: string;
  description: string;
  icon: React.ReactNode;
  onClick: () => void;
  disabled?: boolean;
}

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
              <Upload className="h-4 w-4" />
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
        {importOptions.map((option) => (
          <DropdownMenuItem
            key={option.id}
            onClick={option.onClick}
            disabled={option.disabled}
            className="flex items-center gap-2 cursor-pointer"
          >
            {option.icon}
            <span>{option.description}</span>
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

// Pre-configured Docker Compose import dropdown
interface DockerComposeImportDropdownProps {
  onDockerComposeImport: () => void;
  variant?: VariantProps<typeof buttonVariants>['variant'];
  size?: VariantProps<typeof buttonVariants>['size'];
  children?: React.ReactNode;
  className?: string;
  disabled?: boolean;
}

export function DockerComposeImportDropdown({
  onDockerComposeImport,
  variant = 'outline',
  size = 'default',
  children,
  className,
  disabled = false,
}: DockerComposeImportDropdownProps) {
  const importOptions: ImportOption[] = [
    {
      id: 'docker-compose',
      label: 'Docker Compose',
      description: 'from Docker Compose',
      icon: <img src={dockerIconUrl} alt="Docker" className="h-4 w-4" />,
      onClick: onDockerComposeImport,
    },
    // Future import options can be added here:
    // {
    //   id: 'kubernetes',
    //   label: 'Kubernetes YAML',
    //   description: 'Import from Kubernetes manifests',
    //   icon: <KubernetesIcon className="h-4 w-4 text-info" />,
    //   onClick: onKubernetesImport,
    //   disabled: true, // Coming soon
    // },
  ];

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

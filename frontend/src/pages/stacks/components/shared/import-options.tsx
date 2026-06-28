import type { ReactNode } from 'react';
import { LayoutTemplate } from 'lucide-react';
import dockerIconUrl from '@/assets/docker.svg';

export interface ImportOption {
  id: string;
  label: string;
  description: string;
  icon: ReactNode;
  onClick: () => void;
  disabled?: boolean;
}

export function buildImportOptions({
  onDockerCompose,
  onTemplates,
}: {
  onDockerCompose: () => void;
  onTemplates: () => void;
}): ImportOption[] {
  return [
    {
      id: 'templates',
      label: 'Templates',
      description: 'Start from a curated stack template',
      icon: <LayoutTemplate className="h-4 w-4" />,
      onClick: onTemplates,
    },
    {
      id: 'docker-compose',
      label: 'Docker Compose',
      description: 'Import from a docker-compose.yml file',
      icon: <img src={dockerIconUrl} alt="Docker" className="h-4 w-4" />,
      onClick: onDockerCompose,
    },
  ];
}

import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useToast } from '@/components/ui/use-toast';
import { convertDockerComposeToStackData } from '@/lib/docker-compose-converter';
import { parseAndValidateDockerCompose } from '@/lib/docker-compose-parser';
import type { DockerComposeFile } from '@/types/docker-compose';

export interface ImportState {
  isLoading: boolean;
  error: string | null;
  isDialogOpen: boolean;
}

export interface ImportActions {
  openDialog: () => void;
  closeDialog: () => void;
  /** Returns true when the import succeeded and navigation was triggered; false on any error. */
  handleImport: (yamlContent: string) => Promise<boolean>;
  clearError: () => void;
}

export function useDockerComposeImport(): ImportState & ImportActions {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const navigate = useNavigate();
  const { toast } = useToast();

  const openDialog = () => {
    setIsDialogOpen(true);
    setError(null);
  };

  const closeDialog = () => {
    setIsDialogOpen(false);
    setError(null);
  };

  const clearError = () => {
    setError(null);
  };

  const handleImport = async (yamlContent: string): Promise<boolean> => {
    if (!yamlContent.trim()) {
      setError('Please enter a Docker Compose YAML configuration');
      return false;
    }

    setIsLoading(true);
    setError(null);

    try {
      // Parse and validate the Docker Compose YAML
      const dockerComposeData = parseAndValidateDockerCompose(yamlContent);

      // Convert to StackDome form data
      // Type cast is safe here since parseAndValidateDockerCompose ensures the structure matches
      const conversionResult = convertDockerComposeToStackData(dockerComposeData as DockerComposeFile);

      if (!conversionResult.success || !conversionResult.data) {
        const errorMessages = conversionResult.errors?.map(e => e.message).join(', ') || 'Failed to convert Docker Compose file';
        throw new Error(errorMessages);
      }

      setIsDialogOpen(false);
      navigate('/stacks/new', {
        state: {
          seed: {
            name: conversionResult.data.name ?? "",
            labels: conversionResult.data.labels ?? [],
            resources: conversionResult.data.spec?.stack_resources ?? [],
            volumes: conversionResult.data.spec?.volumes ?? [],
            linkedAddonIds: [],
          },
        },
      });

      // Show simple success message
      toast({
        title: 'Import successful',
        description: 'Docker Compose services imported. Please review and configure as needed.',
        variant: 'success',
      });

      return true;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to import Docker Compose file';
      setError(errorMessage);

      toast({
        title: 'Import failed',
        description: errorMessage,
        variant: 'destructive',
      });

      return false;
    } finally {
      setIsLoading(false);
    }
  };

  return {
    isLoading,
    error,
    isDialogOpen,
    openDialog,
    closeDialog,
    handleImport,
    clearError,
  };
}

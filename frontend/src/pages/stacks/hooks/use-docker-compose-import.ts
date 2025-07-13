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
  handleImport: (yamlContent: string) => Promise<void>;
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

  const handleImport = async (yamlContent: string): Promise<void> => {
    if (!yamlContent.trim()) {
      setError('Please enter a Docker Compose YAML configuration');
      return;
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

      // Close dialog and navigate to create page with imported data
      setIsDialogOpen(false);

      navigate('/stacks/create', {
        state: {
          importedData: conversionResult.data,
          importSource: 'docker-compose',
          importWarnings: conversionResult.warnings
        }
      });

      // Show simple success message
      toast({
        title: 'Import Successful',
        description: 'Docker Compose services imported. Please review and configure as needed.',
      });

    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to import Docker Compose file';
      setError(errorMessage);

      toast({
        title: 'Import Failed',
        description: errorMessage,
        variant: 'destructive',
      });
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

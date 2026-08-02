import { useState, useEffect } from 'react';
import { getSecrets, type Secret } from '../../../api/secrets';
import { getCurrentOrganizationId } from '@/lib/common';
import { useToast } from '@/components/ui/use-toast';

export interface UseSecretsReturn {
  secrets: Secret[];
  isLoading: boolean;
}

export function useSecrets(): UseSecretsReturn {
  const [secrets, setSecrets] = useState<Secret[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const { toast } = useToast();

  useEffect(() => {
    const fetchSecrets = async () => {
      try {
        const orgId = getCurrentOrganizationId();
        if (!orgId) {
          setIsLoading(false);
          return;
        }

        const response = await getSecrets(orgId);
        setSecrets(response.items || []);
      } catch (error) {
        console.error('Failed to fetch secrets:', error);
        toast({
          title: 'Failed to load secrets',
          description: 'Unable to fetch available secrets. Secret selection will be disabled.',
          variant: 'destructive',
        });
      } finally {
        setIsLoading(false);
      }
    };

    fetchSecrets();
  }, [toast]);

  return {
    secrets,
    isLoading,
  };
}

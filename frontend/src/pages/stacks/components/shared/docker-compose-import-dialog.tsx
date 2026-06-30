import {
  Dialog,
  DialogContent,
  DialogTitle,
} from '@/components/ui/dialog';
import { DockerComposeImportPanel } from '../wizard/docker-compose-import-panel';

interface DockerComposeImportDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onImport: (yamlContent: string) => Promise<void>;
  isLoading: boolean;
  error: string | null;
  onClearError: () => void;
}

export default function DockerComposeImportDialog({
  open,
  onOpenChange,
  onImport,
  isLoading,
  error,
  onClearError,
}: DockerComposeImportDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-2xl">
        <DialogTitle className="sr-only">Import from Docker Compose</DialogTitle>
        <DockerComposeImportPanel
          onImport={onImport}
          isLoading={isLoading}
          error={error}
          onClearError={onClearError}
          onCancel={() => onOpenChange(false)}
        />
      </DialogContent>
    </Dialog>
  );
}

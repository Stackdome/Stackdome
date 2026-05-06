import React, { useRef, useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { FieldShell } from '@/components/branded';
import { Loader2, Upload } from 'lucide-react';

interface DockerComposeImportDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onImport: (yamlContent: string) => Promise<void>;
  isLoading: boolean;
  error: string | null;
  onClearError: () => void;
}

const SAMPLE_YAML = `version: '3.8'
services:
  web:
    image: nginx:latest
    ports:
      - "8080:80"
    environment:
      - ENV=production
  api:
    image: node:18
    ports:
      - "3000:3000"
    depends_on:
      - db
  db:
    image: postgres:15
    environment:
      - POSTGRES_PASSWORD=secret`;

export default function DockerComposeImportDialog({
  open,
  onOpenChange,
  onImport,
  isLoading,
  error,
  onClearError,
}: DockerComposeImportDialogProps) {
  const [yamlContent, setYamlContent] = useState('');
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleImport = async () => {
    await onImport(yamlContent);
  };

  const handleCancel = () => {
    setYamlContent('');
    onClearError();
    onOpenChange(false);
  };

  const handleContentChange = (value: string) => {
    setYamlContent(value);
    if (error) onClearError();
  };

  const handleFileUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (e) => {
        const content = e.target?.result as string;
        setYamlContent(content);
        if (error) onClearError();
      };
      reader.readAsText(file);
    }
  };

  const isValidYaml = yamlContent.trim().length > 0;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>Import from Docker Compose</DialogTitle>
          <DialogDescription>
            Upload a <span className="font-mono">docker-compose.yml</span> file or paste its contents to scaffold a new stack.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-5">
          <div>
            <Button
              type="button"
              variant="outline"
              onClick={() => fileInputRef.current?.click()}
              disabled={isLoading}
            >
              <Upload className="h-4 w-4" />
              Choose file…
            </Button>
            <input
              ref={fileInputRef}
              type="file"
              accept=".yml,.yaml"
              onChange={handleFileUpload}
              className="hidden"
              disabled={isLoading}
            />
          </div>

          <FieldShell
            label="YAML"
            htmlFor="yaml-content"
            hint="Paste your docker-compose YAML below."
            error={error}
          >
            <Textarea
              id="yaml-content"
              placeholder={SAMPLE_YAML}
              value={yamlContent}
              onChange={(e) => handleContentChange(e.target.value)}
              className="h-96 font-mono text-sm resize-none"
              disabled={isLoading}
            />
          </FieldShell>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={handleCancel} disabled={isLoading}>
            Cancel
          </Button>
          <Button onClick={handleImport} disabled={!isValidYaml || isLoading}>
            {isLoading && <Loader2 className="h-4 w-4 animate-spin" />}
            Import
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

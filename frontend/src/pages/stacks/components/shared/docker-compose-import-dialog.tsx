import React, { useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Upload } from 'lucide-react';

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
  const [yamlContent, setYamlContent] = useState('');

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
      <DialogContent className="max-w-2xl w-full p-0 overflow-hidden min-h-[600px]">
        <DialogHeader className="px-6 pt-6 pb-2">
          <DialogTitle className="text-lg">Create your stack by</DialogTitle>
        </DialogHeader>
        <div className="px-6 pb-2 flex flex-col gap-4 items-center">
          <label
            htmlFor="file-upload"
            className="cursor-pointer flex items-center gap-2 px-4 py-2 border border-dashed border-gray-300 rounded-md hover:border-gray-400 transition-colors bg-muted/40 text-sm font-medium mb-1"
            style={{ width: 'fit-content' }}
          >
            <Upload className="h-4 w-4" />
            Upload <span className="font-mono">docker-compose.yml</span>
            <input
              id="file-upload"
              type="file"
              accept=".yml,.yaml"
              onChange={handleFileUpload}
              className="hidden"
              disabled={isLoading}
            />
          </label>
          <div className="text-xs text-muted-foreground mb-2">or paste below</div>
          <Textarea
            id="yaml-content"
            placeholder={`version: '3.8'
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
      - POSTGRES_PASSWORD=secret`}
            value={yamlContent}
            onChange={e => handleContentChange(e.target.value)}
            className="h-96 font-mono text-sm resize-none bg-background border border-gray-300 rounded-md px-3 py-2 w-full overflow-auto"
            style={{ minHeight: '24rem', maxHeight: '24rem' }}
            disabled={isLoading}
          />
          {error && (
            <div className="text-red-500 text-xs mt-1 px-1 w-full text-left">{error}</div>
          )}
        </div>
        <DialogFooter className="px-6 pb-6 pt-2 flex-row gap-2">
          <Button variant="outline" onClick={handleCancel} disabled={isLoading}>
            Cancel
          </Button>
          <Button onClick={handleImport} disabled={!isValidYaml || isLoading} className="min-w-[100px]">
            Import
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

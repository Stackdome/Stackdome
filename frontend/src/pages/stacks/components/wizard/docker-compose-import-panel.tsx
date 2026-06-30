import React, { useRef, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { FieldShell } from '@/components/branded';
import { Upload } from 'lucide-react';
import { WizardFooter } from './wizard-footer';

export interface DockerComposeImportPanelProps {
  onImport: (yaml: string) => Promise<void>;
  isLoading: boolean;
  error: string | null;
  onClearError: () => void;
  onBack: () => void;
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

export function DockerComposeImportPanel({
  onImport,
  isLoading,
  error,
  onClearError,
  onBack,
}: DockerComposeImportPanelProps) {
  const [yamlContent, setYamlContent] = useState('');
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleImport = async () => {
    await onImport(yamlContent);
  };

  const handleBack = () => {
    setYamlContent('');
    onClearError();
    onBack();
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
    <div className="flex h-full flex-col">
      <div className="scrollbar-hide flex-1 overflow-y-auto p-6">
        <h2 className="text-lg font-semibold leading-none tracking-tight">
          Import from Docker Compose
        </h2>
        <p className="mt-1.5 text-sm text-muted-foreground">
          Upload a <span className="font-mono">docker-compose.yml</span> file or
          paste its contents to scaffold a new stack.
        </p>

        <div className="mt-5 space-y-5">
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
      </div>

      <WizardFooter
        onBack={handleBack}
        onContinue={handleImport}
        continueDisabled={!isValidYaml}
        loading={isLoading}
        hint="Scaffolds a new stack from your compose file."
      />
    </div>
  );
}

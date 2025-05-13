export interface StackFormState {
  name: string;
  description: string;
  region: string;
  template: string;
  repositoryUrl: string;
  yamlConfig: string;
  environment: Record<string, string>;
}

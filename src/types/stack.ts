export interface Port {
  number: number;
  exposeToPublic: boolean;
  isHttp: boolean;
}

export interface Build {
  sourceVolume: string;
  buildContext: string;
  dockerFilePath: string;
}

export interface Volume {
  size: string;
  source?: {
    localDir?: {
      sync: boolean;
      path: string;
    };
  };
}

export interface StackResource {
  image?: string;
  imageRegistry?: string;
  build?: Build;
  dependsOn?: string[];
  volumeMounts?: Record<string, string>;
  environmentVariables?: Record<string, string>;
  envFiles?: string[];
  ports?: Port[];
  command?: string[];
  args?: string[];
}

// Create a composite type that allows both stack resources and a volumes property
export interface StackCompose {
  [resourceName: string]: StackResource | Record<string, Volume> | undefined;
  volumes?: Record<string, Volume>;
}

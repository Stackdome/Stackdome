import type {
  ComposeSpecification,
  Service,
  Volume1,
  Network,
  Secret,
  Config,
} from './docker-compose-generated';

// Re-export main types from generated Docker Compose types
export type {
  ComposeSpecification,
  Service,
  Volume1,
  Network,
  Secret,
  Config,
};

// Main Docker Compose file type
export type DockerComposeFile = ComposeSpecification;

// Convenience type aliases
export type DockerComposeService = Service;
export type DockerComposeVolume = Volume1;
export type DockerComposeNetwork = Network;
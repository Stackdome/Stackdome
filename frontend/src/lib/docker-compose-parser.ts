/**
 * Docker Compose YAML parser with validation and error handling
 * Converts YAML content to validated Docker Compose objects
 */
import { load } from "js-yaml";
import { ZodError } from "zod";
import {
  DockerComposeSchema,
  DockerComposeServiceSchema,
  type DockerComposeFile,
  type DockerComposeService,
  type DockerComposeVolume,
  type DockerComposeNetwork
} from "@/schemas/docker-compose-schema";

export interface ParseError {
  type: 'yaml' | 'validation';
  message: string;
  details?: string;
  path?: string[];
}

export interface ParseResult {
  success: boolean;
  data?: DockerComposeFile;
  errors?: ParseError[];
}

/**
 * Parse and validate a Docker Compose YAML string
 */
export function parseAndValidateDockerCompose(yamlContent: string): DockerComposeFile {
  const result = parseDockerCompose(yamlContent);

  if (!result.success) {
    const primaryError = result.errors?.[0];
    throw new Error(primaryError?.message || "Failed to parse Docker Compose file");
  }

  return result.data!;
}

/**
 * Parse Docker Compose YAML with detailed error reporting
 */
export function parseDockerCompose(yamlContent: string): ParseResult {
  // Step 1: Parse YAML
  let parsedYaml: unknown;

  try {
    parsedYaml = load(yamlContent);
  } catch (error) {
    return {
      success: false,
      errors: [{
        type: 'yaml',
        message: `Invalid YAML syntax: ${(error as Error).message}`,
        details: (error as Error).message,
      }]
    };
  }

  // Step 2: Validate structure
  if (typeof parsedYaml !== 'object' || parsedYaml === null) {
    return {
      success: false,
      errors: [{
        type: 'validation',
        message: "Docker Compose file must be a YAML object",
        details: `Expected object, got ${typeof parsedYaml}`,
      }]
    };
  }

  // Step 3: Validate against schema
  try {
    const validatedData = DockerComposeSchema.parse(parsedYaml);
    return {
      success: true,
      data: validatedData,
    };
  } catch (error) {
    if (error instanceof ZodError) {
      const errors: ParseError[] = error.issues.map(issue => ({
        type: 'validation',
        message: formatValidationMessage(issue),
        details: issue.message,
        path: issue.path.map(String),
      }));

      return {
        success: false,
        errors,
      };
    }

    return {
      success: false,
      errors: [{
        type: 'validation',
        message: `Validation failed: ${(error as Error).message}`,
        details: (error as Error).message,
      }]
    };
  }
}

/**
 * Format validation error messages to be user-friendly
 */
function formatValidationMessage(issue: ZodError['issues'][0]): string {
  const path = issue.path.length > 0 ? issue.path.join('.') : 'root';

  switch (issue.code) {
    case 'invalid_type':
      return `Invalid type at '${path}': expected ${issue.expected}, got ${issue.received}`;

    case 'invalid_enum_value':
      return `Invalid value at '${path}': must be one of [${issue.options.join(', ')}], got '${issue.received}'`;

    case 'too_small':
      if (issue.type === 'string') {
        return `Field '${path}' must be at least ${issue.minimum} characters long`;
      }
      if (issue.type === 'array') {
        return `Field '${path}' must contain at least ${issue.minimum} items`;
      }
      return `Field '${path}' value is too small: minimum ${issue.minimum}`;

    case 'too_big':
      if (issue.type === 'string') {
        return `Field '${path}' must be at most ${issue.maximum} characters long`;
      }
      if (issue.type === 'array') {
        return `Field '${path}' must contain at most ${issue.maximum} items`;
      }
      return `Field '${path}' value is too big: maximum ${issue.maximum}`;

    case 'invalid_string':
      if (issue.validation === 'email') {
        return `Field '${path}' must be a valid email address`;
      }
      if (issue.validation === 'url') {
        return `Field '${path}' must be a valid URL`;
      }
      return `Field '${path}' has invalid format: ${issue.validation}`;

    case 'unrecognized_keys':
      return `Unrecognized properties at '${path}': ${issue.keys.join(', ')}`;

    case 'invalid_union':
      return `Field '${path}' doesn't match any expected format`;

    case 'custom':
      return issue.message;

    default:
      return `Validation error at '${path}': ${issue.message}`;
  }
}

/**
 * Validate individual Docker Compose service
 */
export function validateDockerComposeService(serviceName: string, serviceConfig: unknown): ParseResult {
  try {
    const validatedService = DockerComposeServiceSchema.parse(serviceConfig);

    return {
      success: true,
      data: { services: { [serviceName]: validatedService } } as DockerComposeFile,
    };
  } catch (error) {
    if (error instanceof ZodError) {
      const errors: ParseError[] = error.issues.map(issue => ({
        type: 'validation',
        message: formatValidationMessage(issue),
        details: issue.message,
        path: ['services', serviceName, ...issue.path.map(String)],
      }));

      return {
        success: false,
        errors,
      };
    }

    return {
      success: false,
      errors: [{
        type: 'validation',
        message: `Service validation failed: ${(error as Error).message}`,
        details: (error as Error).message,
      }]
    };
  }
}

/**
 * Extract services from a Docker Compose file
 */
export function extractServices(dockerCompose: DockerComposeFile): Record<string, DockerComposeService> {
  return dockerCompose.services || {};
}

/**
 * Extract volumes from a Docker Compose file
 */
export function extractVolumes(dockerCompose: DockerComposeFile): Record<string, DockerComposeVolume> {
  return dockerCompose.volumes || {};
}

/**
 * Extract networks from a Docker Compose file
 */
export function extractNetworks(dockerCompose: DockerComposeFile): Record<string, DockerComposeNetwork> {
  return dockerCompose.networks || {};
}

/**
 * Get summary information about a Docker Compose file
 */
export function getDockerComposeSummary(dockerCompose: DockerComposeFile) {
  const services = extractServices(dockerCompose);
  const volumes = extractVolumes(dockerCompose);
  const networks = extractNetworks(dockerCompose);

  return {
    name: dockerCompose.name,
    version: dockerCompose.version,
    serviceCount: Object.keys(services).length,
    volumeCount: Object.keys(volumes).length,
    networkCount: Object.keys(networks).length,
    serviceNames: Object.keys(services),
    volumeNames: Object.keys(volumes),
    networkNames: Object.keys(networks),
  };
}

/**
 * Create a sample Docker Compose file for testing
 */
export function createSampleDockerCompose(): string {
  return `
version: '3.8'
name: 'sample-app'

services:
  web:
    image: nginx:alpine
    ports:
      - "80:80"
    depends_on:
      - api
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
    environment:
      - NODE_ENV=production

  api:
    build: 
      context: ./api
      dockerfile: Dockerfile
    ports:
      - "3000:3000"
    depends_on:
      - db
    environment:
      DATABASE_URL: postgres://user:password@db:5432/myapp
      NODE_ENV: production
    volumes:
      - ./api:/app
      - /app/node_modules

  db:
    image: postgres:14
    environment:
      POSTGRES_DB: myapp
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

volumes:
  postgres_data:

networks:
  default:
    driver: bridge
`.trim();
}

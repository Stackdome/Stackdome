/**
 * Docker Compose to StackResource conversion engine
 * Converts validated Docker Compose data to StackDome's StackResource specifications
 */
import type {
  DockerComposeFile,
  DockerComposeService,
  DockerComposeVolume,
} from "@/types/docker-compose";
import type {
  FormEnvVarData,
  FormStackData,
  FormStackResourceData,
  FormVolumeExtendedData,
} from "@/pages/stacks/schemas/form-schema";

export interface ConversionResult {
  success: boolean;
  data?: FormStackData;
  warnings?: ConversionWarning[];
  errors?: ConversionError[];
}

export interface ServiceConversionResult {
  success: boolean;
  data?: FormStackResourceData;
  warnings?: ConversionWarning[];
  errors?: ConversionError[];
}

export interface VolumeConversionResult {
  success: boolean;
  data?: FormVolumeExtendedData;
  warnings?: ConversionWarning[];
  errors?: ConversionError[];
}

export interface ConversionWarning {
  type: 'unsupported' | 'partial' | 'defaults';
  message: string;
  dockerComposeField?: string;
  service?: string;
  volume?: string;
}

export interface ConversionError {
  type: 'conversion' | 'validation';
  message: string;
  service?: string;
  volume?: string;
}

/**
 * Sanitize a Docker Compose name into a valid Kubernetes resource name (RFC 1123 subdomain).
 * Docker Compose permits characters Kubernetes rejects — most commonly '_' in volume names
 * like "postgres_data". An invalid name makes the Volume CR rejected by the API server, which
 * stalls the entire stack rollout. Lowercase, replace unsupported runs with '-', and trim so
 * the result starts and ends on an alphanumeric character.
 */
export function sanitizeKubernetesName(name: string): string {
  return name
    .toLowerCase()
    .replace(/[^a-z0-9.-]+/g, "-")
    .replace(/^[^a-z0-9]+/, "")
    .replace(/[^a-z0-9]+$/, "");
}

/**
 * Convert Docker Compose file to StackResource format
 */
export function convertDockerComposeToStackData(
  compose: DockerComposeFile
): ConversionResult {
  try {
    const warnings: ConversionWarning[] = [];
    const errors: ConversionError[] = [];

    // Convert services to stack resources
    const services = compose.services || {};
    const stackResources: FormStackResourceData[] = [];

    const serviceNames = Object.keys(services);

    for (const [serviceName, serviceConfig] of Object.entries(services)) {
      const conversionResult = convertServiceToStackResource(serviceName, serviceConfig, serviceNames);

      if (conversionResult.success && conversionResult.data) {
        stackResources.push(conversionResult.data);
        warnings.push(...(conversionResult.warnings || []));
      } else {
        errors.push(...(conversionResult.errors || []));
      }
    }

    // Convert volumes
    const volumes = compose.volumes || {};
    const stackVolumes: FormVolumeExtendedData[] = [];

    for (const [volumeName, volumeConfig] of Object.entries(volumes)) {
      const conversionResult = convertVolumeToStackVolume(volumeName, volumeConfig);

      if (conversionResult.success && conversionResult.data) {
        stackVolumes.push(conversionResult.data);
        warnings.push(...(conversionResult.warnings || []));
      } else {
        errors.push(...(conversionResult.errors || []));
      }
    }

    // Handle unsupported top-level features
    if (compose.networks && Object.keys(compose.networks).length > 0) {
      warnings.push({
        type: 'unsupported',
        message: 'Docker Compose networks are not supported in StackDome. Services will use default networking.',
        dockerComposeField: 'networks',
      });
    }

    if (compose.secrets && Object.keys(compose.secrets).length > 0) {
      warnings.push({
        type: 'unsupported',
        message: 'Docker Compose secrets must be manually configured in StackDome secrets management.',
        dockerComposeField: 'secrets',
      });
    }

    if (compose.configs && Object.keys(compose.configs).length > 0) {
      warnings.push({
        type: 'unsupported',
        message: 'Docker Compose configs are not directly supported. Consider using environment variables or volume mounts.',
        dockerComposeField: 'configs',
      });
    }

    // If no services were successfully converted, that's an error
    if (stackResources.length === 0) {
      errors.push({
        type: 'conversion',
        message: 'No valid services found to convert to StackResources',
      });

      return {
        success: false,
        errors,
        warnings,
      };
    }

    const result: FormStackData = {
      name: compose.name || "imported-stack",
      labels: [],
      spec: {
        stack_resources: stackResources,
        volumes: stackVolumes.length > 0 ? stackVolumes : undefined,
      },
    };

    return {
      success: true,
      data: result,
      warnings,
      errors: errors.length > 0 ? errors : undefined,
    };

  } catch (error) {
    return {
      success: false,
      errors: [{
        type: 'conversion',
        message: `Conversion failed: ${(error as Error).message}`,
      }],
    };
  }
}

/**
 * Convert single Docker Compose service to StackResource
 */
export function convertServiceToStackResource(
  serviceName: string,
  service: DockerComposeService,
  allServiceNames: string[] = []
): ServiceConversionResult {
  const warnings: ConversionWarning[] = [];
  const errors: ConversionError[] = [];

  try {
    // Basic resource structure
    const resource: FormStackResourceData = {
      name: serviceName,
      workload_type: "Service",
      sourceType: service.build ? "git" : "image",
      labels: convertLabels(service.labels, warnings),
      depends_on: Array.isArray(service.depends_on) ? service.depends_on :
        typeof service.depends_on === 'object' && service.depends_on !== null ?
          Object.keys(service.depends_on) : [],
      ports: convertPorts(service.ports, serviceName, warnings, service.expose),
      volume_mounts: convertVolumeMounts(service.volumes, serviceName, warnings),
      execution_config: {
        environment_variables: convertEnvironmentVariables(service.environment, serviceName, allServiceNames, warnings),
        command: convertCommand(service.command, warnings),
      },
      // UI-specific fields
      gitRevisionType: undefined,
      gitRevisionValue: undefined,
    };

    // Handle image vs build specification
    if (service.build) {
      resource.source = convertGitSource(service.build, serviceName, warnings);
    } else if (service.image) {
      resource.source = { image: { ref: service.image } };
    } else {
      errors.push({
        type: 'conversion',
        message: 'Service must have either image or build specification',
        service: serviceName,
      });
      return { success: false, errors };
    }

    // Handle unsupported service features
    handleUnsupportedServiceFeatures(service, serviceName, warnings);

    return {
      success: true,
      data: resource,
      warnings,
    };

  } catch (error) {
    return {
      success: false,
      errors: [{
        type: 'conversion',
        message: `Failed to convert service '${serviceName}': ${(error as Error).message}`,
        service: serviceName,
      }],
    };
  }
}

/**
 * Convert Docker Compose volume to StackVolume
 */
export function convertVolumeToStackVolume(
  volumeName: string,
  volume: DockerComposeVolume
): VolumeConversionResult {
  const warnings: ConversionWarning[] = [];

  try {
    const stackVolume: FormVolumeExtendedData = {
      name: sanitizeKubernetesName(volumeName),
      sourceType: "None", // User must configure manually
      labels: [],
      spec: {
        size: "1Gi", // Default size
        access_mode: "ReadWriteOnce",
        needs_sync_before_use: false,
      },
    };

    // Handle external volumes
    if (volume && typeof volume === 'object' && volume.external) {
      warnings.push({
        type: 'partial',
        message: 'External volume detected. You need to manually configure the volume source in StackDome.',
        volume: volumeName,
        dockerComposeField: 'external',
      });
    }

    // Handle volume driver
    if (volume && typeof volume === 'object' && volume.driver && volume.driver !== 'local') {
      warnings.push({
        type: 'unsupported',
        message: `Volume driver '${volume.driver}' is not supported. Using default storage.`,
        volume: volumeName,
        dockerComposeField: 'driver',
      });
    }

    warnings.push({
      type: 'defaults',
      message: `Volume '${volumeName}' created with default settings. Please configure size, access mode, and source as needed.`,
      volume: volumeName,
    });

    return {
      success: true,
      data: stackVolume,
      warnings,
    };

  } catch (error) {
    return {
      success: false,
      errors: [{
        type: 'conversion',
        message: `Failed to convert volume '${volumeName}': ${(error as Error).message}`,
        volume: volumeName,
      }],
    };
  }
}

/**
 * Convert Docker Compose labels
 */
function convertLabels(
  labels: DockerComposeService['labels'],
  warnings: ConversionWarning[]
): Array<{ key: string; value: string }> {
  if (!labels) return [];

  try {
    if (Array.isArray(labels)) {
      // Handle array format: ["key=value", "key2=value2"]
      return labels.map(label => {
        const [key, ...valueParts] = label.split('=');
        return {
          key: key || '',
          value: valueParts.join('=') || '',
        };
      });
    } else if (typeof labels === 'object') {
      // Handle object format: { key: "value", key2: "value2" }
      return Object.entries(labels).map(([key, value]) => ({
        key,
        value: String(value),
      }));
    }

    return [];
  } catch {
    warnings.push({
      type: 'partial',
      message: 'Failed to convert some labels. Please review and add them manually.',
      dockerComposeField: 'labels',
    });
    return [];
  }
}

/**
 * Convert Docker Compose ports to StackResource port specifications
 * The internal port (container port) is what gets exposed to public, not the external host port
 */
function convertPorts(
  ports: DockerComposeService['ports'],
  serviceName: string,
  warnings: ConversionWarning[],
  expose?: DockerComposeService['expose']
): Array<{ name: string; number: number; protocol: 'tcp' | 'http'; exposed_to_public: boolean }> {
  const convertedPorts: Array<{ name: string; number: number; protocol: 'tcp' | 'http'; exposed_to_public: boolean }> = [];

  // Compose `expose` lists container ports reachable by sibling services but
  // never published on the host — map them to internal (non-public) ports.
  (expose ?? []).forEach((entry, index) => {
    const port = parseInt(String(entry).split('/')[0], 10);
    if (isNaN(port)) {
      warnings.push({
        type: 'partial',
        message: `Invalid expose entry at index ${index}: ${entry}`,
        service: serviceName,
        dockerComposeField: 'expose',
      });
      return;
    }
    convertedPorts.push({
      name: `tcp-${port}`,
      number: port,
      protocol: 'tcp',
      exposed_to_public: false,
    });
  });

  if (!ports || !Array.isArray(ports)) return convertedPorts;

  ports.forEach((port, index) => {
    try {
      if (typeof port === 'string') {
        // Parse string formats like "80:80", "8080:3000", "127.0.0.1:8080:80"
        const parts = port.split(':');
        let internalPort: number;
        const protocol = 'http'; // Always use lowercase http

        if (parts.length === 1) {
          // Format: "8080" - port is both external and internal
          internalPort = parseInt(parts[0], 10);
        } else if (parts.length === 2) {
          // Format: "8080:3000" - use internal port (3000), external port is just Docker host mapping
          internalPort = parseInt(parts[1], 10);
        } else if (parts.length === 3) {
          // Format: "127.0.0.1:8080:3000" - use internal port (3000)
          internalPort = parseInt(parts[2], 10);
          warnings.push({
            type: 'partial',
            message: `Port binding with host IP (${parts[0]}) is not supported. Using default binding.`,
            service: serviceName,
            dockerComposeField: 'ports',
          });
        } else {
          throw new Error(`Invalid port format: ${port}`);
        }

        if (isNaN(internalPort)) {
          throw new Error(`Invalid port number: ${port}`);
        }

        convertedPorts.push({
          name: `${protocol}-${internalPort}`,
          number: internalPort,
          protocol,
          exposed_to_public: true,
        });

      } else if (typeof port === 'number') {
        // Direct port number - treat as both external and internal
        convertedPorts.push({
          name: `http-${port}`,
          number: port,
          protocol: 'http',
          exposed_to_public: true,
        });

      } else if (typeof port === 'object' && port !== null) {
        // Object format (long syntax) - use target port (internal) for public exposure
        const internalPort = typeof port.target === 'number'
          ? port.target
          : parseInt(String(port.target), 10);

        if (isNaN(internalPort)) {
          throw new Error(`Invalid port number in object format`);
        }

        convertedPorts.push({
          name: `http-${internalPort}`,
          number: internalPort,
          protocol: 'http', // Always use http
          exposed_to_public: true,
        });

        if (port.mode && port.mode !== 'ingress') {
          warnings.push({
            type: 'unsupported',
            message: `Port mode '${port.mode}' is not supported. Using default ingress mode.`,
            service: serviceName,
            dockerComposeField: 'ports',
          });
        }
      }

    } catch (error) {
      warnings.push({
        type: 'partial',
        message: `Failed to convert port at index ${index}: ${(error as Error).message}`,
        service: serviceName,
        dockerComposeField: 'ports',
      });
    }
  });

  return convertedPorts;
}

/**
 * Convert Docker Compose volumes to volume mounts
 */
function convertVolumeMounts(
  volumes: DockerComposeService['volumes'],
  serviceName: string,
  warnings: ConversionWarning[]
): Array<{ source_volume_name: string; target_path: string }> {
  if (!volumes || !Array.isArray(volumes)) return [];

  const volumeMounts: Array<{ source_volume_name: string; target_path: string }> = [];

  volumes.forEach((volume, index) => {
    try {
      if (typeof volume === 'string') {
        // Parse string formats like "volume:path", "/host/path:/container/path", "./relative:/container/path:ro"
        const parts = volume.split(':');

        if (parts.length >= 2) {
          const source = parts[0];
          const target = parts[1];
          const options = parts[2];

          // Check if this is a named volume (doesn't start with . or /)
          if (!source.startsWith('/') && !source.startsWith('./') && !source.startsWith('../')) {
            // Named volume - this maps to StackDome volumes
            volumeMounts.push({
              source_volume_name: sanitizeKubernetesName(source),
              target_path: target,
            });

            // Add a warning about read-only mounts since StackDome doesn't support them
            if (options?.includes('ro')) {
              warnings.push({
                type: 'unsupported',
                message: `Read-only volume mount option ':ro' is not supported in StackDome. Volume '${source}' will be mounted as read-write.`,
                service: serviceName,
                dockerComposeField: 'volumes',
              });
            }
          } else {
            // Host path or relative path - not directly supported
            warnings.push({
              type: 'unsupported',
              message: `Host path volume mount '${source}' is not supported. Consider using a StackDome volume instead.`,
              service: serviceName,
              dockerComposeField: 'volumes',
            });
          }
        } else {
          warnings.push({
            type: 'partial',
            message: `Invalid volume format at index ${index}: ${volume}`,
            service: serviceName,
            dockerComposeField: 'volumes',
          });
        }

      } else if (typeof volume === 'object' && volume !== null) {
        // Object format (long syntax)
        if (volume.source && volume.target) {
          if (volume.type === 'volume') {
            // Named volume
            volumeMounts.push({
              source_volume_name: sanitizeKubernetesName(volume.source),
              target_path: volume.target,
            });

            // Add a warning about read-only mounts since StackDome doesn't support them
            if (volume.read_only === true) {
              warnings.push({
                type: 'unsupported',
                message: `Read-only volume mount option is not supported in StackDome. Volume '${volume.source}' will be mounted as read-write.`,
                service: serviceName,
                dockerComposeField: 'volumes',
              });
            }
          } else if (volume.type === 'bind') {
            warnings.push({
              type: 'unsupported',
              message: `Bind mount '${volume.source}' is not supported. Consider using a StackDome volume instead.`,
              service: serviceName,
              dockerComposeField: 'volumes',
            });
          } else if (volume.type === 'tmpfs') {
            warnings.push({
              type: 'unsupported',
              message: `tmpfs volume '${volume.source}' is not supported.`,
              service: serviceName,
              dockerComposeField: 'volumes',
            });
          }
        }
      }

    } catch (error) {
      warnings.push({
        type: 'partial',
        message: `Failed to convert volume at index ${index}: ${(error as Error).message}`,
        service: serviceName,
        dockerComposeField: 'volumes',
      });
    }
  });

  return volumeMounts;
}

/**
 * A whole-value env reference: "${self.public_url}" or "${<service>.<output>}".
 * Output keys follow pkg/models/output_descriptor.go (host, port, url,
 * public_host, public_url, optionally suffixed ".<port-name>").
 */
const ENV_REFERENCE_PATTERN = /^\$\{([a-z0-9][a-z0-9_-]*)\.([a-z0-9][a-z0-9_.-]*)\}$/;

/**
 * Classify one env value: reference values become self/resource rows resolved
 * at deploy; everything else stays a literal row. References to unknown
 * services are kept literal with a warning rather than dropped.
 */
function toEnvRow(
  name: string,
  value: string,
  serviceName: string,
  allServiceNames: string[],
  warnings: ConversionWarning[]
): FormEnvVarData {
  const match = ENV_REFERENCE_PATTERN.exec(value);
  if (!match) return { from: "stack", name, value };

  const [, target, output] = match;
  if (target === "self" || target === serviceName) {
    return { from: "self", name, selfOutput: output };
  }
  if (allServiceNames.includes(target)) {
    return { from: "resource", name, resourceName: target, output };
  }

  warnings.push({
    type: 'partial',
    message: `Environment variable '${name}' references unknown service '${target}'. Kept as literal text.`,
    service: serviceName,
    dockerComposeField: 'environment',
  });
  return { from: "stack", name, value };
}

/**
 * Convert Docker Compose environment variables
 */
function convertEnvironmentVariables(
  environment: DockerComposeService['environment'],
  serviceName: string,
  allServiceNames: string[],
  warnings: ConversionWarning[]
): FormEnvVarData[] {
  if (!environment) return [];

  const envVars: FormEnvVarData[] = [];

  try {
    if (Array.isArray(environment)) {
      // Handle array format: ["KEY=value", "KEY2=value2"]
      environment.forEach((envVar, index) => {
        try {
          if (typeof envVar === 'string') {
            const [name, ...valueParts] = envVar.split('=');
            if (name) {
              envVars.push(toEnvRow(name, valueParts.join('=') || '', serviceName, allServiceNames, warnings));
            }
          }
        } catch {
          warnings.push({
            type: 'partial',
            message: `Failed to convert environment variable at index ${index}`,
            service: serviceName,
            dockerComposeField: 'environment',
          });
        }
      });

    } else if (typeof environment === 'object' && environment !== null) {
      // Handle object format: { KEY: "value", KEY2: "value2" }
      Object.entries(environment).forEach(([name, value]) => {
        envVars.push(toEnvRow(name, String(value ?? ''), serviceName, allServiceNames, warnings));
      });
    }

  } catch {
    warnings.push({
      type: 'partial',
      message: 'Failed to convert some environment variables. Please review and add them manually.',
      service: serviceName,
      dockerComposeField: 'environment',
    });
  }

  // Add warning about secrets
  if (envVars.length > 0) {
    warnings.push({
      type: 'defaults',
      message: 'Environment variables imported as plain text. Consider using StackDome secrets for sensitive values.',
      service: serviceName,
      dockerComposeField: 'environment',
    });
  }

  return envVars;
}

/**
 * Convert Docker Compose command
 */
function convertCommand(
  command: DockerComposeService['command'],
  warnings: ConversionWarning[]
): string[] {
  if (!command) return [];

  if (typeof command === 'string') {
    // Split string command into array
    return command.split(' ').filter(part => part.length > 0);
  } else if (Array.isArray(command)) {
    return command;
  } else {
    warnings.push({
      type: 'partial',
      message: 'Invalid command format. Please review and set manually.',
      dockerComposeField: 'command',
    });
    return [];
  }
}

/**
 * Convert Docker Compose build specification
 */
function convertGitSource(
  build: DockerComposeService['build'],
  serviceName: string,
  warnings: ConversionWarning[]
): FormStackResourceData['source'] {
  if (!build) return undefined;

  try {
    const buildContext = typeof build === 'string' ? (build || '.') : (build.context || '.');
    const dockerfile = typeof build === 'string' ? 'Dockerfile' : (build.dockerfile || 'Dockerfile');

    if (typeof build === 'object' && build !== null) {
      // Handle build args
      if (build.args && Object.keys(build.args).length > 0) {
        warnings.push({
          type: 'unsupported',
          message: 'Build args are not directly supported. Consider using environment variables or configuring them in your Dockerfile.',
          service: serviceName,
          dockerComposeField: 'build.args',
        });
      }

      // Handle target
      if (build.target) {
        warnings.push({
          type: 'unsupported',
          message: `Multi-stage build target '${build.target}' must be configured in your Dockerfile.`,
          service: serviceName,
          dockerComposeField: 'build.target',
        });
      }
    }

    warnings.push({
      type: 'defaults',
      message: 'Build specification imported with a placeholder. You must configure the Git repository URL.',
      service: serviceName,
      dockerComposeField: 'build',
    });

    return {
      git: {
        repo_url: '', // User must provide
        dockerfile_path: dockerfile,
        build_context: buildContext,
      },
    };
  } catch (error) {
    warnings.push({
      type: 'partial',
      message: `Failed to convert build specification: ${(error as Error).message}`,
      service: serviceName,
      dockerComposeField: 'build',
    });
  }

  return undefined;
}

/**
 * Handle unsupported Docker Compose service features
 */
function handleUnsupportedServiceFeatures(
  service: DockerComposeService,
  serviceName: string,
  warnings: ConversionWarning[]
): void {
  // Check for unsupported features
  const unsupportedFeatures: Array<{ field: keyof DockerComposeService; message: string }> = [
    { field: 'networks', message: 'Service networks are not supported. Services will use default networking.' },
    { field: 'restart', message: 'Restart policies are managed by StackDome automatically.' },
    { field: 'privileged', message: 'Privileged mode is not supported for security reasons.' },
    { field: 'user', message: 'User specification is not supported. Use container defaults.' },
    { field: 'working_dir', message: 'Working directory should be configured in your container image.' },
    { field: 'entrypoint', message: 'Entrypoint should be configured in your container image or Dockerfile.' },
    { field: 'healthcheck', message: 'Health checks are managed by StackDome automatically.' },
    { field: 'logging', message: 'Logging configuration is managed by StackDome automatically.' },
    { field: 'ulimits', message: 'Resource limits should be configured in StackDome resource management.' },
    { field: 'sysctls', message: 'System controls are not supported for security reasons.' },
    { field: 'cap_add', message: 'Capabilities are not supported for security reasons.' },
    { field: 'cap_drop', message: 'Capabilities are not supported for security reasons.' },
    { field: 'security_opt', message: 'Security options are managed by StackDome automatically.' },
    { field: 'tmpfs', message: 'tmpfs mounts are not supported. Use regular volumes.' },
    { field: 'external_links', message: 'External links are not supported. Use service discovery.' },
    { field: 'links', message: 'Links are deprecated. Use service names for inter-service communication.' },
  ];

  unsupportedFeatures.forEach(({ field, message }) => {
    if (service[field] !== undefined && service[field] !== null) {
      warnings.push({
        type: 'unsupported',
        message,
        service: serviceName,
        dockerComposeField: field as string,
      });
    }
  });

  // Special handling for some features
  if (service.deploy) {
    warnings.push({
      type: 'unsupported',
      message: 'Deploy configuration is not supported. StackDome manages deployment automatically.',
      service: serviceName,
      dockerComposeField: 'deploy',
    });
  }

  if (service.profiles && service.profiles.length > 0) {
    warnings.push({
      type: 'unsupported',
      message: 'Profiles are not supported. Consider using different stack configurations.',
      service: serviceName,
      dockerComposeField: 'profiles',
    });
  }
}

/**
 * Get conversion summary
 */
export function getConversionSummary(result: ConversionResult): {
  servicesConverted: number;
  volumesConverted: number;
  warningCount: number;
  errorCount: number;
  requiresManualConfiguration: boolean;
} {
  const servicesConverted = result.data?.spec.stack_resources?.length || 0;
  const volumesConverted = result.data?.spec.volumes?.length || 0;
  const warningCount = result.warnings?.length || 0;
  const errorCount = result.errors?.length || 0;

  // Check if manual configuration is required
  const requiresManualConfiguration = (result.warnings || []).some(warning =>
    warning.type === 'defaults' ||
    warning.dockerComposeField === 'build' ||
    warning.message.includes('must configure') ||
    warning.message.includes('placeholder')
  );

  return {
    servicesConverted,
    volumesConverted,
    warningCount,
    errorCount,
    requiresManualConfiguration,
  };
}

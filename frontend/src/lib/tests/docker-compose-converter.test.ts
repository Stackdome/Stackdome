/**
 * Unit tests for Docker Compose to StackResource conversion engine
 * Uses simple fixture file imports for comprehensive conversion testing
 */
import { describe, it, expect } from 'vitest';
import { readFileSync } from 'fs';
import { join } from 'path';
import {
  convertDockerComposeToStackData,
  convertServiceToStackResource,
  convertVolumeToStackVolume,
  getConversionSummary,
  type ConversionResult,
} from '../docker-compose-converter';
import { parseDockerCompose } from '../docker-compose-parser';
import type { DockerComposeFile, DockerComposeService } from '@/types/docker-compose';
import type { FormStackResourceData, FormVolumeExtendedData } from '@/pages/stacks/schemas/form-schema';

// Simple fixture loaders
const simpleComposeYaml = readFileSync(join(__dirname, '__fixtures__/simple-docker-compose.yml'), 'utf-8');
const complexComposeYaml = readFileSync(join(__dirname, '__fixtures__/complex-docker-compose.yml'), 'utf-8');
const invalidComposeYaml = readFileSync(join(__dirname, '__fixtures__/invalid-docker-compose.yml'), 'utf-8');

// Test helper to create minimal stack resource that matches the actual form schema
function createMockStackResource(name: string): FormStackResourceData {
  return {
    name,
    sourceType: 'image' as const,
    labels: [],
    depends_on: [],
    ports: [],
    volume_mounts: [],
    execution_config: {
      command: [],
      args: [],
      environment_variables: [],
    },
    gitRevisionType: undefined,
    gitRevisionValue: undefined,
    useImageSecret: false,
    selectedImageSecretId: undefined,
    useGitSecret: false,
    selectedGitSecretId: undefined,
    image_spec: { image: 'test:latest' },
    build_spec: undefined,
  };
}

// Test helper to create minimal volume that matches the actual form schema
function createMockVolume(name: string): FormVolumeExtendedData {
  return {
    name,
    sourceType: 'None' as const,
    labels: [],
    spec: {
      size: '10Gi',
      storage_class: 'default',
      needs_sync_before_use: false,
      access_mode: 'ReadWriteOnce' as const,
    },
  };
}

describe('convertDockerComposeToStackData', () => {
  it('should convert simple Docker Compose file with image-based service', () => {
    const compose: DockerComposeFile = {
      version: '3.8',
      services: {
        'web-app': {
          image: 'nginx:latest',
          ports: ['80:80'],
          environment: {
            NODE_ENV: 'production',
          },
        },
      },
    };

    const result = convertDockerComposeToStackData(compose);

    expect(result.success).toBe(true);
    expect(result.data).toBeDefined();
    expect(result.data!.name).toBe('imported-stack');
    expect(result.data!.spec.stack_resources).toHaveLength(1);

    const resource = result.data!.spec.stack_resources[0];
    expect(resource.name).toBe('web-app');
    expect(resource.sourceType).toBe('image');
    expect(resource.image_spec?.image).toBe('nginx:latest');
    expect(resource.ports).toHaveLength(1);
    expect(resource.ports[0]).toEqual({
      name: 'http-80',
      number: 80,
      protocol: 'http',
      exposed_to_public: true,
    });
    expect(resource.execution_config.environment_variables).toHaveLength(1);
    expect(resource.execution_config.environment_variables[0]).toEqual({
      name: 'NODE_ENV',
      value: 'production',
      from: 'stack',
    });
  });

  it('should convert Docker Compose file with build-based service', () => {
    const compose: DockerComposeFile = {
      version: '3.8',
      services: {
        'api-server': {
          build: {
            context: './api',
            dockerfile: 'Dockerfile.prod',
            args: {
              NODE_VERSION: '18',
            },
          },
          command: ['npm', 'start'],
          environment: ['PORT=3000', 'DEBUG=true'],
        },
      },
    };

    const result = convertDockerComposeToStackData(compose);

    expect(result.success).toBe(true);
    expect(result.data).toBeDefined();

    const resource = result.data!.spec.stack_resources[0];
    expect(resource.name).toBe('api-server');
    expect(resource.sourceType).toBe('git');
    expect(resource.build_spec).toBeDefined();
    expect(resource.build_spec!.context_path_within_source).toBe('./api');
    expect(resource.build_spec!.dockerfile_path).toBe('Dockerfile.prod');
    expect(resource.execution_config.command).toEqual(['npm', 'start']);
    expect(resource.execution_config.environment_variables).toHaveLength(2);

    // Should have warnings about build args and placeholders
    expect(result.warnings).toBeDefined();
    const buildWarnings = result.warnings!.filter((w) =>
      w.dockerComposeField === 'build.args' || w.dockerComposeField === 'build'
    );
    expect(buildWarnings.length).toBeGreaterThan(0);
  });

  it('should convert volumes correctly', () => {
    const compose: DockerComposeFile = {
      version: '3.8',
      services: {
        database: {
          image: 'postgres:13',
          volumes: ['db-data:/var/lib/postgresql/data'],
        },
      },
      volumes: {
        'db-data': {
          driver: 'local',
        },
      },
    };

    const result = convertDockerComposeToStackData(compose);

    expect(result.success).toBe(true);
    expect(result.data!.spec.volumes).toHaveLength(1);
    expect(result.data!.spec.volumes![0]).toEqual({
      name: 'db-data',
      sourceType: 'None',
      labels: [],
      spec: {
        size: '1Gi',
        access_mode: 'ReadWriteOnce',
        needs_sync_before_use: false,
      },
    });

    const resource = result.data!.spec.stack_resources[0];
    expect(resource.volume_mounts).toHaveLength(1);
    expect(resource.volume_mounts[0]).toEqual({
      source_volume_name: 'db-data',
      target_path: '/var/lib/postgresql/data',
    });
  });

  it('should handle service dependencies', () => {
    const compose: DockerComposeFile = {
      version: '3.8',
      services: {
        database: {
          image: 'postgres:13',
        },
        'web-app': {
          image: 'nginx:latest',
          depends_on: ['database'],
        },
        api: {
          image: 'node:18',
          depends_on: {
            database: {
              condition: 'service_healthy',
            },
          },
        },
      },
    };

    const result = convertDockerComposeToStackData(compose);

    expect(result.success).toBe(true);
    expect(result.data!.spec.stack_resources).toHaveLength(3);

    const webApp = result.data!.spec.stack_resources.find((r: { name: string }) => r.name === 'web-app')!;
    expect(webApp.depends_on).toEqual(['database']);

    const api = result.data!.spec.stack_resources.find((r: { name: string }) => r.name === 'api')!;
    expect(api.depends_on).toEqual(['database']);
  });

  it('should handle unsupported features with warnings', () => {
    const compose: DockerComposeFile = {
      version: '3.8',
      services: {
        'web-app': {
          image: 'nginx:latest',
          restart: 'always',
          privileged: true,
          networks: ['custom-network'],
        },
      },
      networks: {
        'custom-network': {
          driver: 'bridge',
        },
      },
      secrets: {
        'db-password': {
          file: './db_password.txt',
        },
      },
    };

    const result = convertDockerComposeToStackData(compose);

    expect(result.success).toBe(true);
    expect(result.warnings).toBeDefined();
    expect(result.warnings!.length).toBeGreaterThan(0);

    // Check for specific warnings
    const networkWarning = result.warnings!.find((w) => w.dockerComposeField === 'networks');
    expect(networkWarning).toBeDefined();

    const secretsWarning = result.warnings!.find((w) => w.dockerComposeField === 'secrets');
    expect(secretsWarning).toBeDefined();

    const restartWarning = result.warnings!.find((w) => w.dockerComposeField === 'restart');
    expect(restartWarning).toBeDefined();
  });

  it('should fail when no valid services are found', () => {
    const compose: DockerComposeFile = {
      version: '3.8',
      services: {
        'invalid-service': {
          // No image or build specified
        } as DockerComposeService,
      },
    };

    const result = convertDockerComposeToStackData(compose);

    expect(result.success).toBe(false);
    expect(result.errors).toBeDefined();
    expect(result.errors!.some((e) => e.message.includes('No valid services found'))).toBe(true);
  });

  it('should handle empty compose file', () => {
    const compose: DockerComposeFile = {
      version: '3.8',
    };

    const result = convertDockerComposeToStackData(compose);

    expect(result.success).toBe(false);
    expect(result.errors).toBeDefined();
    expect(result.errors!.some((e) => e.message.includes('No valid services found'))).toBe(true);
  });

  it('should use compose name when provided', () => {
    const compose: DockerComposeFile = {
      version: '3.8',
      name: 'my-awesome-stack',
      services: {
        app: {
          image: 'nginx:latest',
        },
      },
    };

    const result = convertDockerComposeToStackData(compose);

    expect(result.success).toBe(true);
    expect(result.data!.name).toBe('my-awesome-stack');
  });
});

describe('convertServiceToStackResource', () => {
  it('should convert service with image specification', () => {
    const service: DockerComposeService = {
      image: 'redis:6-alpine',
      ports: ['6379:6379'],
      environment: {
        REDIS_PASSWORD: 'secret123',
      },
      labels: {
        'app.version': '1.0.0',
        'environment': 'production',
      },
    };

    const result = convertServiceToStackResource('cache', service);

    expect(result.success).toBe(true);
    expect(result.data).toBeDefined();
    expect(result.data!.name).toBe('cache');
    expect(result.data!.sourceType).toBe('image');
    expect(result.data!.image_spec?.image).toBe('redis:6-alpine');
    expect(result.data!.labels).toEqual([
      { key: 'app.version', value: '1.0.0' },
      { key: 'environment', value: 'production' },
    ]);
  });

  it('should convert service with build specification', () => {
    const service: DockerComposeService = {
      build: './backend',
      environment: ['NODE_ENV=production'],
      command: 'npm run start:prod',
    };

    const result = convertServiceToStackResource('backend', service);

    expect(result.success).toBe(true);
    expect(result.data!.sourceType).toBe('git');
    expect(result.data!.build_spec).toBeDefined();
    expect(result.data!.build_spec!.context_path_within_source).toBe('./backend');
    expect(result.data!.execution_config.command).toEqual(['npm', 'run', 'start:prod']);
  });

  it('should fail when neither image nor build is specified', () => {
    const service: DockerComposeService = {
      ports: ['8080:80'],
    };

    const result = convertServiceToStackResource('invalid', service);

    expect(result.success).toBe(false);
    expect(result.errors).toBeDefined();
    expect(result.errors![0].message).toContain('must have either image or build specification');
  });

  it('should convert complex port specifications', () => {
    const service: DockerComposeService = {
      image: 'nginx:latest',
      ports: [
        '80:80',
        '443:443/tcp',
        '8080-8090:8080-8090',
        {
          target: 3000,
          published: 3000,
          protocol: 'tcp',
        },
      ],
    };

    const result = convertServiceToStackResource('web', service);

    expect(result.success).toBe(true);
    expect(result.data!.ports).toHaveLength(4);
    expect(result.data!.ports[0]).toEqual({
      name: 'http-80',
      number: 80,
      protocol: 'http',
      exposed_to_public: true,
    });
    expect(result.data!.ports[3]).toEqual({
      name: 'http-3000',
      number: 3000,
      protocol: 'http',
      exposed_to_public: true,
    });
  });

  it('should convert volume mounts correctly', () => {
    const service: DockerComposeService = {
      image: 'postgres:13',
      volumes: [
        'db-data:/var/lib/postgresql/data',
        'logs:/var/log:ro',
        '/host/path:/container/path', // Should generate warning
        {
          type: 'volume',
          source: 'config-data',
          target: '/etc/config',
          read_only: true,
        },
        {
          type: 'bind',
          source: '/host/bind',
          target: '/container/bind',
        }, // Should generate warning
      ],
    };

    const result = convertServiceToStackResource('database', service);

    expect(result.success).toBe(true);
    expect(result.data!.volume_mounts).toHaveLength(3); // Only named volumes
    expect(result.data!.volume_mounts[0]).toEqual({
      source_volume_name: 'db-data',
      target_path: '/var/lib/postgresql/data',
    });
    expect(result.data!.volume_mounts[1]).toEqual({
      source_volume_name: 'logs',
      target_path: '/var/log',
    });
    expect(result.data!.volume_mounts[2]).toEqual({
      source_volume_name: 'config-data',
      target_path: '/etc/config',
    });

    // Should have warnings about unsupported bind mounts
    expect(result.warnings).toBeDefined();
    const bindWarnings = result.warnings!.filter((w) => w.message.includes('not supported'));
    expect(bindWarnings.length).toBeGreaterThan(0);
  });

  it('should sanitize volume mount source names to match the sanitized volume name', () => {
    // The mount's source_volume_name must use the same RFC 1123 sanitization as the
    // volume definition, otherwise the mount references a volume name that never exists.
    const service: DockerComposeService = {
      image: 'postgres:16',
      volumes: ['postgres_data:/var/lib/postgresql/data'],
    };

    const result = convertServiceToStackResource('postgresql', service);

    expect(result.success).toBe(true);
    expect(result.data!.volume_mounts[0]).toEqual({
      source_volume_name: 'postgres-data',
      target_path: '/var/lib/postgresql/data',
    });
  });

  it('should convert environment variables from different formats', () => {
    const service: DockerComposeService = {
      image: 'node:18',
      environment: [
        'NODE_ENV=production',
        'PORT=3000',
        'DEBUG=', // Empty value
        'FEATURE_FLAG', // No value
      ],
    };

    const result = convertServiceToStackResource('app', service);

    expect(result.success).toBe(true);
    expect(result.data!.execution_config.environment_variables).toHaveLength(4);
    expect(result.data!.execution_config.environment_variables[0]).toEqual({
      name: 'NODE_ENV',
      value: 'production',
      from: 'stack',
    });
    expect(result.data!.execution_config.environment_variables[2]).toEqual({
      name: 'DEBUG',
      value: '',
      from: 'stack',
    });
    expect(result.data!.execution_config.environment_variables[3]).toEqual({
      name: 'FEATURE_FLAG',
      value: '',
      from: 'stack',
    });
  });

  it('should handle command in different formats', () => {
    const serviceString: DockerComposeService = {
      image: 'node:18',
      command: 'npm start',
    };

    const serviceArray: DockerComposeService = {
      image: 'node:18',
      command: ['npm', 'run', 'start:prod'],
    };

    const resultString = convertServiceToStackResource('app1', serviceString);
    const resultArray = convertServiceToStackResource('app2', serviceArray);

    expect(resultString.success).toBe(true);
    expect(resultString.data!.execution_config.command).toEqual(['npm', 'start']);

    expect(resultArray.success).toBe(true);
    expect(resultArray.data!.execution_config.command).toEqual(['npm', 'run', 'start:prod']);
  });

  it('should map internal ports as exposed when external mapping exists', () => {
    const service: DockerComposeService = {
      image: 'nginx:latest',
      ports: [
        '8080:80',       // External port 8080 mapped to internal 80 -> expose internal 80
        '443:443',       // External port 443 mapped to internal 443 -> expose internal 443
        '3000:3000',     // External port 3000 mapped to internal 3000 -> expose internal 3000
        '9000:8000',     // External port 9000 mapped to internal 8000 -> expose internal 8000
      ],
    };

    const result = convertServiceToStackResource('web-server', service);

    expect(result.success).toBe(true);
    expect(result.data!.ports).toHaveLength(4);

    // Check that internal ports are exposed, not external ports
    expect(result.data!.ports[0]).toEqual({
      name: 'http-80',
      number: 80,      // Internal port, not external 8080
      protocol: 'http',
      exposed_to_public: true,
    });

    expect(result.data!.ports[1]).toEqual({
      name: 'http-443',
      number: 443,     // Internal port (same as external in this case)
      protocol: 'http',
      exposed_to_public: true,
    });

    expect(result.data!.ports[2]).toEqual({
      name: 'http-3000',
      number: 3000,    // Internal port (same as external in this case)
      protocol: 'http',
      exposed_to_public: true,
    });

    expect(result.data!.ports[3]).toEqual({
      name: 'http-8000',
      number: 8000,    // Internal port, not external 9000
      protocol: 'http',
      exposed_to_public: true,
    });
  });

  it('should handle internal-only ports without external mapping', () => {
    const service: DockerComposeService = {
      image: 'redis:latest',
      expose: [6379], // Internal port only, no external mapping
    };

    const result = convertServiceToStackResource('cache', service);

    expect(result.success).toBe(true);
    // Since expose doesn't provide external mapping, no ports should be created
    // or if ports are created, they should not be exposed to public
    expect(result.data!.ports).toHaveLength(0);
  });

  it('should use internal port when external mapping differs', () => {
    const service: DockerComposeService = {
      image: 'api:latest',
      ports: [
        '8080:3000',     // External 8080 maps to internal 3000 -> expose internal 3000
        '9090:4000',     // External 9090 maps to internal 4000 -> expose internal 4000
      ],
    };

    const result = convertServiceToStackResource('api-server', service);

    expect(result.success).toBe(true);
    expect(result.data!.ports).toHaveLength(2);

    // The internal port should be exposed, not the external port
    expect(result.data!.ports[0]).toEqual({
      name: 'http-3000',
      number: 3000,    // Internal port, not external 8080
      protocol: 'http',
      exposed_to_public: true,
    });

    expect(result.data!.ports[1]).toEqual({
      name: 'http-4000',
      number: 4000,    // Internal port, not external 9090
      protocol: 'http',
      exposed_to_public: true,
    });
  });


});

describe('convertVolumeToStackVolume', () => {
  it('should convert simple named volume', () => {
    const result = convertVolumeToStackVolume('data-volume', {});

    expect(result.success).toBe(true);
    expect(result.data).toEqual({
      name: 'data-volume',
      sourceType: 'None',
      labels: [],
      spec: {
        size: '1Gi',
        access_mode: 'ReadWriteOnce',
        needs_sync_before_use: false,
      },
    });
  });

  it('should convert volume with driver options', () => {
    const volumeConfig = {
      driver: 'nfs',
      driver_opts: {
        type: 'nfs',
        device: 'nfs-server:/path',
      },
    };

    const result = convertVolumeToStackVolume('nfs-volume', volumeConfig);

    expect(result.success).toBe(true);
    expect(result.data!.name).toBe('nfs-volume');
    expect(result.warnings).toBeDefined();
    expect(result.warnings!.some((w) => w.message.includes('not supported'))).toBe(true);
  });

  it('should handle external volumes with warnings', () => {
    const volumeConfig = {
      external: true,
      name: 'existing-volume',
    };

    const result = convertVolumeToStackVolume('ext-volume', volumeConfig);

    expect(result.success).toBe(true);
    expect(result.warnings).toBeDefined();
    expect(result.warnings!.some((w) => w.message.includes('External volume detected'))).toBe(true);
  });

  it('should handle null volume config', () => {
    const resultNull = convertVolumeToStackVolume('vol1', null);

    expect(resultNull.success).toBe(true);
    expect(resultNull.data!.name).toBe('vol1');
  });

  it('should sanitize volume names to a valid RFC 1123 Kubernetes name', () => {
    // Docker Compose allows underscores (e.g. the ToolJet template's "postgres_data"),
    // but Kubernetes resource names must be RFC 1123 subdomains. An underscore makes the
    // Volume CR invalid, which blocks the release-apply gate and stalls the whole deploy.
    const result = convertVolumeToStackVolume('postgres_data', {});

    expect(result.success).toBe(true);
    expect(result.data!.name).toBe('postgres-data');
  });
});

describe('getConversionSummary', () => {
  it('should generate correct summary for successful conversion', () => {
    const result: ConversionResult = {
      success: true,
      data: {
        name: 'test-stack',
        labels: [],
        spec: {
          stack_resources: [
            createMockStackResource('app1'),
            createMockStackResource('app2'),
          ],
          volumes: [
            createMockVolume('vol1'),
          ],
        },
      },
      warnings: [
        { type: 'unsupported', message: 'Feature not supported' },
        { type: 'defaults', message: 'Using defaults' },
      ],
    };

    const summary = getConversionSummary(result);

    expect(summary.servicesConverted).toBe(2);
    expect(summary.volumesConverted).toBe(1);
    expect(summary.warningCount).toBe(2);
    expect(summary.errorCount).toBe(0);
    expect(summary.requiresManualConfiguration).toBe(true);
  });

  it('should generate correct summary for failed conversion', () => {
    const result: ConversionResult = {
      success: false,
      errors: [
        { type: 'conversion', message: 'Failed to convert' },
        { type: 'validation', message: 'Validation error' },
      ],
    };

    const summary = getConversionSummary(result);

    expect(summary.servicesConverted).toBe(0);
    expect(summary.volumesConverted).toBe(0);
    expect(summary.warningCount).toBe(0);
    expect(summary.errorCount).toBe(2);
    expect(summary.requiresManualConfiguration).toBe(false);
  });

  it('should detect manual configuration requirements', () => {
    const resultWithBuildWarning: ConversionResult = {
      success: true,
      data: {
        name: 'test',
        labels: [],
        spec: { stack_resources: [createMockStackResource('app')] },
      },
      warnings: [
        { type: 'defaults', dockerComposeField: 'build', message: 'Requires manual configuration' },
      ],
    };

    const resultWithoutManual: ConversionResult = {
      success: true,
      data: {
        name: 'test',
        labels: [],
        spec: { stack_resources: [createMockStackResource('app')] },
      },
      warnings: [
        { type: 'unsupported', message: 'Feature not supported' },
      ],
    };

    expect(getConversionSummary(resultWithBuildWarning).requiresManualConfiguration).toBe(true);
    expect(getConversionSummary(resultWithoutManual).requiresManualConfiguration).toBe(false);
  });
});

describe('error handling', () => {
  it('should handle conversion errors gracefully', () => {
    // Test with malformed service that throws during conversion
    const malformedCompose = {
      version: '3.8',
      services: {
        'broken-service': {
          // This will cause errors during conversion processing
          ports: 'invalid-port-format' as unknown as DockerComposeService['ports'],
          image: 'nginx:latest',
        },
      },
    };

    const result = convertDockerComposeToStackData(malformedCompose);

    // Should still succeed but with warnings about port conversion
    expect(result.success).toBe(true);
    expect(result.warnings).toBeDefined();
  });

  it('should handle null/undefined inputs', () => {
    const emptyResult = convertDockerComposeToStackData({} as DockerComposeFile);
    expect(emptyResult.success).toBe(false);

    const serviceResult = convertServiceToStackResource('test', {} as DockerComposeService);
    expect(serviceResult.success).toBe(false);
  });
});

describe('fixture-based integration tests', () => {
  describe('simple Docker Compose fixture', () => {
    it('should convert simple web app Docker Compose to StackData', () => {
      const parseResult = parseDockerCompose(simpleComposeYaml);

      expect(parseResult.success).toBe(true);
      expect(parseResult.data).toBeDefined();

      const conversionResult = convertDockerComposeToStackData(parseResult.data! as DockerComposeFile);

      expect(conversionResult.success).toBe(true);
      expect(conversionResult.data?.name).toBe('simple-web-app');

      const summary = getConversionSummary(conversionResult);
      expect(summary.servicesConverted).toBe(3);
      expect(summary.volumesConverted).toBe(1);
      expect(summary.warningCount).toBeGreaterThanOrEqual(4); // Environment variables + volume defaults
      expect(summary.errorCount).toBe(0);
      expect(summary.requiresManualConfiguration).toBe(true);
    });
  });

  describe('complex Docker Compose fixture', () => {
    it('should convert complex multi-service Docker Compose to StackData', () => {
      const parseResult = parseDockerCompose(complexComposeYaml);

      expect(parseResult.success).toBe(true);
      expect(parseResult.data).toBeDefined();

      const conversionResult = convertDockerComposeToStackData(parseResult.data! as DockerComposeFile);

      expect(conversionResult.success).toBe(true);
      expect(conversionResult.data?.name).toBe('complex-app');

      const summary = getConversionSummary(conversionResult);
      expect(summary.servicesConverted).toBe(4);
      expect(summary.volumesConverted).toBe(3);
      expect(summary.warningCount).toBeGreaterThanOrEqual(15); // Many unsupported features
      expect(summary.errorCount).toBe(0);
      expect(summary.requiresManualConfiguration).toBe(true);

      // Check for specific unsupported feature warnings
      expect(conversionResult.warnings).toBeDefined();
      const warningFields = conversionResult.warnings!.map(w => w.dockerComposeField);
      expect(warningFields).toContain('networks');
      expect(warningFields).toContain('secrets');
      expect(warningFields).toContain('build.args');
    });
  });

  describe('invalid Docker Compose fixture', () => {
    it('should handle invalid Docker Compose with errors', () => {
      const parseResult = parseDockerCompose(invalidComposeYaml);

      expect(parseResult.success).toBe(true);
      expect(parseResult.data).toBeDefined();

      const conversionResult = convertDockerComposeToStackData(parseResult.data! as DockerComposeFile);

      // Should succeed but with errors (partial conversion)
      expect(conversionResult.success).toBe(true);

      const summary = getConversionSummary(conversionResult);
      expect(summary.servicesConverted).toBe(1);
      expect(summary.volumesConverted).toBe(0);
      expect(summary.errorCount).toBe(1);

      // Should have error about missing image/build spec
      expect(conversionResult.errors).toBeDefined();
      expect(conversionResult.errors!.length).toBeGreaterThan(0);
      expect(conversionResult.errors![0].message).toContain('image or build specification');
      expect(conversionResult.errors![0].service).toBe('invalid-service');
    });
  });
});

describe('real-world fixture tests', () => {
  it('should convert simple web app fixture successfully', () => {
    const yamlContent = `version: '3.8'
name: 'simple-web-app'

services:
  web:
    image: nginx:latest
    ports:
      - "80:80"
    environment:
      - NODE_ENV=production
      - PORT=80

  api:
    image: node:18-alpine
    ports:
      - "3000:3000"
    environment:
      NODE_ENV: production
      DATABASE_URL: postgres://db:5432/myapp
    depends_on:
      - database

  database:
    image: postgres:13
    environment:
      POSTGRES_DB: myapp
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
    volumes:
      - db-data:/var/lib/postgresql/data

volumes:
  db-data:
    driver: local`;

    const parseResult = parseDockerCompose(yamlContent);
    expect(parseResult.success).toBe(true);

    const conversionResult = convertDockerComposeToStackData(parseResult.data! as DockerComposeFile);
    expect(conversionResult.success).toBe(true);
    expect(conversionResult.data!.name).toBe('simple-web-app');
    expect(conversionResult.data!.spec.stack_resources).toHaveLength(3);
    expect(conversionResult.data!.spec.volumes).toHaveLength(1);
    expect(conversionResult.warnings!.length).toBeGreaterThan(0);
  });

  it('should convert complex app fixture with warnings', () => {
    const yamlContent = `version: '3.8'
name: 'complex-app'

services:
  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile.prod
      args:
        NODE_VERSION: "18"
    ports:
      - "80:80"
      - "443:443"
    environment:
      - API_URL=http://api:3000
    depends_on:
      - api
    networks:
      - app-network
    restart: always

  api:
    build: ./backend
    ports:
      - "3000:3000"
    environment:
      NODE_ENV: production
      DATABASE_URL: postgres://db:5432/myapp
      REDIS_URL: redis://redis:6379
    volumes:
      - api-logs:/var/log/app
      - ./config:/app/config:ro
    depends_on:
      - database
      - redis
    networks:
      - app-network

networks:
  app-network:
    driver: bridge

volumes:
  api-logs:
    driver: local`;

    const parseResult = parseDockerCompose(yamlContent);
    expect(parseResult.success).toBe(true);

    const conversionResult = convertDockerComposeToStackData(parseResult.data! as DockerComposeFile);
    expect(conversionResult.success).toBe(true);
    expect(conversionResult.data!.name).toBe('complex-app');
    expect(conversionResult.data!.spec.stack_resources).toHaveLength(2);
    expect(conversionResult.data!.spec.volumes).toHaveLength(1);
    expect(conversionResult.warnings!.length).toBeGreaterThan(5); // Many warnings for unsupported features
  });

  it('should handle invalid fixture with partial success', () => {
    const yamlContent = `version: '3.8'

services:
  # This service has no image or build spec - should fail conversion
  invalid-service:
    ports:
      - "8080:80"
    environment:
      - NODE_ENV=production

  # This service should work
  valid-service:
    image: nginx:latest
    ports:
      - "80:80"`;

    const parseResult = parseDockerCompose(yamlContent);
    expect(parseResult.success).toBe(true);

    const conversionResult = convertDockerComposeToStackData(parseResult.data! as DockerComposeFile);
    expect(conversionResult.success).toBe(true); // Partial success
    expect(conversionResult.data!.spec.stack_resources).toHaveLength(1); // Only valid service converted
    expect(conversionResult.errors).toHaveLength(1); // One service failed
  });
});

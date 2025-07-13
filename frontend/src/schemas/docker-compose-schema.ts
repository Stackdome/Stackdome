/**
 * Zod schemas for Docker Compose file validation
 * These provide runtime validation for Docker Compose YAML files
 */
import { z } from "zod";

// Utility schemas for common Docker Compose patterns
const StringOrArraySchema = z.union([
  z.string(),
  z.array(z.string())
]);

const PortMappingSchema = z.union([
  z.string(), // "80:80", "8080"
  z.number(), // 8080
  z.object({
    target: z.number().optional(),
    published: z.union([z.string(), z.number()]).optional(),
    protocol: z.enum(["tcp", "udp"]).optional(),
    mode: z.enum(["host", "ingress"]).optional(),
  })
]);

const EnvironmentSchema = z.union([
  z.array(z.string()), // ["KEY=value", "KEY2=value2"]
  z.record(z.string(), z.union([z.string(), z.number(), z.boolean(), z.null()])) // { KEY: "value" }
]);

const VolumeSchema = z.union([
  z.string(), // "volume:path" or "/host/path:/container/path"
  z.object({
    type: z.enum(["bind", "volume", "tmpfs", "npipe", "cluster"]).optional(),
    source: z.string().optional(),
    target: z.string(),
    read_only: z.boolean().optional(),
    consistency: z.enum(["consistent", "cached", "delegated"]).optional(),
    bind: z.object({
      propagation: z.enum(["rprivate", "private", "rshared", "shared", "rslave", "slave"]).optional(),
      create_host_path: z.boolean().optional(),
      selinux: z.enum(["z", "Z"]).optional(),
    }).optional(),
    volume: z.object({
      nocopy: z.boolean().optional(),
    }).optional(),
    tmpfs: z.object({
      size: z.union([z.string(), z.number()]).optional(),
      mode: z.number().optional(),
    }).optional(),
  })
]);

const BuildSchema = z.union([
  z.string(), // build context path
  z.object({
    context: z.string().optional(),
    dockerfile: z.string().optional(),
    dockerfile_inline: z.string().optional(),
    args: z.union([
      z.array(z.string()),
      z.record(z.string(), z.union([z.string(), z.number(), z.boolean(), z.null()]))
    ]).optional(),
    ssh: z.union([z.string(), z.array(z.string())]).optional(),
    cache_from: z.array(z.string()).optional(),
    cache_to: z.array(z.string()).optional(),
    extra_hosts: z.union([z.array(z.string()), z.record(z.string(), z.string())]).optional(),
    isolation: z.string().optional(),
    privileged: z.boolean().optional(),
    labels: z.union([
      z.array(z.string()),
      z.record(z.string(), z.string())
    ]).optional(),
    network: z.string().optional(),
    no_cache: z.boolean().optional(),
    platform: z.string().optional(),
    pull: z.boolean().optional(),
    shm_size: z.union([z.string(), z.number()]).optional(),
    target: z.string().optional(),
    secrets: z.array(z.union([
      z.string(),
      z.object({
        source: z.string(),
        target: z.string().optional(),
        uid: z.union([z.string(), z.number()]).optional(),
        gid: z.union([z.string(), z.number()]).optional(),
        mode: z.union([z.string(), z.number()]).optional(),
      })
    ])).optional(),
    tags: z.array(z.string()).optional(),
    ulimits: z.record(z.string(), z.union([
      z.number(),
      z.object({
        soft: z.number(),
        hard: z.number(),
      })
    ])).optional(),
  })
]);

const DependsOnSchema = z.union([
  z.array(z.string()), // ["service1", "service2"]
  z.record(z.string(), z.object({
    condition: z.enum(["service_started", "service_healthy", "service_completed_successfully"]).optional(),
    restart: z.boolean().optional(),
    required: z.boolean().optional(),
  })) // { service1: { condition: "service_healthy" } }
]);

// Core service schema - covers the most common Docker Compose service properties
export const DockerComposeServiceSchema = z.object({
  image: z.string().optional(),
  build: BuildSchema.optional(),
  command: StringOrArraySchema.optional(),
  args: StringOrArraySchema.optional(),
  entrypoint: StringOrArraySchema.optional(),
  environment: EnvironmentSchema.optional(),
  env_file: StringOrArraySchema.optional(),
  ports: z.array(PortMappingSchema).optional(),
  expose: z.array(z.union([z.string(), z.number()])).optional(),
  volumes: z.array(VolumeSchema).optional(),
  depends_on: DependsOnSchema.optional(),
  links: z.array(z.string()).optional(),
  external_links: z.array(z.string()).optional(),
  restart: z.enum(["no", "always", "on-failure", "unless-stopped"]).optional(),
  working_dir: z.string().optional(),
  user: z.string().optional(),
  hostname: z.string().optional(),
  domainname: z.string().optional(),
  container_name: z.string().optional(),
  labels: z.union([
    z.array(z.string()), // ["label=value"]
    z.record(z.string(), z.string()) // { label: "value" }
  ]).optional(),
  networks: z.union([
    z.array(z.string()),
    z.record(z.string(), z.object({
      aliases: z.array(z.string()).optional(),
      ipv4_address: z.string().optional(),
      ipv6_address: z.string().optional(),
      link_local_ips: z.array(z.string()).optional(),
      mac_address: z.string().optional(),
      priority: z.number().optional(),
    }).optional())
  ]).optional(),
  extra_hosts: z.union([
    z.array(z.string()),
    z.record(z.string(), z.string())
  ]).optional(),
  privileged: z.boolean().optional(),
  tty: z.boolean().optional(),
  stdin_open: z.boolean().optional(),
  init: z.boolean().optional(),
  stop_grace_period: z.string().optional(),
  stop_signal: z.string().optional(),
  security_opt: z.array(z.string()).optional(),
  tmpfs: z.union([
    z.array(z.string()),
    z.record(z.string(), z.string())
  ]).optional(),
  logging: z.object({
    driver: z.string().optional(),
    options: z.record(z.string(), z.string()).optional(),
  }).optional(),
  healthcheck: z.object({
    test: StringOrArraySchema.optional(),
    interval: z.string().optional(),
    timeout: z.string().optional(),
    retries: z.number().optional(),
    start_period: z.string().optional(),
    start_interval: z.string().optional(),
    disable: z.boolean().optional(),
  }).optional(),
  // Allow additional properties for completeness
}).passthrough();

// Volume definition schema
export const DockerComposeVolumeSchema = z.union([
  z.object({
    driver: z.string().optional(),
    driver_opts: z.record(z.string(), z.string()).optional(),
    external: z.union([z.boolean(), z.object({
      name: z.string(),
    })]).optional(),
    labels: z.union([
      z.array(z.string()),
      z.record(z.string(), z.string())
    ]).optional(),
    name: z.string().optional(),
  }).passthrough(),
  z.null(), // Allow null for empty volume definitions like "postgres_data:"
]);

// Network definition schema
export const DockerComposeNetworkSchema = z.object({
  driver: z.string().optional(),
  driver_opts: z.record(z.string(), z.string()).optional(),
  attachable: z.boolean().optional(),
  enable_ipv6: z.boolean().optional(),
  ipam: z.object({
    driver: z.string().optional(),
    config: z.array(z.object({
      subnet: z.string().optional(),
      ip_range: z.string().optional(),
      gateway: z.string().optional(),
      aux_addresses: z.record(z.string(), z.string()).optional(),
    })).optional(),
    options: z.record(z.string(), z.string()).optional(),
  }).optional(),
  internal: z.boolean().optional(),
  labels: z.union([
    z.array(z.string()),
    z.record(z.string(), z.string())
  ]).optional(),
  external: z.union([z.boolean(), z.object({
    name: z.string(),
  })]).optional(),
  name: z.string().optional(),
}).passthrough();

// Secret definition schema
export const DockerComposeSecretSchema = z.object({
  file: z.string().optional(),
  external: z.union([z.boolean(), z.object({
    name: z.string(),
  })]).optional(),
  labels: z.union([
    z.array(z.string()),
    z.record(z.string(), z.string())
  ]).optional(),
  driver: z.string().optional(),
  driver_opts: z.record(z.string(), z.string()).optional(),
  template_driver: z.string().optional(),
  name: z.string().optional(),
}).passthrough();

// Config definition schema
export const DockerComposeConfigSchema = z.object({
  file: z.string().optional(),
  external: z.union([z.boolean(), z.object({
    name: z.string(),
  })]).optional(),
  labels: z.union([
    z.array(z.string()),
    z.record(z.string(), z.string())
  ]).optional(),
  template_driver: z.string().optional(),
  name: z.string().optional(),
}).passthrough();

// Main Docker Compose file schema
export const DockerComposeSchema = z.object({
  version: z.string().optional(),
  name: z.string().optional(),
  services: z.record(z.string(), DockerComposeServiceSchema).optional(),
  volumes: z.record(z.string(), DockerComposeVolumeSchema).optional(),
  networks: z.record(z.string(), DockerComposeNetworkSchema).optional(),
  secrets: z.record(z.string(), DockerComposeSecretSchema).optional(),
  configs: z.record(z.string(), DockerComposeConfigSchema).optional(),
  // Include support for various extension fields
  include: z.array(z.union([
    z.string(),
    z.object({
      path: StringOrArraySchema.optional(),
      env_file: StringOrArraySchema.optional(),
      project_directory: z.string().optional(),
    })
  ])).optional(),
}).passthrough();

// Convenience type exports
export type DockerComposeFile = z.infer<typeof DockerComposeSchema>;
export type DockerComposeService = z.infer<typeof DockerComposeServiceSchema>;
export type DockerComposeVolume = z.infer<typeof DockerComposeVolumeSchema>;
export type DockerComposeNetwork = z.infer<typeof DockerComposeNetworkSchema>;
export type DockerComposeSecret = z.infer<typeof DockerComposeSecretSchema>;
export type DockerComposeConfig = z.infer<typeof DockerComposeConfigSchema>;
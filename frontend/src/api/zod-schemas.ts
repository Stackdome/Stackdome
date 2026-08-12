import { makeApi, Zodios, type ZodiosOptions } from "@zodios/core";
import { z } from "zod";

const DomainName = z
  .object({ id: z.string(), fqdn: z.string() })
  .partial()
  .passthrough();
const Organisation = z
  .object({
    id: z.string(),
    name: z.string(),
    domains: z.array(DomainName),
    is_platform: z.boolean(),
    created_at: z.string().datetime({ offset: true }),
    updated_at: z.string().datetime({ offset: true }),
  })
  .partial()
  .passthrough();
const UserSignupRequest = z
  .object({
    name: z.string(),
    email: z.string().email(),
    password: z.string(),
    organisation: Organisation.optional(),
    invite_token: z.string().optional(),
    turnstile_token: z.string().optional(),
  })
  .passthrough();
const UserRole = z.enum(["OrgAdmin", "OrgMember"]);
const UserProjectMembership = z
  .object({
    project_id: z.string(),
    project_name: z.string(),
    role: z.enum(["Developer", "Viewer"]),
    default_project: z.boolean(),
  })
  .partial()
  .passthrough();
const User = z
  .object({
    id: z.string(),
    name: z.string(),
    username: z.string(),
    email: z.string().email(),
    organisation: z.string(),
    role: UserRole,
    organisation_id: z.string(),
    projects: z.array(UserProjectMembership),
  })
  .partial()
  .passthrough();
const UserSignupResponse = z
  .object({ user: User, jwt_token: z.string(), refresh_token: z.string() })
  .partial()
  .passthrough();
const ObjectReference = z
  .object({ id: z.string(), kind: z.string(), href: z.string() })
  .partial()
  .passthrough();
const Error = ObjectReference.and(
  z
    .object({
      code: z.string(),
      reason: z.string(),
      operation_id: z.string(),
      details: z.object({}).partial().passthrough(),
    })
    .partial()
    .passthrough()
);
const TurnstileConfigResponse = z
  .object({ enabled: z.boolean(), site_key: z.string(), action: z.string() })
  .passthrough();
const SignupConfigResponse = z
  .object({ turnstile: TurnstileConfigResponse })
  .passthrough();
const AppConfigResponse = z
  .object({ github_oauth: z.boolean(), signup: SignupConfigResponse })
  .partial()
  .passthrough();
const Project = z
  .object({
    id: z.string().optional(),
    name: z.string(),
    organisation_id: z.string().optional(),
    default_project: z.boolean().optional(),
    created_at: z.string().datetime({ offset: true }).optional(),
    updated_at: z.string().datetime({ offset: true }).optional(),
  })
  .passthrough();
const ProjectList = z
  .object({ items: z.array(Project), total: z.number().int() })
  .partial()
  .passthrough();
const LoginRequest = z
  .object({ email: z.string(), password: z.string() })
  .passthrough();
const LoginResponse = z
  .object({
    token: z.string(),
    refresh_token: z.string(),
    user: User,
    expires_in: z.number().int(),
  })
  .partial()
  .passthrough();
const RefreshTokenRequest = z
  .object({ refreshToken: z.string() })
  .passthrough();
const RefreshTokenResponse = z
  .object({ token: z.string(), refreshToken: z.string() })
  .partial()
  .passthrough();
const APITokenCreateRequest = z
  .object({
    name: z.string(),
    scopes: z.array(z.string()),
    resource_ids: z.array(z.string()).optional(),
    expires_at: z.string().datetime({ offset: true }).optional(),
  })
  .passthrough();
const APITokenCreateResponse = z
  .object({
    token: z.string(),
    id: z.string(),
    name: z.string(),
    token_prefix: z.string(),
    expires_at: z.string().datetime({ offset: true }),
  })
  .partial()
  .passthrough();
const APIToken = z
  .object({
    id: z.string(),
    name: z.string(),
    user_id: z.string(),
    token_prefix: z.string(),
    scopes: z.array(z.string()),
    resource_ids: z.array(z.string()),
    org_id: z.string(),
    expires_at: z.string().datetime({ offset: true }),
    last_used_at: z.string().datetime({ offset: true }),
    created_at: z.string().datetime({ offset: true }),
    revoked_at: z.string().datetime({ offset: true }),
  })
  .partial()
  .passthrough();
const APITokenList = z
  .object({ items: z.array(APIToken) })
  .partial()
  .passthrough();
const ScopeResource = z
  .object({ resource: z.string(), actions: z.array(z.string()) })
  .partial()
  .passthrough();
const ScopeList = z
  .object({
    full_access_scope: z.string(),
    items: z.array(ScopeResource),
    total: z.number().int(),
  })
  .partial()
  .passthrough();
const UserList = z
  .object({
    items: z.array(User),
    total: z.number().int(),
    page: z.number().int(),
    page_size: z.number().int(),
    total_pages: z.number().int(),
  })
  .partial()
  .passthrough();
const SecretType = z.enum([
  "Generic",
  "DockerRegistry",
  "GitCredentials",
  "UsernamePassword",
  "Token",
  "SSHKey",
]);
const SecretData = z
  .object({ key: z.string(), value: z.string() })
  .passthrough();
const OutputDescriptor = z
  .object({
    name: z.string(),
    type: z.enum(["string", "integer", "boolean"]),
    sensitive: z.boolean(),
  })
  .passthrough();
const Secret = z.object({
  id: z.string().optional(),
  name: z.string(),
  description: z.string().optional(),
  organisation_id: z.string().optional(),
  project_id: z.string().optional(),
  type: SecretType,
  data: z.array(SecretData),
  outputs: z.array(OutputDescriptor).optional(),
  created_at: z.string().datetime({ offset: true }).optional(),
  updated_at: z.string().datetime({ offset: true }).optional(),
});
const SecretList = z
  .object({ items: z.array(Secret), total: z.number().int() })
  .partial()
  .passthrough();
const RegistryCredentialPurpose = z.enum(["pull", "push", "both"]);
const RegistryCredential = z.object({
  id: z.string().optional(),
  host: z.string(),
  purpose: RegistryCredentialPurpose.optional().default("both"),
  username: z.string(),
  password: z.string().optional(),
  organisation_id: z.string().optional(),
  created_at: z.string().datetime({ offset: true }).optional(),
  updated_at: z.string().datetime({ offset: true }).optional(),
});
const RegistryCredentialList = z
  .object({ items: z.array(RegistryCredential), total: z.number().int() })
  .partial()
  .passthrough();
const AffectedStackRef = z
  .object({ id: z.string(), name: z.string() })
  .partial()
  .passthrough();
const RegistryCredentialDeleteResponse = z
  .object({ affected_stacks: z.array(AffectedStackRef) })
  .partial()
  .passthrough();
const RegistryCredentialVerifyRequest = z.object({
  repository: z.string(),
  purpose: RegistryCredentialPurpose.optional().default("both"),
});
const GitHubAppManifestFlow = z
  .object({
    manifest: z.object({}).partial().passthrough(),
    github_url: z.string(),
    state: z.string(),
  })
  .partial()
  .passthrough();
const GitInstallation = z
  .object({
    id: z.string(),
    installation_id: z.number().int(),
    account_login: z.string(),
    account_type: z.string(),
    repository_selection: z.string(),
    created_at: z.string().datetime({ offset: true }),
  })
  .partial()
  .passthrough();
const GitInstallationList = z
  .object({ items: z.array(GitInstallation), total: z.number().int() })
  .partial()
  .passthrough();
const GitRepository = z
  .object({
    full_name: z.string(),
    clone_url: z.string(),
    default_branch: z.string(),
    private: z.boolean(),
    pushed_at: z.string().datetime({ offset: true }),
    owner: z.string(),
  })
  .partial()
  .passthrough();
const GitRepositoryPage = z
  .object({
    items: z.array(GitRepository),
    page: z.number().int(),
    total_count: z.number().int(),
    has_next: z.boolean(),
  })
  .partial()
  .passthrough();
const GitBranchList = z
  .object({ items: z.array(z.string()), total: z.number().int() })
  .partial()
  .passthrough();
const GitIntegrationType = z.enum(["git_credentials", "github_app"]);
const GitIntegrationBasicAuth = z.object({
  username: z.string(),
  password: z.string(),
});
const GitIntegrationAuth = z
  .object({ token: z.string(), basic: GitIntegrationBasicAuth })
  .partial();
const GitIntegration = z.object({
  id: z.string().optional(),
  type: GitIntegrationType.optional().default("git_credentials"),
  host: z.string(),
  status: z.enum(["active", "pending_install", "installed"]).optional(),
  auth: GitIntegrationAuth.optional(),
  credentials_configured: z.boolean().optional(),
  install_url: z.string().optional(),
  organisation_id: z.string().optional(),
  created_at: z.string().datetime({ offset: true }).optional(),
  updated_at: z.string().datetime({ offset: true }).optional(),
});
const GitIntegrationList = z
  .object({ items: z.array(GitIntegration), total: z.number().int() })
  .partial()
  .passthrough();
const GitIntegrationVerifyRequest = z.object({ repo_url: z.string() });
const SecretReference = z
  .object({ secret_id: z.string().uuid(), key: z.string() })
  .passthrough();
const S3Credentials = z
  .object({
    access_key_id: SecretReference,
    secret_access_key: SecretReference,
    region: z.string(),
    endpoint_url: z.string().optional(),
  })
  .passthrough();
const AzureCredentials = z
  .object({
    connection_string: SecretReference,
    storage_account_name: z.string().optional(),
  })
  .passthrough();
const GCSCredentials = z
  .object({ service_account_credentials: SecretReference })
  .passthrough();
const ObjectStoreConfiguration = z
  .object({
    s3_credentials: S3Credentials,
    azure_credentials: AzureCredentials,
    gcs_credentials: GCSCredentials,
  })
  .partial()
  .passthrough();
const ObjectStoreSpec = z
  .object({
    configuration: ObjectStoreConfiguration,
    destination_path: z.string(),
    retention_policy: z
      .string()
      .regex(/^[1-9][0-9]*[dwm]$/)
      .optional()
      .default("7d"),
  })
  .passthrough();
const ObjectStoreStatus = z
  .object({ state: z.enum(["Pending", "Ready", "Error"]), message: z.string() })
  .partial()
  .passthrough();
const ObjectStore = z
  .object({
    id: z.string().optional(),
    organisation_id: z.string().optional(),
    project_id: z.string().optional(),
    name: z.string(),
    spec: ObjectStoreSpec,
    status: ObjectStoreStatus.optional(),
    created_at: z.string().datetime({ offset: true }).optional(),
    updated_at: z.string().datetime({ offset: true }).optional(),
  })
  .passthrough();
const ObjectStoreList = z
  .object({ items: z.array(ObjectStore), total: z.number().int() })
  .partial()
  .passthrough();
const Label = z.object({ key: z.string(), value: z.string() });
const Annotation = z
  .object({ key: z.string(), value: z.string() })
  .passthrough();
const PushTarget = z.object({
  repository: z.string(),
  registry_credentials_id: z.string().optional(),
});
const GitSource = z.object({
  repo_url: z.string(),
  branch: z.string().optional(),
  tag: z.string().optional(),
  commit: z.string().optional(),
  dockerfile_path: z.string().optional().default("Dockerfile"),
  build_context: z.string().optional().default("."),
  integration_id: z.string().optional(),
  push: PushTarget.optional(),
});
const ImageSource = z.object({
  ref: z.string(),
  registry_credentials_id: z.string().optional(),
});
const VolumeBuildSource = z
  .object({
    volume_id: z.string(),
    volume_name: z.string(),
    current_volume_hash: z.string(),
    dockerfile_path: z.string().default("Dockerfile"),
    build_context: z.string().default("."),
  })
  .partial();
const SourceSpec = z
  .object({ git: GitSource, image: ImageSource, volume: VolumeBuildSource })
  .partial();
const InitSpec = z
  .object({ command: z.array(z.string()), args: z.array(z.string()) })
  .partial();
const EnvVar = z.object({
  name: z.string(),
  value: z.string().optional(),
  self_output: z.string().optional(),
});
const ExecutionConfig = z
  .object({
    command: z.array(z.string()),
    args: z.array(z.string()),
    environment_variables: z.array(EnvVar),
  })
  .partial()
  .passthrough();
const VolumeMountSourceType = z.enum([
  "EmptyVolume",
  "RemoteDirSyncedVolume",
  "BuildArtifactSyncedVolume",
  "GitRepoSyncedVolume",
]);
const VolumeMount = z
  .object({
    stack_resource_id: z.string().optional(),
    source_volume_type: VolumeMountSourceType.optional(),
    source_volume_name: z.string(),
    source_sub_path: z.string().optional(),
    target_path: z.string(),
  })
  .passthrough();
const LifecycleConfig = z
  .object({ restart_request_time: z.string().datetime({ offset: true }) })
  .partial()
  .passthrough();
const Port = z
  .object({
    name: z.string(),
    number: z.number().int(),
    protocol: z.string().optional(),
    exposed_to_public: z.boolean(),
    subdomain_prefix: z.string().optional(),
  })
  .passthrough();
const StackResource = z
  .object({
    id: z.string().optional(),
    stack_id: z.string().optional(),
    name: z.string(),
    labels: z.array(Label).optional(),
    annotations: z.array(Annotation).optional(),
    revision: z.string().optional(),
    source: SourceSpec.optional(),
    init_spec: InitSpec.optional(),
    execution_config: ExecutionConfig.optional(),
    volume_mounts: z.array(VolumeMount).optional(),
    depends_on: z.array(z.string()).optional(),
    lifecycle_config: LifecycleConfig.optional(),
    ports: z.array(Port).optional(),
    outputs: z.array(OutputDescriptor).optional(),
    workload_type: z
      .enum(["Service", "StatefulService", "Worker", "Job", "CronJob"])
      .optional()
      .default("Service"),
    schedule: z.string().optional(),
    replicas: z.number().int().gte(0).optional(),
  })
  .passthrough();
const VolumeAccessMode = z.enum([
  "ReadWriteOnce",
  "ReadWriteMany",
  "ReadOnlyMany",
]);
const GitRepoRevision = z
  .object({ branch: z.string(), tag: z.string(), commit: z.string() })
  .partial()
  .passthrough();
const GitRepoSource = z
  .object({ repo_url: z.string(), revision: GitRepoRevision })
  .passthrough();
const VolumeSourceTypes = z.enum(["RemoteDir", "BuildArtifact", "GitRepo"]);
const RemoteSource = z
  .object({ path: z.string(), current_directory_hash: z.string() })
  .passthrough();
const BuildArtifact = z
  .object({
    resource_ref: z.string(),
    source_path: z.string(),
    destination_path: z.string(),
  })
  .passthrough();
const VolumeSource = z.object({
  git_repo_source: GitRepoSource.optional(),
  source_type: VolumeSourceTypes,
  remote_source: RemoteSource.optional(),
  build_source: z.array(BuildArtifact).optional(),
});
const VolumeSpec = z.object({
  size: z.string(),
  storage_class: z.string().optional(),
  needs_sync_before_use: z.boolean(),
  access_mode: VolumeAccessMode,
  source: VolumeSource.optional(),
});
const Condition = z
  .object({
    type: z.string(),
    status: z.string(),
    observed_generation: z.number().int(),
    last_transition_time: z.string().datetime({ offset: true }),
    reason: z.string(),
    message: z.string(),
  })
  .partial();
const BuildArtifactSyncInfo = z
  .object({
    resource_name: z.string(),
    build_id: z.string(),
    status: z.string(),
  })
  .partial()
  .passthrough();
const VolumeStatus = z
  .object({
    conditions: z.array(Condition),
    phase: z.string(),
    build_artifact_syncs: z.array(BuildArtifactSyncInfo),
    last_synced_git_revision: z.string(),
    last_remote_sync_hash: z.string(),
  })
  .partial()
  .passthrough();
const Volume = z.object({
  id: z.string().optional(),
  project_id: z.string().optional(),
  name: z.string(),
  labels: z.array(Label).optional(),
  annotations: z.array(Annotation).optional(),
  spec: VolumeSpec,
  status: VolumeStatus.optional(),
});
const TopologyNodeRef = z
  .object({
    type: z.enum([
      "stack_resource",
      "addon/postgres",
      "secret",
      "volume",
      "object_store",
    ]),
    id: z.string().optional(),
    name: z.string().optional(),
  })
  .passthrough();
const ConnectionTarget = z
  .object({
    type: z.enum(["env", "file"]),
    name: z.string().optional(),
    path: z.string().optional(),
  })
  .passthrough();
const OutputValueRef = z.object({ output: z.string() }).passthrough();
const ValueRef = z
  .object({
    output: z.string(),
    template: z.string(),
    values: z.record(OutputValueRef),
  })
  .partial()
  .passthrough();
const ConnectionMapping = z
  .object({ target: ConnectionTarget, value: ValueRef })
  .passthrough();
const PostgresEnvConfig = z
  .object({
    database: z.string(),
    credential_scope: z.enum(["owner", "superuser"]),
    superuser: z.boolean(),
  })
  .partial();
const VolumeMountConfig = z.object({
  mount_path: z.string(),
  sub_path: z.string().optional(),
  read_only: z.boolean().optional(),
});
const BuildArtifactSourceConfig = z.object({
  source_path: z.string(),
  destination_path: z.string().optional(),
});
const StackConnectionConfig = z.union([
  PostgresEnvConfig,
  VolumeMountConfig,
  BuildArtifactSourceConfig,
]);
const StackConnection = z
  .object({
    id: z.string().optional(),
    kind: z.enum(["env", "volume_mount", "build_artifact_source"]),
    from: TopologyNodeRef,
    to: TopologyNodeRef,
    mappings: z.array(ConnectionMapping).optional(),
    config: StackConnectionConfig.optional(),
  })
  .passthrough();
const StackSpec = z
  .object({
    stack_resources: z.array(StackResource),
    volumes: z.array(Volume),
    connections: z.array(StackConnection),
  })
  .partial()
  .passthrough();
const StackSettings = z
  .object({
    release_retention_limit: z.number().int().default(10),
    min_successful_releases: z.number().int().default(5),
  })
  .partial()
  .passthrough();
const StackLifecycle = z.enum(["active", "deleting"]);
const StackReleaseState = z.enum([
  "Pending",
  "InProgress",
  "Released",
  "Failed",
  "Superseded",
  "Cancelled",
]);
const ReleaseHealth = z.enum([
  "ok",
  "progressing",
  "degraded",
  "unavailable",
  "failed",
]);
const ReleaseSummary = z
  .object({
    id: z.string(),
    sequence: z.number().int(),
    state: StackReleaseState,
    health: ReleaseHealth,
    message: z.string(),
    created_at: z.string().datetime({ offset: true }),
    completed_at: z.string().datetime({ offset: true }),
  })
  .partial()
  .passthrough();
const Stack = z
  .object({
    id: z.string().optional(),
    organisation_id: z.string().optional(),
    project_id: z.string().optional(),
    user_id: z.string().optional(),
    name: z.string(),
    namespace: z.string().optional(),
    labels: z.array(Label).optional(),
    annotations: z.array(Annotation).optional(),
    revision: z.string().optional(),
    spec: StackSpec,
    settings: StackSettings.optional(),
    lifecycle: StackLifecycle.optional(),
    converged_release: ReleaseSummary.optional(),
    latest_release: ReleaseSummary.optional(),
    created_at: z.string().datetime({ offset: true }).optional(),
    updated_at: z.string().datetime({ offset: true }).optional(),
  })
  .passthrough();
const StackList = z
  .object({ items: z.array(Stack), total: z.number().int() })
  .partial()
  .passthrough();
const PostgresVersion = z
  .object({
    major: z.number().int().gte(13).lte(17),
    minor: z.number().int().optional(),
    enable_auto_minor_upgrade: z.boolean().optional().default(true),
    enable_auto_major_upgrade: z.boolean().optional().default(false),
  })
  .passthrough();
const PostgresInstances = z
  .object({
    count: z.number().int().gte(1).lte(5),
    placement: z
      .object({
        topology_key: z.string().default("kubernetes.io/hostname"),
        policy: z.enum(["preferred", "required"]).default("preferred"),
        node_selector: z.record(z.string()),
        tolerations: z.array(
          z
            .object({
              key: z.string(),
              operator: z.string(),
              value: z.string(),
              effect: z.string(),
            })
            .partial()
            .passthrough()
        ),
      })
      .partial()
      .passthrough()
      .optional(),
  })
  .passthrough();
const PostgresStorage = z
  .object({
    size: z.string().regex(/^[0-9]+[KMGTP]i?$/),
    storage_class: z.string().optional(),
  })
  .passthrough();
const PostgresResources = z
  .object({
    cpu: z
      .object({ request: z.string(), limit: z.string() })
      .partial()
      .passthrough(),
    memory: z
      .object({ request: z.string(), limit: z.string() })
      .partial()
      .passthrough(),
  })
  .partial()
  .passthrough();
const PostgresBackupConfig = z
  .object({
    enabled: z.boolean().default(false),
    object_store_id: z.string(),
    schedule: z.string().default("0 0 0 * * 0"),
    wal_archiving: z.boolean().default(false),
  })
  .partial()
  .passthrough();
const PostgresInitialization = z
  .object({
    type: z
      .enum([
        "new",
        "restore_from_backup",
        "restore_from_object_store",
        "import_from_external",
      ])
      .default("new"),
    restore_from_backup: z
      .object({ backup_id: z.string() })
      .partial()
      .passthrough(),
    restore_from_object_store: z
      .object({
        object_store_id: z.string(),
        source_postgres_addon_id: z.string(),
        recovery_target_time: z.string().datetime({ offset: true }),
      })
      .partial()
      .passthrough(),
    import_from_external: z
      .object({
        host: z.string(),
        port: z.number().int().default(5432),
        database: z.string(),
        username: z.string(),
        password_secret_id: z.string(),
        ssl_mode: z
          .enum(["disable", "require", "verify-ca", "verify-full"])
          .optional()
          .default("require"),
        databases_to_import: z.array(z.string()).optional(),
      })
      .passthrough(),
  })
  .partial()
  .passthrough();
const PostgresDatabase = z
  .object({
    name: z.string(),
    extensions: z.array(z.literal("vector")).optional(),
  })
  .passthrough();
const PostgresConfiguration = z
  .object({
    enable_superuser_access: z.boolean().default(false),
    parameters: z.record(z.string()),
  })
  .partial()
  .passthrough();
const PostgresAddonSpec = z
  .object({
    version: PostgresVersion,
    instances: PostgresInstances,
    storage: PostgresStorage,
    resources: PostgresResources.optional(),
    backup: PostgresBackupConfig.optional(),
    initialization: PostgresInitialization.optional(),
    databases: z.array(PostgresDatabase).optional(),
    configuration: PostgresConfiguration.optional(),
  })
  .passthrough();
const PostgresClusterInfo = z
  .object({ version: z.string() })
  .partial()
  .passthrough();
const PostgresConnectionInfo = z
  .object({
    host: z.string(),
    port: z.number().int().default(5432),
    databases: z.array(
      z.object({ name: z.string(), owner: z.string() }).partial().passthrough()
    ),
    credentials: z
      .object({
        superuser_secret_id: z.string(),
        app_user_secrets: z.record(z.string()),
        ca_certificate_secret_id: z.string(),
      })
      .partial()
      .passthrough(),
  })
  .partial()
  .passthrough();
const PostgresAddonStatus = z
  .object({
    state: z.enum([
      "Pending",
      "Creating",
      "Initializing",
      "Ready",
      "Updating",
      "Backing Up",
      "Restoring",
      "Error",
      "Deleting",
      "Hibernated",
      "Fenced",
    ]),
    message: z.string(),
    phase: z.string(),
    conditions: z.array(Condition),
    observed_revision: z.string(),
    observed_generation: z.number().int(),
    cluster_info: PostgresClusterInfo,
    connection_info: PostgresConnectionInfo,
  })
  .partial()
  .passthrough();
const PostgresAddon = z
  .object({
    id: z.string().optional(),
    organisation_id: z.string().optional(),
    project_id: z.string().optional(),
    user_id: z.string().optional(),
    cluster_id: z.string().optional(),
    name: z.string(),
    namespace: z.string().optional(),
    labels: z.array(Label).optional(),
    annotations: z.array(Annotation).optional(),
    revision: z.string().optional(),
    outputs: z.array(OutputDescriptor).optional(),
    spec: PostgresAddonSpec,
    status: PostgresAddonStatus.optional(),
    created_at: z.string().datetime({ offset: true }).optional(),
    updated_at: z.string().datetime({ offset: true }).optional(),
  })
  .passthrough();
const PostgresAddonList = z
  .object({ items: z.array(PostgresAddon), total: z.number().int() })
  .partial()
  .passthrough();
const ClusterImageRegistrySpec = z
  .object({
    backend_storage_size: z.string(),
    backend_storage_class: z.string(),
  })
  .partial()
  .passthrough();
const ClusterImageRegistryState = z.enum([
  "ImageRegistryPending",
  "ImageRegistryError",
  "ImageRegistryRunning",
  "ImageRegistryDeleting",
]);
const ClusterImageRegistryStatus = z
  .object({ state: ClusterImageRegistryState, conditions: z.array(Condition) })
  .partial()
  .passthrough();
const ClusterImageRegistry = z
  .object({
    id: z.string().optional(),
    name: z.string(),
    organisation_id: z.string().optional(),
    cluster_id: z.string().optional(),
    spec: ClusterImageRegistrySpec.optional(),
    status: ClusterImageRegistryStatus.optional(),
    created_at: z.string().datetime({ offset: true }).optional(),
    updated_at: z.string().datetime({ offset: true }).optional(),
  })
  .passthrough();
const Cluster = z
  .object({
    id: z.string().optional(),
    name: z.string(),
    organisation_id: z.string().optional(),
    shared_compute: z.boolean().optional(),
    platform: z.boolean().optional(),
    cluster_url: z.string(),
    cluster_ca_data: z.string(),
    cluster_sa_token: z.string(),
    cluster_image_registry: ClusterImageRegistry.optional(),
  })
  .passthrough();
const ClusterList = z
  .object({ items: z.array(Cluster), total: z.number().int() })
  .partial()
  .passthrough();
const ClusterImageRegistryList = z
  .object({ items: z.array(ClusterImageRegistry), total: z.number().int() })
  .partial()
  .passthrough();
const ProjectCreateRequest = z.object({ name: z.string() }).passthrough();
const ProjectUpdateRequest = z.object({ name: z.string() }).passthrough();
const ProjectRole = z.enum(["Developer", "Viewer"]);
const AddProjectMemberRequest = z
  .object({ user_id: z.string(), role: ProjectRole })
  .passthrough();
const ProjectMembership = z
  .object({
    id: z.string().optional(),
    project_id: z.string(),
    user_id: z.string(),
    role: z.enum(["Developer", "Viewer"]),
    project: Project.optional(),
    user: User.optional(),
    created_at: z.string().datetime({ offset: true }).optional(),
  })
  .passthrough();
const ProjectMembershipList = z
  .object({ items: z.array(ProjectMembership), total: z.number().int() })
  .partial()
  .passthrough();
const UpdateProjectMemberRoleRequest = z
  .object({ role: ProjectRole })
  .passthrough();
const ProjectRoleList = z
  .object({ roles: z.array(ProjectRole) })
  .partial()
  .passthrough();
const PromoteAdminRequest = z.object({ user_id: z.string() }).passthrough();
const DemoteAdminRequest = z
  .object({
    project_name: z.string(),
    role: ProjectRole.optional().default("Viewer"),
  })
  .passthrough();
const ResourceMetrics = z
  .object({
    assigned_nodes: z.array(z.string()),
    cpu_usage: z.string(),
    memory_usage: z.string(),
    node_capacities: z.array(
      z
        .object({
          node_name: z.string(),
          cpu_capacity: z.string(),
          memory_capacity: z.string(),
          storage_capacity: z.string(),
        })
        .partial()
        .passthrough()
    ),
    timestamp: z.string().datetime({ offset: true }),
  })
  .partial()
  .passthrough();
const StackResourceList = z
  .object({ items: z.array(StackResource), total: z.number().int() })
  .partial()
  .passthrough();
const BuildSourceRevision = z
  .object({
    volume_source_revision: z
      .object({ current_volume_hash: z.string() })
      .passthrough(),
    git_repo_revision: GitRepoRevision,
  })
  .partial()
  .passthrough();
const BuildSourceContext = z
  .object({
    volume: z
      .object({ id: z.string(), name: z.string().optional() })
      .passthrough(),
    git_repo: z.object({ repo_url: z.string() }).passthrough(),
  })
  .partial()
  .passthrough();
const BuildFailureDetail = z
  .object({
    failure_type: z.enum([
      "crash_loop",
      "out_of_memory",
      "image_pull_failed",
      "create_container_error",
      "exit_error",
      "port_not_listening",
    ]),
    reason: z.string(),
    message: z.string(),
    restart_count: z.number().int(),
    exit_code: z.number().int(),
  })
  .partial()
  .passthrough();
const ImageBuildStatus = z
  .object({
    state: z.string(),
    conditions: z.array(Condition),
    image_url: z.string(),
    build_source_revision: z.string(),
    last_build_failure_detail: BuildFailureDetail,
  })
  .partial()
  .passthrough();
const ImageBuild = z
  .object({
    id: z.string().optional(),
    namespace: z.string().optional(),
    stack_id: z.string().optional(),
    stack_resource_id: z.string(),
    stack_resource_name: z.string(),
    source_revision: BuildSourceRevision,
    build_context: BuildSourceContext,
    image_repo: z.string(),
    status: ImageBuildStatus.optional(),
    created_at: z.string().datetime({ offset: true }).optional(),
    updated_at: z.string().datetime({ offset: true }).optional(),
  })
  .passthrough();
const ImageBuildList = z
  .object({ items: z.array(ImageBuild), total: z.number().int() })
  .partial()
  .passthrough();
const TopologyNode = z
  .object({
    ref: TopologyNodeRef,
    label: z.string(),
    outputs: z.array(OutputDescriptor).optional(),
    state: z.string().optional(),
  })
  .passthrough();
const TopologyEdge = z
  .object({
    id: z.string().optional(),
    kind: z.enum([
      "env",
      "volume_mount",
      "build_artifact_source",
      "depends_on",
    ]),
    source: TopologyNodeRef,
    target: TopologyNodeRef,
    mappings: z.array(ConnectionMapping).optional(),
    config: StackConnectionConfig.optional(),
    source_of_truth: z.enum(["connection", "derived"]),
  })
  .passthrough();
const StackTopology = z
  .object({ nodes: z.array(TopologyNode), edges: z.array(TopologyEdge) })
  .passthrough();
const StackConnectionList = z
  .object({ items: z.array(StackConnection), total: z.number().int() })
  .partial()
  .passthrough();
const VolumeList = z
  .object({ items: z.array(Volume), total: z.number().int() })
  .partial()
  .passthrough();
const CreateReleaseRequest = z
  .object({ from_release_id: z.string() })
  .partial()
  .passthrough();
const ReleaseCauseKind = z.enum([
  "manual",
  "rollback",
  "webhook_push",
  "preview_sync",
]);
const ReleaseCause = z
  .object({ kind: ReleaseCauseKind, detail: z.string() })
  .partial()
  .passthrough();
const ResourcePins = z
  .object({
    git_sha: z.string(),
    volume_hash: z.string(),
    image_digest: z.string(),
  })
  .partial()
  .passthrough();
const ReleasePins = z
  .object({ resources: z.record(ResourcePins) })
  .partial()
  .passthrough();
const ResourceOutcome = z
  .object({
    phase: z.string(),
    ready_replicas: z.number().int(),
    replicas: z.number().int(),
    message: z.string(),
  })
  .partial()
  .passthrough();
const ReleaseOutcome = z
  .object({ resources: z.record(ResourceOutcome), duration: z.string() })
  .partial()
  .passthrough();
const ReleaseValidationError = z
  .object({
    resource_name: z.string(),
    field: z.string(),
    code: z.enum([
      "resource_name_required",
      "resource_name_invalid",
      "resource_name_duplicate",
      "source_required",
      "source_conflict",
      "workload_type_invalid",
      "schedule_required",
      "schedule_not_allowed",
      "schedule_invalid",
      "replicas_invalid",
      "ports_not_allowed",
      "public_port_not_http",
      "port_protocol_invalid",
      "port_name_invalid",
      "port_number_invalid",
      "port_name_duplicate",
      "port_number_duplicate",
      "subdomain_duplicate",
      "domain_not_configured",
      "env_name_required",
      "env_name_duplicate",
      "env_value_missing",
      "env_value_conflict",
      "env_self_output_unknown",
      "volume_mount_invalid",
      "volume_not_found",
      "volume_hash_missing",
      "secret_not_found",
      "git_integration_not_found",
      "registry_credential_not_found",
      "self_dependency",
      "duplicate_dependency",
      "unknown_dependency",
      "dependency_cycle",
      "git_repo_url_required",
      "git_branch_tag_conflict",
      "git_commit_invalid",
      "git_commit_requires_ref",
      "image_ref_required",
      "image_ref_invalid",
      "push_target_required",
      "push_target_conflict",
      "push_ref_invalid",
      "git_repo_unreachable",
      "git_auth_failed",
      "git_branch_not_found",
      "git_tag_not_found",
      "git_rate_limited",
      "image_not_found",
      "registry_credentials_required",
      "registry_auth_failed",
      "push_access_denied",
      "stack_name_invalid",
      "stack_settings_invalid",
      "connection_invalid",
    ]),
    message: z.string(),
  })
  .partial()
  .passthrough();
const Ingress = z
  .object({ url: z.string(), target_port: z.number().int() })
  .partial()
  .passthrough();
const ContainerFailureDetail = z
  .object({
    failure_type: z.enum([
      "crash_loop",
      "out_of_memory",
      "image_pull_failed",
      "create_container_error",
      "exit_error",
      "port_not_listening",
    ]),
    reason: z.string(),
    message: z.string(),
    restart_count: z.number().int(),
    exit_code: z.number().int(),
  })
  .partial()
  .passthrough();
const StackResourceFailure = z
  .object({
    type: z.enum(["runtime_crash", "build_failure", "readiness_failure"]),
    container: ContainerFailureDetail,
    init_container: ContainerFailureDetail,
    build: BuildFailureDetail,
  })
  .partial()
  .passthrough();
const StackResourceStatus = z
  .object({
    public_ingress: z.array(Ingress),
    internal_service_name: z.string(),
    last_restart_request_processed_at: z.string().datetime({ offset: true }),
    state: z.string(),
    message: z.string(),
    observed_revision: z.string(),
    conditions: z.array(Condition),
    last_failure: StackResourceFailure,
    replicas: z.number().int(),
    available_replicas: z.number().int(),
    updated_replicas: z.number().int(),
    last_run_time: z.string().datetime({ offset: true }),
    last_run_succeeded: z.boolean(),
  })
  .partial()
  .passthrough();
const ReleaseLiveStatus = z
  .object({
    health: ReleaseHealth,
    resources: z.record(StackResourceStatus),
    conditions: z.array(Condition),
    target_revision: z.string(),
    observed_revision: z.string(),
  })
  .partial()
  .passthrough();
const StackRelease = z
  .object({
    id: z.string(),
    stack_id: z.string(),
    sequence: z.number().int(),
    state: StackReleaseState,
    message: z.string(),
    cause: ReleaseCause,
    snapshot_revision: z.string(),
    manifest_revision: z.string(),
    renderer_version: z.string(),
    pins: ReleasePins,
    outcome: ReleaseOutcome,
    created_by: z.string(),
    created_at: z.string().datetime({ offset: true }),
    updated_at: z.string().datetime({ offset: true }),
    rendered_at: z.string().datetime({ offset: true }),
    completed_at: z.string().datetime({ offset: true }),
    validation_errors: z.array(ReleaseValidationError),
    live_status: ReleaseLiveStatus,
  })
  .partial()
  .passthrough();
const StackReleaseList = z
  .object({
    items: z.array(StackRelease),
    total: z.number().int(),
    page: z.number().int(),
    page_size: z.number().int(),
    total_pages: z.number().int(),
  })
  .partial()
  .passthrough();
const StackReleaseSnapshot = z
  .object({
    stack: z
      .object({
        id: z.string(),
        organisation_id: z.string(),
        project_id: z.string(),
        cluster_id: z.string(),
        user_id: z.string(),
        name: z.string(),
        namespace_id: z.string(),
        namespace: z.string(),
        labels: z.record(z.string()),
        annotations: z.record(z.string()),
      })
      .partial()
      .passthrough(),
    resources: z.array(StackResource),
    volumes: z.array(Volume),
    connections: z.array(StackConnection),
    captured_at: z.string().datetime({ offset: true }),
  })
  .partial()
  .passthrough();
const StackReleaseDetail = StackRelease.and(
  z.object({ snapshot: StackReleaseSnapshot }).partial().passthrough()
);
const ReleaseEventLink = z
  .object({ kind: z.string(), label: z.string(), target: z.record(z.string()) })
  .partial()
  .passthrough();
const ReleaseEvent = z
  .object({
    id: z.string(),
    release_id: z.string(),
    stack_id: z.string(),
    sequence: z.number().int(),
    occurred_at: z.string().datetime({ offset: true }),
    source: z.enum(["hub", "cluster"]),
    scope: z.enum(["release", "resource"]),
    resource_name: z.string(),
    type: z.string(),
    level: z.enum(["info", "success", "warning", "error"]),
    message: z.string(),
    links: z.array(ReleaseEventLink),
    metadata: z.record(z.string()),
  })
  .partial()
  .passthrough();
const ReleaseEventList = z
  .object({
    items: z.array(ReleaseEvent),
    next_after_sequence: z.number().int(),
  })
  .partial()
  .passthrough();
const postApiv1organizationsOrg_idprojectsProject_nameaddonspostgresIdactionsfence_Body =
  z.object({ fence: z.boolean(), reason: z.string().optional() }).passthrough();
const PostgresBackup = z
  .object({
    id: z.string(),
    postgres_addon_id: z.string(),
    name: z.string(),
    description: z.string(),
    type: z.enum(["scheduled", "manual", "pre_upgrade"]),
    phase: z.enum(["pending", "running", "completed", "failed"]),
    started_at: z.string().datetime({ offset: true }),
    completed_at: z.string().datetime({ offset: true }),
    error: z.string(),
    size_bytes: z.number().int(),
    created_at: z.string().datetime({ offset: true }),
  })
  .partial()
  .passthrough();
const PostgresBackupList = z
  .object({ items: z.array(PostgresBackup), total: z.number().int() })
  .partial()
  .passthrough();
const PostgresCredentials = z
  .object({
    database: z.string(),
    host: z.string(),
    port: z.number().int(),
    username: z.string(),
    password: z.string(),
    sslMode: z.string(),
    connectionString: z.string(),
    caCertificate: z.string(),
  })
  .partial()
  .passthrough();
const OrgInviteCreateRequest = z
  .object({
    email: z.string().email(),
    project_name: z.string(),
    role: z.enum(["Developer", "Viewer"]),
    expires_in_days: z.number().int().gte(1).lte(30),
  })
  .passthrough();
const InviteStatus = z.enum(["pending", "accepted", "revoked", "expired"]);
const OrgInviteCreateResponse = z
  .object({
    id: z.string(),
    email: z.string(),
    organisation_id: z.string(),
    project_name: z.string(),
    role: z.string(),
    status: InviteStatus,
    expires_at: z.string().datetime({ offset: true }),
    invited_by: z.string(),
    email_sent: z.boolean(),
    invite_token: z.string(),
    created_at: z.string().datetime({ offset: true }),
  })
  .partial()
  .passthrough();
const OrgInvite = z
  .object({
    id: z.string(),
    email: z.string(),
    organisation_id: z.string(),
    project_name: z.string(),
    role: z.enum(["Developer", "Viewer"]),
    status: InviteStatus,
    expires_at: z.string().datetime({ offset: true }),
    invited_by: z.string(),
    email_sent: z.boolean(),
    email_error: z.string(),
    created_at: z.string().datetime({ offset: true }),
    accepted_at: z.string().datetime({ offset: true }),
  })
  .partial()
  .passthrough();
const OrgInviteList = z
  .object({ items: z.array(OrgInvite), total: z.number().int() })
  .partial()
  .passthrough();
const OrgInviteInfo = z
  .object({
    org_name: z.string(),
    project_name: z.string(),
    inviter_name: z.string(),
    expires_at: z.string().datetime({ offset: true }),
  })
  .partial()
  .passthrough();
const PreviewGitRepository = z.object({
  repo_url: z.string(),
  base_branch: z.string().optional(),
  integration_id: z.string().optional(),
});
const StackPreviewConfigCreate = z
  .object({
    name: z.string(),
    git_repository: PreviewGitRepository,
    description: z.string().optional(),
    stackfile_path: z.string().optional(),
    max_active_previews: z.number().int().optional(),
    env: z.array(EnvVar).optional(),
    labels: z.array(Label).optional(),
    annotations: z.array(Annotation).optional(),
  })
  .passthrough();
const StackPreviewConfig = z
  .object({
    id: z.string(),
    organisation_id: z.string(),
    project_id: z.string(),
    user_id: z.string(),
    name: z.string(),
    description: z.string(),
    git_repository: PreviewGitRepository,
    stackfile_path: z.string(),
    max_active_previews: z.number().int(),
    env: z.array(EnvVar),
    labels: z.array(Label),
    annotations: z.array(Annotation),
    created_at: z.string().datetime({ offset: true }),
    updated_at: z.string().datetime({ offset: true }),
  })
  .partial()
  .passthrough();
const StackPreviewConfigList = z
  .object({
    items: z.array(StackPreviewConfig),
    total: z.number().int(),
    page: z.number().int(),
    page_size: z.number().int(),
    total_pages: z.number().int(),
  })
  .partial()
  .passthrough();
const StackPreviewConfigUpdate = z
  .object({
    description: z.string(),
    stackfile_path: z.string(),
    max_active_previews: z.number().int(),
    git_repository: PreviewGitRepository,
    env: z.array(EnvVar),
    labels: z.array(Label),
    annotations: z.array(Annotation),
  })
  .partial()
  .passthrough();
const PreviewStackCreate = z
  .object({
    config_id: z.string(),
    pr_number: z.string(),
    branch: z.string(),
    commit: z.string().optional(),
    stackfile_content: z.string().optional(),
    image_overrides: z.record(z.string()).optional(),
  })
  .passthrough();
const PreviewStack = z
  .object({
    id: z.string(),
    organisation_id: z.string(),
    project_id: z.string(),
    user_id: z.string(),
    config_id: z.string(),
    stack_id: z.string(),
    name: z.string(),
    pr_number: z.string(),
    branch: z.string(),
    commit: z.string(),
    source: z.enum(["manual", "webhook"]),
    status: z
      .object({
        phase: z.enum([
          "Provisioning",
          "Deploying",
          "Ready",
          "Failed",
          "Deleting",
        ]),
        reason: z.string(),
        message: z.string(),
        outputs: z
          .object({
            commit_sha: z.string(),
            urls: z.array(
              z
                .object({ resource: z.string(), url: z.string() })
                .partial()
                .passthrough()
            ),
          })
          .partial()
          .passthrough(),
      })
      .partial()
      .passthrough(),
    image_overrides: z.record(z.string()),
    labels: z.array(Label),
    annotations: z.array(Annotation),
    deletion_timestamp: z.string().datetime({ offset: true }),
    created_at: z.string().datetime({ offset: true }),
    updated_at: z.string().datetime({ offset: true }),
  })
  .partial()
  .passthrough();
const PreviewStackList = z
  .object({
    items: z.array(PreviewStack),
    total: z.number().int(),
    page: z.number().int(),
    page_size: z.number().int(),
    total_pages: z.number().int(),
  })
  .partial()
  .passthrough();
const PreviewStackSync = z
  .object({
    commit: z.string(),
    stackfile_content: z.string(),
    force_sync: z.boolean(),
    image_overrides: z.record(z.string()),
  })
  .partial()
  .passthrough();
const SSHConfig = z.object({ public_key: z.string() });
const FieldValidationError = z
  .object({
    field: z.string(),
    code: z.enum([
      "resource_name_required",
      "resource_name_invalid",
      "resource_name_duplicate",
      "source_required",
      "source_conflict",
      "workload_type_invalid",
      "schedule_required",
      "schedule_not_allowed",
      "schedule_invalid",
      "replicas_invalid",
      "ports_not_allowed",
      "public_port_not_http",
      "port_protocol_invalid",
      "port_name_invalid",
      "port_number_invalid",
      "port_name_duplicate",
      "port_number_duplicate",
      "subdomain_duplicate",
      "domain_not_configured",
      "env_name_required",
      "env_name_duplicate",
      "env_value_missing",
      "env_value_conflict",
      "env_self_output_unknown",
      "volume_mount_invalid",
      "volume_not_found",
      "volume_hash_missing",
      "secret_not_found",
      "git_integration_not_found",
      "registry_credential_not_found",
      "self_dependency",
      "duplicate_dependency",
      "unknown_dependency",
      "dependency_cycle",
      "git_repo_url_required",
      "git_branch_tag_conflict",
      "git_commit_invalid",
      "git_commit_requires_ref",
      "image_ref_required",
      "image_ref_invalid",
      "push_target_required",
      "push_target_conflict",
      "push_ref_invalid",
      "git_repo_unreachable",
      "git_auth_failed",
      "git_branch_not_found",
      "git_tag_not_found",
      "git_rate_limited",
      "image_not_found",
      "registry_credentials_required",
      "registry_auth_failed",
      "push_access_denied",
      "stack_name_invalid",
      "stack_settings_invalid",
      "connection_invalid",
    ]),
    message: z.string(),
  })
  .partial()
  .passthrough();
const ValidationErrorDetail = z
  .object({ errors: z.array(FieldValidationError) })
  .partial()
  .passthrough();
const List = z
  .object({
    kind: z.string(),
    page: z.number().int(),
    size: z.number().int(),
    total: z.number().int(),
  })
  .passthrough();
const ErrorList = List;
const WALConfiguration = z
  .object({ compression: z.enum(["gzip", "lz4", "zstd"]).default("gzip") })
  .partial()
  .passthrough();

export const schemas = {
  DomainName,
  Organisation,
  UserSignupRequest,
  UserRole,
  UserProjectMembership,
  User,
  UserSignupResponse,
  ObjectReference,
  Error,
  TurnstileConfigResponse,
  SignupConfigResponse,
  AppConfigResponse,
  Project,
  ProjectList,
  LoginRequest,
  LoginResponse,
  RefreshTokenRequest,
  RefreshTokenResponse,
  APITokenCreateRequest,
  APITokenCreateResponse,
  APIToken,
  APITokenList,
  ScopeResource,
  ScopeList,
  UserList,
  SecretType,
  SecretData,
  OutputDescriptor,
  Secret,
  SecretList,
  RegistryCredentialPurpose,
  RegistryCredential,
  RegistryCredentialList,
  AffectedStackRef,
  RegistryCredentialDeleteResponse,
  RegistryCredentialVerifyRequest,
  GitHubAppManifestFlow,
  GitInstallation,
  GitInstallationList,
  GitRepository,
  GitRepositoryPage,
  GitBranchList,
  GitIntegrationType,
  GitIntegrationBasicAuth,
  GitIntegrationAuth,
  GitIntegration,
  GitIntegrationList,
  GitIntegrationVerifyRequest,
  SecretReference,
  S3Credentials,
  AzureCredentials,
  GCSCredentials,
  ObjectStoreConfiguration,
  ObjectStoreSpec,
  ObjectStoreStatus,
  ObjectStore,
  ObjectStoreList,
  Label,
  Annotation,
  PushTarget,
  GitSource,
  ImageSource,
  VolumeBuildSource,
  SourceSpec,
  InitSpec,
  EnvVar,
  ExecutionConfig,
  VolumeMountSourceType,
  VolumeMount,
  LifecycleConfig,
  Port,
  StackResource,
  VolumeAccessMode,
  GitRepoRevision,
  GitRepoSource,
  VolumeSourceTypes,
  RemoteSource,
  BuildArtifact,
  VolumeSource,
  VolumeSpec,
  Condition,
  BuildArtifactSyncInfo,
  VolumeStatus,
  Volume,
  TopologyNodeRef,
  ConnectionTarget,
  OutputValueRef,
  ValueRef,
  ConnectionMapping,
  PostgresEnvConfig,
  VolumeMountConfig,
  BuildArtifactSourceConfig,
  StackConnectionConfig,
  StackConnection,
  StackSpec,
  StackSettings,
  StackLifecycle,
  StackReleaseState,
  ReleaseHealth,
  ReleaseSummary,
  Stack,
  StackList,
  PostgresVersion,
  PostgresInstances,
  PostgresStorage,
  PostgresResources,
  PostgresBackupConfig,
  PostgresInitialization,
  PostgresDatabase,
  PostgresConfiguration,
  PostgresAddonSpec,
  PostgresClusterInfo,
  PostgresConnectionInfo,
  PostgresAddonStatus,
  PostgresAddon,
  PostgresAddonList,
  ClusterImageRegistrySpec,
  ClusterImageRegistryState,
  ClusterImageRegistryStatus,
  ClusterImageRegistry,
  Cluster,
  ClusterList,
  ClusterImageRegistryList,
  ProjectCreateRequest,
  ProjectUpdateRequest,
  ProjectRole,
  AddProjectMemberRequest,
  ProjectMembership,
  ProjectMembershipList,
  UpdateProjectMemberRoleRequest,
  ProjectRoleList,
  PromoteAdminRequest,
  DemoteAdminRequest,
  ResourceMetrics,
  StackResourceList,
  BuildSourceRevision,
  BuildSourceContext,
  BuildFailureDetail,
  ImageBuildStatus,
  ImageBuild,
  ImageBuildList,
  TopologyNode,
  TopologyEdge,
  StackTopology,
  StackConnectionList,
  VolumeList,
  CreateReleaseRequest,
  ReleaseCauseKind,
  ReleaseCause,
  ResourcePins,
  ReleasePins,
  ResourceOutcome,
  ReleaseOutcome,
  ReleaseValidationError,
  Ingress,
  ContainerFailureDetail,
  StackResourceFailure,
  StackResourceStatus,
  ReleaseLiveStatus,
  StackRelease,
  StackReleaseList,
  StackReleaseSnapshot,
  StackReleaseDetail,
  ReleaseEventLink,
  ReleaseEvent,
  ReleaseEventList,
  postApiv1organizationsOrg_idprojectsProject_nameaddonspostgresIdactionsfence_Body,
  PostgresBackup,
  PostgresBackupList,
  PostgresCredentials,
  OrgInviteCreateRequest,
  InviteStatus,
  OrgInviteCreateResponse,
  OrgInvite,
  OrgInviteList,
  OrgInviteInfo,
  PreviewGitRepository,
  StackPreviewConfigCreate,
  StackPreviewConfig,
  StackPreviewConfigList,
  StackPreviewConfigUpdate,
  PreviewStackCreate,
  PreviewStack,
  PreviewStackList,
  PreviewStackSync,
  SSHConfig,
  FieldValidationError,
  ValidationErrorDetail,
  List,
  ErrorList,
  WALConfiguration,
};

const endpoints = makeApi([
  {
    method: "post",
    path: "/api/v1/api-tokens",
    alias: "postApiv1apiTokens",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: APITokenCreateRequest,
      },
    ],
    response: APITokenCreateResponse,
    errors: [
      {
        status: 400,
        description: `Invalid request data`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/api-tokens",
    alias: "getApiv1apiTokens",
    requestFormat: "json",
    response: APITokenList,
  },
  {
    method: "get",
    path: "/api/v1/api-tokens/:id",
    alias: "getApiv1apiTokensId",
    requestFormat: "json",
    parameters: [
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: APIToken,
    errors: [
      {
        status: 404,
        description: `API token not found`,
        schema: Error,
      },
    ],
  },
  {
    method: "delete",
    path: "/api/v1/api-tokens/:id",
    alias: "deleteApiv1apiTokensId",
    requestFormat: "json",
    parameters: [
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: z.void(),
    errors: [
      {
        status: 404,
        description: `API token not found`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/api-tokens/scopes",
    alias: "getApiv1apiTokensscopes",
    description: `Returns the list of resources and their allowed actions that can be used when creating API tokens`,
    requestFormat: "json",
    response: ScopeList,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/auth/github",
    alias: "getApiv1authgithub",
    description: `Redirects the user to GitHub for OAuth authorization`,
    requestFormat: "json",
    response: z.void(),
    errors: [
      {
        status: 302,
        description: `Redirect to GitHub OAuth authorization page`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/auth/github/callback",
    alias: "getApiv1authgithubcallback",
    description: `Handles the callback from GitHub after OAuth authorization`,
    requestFormat: "json",
    parameters: [
      {
        name: "code",
        type: "Query",
        schema: z.string(),
      },
      {
        name: "state",
        type: "Query",
        schema: z.string(),
      },
    ],
    response: LoginResponse,
    errors: [
      {
        status: 401,
        description: `OAuth authorization failed`,
        schema: Error,
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "post",
    path: "/api/v1/auth/login",
    alias: "postApiv1authlogin",
    description: `Authenticate user and generate an access token`,
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: LoginRequest,
      },
    ],
    response: LoginResponse,
    errors: [
      {
        status: 401,
        description: `Invalid credentials`,
        schema: Error,
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "post",
    path: "/api/v1/auth/refresh",
    alias: "postApiv1authrefresh",
    description: `Exchange a refresh token for a new access token`,
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: z.object({ refreshToken: z.string() }).passthrough(),
      },
    ],
    response: RefreshTokenResponse,
    errors: [
      {
        status: 401,
        description: `Invalid or expired refresh token`,
        schema: Error,
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/config",
    alias: "getApiv1config",
    description: `Returns feature flags the web client needs before authentication, such as whether GitHub OAuth is enabled.`,
    requestFormat: "json",
    response: AppConfigResponse,
    errors: [
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/git-integrations/github/manifest/callback",
    alias: "getApiv1gitIntegrationsgithubmanifestcallback",
    requestFormat: "json",
    parameters: [
      {
        name: "code",
        type: "Query",
        schema: z.string(),
      },
      {
        name: "state",
        type: "Query",
        schema: z.string(),
      },
    ],
    response: z.void(),
    errors: [
      {
        status: 302,
        description: `Redirects the browser to the GitHub App install page`,
        schema: z.void(),
      },
      {
        status: 400,
        description: `Invalid or expired state`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/git-integrations/github/setup",
    alias: "getApiv1gitIntegrationsgithubsetup",
    requestFormat: "json",
    parameters: [
      {
        name: "installation_id",
        type: "Query",
        schema: z.number().int(),
      },
      {
        name: "state",
        type: "Query",
        schema: z.string(),
      },
    ],
    response: z.void(),
    errors: [
      {
        status: 302,
        description: `Redirects the browser back to the git integrations page`,
        schema: z.void(),
      },
      {
        status: 400,
        description: `Invalid or expired state`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `The installation was not found on the platform app`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/invites/:token/info",
    alias: "getApiv1invitesTokeninfo",
    requestFormat: "json",
    parameters: [
      {
        name: "token",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: OrgInviteInfo,
    errors: [
      {
        status: 404,
        description: `Invite not found`,
        schema: z.void(),
      },
      {
        status: 410,
        description: `Invite expired or revoked`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:id",
    alias: "getApiv1organizationsId",
    requestFormat: "json",
    parameters: [
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: Organisation,
    errors: [
      {
        status: 401,
        description: `Auth token is invalid`,
        schema: Error,
      },
      {
        status: 403,
        description: `Unauthorized to perform operation`,
        schema: Error,
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "put",
    path: "/api/v1/organizations/:id",
    alias: "putApiv1organizationsId",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: Organisation,
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: Organisation,
    errors: [
      {
        status: 401,
        description: `Auth token is invalid`,
        schema: Error,
      },
      {
        status: 403,
        description: `Unauthorized to perform operation`,
        schema: Error,
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "post",
    path: "/api/v1/organizations/:org_id/admins",
    alias: "postApiv1organizationsOrg_idadmins",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: z.object({ user_id: z.string() }).passthrough(),
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: z.void(),
    errors: [
      {
        status: 400,
        description: `Invalid request data`,
        schema: Error,
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/admins",
    alias: "getApiv1organizationsOrg_idadmins",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: UserList,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "post",
    path: "/api/v1/organizations/:org_id/admins/:user_id/demote",
    alias: "postApiv1organizationsOrg_idadminsUser_iddemote",
    description: `Demotes an OrgAdmin and places them in the specified project with the given role (defaults to Viewer).`,
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: DemoteAdminRequest,
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "user_id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: z.void(),
    errors: [
      {
        status: 400,
        description: `Bad request`,
        schema: z.void(),
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `User not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "post",
    path: "/api/v1/organizations/:org_id/clusters",
    alias: "postApiv1organizationsOrg_idclusters",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: Cluster,
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: Cluster,
    errors: [
      {
        status: 400,
        description: `Invalid request payload`,
        schema: z.void(),
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/clusters",
    alias: "getApiv1organizationsOrg_idclusters",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: ClusterList,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "post",
    path: "/api/v1/organizations/:org_id/clusters/:cluster_id/image_registries",
    alias: "postApiv1organizationsOrg_idclustersCluster_idimage_registries",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: ClusterImageRegistry,
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "cluster_id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: ClusterImageRegistry,
    errors: [
      {
        status: 400,
        description: `Invalid request payload`,
        schema: z.void(),
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/clusters/:cluster_id/image_registries",
    alias: "getApiv1organizationsOrg_idclustersCluster_idimage_registries",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "cluster_id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: ClusterImageRegistryList,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/clusters/:cluster_id/image_registries/:id",
    alias: "getApiv1organizationsOrg_idclustersCluster_idimage_registriesId",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "cluster_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: ClusterImageRegistry,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `ImageRegistry not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "delete",
    path: "/api/v1/organizations/:org_id/clusters/:cluster_id/image_registries/:id",
    alias: "deleteApiv1organizationsOrg_idclustersCluster_idimage_registriesId",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "cluster_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: z.void(),
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `ImageRegistry not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/clusters/:id",
    alias: "getApiv1organizationsOrg_idclustersId",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: Cluster,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Cluster not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "delete",
    path: "/api/v1/organizations/:org_id/clusters/:id",
    alias: "deleteApiv1organizationsOrg_idclustersId",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: z.void(),
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Cluster not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "post",
    path: "/api/v1/organizations/:org_id/git-integrations",
    alias: "postApiv1organizationsOrg_idgitIntegrations",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: GitIntegration,
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: GitIntegration,
    errors: [
      {
        status: 400,
        description: `Invalid request payload`,
        schema: z.void(),
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 409,
        description: `An integration for this host already exists`,
        schema: Error,
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/git-integrations",
    alias: "getApiv1organizationsOrg_idgitIntegrations",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: GitIntegrationList,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/git-integrations/:id",
    alias: "getApiv1organizationsOrg_idgitIntegrationsId",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: GitIntegration,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Git integration not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "put",
    path: "/api/v1/organizations/:org_id/git-integrations/:id",
    alias: "putApiv1organizationsOrg_idgitIntegrationsId",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: GitIntegration,
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: GitIntegration,
    errors: [
      {
        status: 400,
        description: `Invalid request payload`,
        schema: z.void(),
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Git integration not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "delete",
    path: "/api/v1/organizations/:org_id/git-integrations/:id",
    alias: "deleteApiv1organizationsOrg_idgitIntegrationsId",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: z.void(),
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Git integration not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/git-integrations/:id/installations",
    alias: "getApiv1organizationsOrg_idgitIntegrationsIdinstallations",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "refresh",
        type: "Query",
        schema: z.boolean().optional().default(false),
      },
    ],
    response: GitInstallationList,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Git integration not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/git-integrations/:id/repositories",
    alias: "getApiv1organizationsOrg_idgitIntegrationsIdrepositories",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "page",
        type: "Query",
        schema: z.number().int().optional(),
      },
      {
        name: "installation_id",
        type: "Query",
        schema: z.string().optional(),
      },
    ],
    response: GitRepositoryPage,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Git integration not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/git-integrations/:id/repositories/:owner/:repo",
    alias: "getApiv1organizationsOrg_idgitIntegrationsIdrepositoriesOwnerRepo",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "owner",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "repo",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: GitRepository,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/git-integrations/:id/repositories/:owner/:repo/branches",
    alias:
      "getApiv1organizationsOrg_idgitIntegrationsIdrepositoriesOwnerRepobranches",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "owner",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "repo",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: GitBranchList,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "post",
    path: "/api/v1/organizations/:org_id/git-integrations/:id/verify",
    alias: "postApiv1organizationsOrg_idgitIntegrationsIdverify",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: z.object({ repo_url: z.string() }),
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: z.void(),
    errors: [
      {
        status: 400,
        description: `Verification failed`,
        schema: Error,
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Git integration not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "post",
    path: "/api/v1/organizations/:org_id/git-integrations/github/manifest",
    alias: "postApiv1organizationsOrg_idgitIntegrationsgithubmanifest",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: GitHubAppManifestFlow,
    errors: [
      {
        status: 400,
        description: `The hub external URL is not configured`,
        schema: z.void(),
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 409,
        description: `A GitHub App is already installed`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/image_registries",
    alias: "getApiv1organizationsOrg_idimage_registries",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: ClusterImageRegistryList,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "post",
    path: "/api/v1/organizations/:org_id/invites",
    alias: "postApiv1organizationsOrg_idinvites",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: OrgInviteCreateRequest,
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: OrgInviteCreateResponse,
    errors: [
      {
        status: 400,
        description: `Invalid request data`,
        schema: Error,
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 409,
        description: `Conflict - user exists or duplicate pending invite`,
        schema: Error,
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/invites",
    alias: "getApiv1organizationsOrg_idinvites",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "status",
        type: "Query",
        schema: z.string().optional(),
      },
    ],
    response: OrgInviteList,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/invites/:id",
    alias: "getApiv1organizationsOrg_idinvitesId",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: OrgInvite,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Invite not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "delete",
    path: "/api/v1/organizations/:org_id/invites/:id",
    alias: "deleteApiv1organizationsOrg_idinvitesId",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: z.void(),
    errors: [
      {
        status: 400,
        description: `Can only revoke pending invites`,
        schema: z.void(),
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Invite not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "post",
    path: "/api/v1/organizations/:org_id/invites/:id/resend",
    alias: "postApiv1organizationsOrg_idinvitesIdresend",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: z.void(),
    errors: [
      {
        status: 400,
        description: `Can only resend pending invites`,
        schema: z.void(),
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Invite not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/object-stores",
    alias: "getApiv1organizationsOrg_idobjectStores",
    description: `Returns object stores from all projects the user belongs to. OrgAdmins see all object stores in the org.`,
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: ObjectStoreList,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/postgres-addons",
    alias: "getApiv1organizationsOrg_idpostgresAddons",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: PostgresAddonList,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "post",
    path: "/api/v1/organizations/:org_id/projects",
    alias: "postApiv1organizationsOrg_idprojects",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: z.object({ name: z.string() }).passthrough(),
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: Project,
    errors: [
      {
        status: 400,
        description: `Invalid request data`,
        schema: Error,
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects",
    alias: "getApiv1organizationsOrg_idprojects",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: ProjectList,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name",
    alias: "getApiv1organizationsOrg_idprojectsProject_name",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: Project,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Project not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "put",
    path: "/api/v1/organizations/:org_id/projects/:project_name",
    alias: "putApiv1organizationsOrg_idprojectsProject_name",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: z.object({ name: z.string() }).passthrough(),
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: Project,
    errors: [
      {
        status: 400,
        description: `Invalid request data`,
        schema: Error,
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Project not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "delete",
    path: "/api/v1/organizations/:org_id/projects/:project_name",
    alias: "deleteApiv1organizationsOrg_idprojectsProject_name",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: z.void(),
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Project not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "post",
    path: "/api/v1/organizations/:org_id/projects/:project_name/addons/postgres",
    alias: "postApiv1organizationsOrg_idprojectsProject_nameaddonspostgres",
    description: `Create a new PostgreSQL database cluster addon`,
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: PostgresAddon,
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: PostgresAddon,
    errors: [
      {
        status: 400,
        description: `Invalid request data`,
        schema: Error,
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 409,
        description: `PostgresAddon already exists`,
        schema: Error,
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name/addons/postgres",
    alias: "getApiv1organizationsOrg_idprojectsProject_nameaddonspostgres",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: PostgresAddonList,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name/addons/postgres/:id",
    alias: "getApiv1organizationsOrg_idprojectsProject_nameaddonspostgresId",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: PostgresAddon,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `PostgresAddon not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "put",
    path: "/api/v1/organizations/:org_id/projects/:project_name/addons/postgres/:id",
    alias: "putApiv1organizationsOrg_idprojectsProject_nameaddonspostgresId",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: PostgresAddon,
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: PostgresAddon,
    errors: [
      {
        status: 400,
        description: `Invalid request data`,
        schema: Error,
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `PostgresAddon not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "delete",
    path: "/api/v1/organizations/:org_id/projects/:project_name/addons/postgres/:id",
    alias: "deleteApiv1organizationsOrg_idprojectsProject_nameaddonspostgresId",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: PostgresAddon,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `PostgresAddon not found`,
        schema: z.void(),
      },
      {
        status: 409,
        description: `PostgreSQL addon is in use and cannot be deleted`,
        schema: Error,
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "post",
    path: "/api/v1/organizations/:org_id/projects/:project_name/addons/postgres/:id/actions/backup",
    alias:
      "postApiv1organizationsOrg_idprojectsProject_nameaddonspostgresIdactionsbackup",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: z.object({ description: z.string() }).partial().passthrough(),
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: z
      .object({ message: z.string(), backup_id: z.string() })
      .partial()
      .passthrough(),
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `PostgresAddon not found`,
        schema: z.void(),
      },
      {
        status: 409,
        description: `Backup already in progress`,
        schema: Error,
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "post",
    path: "/api/v1/organizations/:org_id/projects/:project_name/addons/postgres/:id/actions/fence",
    alias:
      "postApiv1organizationsOrg_idprojectsProject_nameaddonspostgresIdactionsfence",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema:
          postApiv1organizationsOrg_idprojectsProject_nameaddonspostgresIdactionsfence_Body,
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: z
      .object({ message: z.string(), fenced: z.boolean() })
      .partial()
      .passthrough(),
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `PostgresAddon not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "post",
    path: "/api/v1/organizations/:org_id/projects/:project_name/addons/postgres/:id/actions/hibernate",
    alias:
      "postApiv1organizationsOrg_idprojectsProject_nameaddonspostgresIdactionshibernate",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: z.object({ hibernate: z.boolean() }).passthrough(),
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: z
      .object({ message: z.string(), hibernated: z.boolean() })
      .partial()
      .passthrough(),
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `PostgresAddon not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name/addons/postgres/:id/backups",
    alias:
      "getApiv1organizationsOrg_idprojectsProject_nameaddonspostgresIdbackups",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "limit",
        type: "Query",
        schema: z.number().int().optional().default(20),
      },
      {
        name: "offset",
        type: "Query",
        schema: z.number().int().optional().default(0),
      },
    ],
    response: PostgresBackupList,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `PostgresAddon not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name/addons/postgres/:id/credentials/:database",
    alias:
      "getApiv1organizationsOrg_idprojectsProject_nameaddonspostgresIdcredentialsDatabase",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "database",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "superuser",
        type: "Query",
        schema: z.boolean().optional().default(false),
      },
    ],
    response: PostgresCredentials,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: Error,
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: Error,
      },
      {
        status: 404,
        description: `Addon or database not found`,
        schema: Error,
      },
      {
        status: 503,
        description: `Addon not ready or cluster unreachable`,
        schema: Error,
      },
    ],
  },
  {
    method: "post",
    path: "/api/v1/organizations/:org_id/projects/:project_name/members",
    alias: "postApiv1organizationsOrg_idprojectsProject_namemembers",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: AddProjectMemberRequest,
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: ProjectMembership,
    errors: [
      {
        status: 400,
        description: `Invalid request data`,
        schema: Error,
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Project not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name/members",
    alias: "getApiv1organizationsOrg_idprojectsProject_namemembers",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: ProjectMembershipList,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Project not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "put",
    path: "/api/v1/organizations/:org_id/projects/:project_name/members/:id",
    alias: "putApiv1organizationsOrg_idprojectsProject_namemembersId",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: UpdateProjectMemberRoleRequest,
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: ProjectMembership,
    errors: [
      {
        status: 400,
        description: `Invalid request data`,
        schema: Error,
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Membership not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "delete",
    path: "/api/v1/organizations/:org_id/projects/:project_name/members/:id",
    alias: "deleteApiv1organizationsOrg_idprojectsProject_namemembersId",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: z.void(),
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Membership not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "post",
    path: "/api/v1/organizations/:org_id/projects/:project_name/object-stores",
    alias: "postApiv1organizationsOrg_idprojectsProject_nameobjectStores",
    description: `Add a new ObjectStore configuration for storing PostgreSQL backups and WAL files`,
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: ObjectStore,
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: ObjectStore,
    errors: [
      {
        status: 400,
        description: `Invalid request data`,
        schema: Error,
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 409,
        description: `ObjectStore already exists`,
        schema: Error,
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name/object-stores",
    alias: "getApiv1organizationsOrg_idprojectsProject_nameobjectStores",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: ObjectStoreList,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name/object-stores/:id",
    alias: "getApiv1organizationsOrg_idprojectsProject_nameobjectStoresId",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: ObjectStore,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `ObjectStore not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "put",
    path: "/api/v1/organizations/:org_id/projects/:project_name/object-stores/:id",
    alias: "putApiv1organizationsOrg_idprojectsProject_nameobjectStoresId",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: ObjectStore,
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: ObjectStore,
    errors: [
      {
        status: 400,
        description: `Invalid request data`,
        schema: Error,
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `ObjectStore not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "delete",
    path: "/api/v1/organizations/:org_id/projects/:project_name/object-stores/:id",
    alias: "deleteApiv1organizationsOrg_idprojectsProject_nameobjectStoresId",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: z.void(),
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `ObjectStore not found`,
        schema: z.void(),
      },
      {
        status: 409,
        description: `ObjectStore is in use by PostgresAddons`,
        schema: Error,
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "post",
    path: "/api/v1/organizations/:org_id/projects/:project_name/preview-stacks",
    alias: "createPreviewStack",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: PreviewStackCreate,
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: PreviewStack,
    errors: [
      {
        status: 400,
        description: `Invalid request data`,
        schema: Error,
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 409,
        description: `Preview stack already exists`,
        schema: Error,
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name/preview-stacks",
    alias: "listPreviewStacks",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "page",
        type: "Query",
        schema: z.number().int().optional().default(1),
      },
      {
        name: "page_size",
        type: "Query",
        schema: z.number().int().optional().default(20),
      },
      {
        name: "config_id",
        type: "Query",
        schema: z.string().optional(),
      },
    ],
    response: PreviewStackList,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name/preview-stacks/:id",
    alias: "getPreviewStack",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: PreviewStack,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Preview stack not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "delete",
    path: "/api/v1/organizations/:org_id/projects/:project_name/preview-stacks/:id",
    alias: "deletePreviewStack",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: PreviewStack,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Preview stack not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "post",
    path: "/api/v1/organizations/:org_id/projects/:project_name/preview-stacks/:id/sync",
    alias: "syncPreviewStack",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: PreviewStackSync,
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: PreviewStack,
    errors: [
      {
        status: 400,
        description: `Invalid request data`,
        schema: Error,
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Preview stack not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "post",
    path: "/api/v1/organizations/:org_id/projects/:project_name/secrets",
    alias: "postApiv1organizationsOrg_idprojectsProject_namesecrets",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: Secret,
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: Secret,
    errors: [
      {
        status: 400,
        description: `Invalid request payload`,
        schema: z.void(),
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name/secrets",
    alias: "getApiv1organizationsOrg_idprojectsProject_namesecrets",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "name",
        type: "Query",
        schema: z.string().optional(),
      },
    ],
    response: SecretList,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name/secrets/:id",
    alias: "getApiv1organizationsOrg_idprojectsProject_namesecretsId",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: Secret,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Secret not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "put",
    path: "/api/v1/organizations/:org_id/projects/:project_name/secrets/:id",
    alias: "putApiv1organizationsOrg_idprojectsProject_namesecretsId",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: Secret,
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: Secret,
    errors: [
      {
        status: 400,
        description: `Invalid request payload`,
        schema: z.void(),
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Secret not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "delete",
    path: "/api/v1/organizations/:org_id/projects/:project_name/secrets/:id",
    alias: "deleteApiv1organizationsOrg_idprojectsProject_namesecretsId",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: z.void(),
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Secret not found`,
        schema: z.void(),
      },
      {
        status: 409,
        description: `Secret is in use and cannot be deleted`,
        schema: Error,
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "post",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stack-preview-configs",
    alias: "createPreviewConfig",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: StackPreviewConfigCreate,
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: StackPreviewConfig,
    errors: [
      {
        status: 400,
        description: `Invalid request data`,
        schema: Error,
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 409,
        description: `Preview config already exists`,
        schema: Error,
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stack-preview-configs",
    alias: "listPreviewConfigs",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "page",
        type: "Query",
        schema: z.number().int().optional().default(1),
      },
      {
        name: "page_size",
        type: "Query",
        schema: z.number().int().optional().default(20),
      },
    ],
    response: StackPreviewConfigList,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stack-preview-configs/:id",
    alias: "getPreviewConfig",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: StackPreviewConfig,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Preview config not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "put",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stack-preview-configs/:id",
    alias: "updatePreviewConfig",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: StackPreviewConfigUpdate,
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: StackPreviewConfig,
    errors: [
      {
        status: 400,
        description: `Invalid request data`,
        schema: Error,
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Preview config not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "delete",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stack-preview-configs/:id",
    alias: "deletePreviewConfig",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: z.void(),
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Preview config not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "post",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks",
    alias: "postApiv1organizationsOrg_idprojectsProject_namestacks",
    description: `Creates a thin stack shell (name, labels, annotations, settings). Any inline
&#x60;stack_resources&#x60;, &#x60;volumes&#x60;, or &#x60;connections&#x60; in the body are ignored — add
children via &#x60;PUT /stacks/{id}/apply&#x60; or the individual sub-resource endpoints.
`,
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: Stack,
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: Stack,
    errors: [
      {
        status: 400,
        description: `Invalid request data`,
        schema: Error,
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 409,
        description: `Stack already exists`,
        schema: Error,
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks",
    alias: "getApiv1organizationsOrg_idprojectsProject_namestacks",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "limit",
        type: "Query",
        schema: z.number().int().optional().default(20),
      },
      {
        name: "offset",
        type: "Query",
        schema: z.number().int().optional().default(0),
      },
    ],
    response: StackList,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks/:id",
    alias: "getApiv1organizationsOrg_idprojectsProject_namestacksId",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: Stack,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "put",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks/:id",
    alias: "putApiv1organizationsOrg_idprojectsProject_namestacksId",
    description: `Updates only shell fields (name, labels, annotations, settings). &#x60;namespace&#x60; is
immutable. Child collections (&#x60;stack_resources&#x60;, &#x60;volumes&#x60;, &#x60;connections&#x60;) in the
body are ignored — use &#x60;PUT /stacks/{id}/apply&#x60; for a full reconcile.
`,
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: Stack,
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: Stack,
    errors: [
      {
        status: 400,
        description: `Invalid request data`,
        schema: Error,
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "delete",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks/:id",
    alias: "deleteApiv1organizationsOrg_idprojectsProject_namestacksId",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: Stack,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "put",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks/:id/apply",
    alias: "applyStack",
    description: `Declarative whole-document apply. Reconciles the stack against the supplied
document: resources and connections not present in the body are deleted, while
volumes are add-only and are never deleted. This is the only endpoint that
accepts a full stack document.
`,
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: Stack,
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: Stack,
    errors: [
      {
        status: 400,
        description: `Invalid request data. &#x60;details&#x60; carries a &#x60;ValidationErrorDetail&#x60; payload when the failure is an aggregated field validation error.`,
        schema: Error,
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks/:id/builds",
    alias: "getApiv1organizationsOrg_idprojectsProject_namestacksIdbuilds",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: ImageBuildList,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks/:id/builds/:build_id",
    alias:
      "getApiv1organizationsOrg_idprojectsProject_namestacksIdbuildsBuild_id",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "build_id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: ImageBuild,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Build not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks/:id/builds/:build_id/logs",
    alias:
      "getApiv1organizationsOrg_idprojectsProject_namestacksIdbuildsBuild_idlogs",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "build_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "follow",
        type: "Query",
        schema: z.boolean().optional().default(false),
      },
      {
        name: "tail",
        type: "Query",
        schema: z.number().int().optional().default(200),
      },
      {
        name: "since",
        type: "Query",
        schema: z.string().optional(),
      },
    ],
    response: z.void(),
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Build not found`,
        schema: z.void(),
      },
      {
        status: 409,
        description: `Build job not created yet, or build pod not started — retry later`,
        schema: Error,
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks/:id/connections",
    alias: "getApiv1organizationsOrg_idprojectsProject_namestacksIdconnections",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: StackConnectionList,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Stack not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "post",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks/:id/connections",
    alias:
      "postApiv1organizationsOrg_idprojectsProject_namestacksIdconnections",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: StackConnection,
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: StackConnection,
    errors: [
      {
        status: 400,
        description: `Invalid request data`,
        schema: Error,
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Stack not found`,
        schema: z.void(),
      },
      {
        status: 409,
        description: `Stack connection already exists`,
        schema: Error,
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "put",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks/:id/connections/:connection_id",
    alias:
      "putApiv1organizationsOrg_idprojectsProject_namestacksIdconnectionsConnection_id",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: StackConnection,
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "connection_id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: StackConnection,
    errors: [
      {
        status: 400,
        description: `Invalid request data`,
        schema: Error,
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Stack or connection not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "delete",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks/:id/connections/:connection_id",
    alias:
      "deleteApiv1organizationsOrg_idprojectsProject_namestacksIdconnectionsConnection_id",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "connection_id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: z.void(),
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Stack or connection not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks/:id/logs",
    alias: "getApiv1organizationsOrg_idprojectsProject_namestacksIdlogs",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "follow",
        type: "Query",
        schema: z.boolean().optional().default(false),
      },
      {
        name: "tail",
        type: "Query",
        schema: z.number().int().optional().default(100),
      },
      {
        name: "since",
        type: "Query",
        schema: z.string().optional(),
      },
    ],
    response: z.void(),
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks/:id/metrics",
    alias: "getApiv1organizationsOrg_idprojectsProject_namestacksIdmetrics",
    description: `Returns metrics for a stack. If &#x60;stream&#x3D;true&#x60; is passed, the server responds using Server-Sent Events (SSE).
`,
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "stream",
        type: "Query",
        schema: z.boolean().optional().default(false),
      },
    ],
    response: ResourceMetrics,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "post",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks/:id/releases",
    alias: "createRelease",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: z
          .object({ from_release_id: z.string() })
          .partial()
          .passthrough(),
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: StackRelease,
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks/:id/releases",
    alias: "listReleases",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "state",
        type: "Query",
        schema: z
          .enum([
            "Pending",
            "InProgress",
            "Released",
            "Failed",
            "Superseded",
            "Cancelled",
          ])
          .optional(),
      },
      {
        name: "page",
        type: "Query",
        schema: z.number().int().optional().default(1),
      },
      {
        name: "page_size",
        type: "Query",
        schema: z.number().int().optional().default(20),
      },
    ],
    response: StackReleaseList,
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks/:id/releases/:release_id",
    alias: "getRelease",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "release_id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: StackReleaseDetail,
  },
  {
    method: "post",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks/:id/releases/:release_id/cancel",
    alias: "cancelRelease",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "release_id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: z.void(),
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks/:id/releases/:release_id/events",
    alias: "listReleaseEvents",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "release_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "after_sequence",
        type: "Query",
        schema: z.number().int().optional().default(0),
      },
      {
        name: "limit",
        type: "Query",
        schema: z.number().int().lte(500).optional().default(100),
      },
    ],
    response: ReleaseEventList,
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks/:id/releases/:release_id/events/stream",
    alias: "streamReleaseEvents",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "release_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "after_sequence",
        type: "Query",
        schema: z.number().int().optional().default(0),
      },
    ],
    response: z.void(),
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks/:id/resources",
    alias: "getApiv1organizationsOrg_idprojectsProject_namestacksIdresources",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: StackResourceList,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "post",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks/:id/resources",
    alias: "createStackResource",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: StackResource,
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: StackResource,
    errors: [
      {
        status: 400,
        description: `Invalid request data. &#x60;details&#x60; carries a &#x60;ValidationErrorDetail&#x60; payload when the failure is an aggregated field validation error.`,
        schema: Error,
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Stack not found`,
        schema: z.void(),
      },
      {
        status: 409,
        description: `Stack resource already exists`,
        schema: Error,
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks/:id/resources/:resource_name",
    alias:
      "getApiv1organizationsOrg_idprojectsProject_namestacksIdresourcesResource_name",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "resource_name",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: StackResource,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "put",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks/:id/resources/:resource_name",
    alias: "updateStackResource",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: StackResource,
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "resource_name",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: StackResource,
    errors: [
      {
        status: 400,
        description: `Invalid request data. &#x60;details&#x60; carries a &#x60;ValidationErrorDetail&#x60; payload when the failure is an aggregated field validation error.`,
        schema: Error,
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Stack or resource not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "delete",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks/:id/resources/:resource_name",
    alias: "deleteStackResource",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "resource_name",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: z.void(),
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Stack or resource not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "post",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks/:id/resources/:resource_name/actions/restart",
    alias:
      "postApiv1organizationsOrg_idprojectsProject_namestacksIdresourcesResource_nameactionsrestart",
    description: `Triggers a rolling restart of the stack resource by setting a new restart request timestamp.`,
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "resource_name",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: StackResource,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Stack resource not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks/:id/resources/:resource_name/builds",
    alias:
      "getApiv1organizationsOrg_idprojectsProject_namestacksIdresourcesResource_namebuilds",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "resource_name",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: ImageBuildList,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks/:id/resources/:resource_name/logs",
    alias:
      "getApiv1organizationsOrg_idprojectsProject_namestacksIdresourcesResource_namelogs",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "resource_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "follow",
        type: "Query",
        schema: z.boolean().optional().default(false),
      },
      {
        name: "tail",
        type: "Query",
        schema: z.number().int().optional().default(100),
      },
      {
        name: "since",
        type: "Query",
        schema: z.string().optional(),
      },
    ],
    response: z.void(),
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks/:id/resources/:resource_name/metrics",
    alias:
      "getApiv1organizationsOrg_idprojectsProject_namestacksIdresourcesResource_namemetrics",
    description: `Returns metrics for a stack resource. If &#x60;stream&#x3D;true&#x60; is passed, the server responds using Server-Sent Events (SSE).
`,
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "resource_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "stream",
        type: "Query",
        schema: z.boolean().optional().default(false),
      },
    ],
    response: ResourceMetrics,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks/:id/topology",
    alias: "getApiv1organizationsOrg_idprojectsProject_namestacksIdtopology",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: StackTopology,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Stack not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks/:id/volumes",
    alias: "getApiv1organizationsOrg_idprojectsProject_namestacksIdvolumes",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: VolumeList,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Stack not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "post",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks/:id/volumes",
    alias: "postApiv1organizationsOrg_idprojectsProject_namestacksIdvolumes",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: Volume,
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: Volume,
    errors: [
      {
        status: 400,
        description: `Invalid request payload`,
        schema: z.void(),
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Stack not found`,
        schema: z.void(),
      },
      {
        status: 409,
        description: `A volume with this name already exists in the stack`,
        schema: Error,
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "put",
    path: "/api/v1/organizations/:org_id/projects/:project_name/stacks/apply",
    alias: "applyStackByName",
    description: `Name-addressed declarative whole-document apply. Stack identity is the
&#x60;name&#x60; in the request body (unique per project). If a stack with that name
exists in the project it is reconciled exactly like the id-addressed apply
(resources and connections not present in the body are deleted, volumes
are add-only); otherwise the stack and its children are created
atomically after full validation. Idempotent — clients need not know
whether the stack already exists.
`,
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: Stack,
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: Stack,
    errors: [
      {
        status: 400,
        description: `Invalid request data. &#x60;details&#x60; carries a &#x60;ValidationErrorDetail&#x60; payload when the failure is an aggregated field validation error.`,
        schema: Error,
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/projects/:project_name/volumes/:id",
    alias: "getApiv1organizationsOrg_idprojectsProject_namevolumesId",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: Volume,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "delete",
    path: "/api/v1/organizations/:org_id/projects/:project_name/volumes/:id",
    alias: "deleteApiv1organizationsOrg_idprojectsProject_namevolumesId",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "project_name",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: z.void(),
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Volume not found`,
        schema: z.void(),
      },
      {
        status: 409,
        description: `Volume is in use and cannot be deleted`,
        schema: Error,
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "post",
    path: "/api/v1/organizations/:org_id/registry-credentials",
    alias: "postApiv1organizationsOrg_idregistryCredentials",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: RegistryCredential,
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: RegistryCredential,
    errors: [
      {
        status: 400,
        description: `Invalid request payload`,
        schema: z.void(),
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 409,
        description: `A credential for this host and purpose already exists`,
        schema: Error,
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/registry-credentials",
    alias: "getApiv1organizationsOrg_idregistryCredentials",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: RegistryCredentialList,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/registry-credentials/:id",
    alias: "getApiv1organizationsOrg_idregistryCredentialsId",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: RegistryCredential,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Registry credential not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "put",
    path: "/api/v1/organizations/:org_id/registry-credentials/:id",
    alias: "putApiv1organizationsOrg_idregistryCredentialsId",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: RegistryCredential,
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: RegistryCredential,
    errors: [
      {
        status: 400,
        description: `Invalid request payload`,
        schema: z.void(),
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Registry credential not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "delete",
    path: "/api/v1/organizations/:org_id/registry-credentials/:id",
    alias: "deleteApiv1organizationsOrg_idregistryCredentialsId",
    description: `Deletion is never blocked; the response lists stacks that were implicitly resolving against this credential so rotations are visible.`,
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: RegistryCredentialDeleteResponse,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Registry credential not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "post",
    path: "/api/v1/organizations/:org_id/registry-credentials/:id/verify",
    alias: "postApiv1organizationsOrg_idregistryCredentialsIdverify",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: RegistryCredentialVerifyRequest,
      },
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: z.void(),
    errors: [
      {
        status: 400,
        description: `Verification failed`,
        schema: Error,
      },
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `Registry credential not found`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/secrets",
    alias: "getApiv1organizationsOrg_idsecrets",
    description: `Returns secrets from all projects the user belongs to. OrgAdmins see all secrets in the org.`,
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "name",
        type: "Query",
        schema: z.string().optional(),
      },
    ],
    response: SecretList,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/stacks",
    alias: "getApiv1organizationsOrg_idstacks",
    description: `Returns stacks from all projects the user belongs to. OrgAdmins see all stacks in the org.`,
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: StackList,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/users",
    alias: "getApiv1organizationsOrg_idusers",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "page",
        type: "Query",
        schema: z.number().int().optional().default(1),
      },
      {
        name: "page_size",
        type: "Query",
        schema: z.number().int().optional().default(20),
      },
    ],
    response: UserList,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 403,
        description: `Forbidden`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/project-roles",
    alias: "getApiv1projectRoles",
    description: `Returns the list of roles that can be assigned to project members.`,
    requestFormat: "json",
    response: ProjectRoleList,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "post",
    path: "/api/v1/user-signup",
    alias: "postApiv1userSignup",
    description: `Create a new user`,
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: UserSignupRequest,
      },
    ],
    response: UserSignupResponse,
    errors: [
      {
        status: 400,
        description: `Invalid request data`,
        schema: Error,
      },
      {
        status: 409,
        description: `User already exists`,
        schema: Error,
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/users/:id",
    alias: "getApiv1usersId",
    description: `Get a user`,
    requestFormat: "json",
    parameters: [
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: User,
    errors: [
      {
        status: 401,
        description: `Auth token is invalid`,
        schema: Error,
      },
      {
        status: 403,
        description: `Unauthorized to perform operation`,
        schema: Error,
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/users/current",
    alias: "getApiv1userscurrent",
    requestFormat: "json",
    response: User,
    errors: [
      {
        status: 401,
        description: `Auth token is invalid`,
        schema: Error,
      },
      {
        status: 403,
        description: `Unauthorized to perform operation`,
        schema: Error,
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/users/current/projects",
    alias: "getApiv1userscurrentprojects",
    requestFormat: "json",
    response: ProjectList,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: Error,
      },
    ],
  },
  {
    method: "post",
    path: "/api/v1/webhooks/github",
    alias: "postApiv1webhooksgithub",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: z.object({}).partial().passthrough(),
      },
    ],
    response: z.void(),
    errors: [
      {
        status: 403,
        description: `Signature verification failed`,
        schema: z.void(),
      },
      {
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
]);

export const api = new Zodios(endpoints);

export function createApiClient(baseUrl: string, options?: ZodiosOptions) {
  return new Zodios(baseUrl, endpoints, options);
}

import { makeApi, Zodios, type ZodiosOptions } from "@zodios/core";
import { z } from "zod";

const UserRole = z.enum(["OrganisationAdmin", "PlatformAdmin", "User"]);
const DomainName = z
  .object({ id: z.string(), fqdn: z.string() })
  .partial()
  .passthrough();
const Organisation = z
  .object({
    id: z.string(),
    name: z.string(),
    domains: z.array(DomainName),
    is_default: z.boolean(),
    created_at: z.string().datetime({ offset: true }),
    updated_at: z.string().datetime({ offset: true }),
  })
  .partial()
  .passthrough();
const UserSignupRequest = z
  .object({
    name: z.string(),
    username: z.string().optional(),
    email: z.string().email(),
    password: z.string(),
    Role: UserRole.optional(),
    organisation: Organisation.optional(),
    organisation_id: z.string().optional(),
  })
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
  })
  .partial()
  .passthrough();
const UserSignupResponse = z
  .object({ user: User, jwt_token: z.string() })
  .partial()
  .passthrough();
const ObjectReference = z
  .object({ id: z.string(), kind: z.string(), href: z.string() })
  .partial()
  .passthrough();
const Error = ObjectReference.and(
  z
    .object({ code: z.string(), reason: z.string(), operation_id: z.string() })
    .partial()
    .passthrough()
);
const LoginRequest = z
  .object({ email: z.string(), password: z.string() })
  .passthrough();
const LoginResponse = z
  .object({ token: z.string(), user: User, expires_in: z.number().int() })
  .partial()
  .passthrough();
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
const WorkspaceUserStatus = z
  .object({
    observed_version: z.number().int(),
    provisioned_namespaces: z.array(
      z
        .object({ workspace_name: z.string(), namespace: z.string() })
        .partial()
        .passthrough()
    ),
    service_account_name: z.string().nullable(),
    serviceaccount_token: z.string().nullable(),
    cluster_ca_cert: z.string().nullable(),
    cluster_url: z.string().nullable(),
    conditions: z.array(Condition),
  })
  .partial()
  .passthrough();
const WorkspaceUserState = z.enum(["Completed", "Error", "Pending"]);
const WorkspaceUser = z
  .object({
    id: z.string().uuid().optional(),
    user_id: z.string().optional(),
    org_id: z.string().optional(),
    workspaces: z.array(z.string()).min(1),
    version: z.number().int().optional(),
    status: WorkspaceUserStatus.optional(),
    state: WorkspaceUserState.optional(),
    message: z.string().optional(),
    created_at: z.string().datetime({ offset: true }).optional(),
    updated_at: z.string().datetime({ offset: true }).optional(),
  })
  .passthrough();
const ClusterImageRegistrySpec = z
  .object({
    backend_storage_size: z.string(),
    backend_storage_class: z.string(),
    max_repositories: z.number().int(),
    tags_per_repository: z.number().int(),
    delete_untagged: z.boolean(),
  })
  .partial()
  .passthrough();
const ClusterImageRegistryState = z.enum([
  "ImageRegistryPending",
  "ImageRegistryError",
  "ImageRegistryRunning",
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
    default: z.boolean().optional(),
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
const Secret = z.object({
  id: z.string().optional(),
  name: z.string(),
  description: z.string().optional(),
  organisation_id: z.string().optional(),
  type: SecretType,
  data: z.array(SecretData),
  created_at: z.string().datetime({ offset: true }).optional(),
  updated_at: z.string().datetime({ offset: true }).optional(),
});
const SecretList = z
  .object({ items: z.array(Secret), total: z.number().int() })
  .partial()
  .passthrough();
const Label = z.object({ key: z.string(), value: z.string() });
const Annotation = z
  .object({ key: z.string(), value: z.string() })
  .passthrough();
const RemoteSyncServerSpec = z.object({
  volumeName: z.string().optional(),
  ssh_public_key: z.string().optional(),
});
const RemoteSyncServerState = z.enum([
  "RemoteSyncServerPending",
  "RemoteSyncServerCreating",
  "RemoteSyncServerCreated",
  "RemoteSyncServerReady",
  "RemoteSyncServerFailed",
]);
const RemoteSyncServerStatus = z
  .object({
    observed_version: z.number().int(),
    conditions: z.array(Condition),
    state: RemoteSyncServerState,
    serviceName: z.string(),
  })
  .partial()
  .passthrough();
const RemoteSyncServer = z.object({
  id: z.string().optional(),
  organisation_id: z.string().optional(),
  name: z.string(),
  namespace: z.string().optional(),
  labels: z.array(Label).optional(),
  annotations: z.array(Annotation).optional(),
  spec: RemoteSyncServerSpec,
  status: RemoteSyncServerStatus.optional(),
  created_at: z.string().datetime({ offset: true }).optional(),
  updated_at: z.string().datetime({ offset: true }).optional(),
});
const VolumeAccessMode = z.enum([
  "ReadWriteOnce",
  "ReadWriteMany",
  "ReadOnlyMany",
]);
const GitRepoRevision = z
  .object({
    branch: z
      .object({ name: z.string(), head_sha: z.string() })
      .partial()
      .passthrough(),
    commit: z.string(),
    tag: z.string(),
  })
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
  name: z.string(),
  labels: z.array(Label).optional(),
  annotations: z.array(Annotation).optional(),
  spec: VolumeSpec,
  status: VolumeStatus.optional(),
});
const VolumeList = z
  .object({ items: z.array(Volume), total: z.number().int() })
  .partial()
  .passthrough();
const SecretRef = z.object({ secret_id: z.string() });
const BuildSourceContext = z
  .object({
    volume: z
      .object({ id: z.string(), name: z.string().optional() })
      .passthrough(),
    git_repo: z
      .object({ repo_url: z.string(), git_secret: SecretRef.optional() })
      .passthrough(),
  })
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
const ImageRepository = z
  .object({
    external_image_repo_url: z.string(),
    use_internal_registry: z.boolean(),
  })
  .partial();
const StackResourceBuildSpec = z.object({
  source_context: BuildSourceContext,
  context_path_within_source: z.string(),
  dockerfile_path: z.string(),
  source_revision: BuildSourceRevision,
  image_repository: ImageRepository,
  registry_push_secret: SecretRef.optional(),
});
const ImageSpec = z.object({
  image: z.string(),
  pull_secret: SecretRef.optional(),
});
const InitSpec = z
  .object({
    image_spec: ImageSpec,
    command: z.array(z.string()),
    args: z.array(z.string()),
  })
  .partial();
const EnvVar = z.object({ name: z.string(), value: z.string() });
const EnvVarFromSecret = z.object({
  name: z.string(),
  secret_ref: SecretRef,
  key: z.string(),
});
const PostgresAddonEnvSource = z
  .object({
    addon_id: z.string(),
    database: z.string().optional(),
    superuser: z.boolean().optional().default(false),
    env_mapping: z.record(z.string()),
  })
  .passthrough();
const AddonEnvSource = z
  .object({ postgres: PostgresAddonEnvSource })
  .partial()
  .passthrough();
const ExecutionConfig = z
  .object({
    command: z.array(z.string()),
    args: z.array(z.string()),
    environment_variables: z.array(EnvVar),
    environment_variables_from_secret: z.array(EnvVarFromSecret),
    env_from_addons: z.array(AddonEnvSource),
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
    number: z.number().int(),
    protocol: z.string().optional(),
    exposed_to_public: z.boolean(),
    subdomain_prefix: z.string().optional(),
  })
  .passthrough();
const Ingress = z
  .object({ url: z.string(), target_port: z.number().int() })
  .partial()
  .passthrough();
const StackResourceStatus = z
  .object({
    public_ingress: z.array(Ingress),
    internal_service_name: z.string(),
    last_restart_request_processed_at: z.string().datetime({ offset: true }),
    state: z.string(),
    observed_revision: z.string(),
    conditions: z.array(Condition),
  })
  .partial()
  .passthrough();
const StackResource = z
  .object({
    id: z.string().optional(),
    stack_id: z.string().optional(),
    name: z.string(),
    labels: z.array(Label).optional(),
    annotations: z.array(Annotation).optional(),
    revision: z.string().optional(),
    build_spec: StackResourceBuildSpec.optional(),
    image_spec: ImageSpec.optional(),
    init_spec: InitSpec.optional(),
    execution_config: ExecutionConfig.optional(),
    volume_mounts: z.array(VolumeMount).optional(),
    depends_on: z.array(z.string()).optional(),
    lifecycle_config: LifecycleConfig.optional(),
    ports: z.array(Port).optional(),
    stateful: z.boolean().optional(),
    status: StackResourceStatus.optional(),
  })
  .passthrough();
const StackSpec = z
  .object({
    stack_resources: z.array(StackResource),
    volumes: z.array(Volume).optional(),
  })
  .passthrough();
const StackStatus = z
  .object({
    state: z.string(),
    message: z.string(),
    observed_revision: z.string(),
    conditions: z.array(Condition),
  })
  .partial()
  .passthrough();
const Stack = z
  .object({
    id: z.string().optional(),
    organisation_id: z.string().optional(),
    user_id: z.string().optional(),
    name: z.string(),
    namespace: z.string().optional(),
    labels: z.array(Label).optional(),
    annotations: z.array(Annotation).optional(),
    revision: z.string().optional(),
    spec: StackSpec,
    status: StackStatus.optional(),
    created_at: z.string().datetime({ offset: true }).optional(),
    updated_at: z.string().datetime({ offset: true }).optional(),
  })
  .passthrough();
const StackList = z
  .object({ items: z.array(Stack), total: z.number().int() })
  .partial()
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
const ImageBuildStatus = z
  .object({
    state: z.string(),
    conditions: z.array(Condition),
    image_url: z.string(),
    build_source_revision: z.string(),
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
    storage_class: z.string(),
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
    user_id: z.string().optional(),
    cluster_id: z.string().optional(),
    name: z.string(),
    namespace: z.string().optional(),
    labels: z.array(Label).optional(),
    annotations: z.array(Annotation).optional(),
    revision: z.string().optional(),
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
const postApiv1organizationsOrg_idaddonspostgresIdactionsfence_Body = z
  .object({ fence: z.boolean(), reason: z.string().optional() })
  .passthrough();
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
const RemoteSyncServerList = z
  .object({ items: z.array(RemoteSyncServer), total: z.number().int() })
  .partial()
  .passthrough();
const SSHConfig = z.object({ public_key: z.string() });
const ClusterImageRegistryList = z
  .object({ items: z.array(ClusterImageRegistry), total: z.number().int() })
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
  UserRole,
  DomainName,
  Organisation,
  UserSignupRequest,
  User,
  UserSignupResponse,
  ObjectReference,
  Error,
  LoginRequest,
  LoginResponse,
  Condition,
  WorkspaceUserStatus,
  WorkspaceUserState,
  WorkspaceUser,
  ClusterImageRegistrySpec,
  ClusterImageRegistryState,
  ClusterImageRegistryStatus,
  ClusterImageRegistry,
  Cluster,
  ClusterList,
  SecretType,
  SecretData,
  Secret,
  SecretList,
  Label,
  Annotation,
  RemoteSyncServerSpec,
  RemoteSyncServerState,
  RemoteSyncServerStatus,
  RemoteSyncServer,
  VolumeAccessMode,
  GitRepoRevision,
  GitRepoSource,
  VolumeSourceTypes,
  RemoteSource,
  BuildArtifact,
  VolumeSource,
  VolumeSpec,
  BuildArtifactSyncInfo,
  VolumeStatus,
  Volume,
  VolumeList,
  SecretRef,
  BuildSourceContext,
  BuildSourceRevision,
  ImageRepository,
  StackResourceBuildSpec,
  ImageSpec,
  InitSpec,
  EnvVar,
  EnvVarFromSecret,
  PostgresAddonEnvSource,
  AddonEnvSource,
  ExecutionConfig,
  VolumeMountSourceType,
  VolumeMount,
  LifecycleConfig,
  Port,
  Ingress,
  StackResourceStatus,
  StackResource,
  StackSpec,
  StackStatus,
  Stack,
  StackList,
  ResourceMetrics,
  StackResourceList,
  ImageBuildStatus,
  ImageBuild,
  ImageBuildList,
  SecretReference,
  S3Credentials,
  AzureCredentials,
  GCSCredentials,
  ObjectStoreConfiguration,
  ObjectStoreSpec,
  ObjectStoreStatus,
  ObjectStore,
  ObjectStoreList,
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
  postApiv1organizationsOrg_idaddonspostgresIdactionsfence_Body,
  PostgresBackup,
  PostgresBackupList,
  PostgresCredentials,
  RemoteSyncServerList,
  SSHConfig,
  ClusterImageRegistryList,
  List,
  ErrorList,
  WALConfiguration,
};

const endpoints = makeApi([
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
    path: "/api/v1/organizations",
    alias: "postApiv1organizations",
    description: `Create a new organization`,
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: Organisation,
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
    method: "get",
    path: "/api/v1/organizations/:id",
    alias: "getApiv1organizationsId",
    description: `Get an organization`,
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
    description: `Update an organization`,
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
    path: "/api/v1/organizations/:id/clusters",
    alias: "postApiv1organizationsIdclusters",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: Cluster,
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
    path: "/api/v1/organizations/:id/clusters",
    alias: "getApiv1organizationsIdclusters",
    requestFormat: "json",
    parameters: [
      {
        name: "id",
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
    path: "/api/v1/organizations/:id/remote-sync-servers",
    alias: "postApiv1organizationsIdremoteSyncServers",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: RemoteSyncServer,
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: RemoteSyncServer,
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
    path: "/api/v1/organizations/:id/remote-sync-servers",
    alias: "getApiv1organizationsIdremoteSyncServers",
    requestFormat: "json",
    parameters: [
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: z
      .object({ items: z.array(RemoteSyncServer), total: z.number().int() })
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
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "post",
    path: "/api/v1/organizations/:org_id/addons/postgres",
    alias: "postApiv1organizationsOrg_idaddonspostgres",
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
    path: "/api/v1/organizations/:org_id/addons/postgres",
    alias: "getApiv1organizationsOrg_idaddonspostgres",
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
    method: "get",
    path: "/api/v1/organizations/:org_id/addons/postgres/:id",
    alias: "getApiv1organizationsOrg_idaddonspostgresId",
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
    path: "/api/v1/organizations/:org_id/addons/postgres/:id",
    alias: "putApiv1organizationsOrg_idaddonspostgresId",
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
    path: "/api/v1/organizations/:org_id/addons/postgres/:id",
    alias: "deleteApiv1organizationsOrg_idaddonspostgresId",
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
    method: "post",
    path: "/api/v1/organizations/:org_id/addons/postgres/:id/actions/backup",
    alias: "postApiv1organizationsOrg_idaddonspostgresIdactionsbackup",
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
    path: "/api/v1/organizations/:org_id/addons/postgres/:id/actions/fence",
    alias: "postApiv1organizationsOrg_idaddonspostgresIdactionsfence",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: postApiv1organizationsOrg_idaddonspostgresIdactionsfence_Body,
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
    path: "/api/v1/organizations/:org_id/addons/postgres/:id/actions/hibernate",
    alias: "postApiv1organizationsOrg_idaddonspostgresIdactionshibernate",
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
    path: "/api/v1/organizations/:org_id/addons/postgres/:id/backups",
    alias: "getApiv1organizationsOrg_idaddonspostgresIdbackups",
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
    path: "/api/v1/organizations/:org_id/addons/postgres/:id/credentials/:database",
    alias: "getApiv1organizationsOrg_idaddonspostgresIdcredentialsDatabase",
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
    response: z
      .object({ items: z.array(ClusterImageRegistry), total: z.number().int() })
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
    method: "put",
    path: "/api/v1/organizations/:org_id/clusters/:id",
    alias: "putApiv1organizationsOrg_idclustersId",
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
      {
        name: "id",
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
    path: "/api/v1/organizations/:org_id/object-stores",
    alias: "postApiv1organizationsOrg_idobjectStores",
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
    path: "/api/v1/organizations/:org_id/object-stores",
    alias: "getApiv1organizationsOrg_idobjectStores",
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
    path: "/api/v1/organizations/:org_id/object-stores/:id",
    alias: "getApiv1organizationsOrg_idobjectStoresId",
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
    path: "/api/v1/organizations/:org_id/object-stores/:id",
    alias: "putApiv1organizationsOrg_idobjectStoresId",
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
    path: "/api/v1/organizations/:org_id/object-stores/:id",
    alias: "deleteApiv1organizationsOrg_idobjectStoresId",
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
    method: "get",
    path: "/api/v1/organizations/:org_id/remote-sync-servers/:id",
    alias: "getApiv1organizationsOrg_idremoteSyncServersId",
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
    response: RemoteSyncServer,
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
        description: `RemoteSyncServer not found`,
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
    path: "/api/v1/organizations/:org_id/remote-sync-servers/:id",
    alias: "putApiv1organizationsOrg_idremoteSyncServersId",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: RemoteSyncServer,
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
    response: RemoteSyncServer,
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
        description: `RemoteSyncServer not found`,
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
    path: "/api/v1/organizations/:org_id/remote-sync-servers/:id",
    alias: "deleteApiv1organizationsOrg_idremoteSyncServersId",
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
        description: `RemoteSyncServer not found`,
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
    path: "/api/v1/organizations/:org_id/remote-sync-servers/current",
    alias: "getApiv1organizationsOrg_idremoteSyncServerscurrent",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: RemoteSyncServer,
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
    path: "/api/v1/organizations/:org_id/secrets",
    alias: "postApiv1organizationsOrg_idsecrets",
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
    path: "/api/v1/organizations/:org_id/secrets",
    alias: "getApiv1organizationsOrg_idsecrets",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
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
    method: "put",
    path: "/api/v1/organizations/:org_id/secrets/:id",
    alias: "putApiv1organizationsOrg_idsecretsId",
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
    method: "get",
    path: "/api/v1/organizations/:org_id/secrets/:id",
    alias: "getApiv1organizationsOrg_idsecretsId",
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
    method: "delete",
    path: "/api/v1/organizations/:org_id/secrets/:id",
    alias: "deleteApiv1organizationsOrg_idsecretsId",
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
    method: "post",
    path: "/api/v1/organizations/:org_id/stacks",
    alias: "postApiv1organizationsOrg_idstacks",
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
    path: "/api/v1/organizations/:org_id/stacks",
    alias: "getApiv1organizationsOrg_idstacks",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
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
    path: "/api/v1/organizations/:org_id/stacks/:id",
    alias: "getApiv1organizationsOrg_idstacksId",
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
    path: "/api/v1/organizations/:org_id/stacks/:id",
    alias: "putApiv1organizationsOrg_idstacksId",
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
    path: "/api/v1/organizations/:org_id/stacks/:id",
    alias: "deleteApiv1organizationsOrg_idstacksId",
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
    method: "get",
    path: "/api/v1/organizations/:org_id/stacks/:id/logs",
    alias: "getApiv1organizationsOrg_idstacksIdlogs",
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
    path: "/api/v1/organizations/:org_id/stacks/:id/metrics",
    alias: "getApiv1organizationsOrg_idstacksIdmetrics",
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
    method: "get",
    path: "/api/v1/organizations/:org_id/stacks/:stack_id/builds",
    alias: "getApiv1organizationsOrg_idstacksStack_idbuilds",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "stack_id",
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
    path: "/api/v1/organizations/:org_id/stacks/:stack_id/builds/:build_id",
    alias: "getApiv1organizationsOrg_idstacksStack_idbuildsBuild_id",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "stack_id",
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
    path: "/api/v1/organizations/:org_id/stacks/:stack_id/resources",
    alias: "getApiv1organizationsOrg_idstacksStack_idresources",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "stack_id",
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
    method: "put",
    path: "/api/v1/organizations/:org_id/stacks/:stack_id/resources/:id",
    alias: "putApiv1organizationsOrg_idstacksStack_idresourcesId",
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
        name: "stack_id",
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
    method: "get",
    path: "/api/v1/organizations/:org_id/stacks/:stack_id/resources/:id",
    alias: "getApiv1organizationsOrg_idstacksStack_idresourcesId",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "stack_id",
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
    path: "/api/v1/organizations/:org_id/stacks/:stack_id/resources/:id/builds",
    alias: "getApiv1organizationsOrg_idstacksStack_idresourcesIdbuilds",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "stack_id",
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
    path: "/api/v1/organizations/:org_id/stacks/:stack_id/resources/:resource_name/logs",
    alias:
      "getApiv1organizationsOrg_idstacksStack_idresourcesResource_namelogs",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "stack_id",
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
    path: "/api/v1/organizations/:org_id/stacks/:stack_id/resources/:resource_name/metrics",
    alias:
      "getApiv1organizationsOrg_idstacksStack_idresourcesResource_namemetrics",
    description: `Returns metrics for a StackResource. If &#x60;stream&#x3D;true&#x60; is passed, the server responds using Server-Sent Events (SSE).
`,
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
        type: "Path",
        schema: z.string(),
      },
      {
        name: "stack_id",
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
    path: "/api/v1/organizations/:org_id/stacks/current",
    alias: "getApiv1organizationsOrg_idstackscurrent",
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
    method: "post",
    path: "/api/v1/organizations/:org_id/volumes",
    alias: "postApiv1organizationsOrg_idvolumes",
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
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/volumes",
    alias: "getApiv1organizationsOrg_idvolumes",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
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
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/:org_id/volumes/:id",
    alias: "getApiv1organizationsOrg_idvolumesId",
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
    method: "put",
    path: "/api/v1/organizations/:org_id/volumes/:id",
    alias: "putApiv1organizationsOrg_idvolumesId",
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
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "delete",
    path: "/api/v1/organizations/:org_id/volumes/:id",
    alias: "deleteApiv1organizationsOrg_idvolumesId",
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
        description: `Volume not found`,
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
    path: "/api/v1/organizations/:org_id/volumes/current",
    alias: "getApiv1organizationsOrg_idvolumescurrent",
    requestFormat: "json",
    parameters: [
      {
        name: "org_id",
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
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/organizations/default",
    alias: "getApiv1organizationsdefault",
    description: `Get the default organization`,
    requestFormat: "json",
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
    description: `Get a the current authenticated user`,
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
    method: "post",
    path: "/api/v1/workspace-users",
    alias: "postApiv1workspaceUsers",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: WorkspaceUser,
      },
    ],
    response: WorkspaceUser,
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
        status: 500,
        description: `Internal server error`,
        schema: z.void(),
      },
    ],
  },
  {
    method: "get",
    path: "/api/v1/workspace-users/:id",
    alias: "getApiv1workspaceUsersId",
    requestFormat: "json",
    parameters: [
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: WorkspaceUser,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `WorkspaceUser object not found`,
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
    path: "/api/v1/workspace-users/:id",
    alias: "putApiv1workspaceUsersId",
    requestFormat: "json",
    parameters: [
      {
        name: "body",
        type: "Body",
        schema: WorkspaceUser,
      },
      {
        name: "id",
        type: "Path",
        schema: z.string(),
      },
    ],
    response: WorkspaceUser,
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
        status: 404,
        description: `WorkspaceUser not found`,
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
    path: "/api/v1/workspace-users/:id",
    alias: "deleteApiv1workspaceUsersId",
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
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `WorkspaceUser not found`,
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
    path: "/api/v1/workspace-users/current",
    alias: "getApiv1workspaceUserscurrent",
    requestFormat: "json",
    response: WorkspaceUser,
    errors: [
      {
        status: 401,
        description: `Unauthorized`,
        schema: z.void(),
      },
      {
        status: 404,
        description: `WorkspaceUser object not found`,
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

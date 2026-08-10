# Backend Glossary

The API-server hub (`cmd/`, `pkg/`) is a Go control plane that reconciles tenant
workloads onto remote Kubernetes clusters. Terms below are the canonical domain
nouns; prefer them over synonyms in code, APIs, and docs.

## Tenancy & identity

| term | definition | source |
|---|---|---|
| Organisation | The top-level tenant that owns all projects, clusters, and resources; the billing and isolation boundary. | `pkg/models/organisation.go` |
| Project | A named group within an Organisation that owns stacks, secrets, object stores, and addons; the access-scoping unit (one may be the DefaultProject). | `pkg/models/project.go` |
| ProjectMembership | A user's assignment to a Project with a ProjectRole (e.g. Viewer); the join between User and Project. | `pkg/models/project.go` |
| ProjectRole | The permission level a member holds within a Project. | `pkg/models/project.go` |
| User | An authentication identity that can belong to projects and act within an organisation; carries a UserRole. | `pkg/models/user.go` |
| OrgInvite | A pending, tokenised invitation for an email to join an Organisation and Project at a given ProjectRole. | `pkg/models/org_invite.go` |
| OrganisationDomain | A DNS domain claimed and verified by an Organisation for use by stack domains. | `pkg/models/organisation_domain.go` |

## Auth tokens

| term | definition | source |
|---|---|---|
| APIToken | A hashed, scoped, optionally resource-restricted token a user creates for programmatic API access. | `pkg/models/api_token.go` |
| RefreshToken | A hashed long-lived token exchanged for a new short-lived access token. | `pkg/models/refresh_token.go` |
| OAuthState | A short-lived anti-CSRF state value for an external OAuth (e.g. GitHub) authorization round-trip. | `pkg/models/oauth_state.go` |
| Scope | A capability grant attached to an APIToken limiting which actions/resources it may use. | `pkg/models/api_token.go` |

## Workloads — stacks

| term | definition | source |
|---|---|---|
| Stack | A project-owned deployable application unit (a set of resources) reconciled onto a target cluster; the primary workload aggregate. | `pkg/models/stack.go`, `pkg/controllers/stack` |
| StackResource | An individual deployable component within a Stack (e.g. a service) with its own ports, dependencies, and status. | `pkg/models/stack_resource.go`, `pkg/controllers/stackresource` |
| StackStatus | The reconciled state of a Stack: state, conditions, observed CR revision, last validation run. | `pkg/models/stack.go` |
| ValidationRun | The outcome of running validation checks against a Stack revision before/while reconciling. | `pkg/models/stack.go` |
| StackDomain | A custom domain bound to a Stack so its public ingress is reachable at that hostname. | `pkg/models/stack_domain.go` |
| StackStorage | Persistent storage attached to a Stack and reconciled onto the cluster. | `pkg/models/stack_storage.go`, `pkg/controllers/stackstorage` |
| Ingress | A public URL/target-port pairing exposing a StackResource externally. | `pkg/models/stack_resource.go` |
| Port | A network port declaration on a StackResource, optionally exposed to the public with a subdomain prefix. | `pkg/models/stack_resource.go` |

## Builds & images

| term | definition | source |
|---|---|---|
| ImageBuild | A reconciled job that builds a container image from a source context (git, volume) and pushes it to a registry. | `pkg/models/image_build.go`, `pkg/controllers/imagebuild` |
| BuildConfigSpec | The build specification: source context, dockerfile path, source revision, target image repository. | `pkg/models/image_build.go` |
| ClusterImageRegistry | An in-cluster container image registry provisioned for a Cluster within an Organisation, with backing storage. | `pkg/models/cluster_image_registry.go`, `pkg/controllers/clusterimageregistry` |
| RegistryState | The lifecycle state of a ClusterImageRegistry. | `pkg/models/cluster_image_registry.go` |

## Storage & volumes

| term | definition | source |
|---|---|---|
| Volume | A persistent volume owned by a project/org, sized and class-bound, optionally hydrated from a source (remote dir, build artifact, git repo). | `pkg/models/volume.go`, `pkg/controllers/volume` |
| VolumeMount | A binding that mounts a Volume into a workload at a target path from a source sub-path. | `pkg/models/volume_mount.go` |
| ObjectStore | A project-owned external object-storage configuration (S3 / Azure / GCS) used to store backups and WAL files. | `pkg/models/object_store.go` |
| Namespace | A reconciled Kubernetes namespace allocated for a tenant's workloads on a cluster. | `pkg/models/namespace.go` |

## Secrets

| term | definition | source |
|---|---|---|
| Secret | A project-owned, typed bag of key→value sensitive data referenced by stacks, addons, and object stores. | `pkg/models/secret.go` |
| SecretReference | A pointer to one key within a Secret (secret_id + key); how other resources consume secret data. | `pkg/models/secret.go` |
| SecretType | The classification of a Secret that governs its expected keys and use. | `pkg/models/secret.go` |

## Addons (managed PostgreSQL)

| term | definition | source |
|---|---|---|
| Addon | A managed backing service provisioned for a project (currently PostgreSQL). | `pkg/models/addon.go` |
| PostgresAddon | A managed PostgreSQL database cluster addon: version, instances, placement, databases, backup config. | `pkg/models/postgres_addon.go`, `pkg/controllers/postgres_addon` |
| PostgresBackup | A backup of a PostgresAddon (scheduled or on-demand) used for restore. | `pkg/models/postgres_addon.go`, `pkg/controllers/postgres_backup` |
| PostgresInstances | The replica count and placement (topology, tolerations, node selector) of a PostgresAddon. | `pkg/models/postgres_addon.go` |
| AddonUsage | A record of an Addon's consumption/metering for an organisation. | `pkg/models/addon_usage.go` |

## Cluster control plane

| term | definition | source |
|---|---|---|
| Cluster | A registered remote Kubernetes cluster onto which the hub reconciles tenant resources; holds (encrypted) connection credentials. | `pkg/models/cluster.go` |
| Reconciler / Controller | A backend control loop that drives a model's desired state onto a cluster and reports observed status (one per resource type). | `pkg/controllers/*` |
| DeployedClusterInfo | The per-cluster placement/status of a resource (e.g. an ObjectStore) across the clusters it is deployed to. | `pkg/models/object_store.go` |
| Condition | A standard status sub-record (type/state/message) attached to a resource's reconciled status. | `pkg/models/meta.go` |

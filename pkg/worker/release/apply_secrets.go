package release

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Stackdome/stackdome/pkg/credentials"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stackdeploy"
	corev1 "k8s.io/api/core/v1"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// syncHubSecrets ensures image pull, image push, and git credential secrets
// are synced to the target cluster namespace before the stack CR is applied.
// Registry credentials resolve through the credential resolver: explicit
// secret refs sync from the backing hub secret, org-level registry credentials
// are synthesized directly from resolver output, anonymous targets sync
// nothing.
func (r *applyReconciler) syncHubSecrets(ctx context.Context, clusterClient client.Client, stack *models.Stack) error {
	// desired accumulates the cluster secrets to write, keyed by name.
	// Multiple resources can resolve to the same secret name with different
	// dockerconfig auth entries (e.g. one hub secret used for a pull on host A
	// and a push on host B); those are merged rather than skipped, so no
	// needed auth entry is dropped.
	desired := make(map[string]*corev1.Secret)

	for _, resource := range stack.StackResources {
		if resource.ImageConfig == nil {
			continue
		}
		imageUrl := resource.ImageConfig.Image
		resolved, serr := r.credentialResolver.RegistryCredentials(ctx, stack.OrganisationID, imageUrl,
			credentials.RegistryPurposePull, credentials.RegistryAuthSelector{
				RegistryCredentialID: resource.ImageConfig.RegistryCredentialID,
			})
		if serr != nil {
			return fmt.Errorf("failed to resolve image pull credentials for '%s': %w", imageUrl, serr)
		}
		if err := r.collectRegistrySecret(ctx, stack, resolved, imageUrl, false, false, desired); err != nil {
			return fmt.Errorf("failed to build image pull secret for '%s': %w", imageUrl, err)
		}
	}

	for _, resource := range stack.StackResources {
		if resource.BuildConfig == nil {
			continue
		}
		repo := resource.BuildConfig.BuildImageRepository
		if repo.UseInClusterRegistry || repo.ExternalImageRef == "" {
			continue
		}
		resolved, serr := r.credentialResolver.RegistryCredentials(ctx, stack.OrganisationID, repo.ExternalImageRef,
			credentials.RegistryPurposePush, credentials.RegistryAuthSelector{
				RegistryCredentialID: resource.BuildConfig.PushRegistryCredentialID,
			})
		if serr != nil {
			return fmt.Errorf("failed to resolve image push credentials for '%s': %w", resource.Name, serr)
		}
		if err := r.collectRegistrySecret(ctx, stack, resolved, repo.ExternalImageRef, true, repo.InsecureRegistry, desired); err != nil {
			return fmt.Errorf("failed to build image push secret for '%s': %w", resource.Name, err)
		}
	}

	for _, resource := range stack.StackResources {
		if resource.BuildConfig == nil || resource.BuildConfig.SourceContext.Git == nil {
			continue
		}
		git := resource.BuildConfig.SourceContext.Git
		resolved, serr := r.credentialResolver.GitCredentials(ctx, stack.OrganisationID, git.RepoURL,
			credentials.GitAuthSelector{IntegrationID: git.IntegrationID})
		if serr != nil {
			return fmt.Errorf("failed to resolve git credentials for '%s': %w", git.RepoURL, serr)
		}
		if err := r.collectGitCredentialSecret(ctx, stack, resolved, git.RepoURL, desired); err != nil {
			return fmt.Errorf("failed to build git credential secret for '%s': %w", git.RepoURL, err)
		}
	}

	for _, secret := range desired {
		if err := createOrUpdateSecret(ctx, clusterClient, secret); err != nil {
			return fmt.Errorf("failed to sync secret '%s': %w", secret.Name, err)
		}
	}

	return nil
}

// collectGitCredentialSecret builds the cluster secret for resolved git
// credentials and accumulates it into desired. Explicit secret refs use the
// backing hub secret; org-level integrations (including minted GitHub App
// tokens) are synthesized from raw resolver output under a deterministic
// host-keyed name; anonymous resolutions contribute nothing. A given name
// always maps to the same content (same secret/integration/host), so a repeat
// is a no-op dedup.
func (r *applyReconciler) collectGitCredentialSecret(
	ctx context.Context,
	stack *models.Stack,
	resolved *credentials.ResolvedGitCredential,
	repoURL string,
	desired map[string]*corev1.Secret,
) error {
	var clusterSecret *corev1.Secret
	var err error

	switch resolved.Source {
	case credentials.SourceIntegration:
		secretName := credentials.ClusterSecretNameForGitHost(resolved.Host)
		clusterSecret, err = r.secretBuilder.BuildGitCredentialsSecretFromCredentials(ctx, secretName, resolved.Credentials)
		if err != nil {
			return err
		}
		clusterSecret.Annotations[models.GitIntegrationIDAnnotation] = resolved.IntegrationID
		if resolved.TokenMintedAt != nil {
			clusterSecret.Annotations[models.GitTokenMintedAtAnnotation] = resolved.TokenMintedAt.UTC().Format(time.RFC3339)
		}
		if resolved.TokenExpiresAt != nil {
			clusterSecret.Annotations[models.GitTokenExpiresAtAnnotation] = resolved.TokenExpiresAt.UTC().Format(time.RFC3339)
		}
	default:
		return nil
	}

	clusterSecret.Namespace = stack.Namespace
	clusterSecret.Annotations[models.SecretDataHashAnnotation] = resolved.DataHash
	clusterSecret.Labels[models.StackIDLabel] = stack.ID

	if _, ok := desired[clusterSecret.Name]; ok {
		// Same name implies same content for git credentials; keep the first.
		return nil
	}
	desired[clusterSecret.Name] = clusterSecret
	return nil
}

// collectRegistrySecret builds the cluster secret for a resolved registry
// credential and accumulates it into desired. Explicit secret refs use the
// backing hub secret; org-level registry credentials are synthesized from raw
// resolver output under a deterministic purpose- and host-keyed name;
// anonymous resolutions contribute nothing. When the same name is built more
// than once (e.g. one hub secret referenced for both a pull and a push on
// different hosts), the dockerconfig auth entries are merged so every host's
// auth lands in a single secret.
func (r *applyReconciler) collectRegistrySecret(
	ctx context.Context,
	stack *models.Stack,
	resolved *credentials.ResolvedRegistryCredential,
	ref string,
	forRepository bool,
	insecure bool,
	desired map[string]*corev1.Secret,
) error {
	var clusterSecret *corev1.Secret
	var err error

	switch resolved.Source {
	case credentials.SourceIntegration:
		secretName := credentials.ClusterSecretNameForRegistryHost(resolved.Host, resolved.Purpose)
		clusterSecret, err = r.secretBuilder.BuildDockerConfigJsonSecret(ctx, secretName, resolved.Username, resolved.Password, ref, insecure)
		if err != nil {
			return err
		}
		clusterSecret.Annotations[models.RegistryCredentialIDAnnotation] = resolved.CredentialID
	default:
		return nil
	}

	clusterSecret.Namespace = stack.Namespace
	clusterSecret.Annotations[models.SecretDataHashAnnotation] = resolved.DataHash
	clusterSecret.Labels[models.StackIDLabel] = stack.ID

	existing, ok := desired[clusterSecret.Name]
	if !ok {
		desired[clusterSecret.Name] = clusterSecret
		return nil
	}
	// Same name, different auth target: merge the dockerconfig auth entries.
	// The annotations reference the same hub secret / credential, so they are
	// already consistent — leave the accumulated secret's annotations intact.
	return mergeDockerConfigAuths(existing, clusterSecret)
}

// mergeDockerConfigAuths unions the `auths` entries of src's dockerconfigjson
// payload into dst's, leaving dst's metadata and annotations intact. Entries
// for a host present in both are identical (same backing credential), so the
// union is well-defined.
func mergeDockerConfigAuths(dst, src *corev1.Secret) error {
	dstJSON, ok := dst.Data[corev1.DockerConfigJsonKey]
	if !ok {
		return fmt.Errorf("secret '%s' has no dockerconfigjson payload to merge into", dst.Name)
	}
	srcJSON, ok := src.Data[corev1.DockerConfigJsonKey]
	if !ok {
		return fmt.Errorf("secret '%s' has no dockerconfigjson payload to merge", src.Name)
	}

	var dstConfig, srcConfig dockerConfigJSON
	if err := json.Unmarshal(dstJSON, &dstConfig); err != nil {
		return fmt.Errorf("failed to unmarshal dockerconfigjson for '%s': %w", dst.Name, err)
	}
	if err := json.Unmarshal(srcJSON, &srcConfig); err != nil {
		return fmt.Errorf("failed to unmarshal dockerconfigjson for '%s': %w", src.Name, err)
	}

	if dstConfig.Auths == nil {
		dstConfig.Auths = make(map[string]json.RawMessage, len(srcConfig.Auths))
	}
	for host, auth := range srcConfig.Auths {
		dstConfig.Auths[host] = auth
	}

	merged, err := json.Marshal(dstConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal merged dockerconfigjson for '%s': %w", dst.Name, err)
	}
	dst.Data[corev1.DockerConfigJsonKey] = merged
	return nil
}

// dockerConfigJSON is the minimal shape needed to union auth entries across
// secrets that share a cluster-secret name.
type dockerConfigJSON struct {
	Auths map[string]json.RawMessage `json:"auths"`
}

// syncPostgresCredentialSecrets ensures postgres connection credentials are
// available as K8s secrets in the stack namespace.
func (r *applyReconciler) syncPostgresCredentialSecrets(ctx context.Context, clusterClient client.Client, stack *models.Stack) error {
	pgConnections := stack.Connections.FromType(models.TopologyNodeTypePostgresAddon)
	if len(pgConnections) == 0 {
		return nil
	}

	for _, connection := range pgConnections {
		addonID := connection.From.Id
		database, superuser, err := stackdeploy.PostgresConnectionConfig(connection)
		if err != nil {
			return fmt.Errorf("failed to parse postgres connection config: %w", err)
		}

		creds, credErr := r.postgresAddonService.InternalGetCredentials(ctx, addonID, database, superuser)
		if credErr != nil {
			return fmt.Errorf("failed to get postgres credentials for addon '%s' db '%s': %w", addonID, database, credErr)
		}

		secretName := stackdeploy.PostgresCredentialSecretName(addonID, database)
		outputMap := creds.ToOutputMap()
		secretData := make(map[string][]byte, len(outputMap))
		for k, v := range outputMap {
			secretData[k] = []byte(v)
		}

		desired := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: stack.Namespace,
				Labels: map[string]string{
					models.StackIDLabel: stack.ID,
				},
			},
			Data: secretData,
		}
		if err := createOrUpdateSecret(ctx, clusterClient, desired); err != nil {
			return fmt.Errorf("failed to sync postgres credential secret '%s': %w", secretName, err)
		}
	}

	return nil
}

// syncGenericSecrets ensures user-created secrets (Generic, Token, etc.) connected
// to the stack are available as K8s Opaque secrets in the namespace.
func (r *applyReconciler) syncGenericSecrets(ctx context.Context, clusterClient client.Client, stack *models.Stack) error {
	secretConnections := stack.Connections.FromType(models.TopologyNodeTypeSecret)
	if len(secretConnections) == 0 {
		return nil
	}

	seen := make(map[string]struct{})
	for _, connection := range secretConnections {
		secretID := connection.From.Id
		if _, ok := seen[secretID]; ok {
			continue
		}
		seen[secretID] = struct{}{}

		secret, serr := r.secretService.InternalGetByID(ctx, secretID)
		if serr != nil {
			return fmt.Errorf("failed to get secret '%s': %w", secretID, serr)
		}

		secretData := make(map[string][]byte, len(secret.Data))
		for k, v := range secret.Data {
			secretData[k] = []byte(v)
		}

		desired := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secret.ClusterSecretName(),
				Namespace: stack.Namespace,
				Labels: map[string]string{
					models.StackIDLabel: stack.ID,
				},
				Annotations: map[string]string{
					models.SecretDataHashAnnotation: secret.DataHash,
					models.SecretIDAnnotation:       secret.ID,
				},
			},
			Data: secretData,
		}
		if err := createOrUpdateSecret(ctx, clusterClient, desired); err != nil {
			return fmt.Errorf("failed to sync generic secret '%s': %w", secret.Name, err)
		}
	}

	return nil
}

func createOrUpdateSecret(ctx context.Context, clusterClient client.Client, desired *corev1.Secret) error {
	existing := &corev1.Secret{}
	err := clusterClient.Get(ctx, client.ObjectKey{Name: desired.Name, Namespace: desired.Namespace}, existing)
	if err != nil {
		if k8sapierrors.IsNotFound(err) {
			return clusterClient.Create(ctx, desired)
		}
		return err
	}
	desired.ResourceVersion = existing.ResourceVersion
	return clusterClient.Update(ctx, desired)
}

func failRelease(ctx context.Context, svc releaseService, log logger.Logger, release *models.StackRelease, msg string) {
	log.Error(ctx, "release %s failed: %s", release.ID, msg)
	if _, err := svc.MarkFailed(ctx, release.ID, msg, nil); err != nil {
		log.Error(ctx, "failed to mark release %s as failed: %v", release.ID, err)
	}
}

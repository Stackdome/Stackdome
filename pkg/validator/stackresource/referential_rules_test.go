package stackresource

import (
	"context"

	"go.uber.org/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/credentials"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
)

func testStack() *models.Stack {
	return &models.Stack{ID: "stack-1", OrganisationID: "org-1", Namespace: "ns-1"}
}

var _ = Describe("validateReferences", func() {
	var ctrl *gomock.Controller

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("reports a missing mounted volume", func() {
		volumes := NewMockvolumeGetter(ctrl)
		volumes.EXPECT().
			GetByVolumeNameAndNamespace(gomock.Any(), "data", "ns-1").
			Return(nil, errors.NotFound("volume not found"))

		v := &validator{volumes: volumes}
		r := validImageResource()
		r.VolumeMounts = []*models.VolumeMount{{SourceVolumeName: "data", TargetPath: "/data"}}

		errs, serr := v.validateReferences(context.Background(), testStack(), r)

		Expect(serr).To(BeNil())
		Expect(codes(errs)).To(HaveKey(errors.VErrVolumeNotFound))
		Expect(fields(errs)).To(HaveKey("volume_mounts[0].source_volume"))
	})

	It("propagates a non-404 volume lookup error instead of masking it as not-found", func() {
		volumes := NewMockvolumeGetter(ctrl)
		volumes.EXPECT().
			GetByVolumeNameAndNamespace(gomock.Any(), "data", "ns-1").
			Return(nil, errors.GeneralError("db unavailable"))

		v := &validator{volumes: volumes}
		r := validImageResource()
		r.VolumeMounts = []*models.VolumeMount{{SourceVolumeName: "data", TargetPath: "/data"}}

		errs, serr := v.validateReferences(context.Background(), testStack(), r)

		Expect(serr).NotTo(BeNil())
		Expect(errs).To(BeEmpty())
	})

	It("reports a missing env secret", func() {
		secrets := NewMocksecretGetter(ctrl)
		secrets.EXPECT().
			GetByName(gomock.Any(), "org-1", "api-secrets").
			Return(nil, errors.NotFound("secret not found"))

		v := &validator{secrets: secrets}
		r := validImageResource()
		r.ExecutionConfig = &models.ExecutionConfig{Env: []models.EnvVar{{
			Name: "KEY", SecretKeyRef: &models.EnvSecretRef{SecretName: "api-secrets", Key: "k"},
		}}}

		errs, serr := v.validateReferences(context.Background(), testStack(), r)

		Expect(serr).To(BeNil())
		Expect(codes(errs)).To(HaveKey(errors.VErrSecretNotFound))
		Expect(fields(errs)).To(HaveKey("execution_config.env[0].secret_key_ref.secret_name"))
	})

	It("propagates a non-404 secret lookup error instead of masking it as not-found", func() {
		secrets := NewMocksecretGetter(ctrl)
		secrets.EXPECT().
			GetByName(gomock.Any(), "org-1", "api-secrets").
			Return(nil, errors.GeneralError("db unavailable"))

		v := &validator{secrets: secrets}
		r := validImageResource()
		r.ExecutionConfig = &models.ExecutionConfig{Env: []models.EnvVar{{
			Name: "KEY", SecretKeyRef: &models.EnvSecretRef{SecretName: "api-secrets", Key: "k"},
		}}}

		errs, serr := v.validateReferences(context.Background(), testStack(), r)

		Expect(serr).NotTo(BeNil())
		Expect(errs).To(BeEmpty())
	})

	It("reports a public port with no organisation domain configured", func() {
		domains := NewMockdomainLister(ctrl)
		domains.EXPECT().
			ListByOrganisationID(gomock.Any(), "org-1").
			Return([]*models.OrganisationDomain{}, nil)

		v := &validator{domains: domains}
		r := validImageResource()
		r.Ports[0].ExposedToPublic = true

		errs, serr := v.validateReferences(context.Background(), testStack(), r)

		Expect(serr).To(BeNil())
		Expect(codes(errs)).To(HaveKey(errors.VErrDomainNotConfigured))
		Expect(fields(errs)).To(HaveKey("ports[0]"))
	})

	It("propagates a non-404 domain lookup error instead of masking it as not-found", func() {
		domains := NewMockdomainLister(ctrl)
		domains.EXPECT().
			ListByOrganisationID(gomock.Any(), "org-1").
			Return(nil, errors.GeneralError("db unavailable"))

		v := &validator{domains: domains}
		r := validImageResource()
		r.Ports[0].ExposedToPublic = true

		errs, serr := v.validateReferences(context.Background(), testStack(), r)

		Expect(serr).NotTo(BeNil())
		Expect(errs).To(BeEmpty())
	})

	It("reports a missing pull registry credential", func() {
		mockCreds := mocks.NewMockCredentialResolver(ctrl)
		mockCreds.EXPECT().
			RegistryCredentials(gomock.Any(), "org-1", "nginx:latest", credentials.RegistryPurposePull,
				credentials.RegistryAuthSelector{RegistryCredentialID: "cred-1"}).
			Return(nil, errors.NotFound("registry credential not found"))

		v := &validator{credentials: mockCreds}
		r := validImageResource()
		r.ImageConfig.RegistryCredentialID = "cred-1"

		errs, serr := v.validateReferences(context.Background(), testStack(), r)

		Expect(serr).To(BeNil())
		Expect(codes(errs)).To(HaveKey(errors.VErrRegistryCredentialNotFound))
		Expect(fields(errs)).To(HaveKey("source.image.registry_credentials_id"))
	})

	It("propagates a non-404 pull registry credential lookup error instead of masking it as not-found", func() {
		mockCreds := mocks.NewMockCredentialResolver(ctrl)
		mockCreds.EXPECT().
			RegistryCredentials(gomock.Any(), "org-1", "nginx:latest", credentials.RegistryPurposePull,
				credentials.RegistryAuthSelector{RegistryCredentialID: "cred-1"}).
			Return(nil, errors.GeneralError("db unavailable"))

		v := &validator{credentials: mockCreds}
		r := validImageResource()
		r.ImageConfig.RegistryCredentialID = "cred-1"

		errs, serr := v.validateReferences(context.Background(), testStack(), r)

		Expect(serr).NotTo(BeNil())
		Expect(errs).To(BeEmpty())
	})

	It("reports a missing push registry credential", func() {
		mockCreds := mocks.NewMockCredentialResolver(ctrl)
		mockCreds.EXPECT().
			RegistryCredentials(gomock.Any(), "org-1", "registry.example.com/app", credentials.RegistryPurposePush,
				credentials.RegistryAuthSelector{RegistryCredentialID: "cred-2"}).
			Return(nil, errors.NotFound("registry credential not found"))

		v := &validator{credentials: mockCreds}
		r := validBuildResource()
		r.BuildConfig.PushRegistryCredentialID = "cred-2"
		r.BuildConfig.BuildImageRepository.ExternalImageRef = "registry.example.com/app"

		errs, serr := v.validateReferences(context.Background(), testStack(), r)

		Expect(serr).To(BeNil())
		Expect(codes(errs)).To(HaveKey(errors.VErrRegistryCredentialNotFound))
		Expect(fields(errs)).To(HaveKey("source.git.push.registry_credentials_id"))
	})

	It("reports a missing git integration", func() {
		gitIntegrations := NewMockgitIntegrationGetter(ctrl)
		gitIntegrations.EXPECT().
			InternalGetByID(gomock.Any(), "gi-1").
			Return(nil, errors.NotFound("git integration not found"))

		v := &validator{gitIntegrations: gitIntegrations}
		r := validBuildResource()
		r.BuildConfig.SourceContext.Git.IntegrationID = "gi-1"

		errs, serr := v.validateReferences(context.Background(), testStack(), r)

		Expect(serr).To(BeNil())
		Expect(codes(errs)).To(HaveKey(errors.VErrGitIntegrationNotFound))
		Expect(fields(errs)).To(HaveKey("source.git.integration_id"))
	})

	It("reports a git integration that belongs to a different organisation as not found", func() {
		gitIntegrations := NewMockgitIntegrationGetter(ctrl)
		gitIntegrations.EXPECT().
			InternalGetByID(gomock.Any(), "gi-1").
			Return(&models.GitIntegration{ID: "gi-1", OrganisationID: "org-other"}, nil)

		v := &validator{gitIntegrations: gitIntegrations}
		r := validBuildResource()
		r.BuildConfig.SourceContext.Git.IntegrationID = "gi-1"

		errs, serr := v.validateReferences(context.Background(), testStack(), r)

		Expect(serr).To(BeNil())
		Expect(codes(errs)).To(HaveKey(errors.VErrGitIntegrationNotFound))
		Expect(fields(errs)).To(HaveKey("source.git.integration_id"))
	})

	It("propagates a non-404 git integration lookup error instead of masking it as not-found", func() {
		gitIntegrations := NewMockgitIntegrationGetter(ctrl)
		gitIntegrations.EXPECT().
			InternalGetByID(gomock.Any(), "gi-1").
			Return(nil, errors.GeneralError("db unavailable"))

		v := &validator{gitIntegrations: gitIntegrations}
		r := validBuildResource()
		r.BuildConfig.SourceContext.Git.IntegrationID = "gi-1"

		errs, serr := v.validateReferences(context.Background(), testStack(), r)

		Expect(serr).NotTo(BeNil())
		Expect(errs).To(BeEmpty())
	})

	It("reports a mounted volume referenced by ID that does not exist", func() {
		volumes := NewMockvolumeGetter(ctrl)
		volumes.EXPECT().
			GetByID(gomock.Any(), "vol-bogus").
			Return(nil, errors.NotFound("volume not found"))

		v := &validator{volumes: volumes}
		r := validImageResource()
		r.VolumeMounts = []*models.VolumeMount{{SourceVolumeID: "vol-bogus", TargetPath: "/data"}}

		errs, serr := v.validateReferences(context.Background(), testStack(), r)

		Expect(serr).To(BeNil())
		Expect(codes(errs)).To(HaveKey(errors.VErrVolumeNotFound))
		Expect(fields(errs)).To(HaveKey("volume_mounts[0].source_volume"))
	})

	It("resolves a mounted volume referenced by ID with no error", func() {
		volumes := NewMockvolumeGetter(ctrl)
		volumes.EXPECT().
			GetByID(gomock.Any(), "vol-1").
			Return(&models.Volume{}, nil)

		v := &validator{volumes: volumes}
		r := validImageResource()
		r.VolumeMounts = []*models.VolumeMount{{SourceVolumeID: "vol-1", TargetPath: "/data"}}

		errs, serr := v.validateReferences(context.Background(), testStack(), r)

		Expect(serr).To(BeNil())
		Expect(errs).To(BeEmpty())
	})

	It("propagates a non-404 error resolving a mounted volume by ID", func() {
		volumes := NewMockvolumeGetter(ctrl)
		volumes.EXPECT().
			GetByID(gomock.Any(), "vol-1").
			Return(nil, errors.GeneralError("db unavailable"))

		v := &validator{volumes: volumes}
		r := validImageResource()
		r.VolumeMounts = []*models.VolumeMount{{SourceVolumeID: "vol-1", TargetPath: "/data"}}

		errs, serr := v.validateReferences(context.Background(), testStack(), r)

		Expect(serr).NotTo(BeNil())
		Expect(errs).To(BeEmpty())
	})

	It("reports a missing build-context volume", func() {
		volumes := NewMockvolumeGetter(ctrl)
		volumes.EXPECT().
			GetByVolumeNameAndNamespace(gomock.Any(), "build-src", "ns-1").
			Return(nil, errors.NotFound("volume not found"))

		v := &validator{volumes: volumes}
		r := validBuildResourceWithVolumeSource()
		r.BuildConfig.SourceContext.Volume.SourceVolumeName = "build-src"

		errs, serr := v.validateReferences(context.Background(), testStack(), r)

		Expect(serr).To(BeNil())
		Expect(codes(errs)).To(HaveKey(errors.VErrVolumeNotFound))
		Expect(fields(errs)).To(HaveKey("source.volume"))
	})

	It("reports a missing build-context volume referenced by ID", func() {
		volumes := NewMockvolumeGetter(ctrl)
		volumes.EXPECT().
			GetByID(gomock.Any(), "vol-bogus").
			Return(nil, errors.NotFound("volume not found"))

		v := &validator{volumes: volumes}
		r := validBuildResourceWithVolumeSource()
		r.BuildConfig.SourceContext.Volume.SourceVolumeName = ""
		r.BuildConfig.SourceContext.Volume.SourceVolumeID = "vol-bogus"

		errs, serr := v.validateReferences(context.Background(), testStack(), r)

		Expect(serr).To(BeNil())
		Expect(codes(errs)).To(HaveKey(errors.VErrVolumeNotFound))
		Expect(fields(errs)).To(HaveKey("source.volume"))
	})

	It("returns no errors when every reference resolves", func() {
		volumes := NewMockvolumeGetter(ctrl)
		volumes.EXPECT().
			GetByVolumeNameAndNamespace(gomock.Any(), "data", "ns-1").
			Return(&models.Volume{}, nil)

		secrets := NewMocksecretGetter(ctrl)
		secrets.EXPECT().
			GetByName(gomock.Any(), "org-1", "api-secrets").
			Return(&models.Secret{}, nil)

		domains := NewMockdomainLister(ctrl)
		domains.EXPECT().
			ListByOrganisationID(gomock.Any(), "org-1").
			Return([]*models.OrganisationDomain{{Domain: "example.com"}}, nil)

		mockCreds := mocks.NewMockCredentialResolver(ctrl)
		mockCreds.EXPECT().
			RegistryCredentials(gomock.Any(), "org-1", "nginx:latest", credentials.RegistryPurposePull,
				credentials.RegistryAuthSelector{RegistryCredentialID: "cred-1"}).
			Return(&credentials.ResolvedRegistryCredential{}, nil)

		v := &validator{volumes: volumes, secrets: secrets, domains: domains, credentials: mockCreds}
		r := validImageResource()
		r.ImageConfig.RegistryCredentialID = "cred-1"
		r.Ports[0].ExposedToPublic = true
		r.VolumeMounts = []*models.VolumeMount{{SourceVolumeName: "data", TargetPath: "/data"}}
		r.ExecutionConfig = &models.ExecutionConfig{Env: []models.EnvVar{{
			Name: "KEY", SecretKeyRef: &models.EnvSecretRef{SecretName: "api-secrets", Key: "k"},
		}}}

		errs, serr := v.validateReferences(context.Background(), testStack(), r)

		Expect(serr).To(BeNil())
		Expect(errs).To(BeEmpty())
	})
})

// validBuildResource returns a minimal build-sourced resource that passes
// input rules, for exercising push/git credential referential rules.
func validBuildResource() *models.StackResource {
	return &models.StackResource{
		Name:         "web",
		WorkloadType: models.WorkloadTypeService,
		BuildConfig: &models.BuildConfigSpec{
			SourceContext: models.BuildContextSource{
				Git: &models.GitBuildSource{RepoURL: "https://github.com/example/repo.git"},
			},
			SourceRevision: models.BuildSourceRevision{
				Git: &models.GitRevision{Branch: "main"},
			},
			BuildImageRepository: models.BuildImageRepository{ExternalImageRef: "registry.example.com/app"},
		},
	}
}

// validBuildResourceWithVolumeSource is validBuildResource but sourced from a
// volume build context instead of git.
func validBuildResourceWithVolumeSource() *models.StackResource {
	r := validBuildResource()
	r.BuildConfig.SourceContext = models.BuildContextSource{
		Volume: &models.VolumeBuildSource{SourceVolumeName: "build-src"},
	}
	r.BuildConfig.SourceRevision = models.BuildSourceRevision{
		Volume: &models.VolumeRevision{CurrentVolumeHash: "hash"},
	}
	return r
}

func fields(errs []errors.FieldError) map[string]bool {
	out := map[string]bool{}
	for _, e := range errs {
		out[e.Field] = true
	}
	return out
}

package stackresource

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

func int32Ptr(i int32) *int32 { return &i }

// validImageResource returns a minimal resource that passes all input rules.
func validImageResource() *models.StackResource {
	return &models.StackResource{
		Name:         "web",
		ImageConfig:  &models.ImageConfigSpec{Image: "nginx:latest"},
		WorkloadType: models.WorkloadTypeService,
		Ports: models.Ports{
			{Name: "http", Number: 8080, Protocol: "http", ExposedToPublic: true},
		},
	}
}

func codes(errs []errors.FieldError) map[string]bool {
	out := map[string]bool{}
	for _, e := range errs {
		out[e.Code] = true
	}
	return out
}

var _ = Describe("validateInputRules", func() {
	DescribeTable("rule failures",
		func(mutate func(r *models.StackResource), wantCode, wantField string) {
			r := validImageResource()
			mutate(r)
			errs := validateInputRules(r)

			Expect(codes(errs)).To(HaveKey(wantCode), "expected code %s in %v", wantCode, errs)

			found := false
			for _, e := range errs {
				if e.Code == wantCode && e.Field == wantField {
					found = true
				}
			}
			Expect(found).To(BeTrue(), "expected field %s for code %s, got %v", wantField, wantCode, errs)
		},
		Entry("no source",
			func(r *models.StackResource) { r.ImageConfig = nil },
			errors.VErrSourceRequired, "source"),
		Entry("both sources",
			func(r *models.StackResource) {
				r.BuildConfig = &models.BuildConfigSpec{}
			},
			errors.VErrSourceConflict, "source"),
		Entry("bad workload type",
			func(r *models.StackResource) { r.WorkloadType = "Bogus" },
			errors.VErrWorkloadTypeInvalid, "workload_type"),
		Entry("cronjob without schedule",
			func(r *models.StackResource) {
				r.WorkloadType = models.WorkloadTypeCronJob
				r.Ports = nil
			},
			errors.VErrScheduleRequired, "schedule"),
		Entry("schedule on non-cronjob",
			func(r *models.StackResource) {
				r.Schedule = "*/5 * * * *"
			},
			errors.VErrScheduleNotAllowed, "schedule"),
		Entry("invalid cron expression",
			func(r *models.StackResource) {
				r.WorkloadType = models.WorkloadTypeCronJob
				r.Ports = nil
				r.Schedule = "not-a-cron"
			},
			errors.VErrScheduleInvalid, "schedule"),
		Entry("ports on worker",
			func(r *models.StackResource) {
				r.WorkloadType = models.WorkloadTypeWorker
			},
			errors.VErrPortsNotAllowed, "ports"),
		Entry("public tcp port",
			func(r *models.StackResource) {
				r.Ports[0].Protocol = "tcp"
			},
			errors.VErrPublicPortNotHTTP, "ports[0].protocol"),
		Entry("bad protocol",
			func(r *models.StackResource) {
				r.Ports[0].Protocol = "udp"
				r.Ports[0].ExposedToPublic = false
			},
			errors.VErrPortProtocolInvalid, "ports[0].protocol"),
		Entry("bad port name",
			func(r *models.StackResource) {
				r.Ports[0].Name = "Bad_Name_Way_Too_Long"
			},
			errors.VErrPortNameInvalid, "ports[0].name"),
		Entry("port number out of range",
			func(r *models.StackResource) {
				r.Ports[0].Number = 70000
			},
			errors.VErrPortNumberInvalid, "ports[0].number"),
		Entry("duplicate port name",
			func(r *models.StackResource) {
				r.Ports = append(r.Ports, models.Port{Name: "http", Number: 9090, Protocol: "http"})
			},
			errors.VErrPortNameDuplicate, "ports[1].name"),
		Entry("duplicate port number",
			func(r *models.StackResource) {
				r.Ports = append(r.Ports, models.Port{Name: "http2", Number: 8080, Protocol: "http"})
			},
			errors.VErrPortNumberDuplicate, "ports[1].number"),
		Entry("env without value",
			func(r *models.StackResource) {
				r.ExecutionConfig = &models.ExecutionConfig{Env: []models.EnvVar{{Name: "API_KEY"}}}
			},
			errors.VErrEnvValueMissing, "execution_config.env[0]"),
		Entry("env with value and secret ref",
			func(r *models.StackResource) {
				r.ExecutionConfig = &models.ExecutionConfig{Env: []models.EnvVar{{
					Name: "API_KEY", Value: "x",
					SecretKeyRef: &models.EnvSecretRef{SecretName: "s", Key: "k"},
				}}}
			},
			errors.VErrEnvValueConflict, "execution_config.env[0]"),
		Entry("duplicate env names",
			func(r *models.StackResource) {
				r.ExecutionConfig = &models.ExecutionConfig{Env: []models.EnvVar{
					{Name: "A", Value: "1"}, {Name: "A", Value: "2"},
				}}
			},
			errors.VErrEnvNameDuplicate, "execution_config.env[1].name"),
		Entry("env without name",
			func(r *models.StackResource) {
				r.ExecutionConfig = &models.ExecutionConfig{Env: []models.EnvVar{{Value: "1"}}}
			},
			errors.VErrEnvNameRequired, "execution_config.env[0].name"),
		Entry("volume mount missing target path",
			func(r *models.StackResource) {
				r.VolumeMounts = []*models.VolumeMount{{SourceVolumeName: "data"}}
			},
			errors.VErrVolumeMountInvalid, "volume_mounts[0]"),
		Entry("depends_on self reference",
			func(r *models.StackResource) {
				r.DependsOn = models.Dependencies{"web"}
			},
			errors.VErrDependencySelf, "depends_on[0]"),
		Entry("depends_on duplicate",
			func(r *models.StackResource) {
				r.DependsOn = models.Dependencies{"db", "db"}
			},
			errors.VErrDependencyDuplicate, "depends_on[1]"),
		Entry("git branch and tag both set",
			func(r *models.StackResource) {
				r.ImageConfig = nil
				r.BuildConfig = &models.BuildConfigSpec{
					SourceContext:        models.BuildContextSource{Git: &models.GitBuildSource{RepoURL: "https://github.com/o/r"}},
					SourceRevision:       models.BuildSourceRevision{Git: &models.GitRevision{Branch: "main", Tag: "v1"}},
					BuildImageRepository: models.BuildImageRepository{UseInClusterRegistry: true},
				}
			},
			errors.VErrGitBranchTagConflict, "source.git"),
		Entry("git repo url missing",
			func(r *models.StackResource) {
				r.ImageConfig = nil
				r.BuildConfig = &models.BuildConfigSpec{
					SourceContext:        models.BuildContextSource{Git: &models.GitBuildSource{}},
					SourceRevision:       models.BuildSourceRevision{Git: &models.GitRevision{}},
					BuildImageRepository: models.BuildImageRepository{UseInClusterRegistry: true},
				}
			},
			errors.VErrGitRepoURLRequired, "source.git.repo_url"),
		Entry("invalid commit sha",
			func(r *models.StackResource) {
				r.ImageConfig = nil
				r.BuildConfig = &models.BuildConfigSpec{
					SourceContext:        models.BuildContextSource{Git: &models.GitBuildSource{RepoURL: "https://github.com/o/r"}},
					SourceRevision:       models.BuildSourceRevision{Git: &models.GitRevision{Branch: "main", Commit: "ZZZ"}},
					BuildImageRepository: models.BuildImageRepository{UseInClusterRegistry: true},
				}
			},
			errors.VErrGitCommitInvalid, "source.git.commit"),
		Entry("push target both in-cluster and external",
			func(r *models.StackResource) {
				r.ImageConfig = nil
				r.BuildConfig = &models.BuildConfigSpec{
					SourceContext:  models.BuildContextSource{Git: &models.GitBuildSource{RepoURL: "https://github.com/o/r"}},
					SourceRevision: models.BuildSourceRevision{Git: &models.GitRevision{Branch: "main"}},
					BuildImageRepository: models.BuildImageRepository{
						UseInClusterRegistry: true, ExternalImageRef: "docker.io/o/r",
					},
				}
			},
			errors.VErrPushTargetConflict, "source.git.push"),
		Entry("unparseable external push ref",
			func(r *models.StackResource) {
				r.ImageConfig = nil
				r.BuildConfig = &models.BuildConfigSpec{
					SourceContext:        models.BuildContextSource{Git: &models.GitBuildSource{RepoURL: "https://github.com/o/r"}},
					SourceRevision:       models.BuildSourceRevision{Git: &models.GitRevision{Branch: "main"}},
					BuildImageRepository: models.BuildImageRepository{ExternalImageRef: "UPPER CASE bad ref!!"},
				}
			},
			errors.VErrPushRefInvalid, "source.git.push.repository"),
		Entry("empty image ref",
			func(r *models.StackResource) {
				r.ImageConfig = &models.ImageConfigSpec{Image: ""}
			},
			errors.VErrImageRefRequired, "source.image.ref"),
	)

	It("has no errors for a valid resource", func() {
		Expect(validateInputRules(validImageResource())).To(BeEmpty())
	})

	It("aggregates multiple failures", func() {
		r := validImageResource()
		r.Ports[0].Protocol = "tcp"              // public_port_not_http
		r.DependsOn = models.Dependencies{"web"} // self_dependency
		errs := validateInputRules(r)
		Expect(len(errs)).To(BeNumerically(">=", 2))
	})
})

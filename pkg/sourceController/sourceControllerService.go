package sourceController

import (
	"bytes"
	"context"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/pkg/sourceController/bean"
	"github.com/devtron-labs/devtron/pkg/sourceController/oci"
	repository "github.com/devtron-labs/devtron/pkg/sourceController/repo"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	gcrv1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/json"
	kuberecorder "k8s.io/client-go/tools/record"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"strings"
)

type SourceControllerService interface {
	CallExternalCIWebHook(digest, tag string) error
	ReconcileSource(ctx context.Context) (bean.Result, error)
	ReconcileSourceWrapper()
}

type SourceControllerServiceImpl struct {
	logger               *zap.SugaredLogger
	SCSconfig            *SourceControllerConfig
	ciArtifactRepository repository.CiArtifactRepository
	client.Client
	kuberecorder.EventRecorder
}

type SourceControllerConfig struct {
	ImageShowCount int    `env:"IMAGE_COUNT_FROM_REPO" envDefault:"20"`
	AppId          int    `env:"APP_ID_EXTERNAL_CI" envDefault:"18"`
	EnvId          int    `env:"ENV_ID_EXTERNAL_CI" envDefault:"1"`
	ExternalCiId   int    `env:"EXTERNAL_CI_ID" envDefault:"6"`
	RepoName       string `env:"REPO_NAME_EXTERNAL_CI" envDefault:"stefanprodan/manifests/podinfo"`
	RegistryURL    string `env:"REGISTRY_URL_EXTERNAL_CI" envDefault:"ghcr.io"`
	Insecure       bool   `env:"INSECURE_EXTERNAL_CI" envDefault:"true"`
}

var UserAgent = "flux/v2"

type invalidOCIURLError struct {
	err error
}

func (e invalidOCIURLError) Error() string {
	return e.err.Error()
}

func NewSourceControllerServiceImpl(logger *zap.SugaredLogger,
	cfg *SourceControllerConfig,
	ciArtifactRepository repository.CiArtifactRepository) *SourceControllerServiceImpl {
	sourceControllerServiceimpl := &SourceControllerServiceImpl{
		logger:               logger,
		SCSconfig:            cfg,
		ciArtifactRepository: ciArtifactRepository,
	}

	return sourceControllerServiceimpl
}

func GetSourceControllerConfig() (*SourceControllerConfig, error) {
	cfg := &SourceControllerConfig{}
	err := env.Parse(cfg)
	if err != nil {
		fmt.Println("failed to parse server cluster status config: " + err.Error())
		return nil, err
	}
	return cfg, err
}

//// OCIRepository is the Schema for the ocirepositories API
//type OCIRepository struct {
//	Spec bean.OCIRepositorySpec `json:"spec,omitempty"`
//	// +kubebuilder:default={"observedGeneration":-1}
//	Status bean.OCIRepositoryStatus `json:"status,omitempty"`
//}

type ExternalCI struct {
	DockerImage  string `json:"dockerImage" validate:"required,image-validator"`
	Digest       string `json:"digest"`
	DataSource   string `json:"dataSource"`
	MaterialType string `json:"materialType"`
}

//type CiCompleteEvent struct {
//	CiProjectDetails   []pipeline.CiProjectDetails `json:"ciProjectDetails"`
//	DockerImage        string                      `json:"dockerImage" validate:"required,image-validator"`
//	Digest             string                      `json:"digest"`
//	PipelineId         int                         `json:"pipelineId"`
//	WorkflowId         *int                        `json:"workflowId"`
//	TriggeredBy        int32                       `json:"triggeredBy"`
//	PipelineName       string                      `json:"pipelineName"`
//	DataSource         string                      `json:"dataSource"`
//	MaterialType       string                      `json:"materialType"`
//	Metrics            util.CIMetrics              `json:"metrics"`
//	AppName            string                      `json:"appName"`
//	IsArtifactUploaded bool                        `json:"isArtifactUploaded"`
//	FailureReason      string                      `json:"failureReason"`
//}

func getPayloadForExternalCi(image, digest string) *ExternalCI {
	payload := &ExternalCI{
		DockerImage:  image,
		Digest:       digest,
		DataSource:   bean.External,
		MaterialType: bean.MaterialTypeGit,
	}
	return payload
}

func (impl *SourceControllerServiceImpl) ReconcileSourceWrapper() {
	result, err := impl.ReconcileSource(context.Background())
	if err != nil {
		impl.logger.Errorw("error in reconciling sources", "err", err, "result", result)

	}
}

func (impl *SourceControllerServiceImpl) ReconcileSource(ctx context.Context) (bean.Result, error) {
	var auth authn.Authenticator
	keychain := oci.Anonymous{}
	transport := remote.DefaultTransport.(*http.Transport).Clone()
	opts := makeRemoteOptions(ctx, transport, keychain, auth, impl.SCSconfig.Insecure)

	url, err := parseRepositoryURLInValidFormat(impl.SCSconfig.RegistryURL, impl.SCSconfig.RepoName)
	if err != nil {
		impl.logger.Errorw("error in parsung repository utl in valid format", "err", err)
		return bean.ResultEmpty, invalidOCIURLError{err}
	}
	tags, err := getAllTags(url, opts.craneOpts)
	if err != nil {
		impl.logger.Errorw("error in getting all tags ", "err", err, "url", url)
		return bean.ResultEmpty, err
	}
	digests := make([]string, 0, len(tags))
	digestTagMap := make(map[string]string)
	for i := 0; i < len(tags) && i < impl.SCSconfig.ImageShowCount; i++ {
		tag := tags[i]
		// Determine which artifact revision to pull
		tagUrl, err := getArtifactURLForTag(url, tag)
		if err != nil {
			impl.logger.Errorw("error in getting artifact url", "err", err, "tag", tag)
			return bean.ResultEmpty, err
		}
		digest, err := crane.Digest(tagUrl, opts.craneOpts...)
		if err != nil {
			fmt.Errorf("error")

		}
		digestTagMap[digest] = tag
		digests = append(digests, digest)
	}

	err = impl.filterAlreadyPresentArtifacts(digests, digestTagMap)
	if err != nil {
		impl.logger.Errorw("error in filtering artifacts", "err", err)
		return bean.ResultEmpty, err
	}
	for digest, tag := range digestTagMap {
		impl.CallExternalCIWebHook(digest, tag)
	}
	return bean.ResultSuccess, err
}

// CallingExternalCiWebhook
func (impl *SourceControllerServiceImpl) CallExternalCIWebHook(digest, tag string) error {
	host := impl.SCSconfig.RegistryURL
	repoName := impl.SCSconfig.RepoName
	image := fmt.Sprintf("%s/%s:%s", host, repoName, tag)
	url := bean.WebHookHostUrl + strconv.Itoa(impl.SCSconfig.ExternalCiId)
	payload := getPayloadForExternalCi(image, digest)
	b, err := json.Marshal(payload)
	if err != nil {
		impl.logger.Errorw("error in marshalling golang struct", "err", err)
		return err
	}
	bearer := ""
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	req.Header.Set("api-token", bearer)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		impl.logger.Errorw("error in hitting http request to web hook", "err", err)
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (impl *SourceControllerServiceImpl) filterAlreadyPresentArtifacts(imageDigests []string, digestTagMap map[string]string) error {
	ciArtifacts, err := impl.ciArtifactRepository.GetByImageDigests(imageDigests)
	if err != nil {
		impl.logger.Errorw("error in getting ci artifact by image digests ", "err", err)
		return err
	}
	for _, ciArtifact := range ciArtifacts {
		delete(digestTagMap, ciArtifact.ImageDigest)
	}
	return nil

}

// getAllTags call the remote container registry, fetches all the tags from the repository
func getAllTags(url string, options []crane.Option) ([]string, error) {
	tags, err := crane.ListTags(url, options...)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

// getArtifactURL determines which tag or revision should be used and returns the OCI artifact FQN.
func getArtifactURLForTag(tag, url string) (string, error) {
	if tag != "" {
		return fmt.Sprintf("%s:%s", tag, url), nil
	}
	return url, nil
}

// parseRepositoryURL validates and extracts the repository URL.
func parseRepositoryURLInValidFormat(registryUrl, repo string) (string, error) {
	url := fmt.Sprintf("%s/%s", registryUrl, repo)
	ref, err := name.ParseReference(url)
	if err != nil {
		return "", err
	}

	imageName := strings.TrimPrefix(url, ref.Context().RegistryStr())
	if s := strings.Split(imageName, ":"); len(s) > 1 {
		return "", fmt.Errorf("URL must not contain a tag; remove ':%s'", s[1])
	}
	return ref.Context().Name(), nil
}

// getRevision fetches the upstream digest, returning the revision in the
// format '<tag>@<digest>'.
func (r *SourceControllerServiceImpl) getRevision(url string, options []crane.Option) (string, error) {
	ref, err := name.ParseReference(url)
	if err != nil {
		return "", err
	}

	repoTag := ""
	repoName := strings.TrimPrefix(url, ref.Context().RegistryStr())
	if s := strings.Split(repoName, ":"); len(s) == 2 && !strings.Contains(repoName, "@") {
		repoTag = s[1]
	}

	if repoTag == "" && !strings.Contains(repoName, "@") {
		repoTag = "latest"
	}

	digest, err := crane.Digest(url, options...)
	if err != nil {
		return "", err
	}

	digestHash, err := gcrv1.NewHash(digest)
	if err != nil {
		return "", err
	}

	revision := digestHash.String()
	if repoTag != "" {
		revision = fmt.Sprintf("%s@%s", repoTag, revision)
	}
	return revision, nil
}

//// getArtifactURL determines which tag or revision should be used and returns the OCI artifact FQN.
//func (r *SourceControllerServiceImpl) getArtifactURL(obj *OCIRepository, options []crane.Option) (string, error) {
//	url, err := r.parseRepositoryURL(obj)
//	if err != nil {
//		return "", invalidOCIURLError{err}
//	}
//
//	if obj.Spec.Reference != nil {
//		if obj.Spec.Reference.Digest != "" {
//			return fmt.Sprintf("%s@%s", url, obj.Spec.Reference.Digest), nil
//		}
//
//		if obj.Spec.Reference.SemVer != "" {
//			tag, err := r.getTagBySemver(url, obj.Spec.Reference.SemVer, options)
//			if err != nil {
//				return "", err
//			}
//			return fmt.Sprintf("%s:%s", url, tag), nil
//		}
//
//		if obj.Spec.Reference.Tag != "" {
//			return fmt.Sprintf("%s:%s", url, obj.Spec.Reference.Tag), nil
//		}
//	}
//
//	return url, nil
//}

//// getTagBySemver call the remote container registry, fetches all the tags from the repository,
//// and returns the latest tag according to the semver expression.
//func (r *SourceControllerServiceImpl) getTagBySemver(url, exp string, options []crane.Option) (string, error) {
//	tags, err := crane.ListTags(url, options...)
//	if err != nil {
//		return "", err
//	}
//
//	constraint, err := semver.NewConstraint(exp)
//	if err != nil {
//		return "", fmt.Errorf("semver '%s' parse error: %w", exp, err)
//	}
//
//	var matchingVersions []*semver.Version
//	for _, t := range tags {
//		v, err := bean.ParseVersion(t)
//		if err != nil {
//			continue
//		}
//
//		if constraint.Check(v) {
//			matchingVersions = append(matchingVersions, v)
//		}
//	}
//
//	if len(matchingVersions) == 0 {
//		return "", fmt.Errorf("no match found for semver: %s", exp)
//	}
//
//	sort.Sort(sort.Reverse(semver.Collection(matchingVersions)))
//	return matchingVersions[0].Original(), nil
//}

//// parseRepositoryURL validates and extracts the repository URL.
//func (r *SourceControllerServiceImpl) parseRepositoryURL(obj *OCIRepository) (string, error) {
//	if !strings.HasPrefix(obj.Spec.URL, oci.OCIRepositoryPrefix) {
//		return "", fmt.Errorf("URL must be in format 'oci://<domain>/<org>/<repo>'")
//	}
//
//	url := strings.TrimPrefix(obj.Spec.URL, oci.OCIRepositoryPrefix)
//	ref, err := name.ParseReference(url)
//	if err != nil {
//		return "", err
//	}
//
//	imageName := strings.TrimPrefix(url, ref.Context().RegistryStr())
//	if s := strings.Split(imageName, ":"); len(s) > 1 {
//		return "", fmt.Errorf("URL must not contain a tag; remove ':%s'", s[1])
//	}
//
//	return ref.Context().Name(), nil
//}

// remoteOptions contains the options to interact with a remote registry.
// It can be used to pass options to go-containerregistry based libraries.
type remoteOptions struct {
	craneOpts  []crane.Option
	verifyOpts []remote.Option
}

// makeRemoteOptions returns a remoteOptions struct with the authentication and transport options set.
// The returned struct can be used to interact with a remote registry using go-containerregistry based libraries.
func makeRemoteOptions(ctxTimeout context.Context, transport http.RoundTripper, keychain authn.Keychain, auth authn.Authenticator, insecure bool) remoteOptions {
	// have to make it configurable insecure in future iterations
	o := remoteOptions{
		craneOpts:  craneOptions(ctxTimeout, insecure),
		verifyOpts: []remote.Option{},
	}

	if transport != nil {
		o.craneOpts = append(o.craneOpts, crane.WithTransport(transport))
		o.verifyOpts = append(o.verifyOpts, remote.WithTransport(transport))
	}

	if auth != nil {
		// auth take precedence over keychain here as we expect the caller to set
		// the auth only if it is required.
		o.verifyOpts = append(o.verifyOpts, remote.WithAuth(auth))
		o.craneOpts = append(o.craneOpts, crane.WithAuth(auth))
		return o
	}

	o.verifyOpts = append(o.verifyOpts, remote.WithAuthFromKeychain(keychain))
	o.craneOpts = append(o.craneOpts, crane.WithAuthFromKeychain(keychain))

	return o
}

// craneOptions sets the auth headers, timeout and user agent
// for all operations against remote container registries.
func craneOptions(ctx context.Context, insecure bool) []crane.Option {
	options := []crane.Option{
		crane.WithContext(ctx),
		crane.WithUserAgent(UserAgent),
	}

	if insecure {
		options = append(options, crane.Insecure)
	}

	return options
}

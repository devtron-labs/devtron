package bitbucket

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/k0kubun/pp"
	"github.com/mitchellh/mapstructure"
)

type Repository struct {
	c *Client

	Project     Project
	Uuid        string
	Name        string
	Slug        string
	Full_name   string
	Description string
	Fork_policy string
	Language    string
	Is_private  bool
	Has_issues  bool
	Mainbranch  RepositoryBranch
	Type        string
	Owner       map[string]interface{}
	Links       map[string]interface{}
	Parent      *Repository
}

type RepositoryFile struct {
	Mimetype   string
	Links      map[string]interface{}
	Path       string
	Commit     map[string]interface{}
	Attributes []string
	Type       string
	Size       int
}

type RepositoryBlob struct {
	Content []byte
}

type RepositoryRefs struct {
	Page     int
	Pagelen  int
	MaxDepth int
	Size     int
	Next     string
	Refs     []map[string]interface{}
}

type RepositoryBranches struct {
	Page     int
	Pagelen  int
	MaxDepth int
	Size     int
	Next     string
	Branches []RepositoryBranch
}

type RepositoryBranch struct {
	Type                   string
	Name                   string
	Default_Merge_Strategy string
	Merge_Strategies       []string
	Links                  map[string]interface{}
	Target                 map[string]interface{}
	Heads                  []map[string]interface{}
}

type RepositoryTags struct {
	Page     int
	Pagelen  int
	MaxDepth int
	Size     int
	Next     string
	Tags     []RepositoryTag
}

type RepositoryTag struct {
	Type   string
	Name   string
	Links  map[string]interface{}
	Target map[string]interface{}
	Heads  []map[string]interface{}
}

type Pipeline struct {
	Type       string
	Enabled    bool
	Repository Repository
}

type PipelineVariables struct {
	Page      int
	Pagelen   int
	MaxDepth  int
	Size      int
	Next      string
	Variables []PipelineVariable
}

type PipelineVariable struct {
	Type    string
	Uuid    string
	Key     string
	Value   string
	Secured bool
}

type PipelineKeyPair struct {
	Type       string
	Uuid       string
	PublicKey  string
	PrivateKey string
}

type PipelineBuildNumber struct {
	Type string
	Next int
}

type BranchingModel struct {
	Type         string
	Branch_Types []BranchType
	Development  BranchModel
	Production   BranchModel
}

type BranchType struct {
	Kind   string
	Prefix string
}

type BranchModel struct {
	Name           string
	Branch         RepositoryBranch
	Use_Mainbranch bool
}

type Environments struct {
	Page         int
	Pagelen      int
	MaxDepth     int
	Size         int
	Next         string
	Environments []Environment
}

type EnvironmentType struct {
	Name string
	Rank int
	Type string
}

type Environment struct {
	Uuid            string
	Name            string
	EnvironmentType EnvironmentType
	Rank            int
	Type            string
}

type DeploymentVariables struct {
	Page      int
	Pagelen   int
	MaxDepth  int
	Size      int
	Next      string
	Variables []DeploymentVariable
}

type DeploymentVariable struct {
	Type    string
	Uuid    string
	Key     string
	Value   string
	Secured bool
}

type DefaultReviewer struct {
	Nickname    string
	DisplayName string `mapstructure:"display_name"`
	Type        string
	Uuid        string
	AccountId   string `mapstructure:"account_id"`
	Links       map[string]map[string]string
}

type DefaultReviewers struct {
	Page             int
	Pagelen          int
	MaxDepth         int
	Size             int
	Next             string
	DefaultReviewers []DefaultReviewer
}

func (r *Repository) Create(ro *RepositoryOptions) (*Repository, error) {
	data := r.buildRepositoryBody(ro)
	urlStr := r.c.requestUrl("/repositories/%s/%s", ro.Owner, ro.RepoSlug)
	response, err := r.c.execute("POST", urlStr, data)
	if err != nil {
		return nil, err
	}

	return decodeRepository(response)
}

func (r *Repository) Fork(fo *RepositoryForkOptions) (*Repository, error) {
	data := r.buildForkBody(fo)
	urlStr := r.c.requestUrl("/repositories/%s/%s/forks", fo.FromOwner, fo.FromSlug)
	response, err := r.c.execute("POST", urlStr, data)
	if err != nil {
		return nil, err
	}

	return decodeRepository(response)
}

func (r *Repository) Get(ro *RepositoryOptions) (*Repository, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s", ro.Owner, ro.RepoSlug)
	response, err := r.c.execute("GET", urlStr, "")
	if err != nil {
		return nil, err
	}

	return decodeRepository(response)
}

func (r *Repository) ListFiles(ro *RepositoryFilesOptions) ([]RepositoryFile, error) {
	filePath := path.Join("/repositories", ro.Owner, ro.RepoSlug, "src", ro.Ref, ro.Path) + "/"
	urlStr := r.c.requestUrl(filePath)
	response, err := r.c.execute("GET", urlStr, "")
	if err != nil {
		return nil, err
	}

	return decodeRepositoryFiles(response)
}

func (r *Repository) GetFileBlob(ro *RepositoryBlobOptions) (*RepositoryBlob, error) {
	filePath := path.Join("/repositories", ro.Owner, ro.RepoSlug, "src", ro.Ref, ro.Path)
	urlStr := r.c.requestUrl(filePath)
	response, err := r.c.executeRaw("GET", urlStr, "")
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadAll(response)
	if err != nil {
		return nil, err
	}

	blob := RepositoryBlob{Content: content}

	return &blob, nil
}

func (r *Repository) WriteFileBlob(ro *RepositoryBlobWriteOptions) error {
	m := make(map[string]string)

	if ro.Author != "" {
		m["author"] = ro.Author
	}

	if ro.Message != "" {
		m["message"] = ro.Message
	}

	if ro.Branch != "" {
		m["branch"] = ro.Branch
	}

	urlStr := r.c.requestUrl("/repositories/%s/%s/src", ro.Owner, ro.RepoSlug)

	_, err := r.c.executeFileUpload("POST", urlStr, ro.FilePath, ro.FileName, ro.FileName, m)
	return err
}

// ListRefs gets all refs in the Bitbucket repository and returns them as a RepositoryRefs.
// It takes in a RepositoryRefOptions instance as its only parameter.
func (r *Repository) ListRefs(rbo *RepositoryRefOptions) (*RepositoryRefs, error) {

	params := url.Values{}
	if rbo.Query != "" {
		params.Add("q", rbo.Query)
	}

	if rbo.Sort != "" {
		params.Add("sort", rbo.Sort)
	}

	if rbo.PageNum > 0 {
		params.Add("page", strconv.Itoa(rbo.PageNum))
	}

	if rbo.Pagelen > 0 {
		params.Add("pagelen", strconv.Itoa(rbo.Pagelen))
	}

	if rbo.MaxDepth > 0 {
		params.Add("max_depth", strconv.Itoa(rbo.MaxDepth))
	}

	urlStr := r.c.requestUrl("/repositories/%s/%s/refs?%s", rbo.Owner, rbo.RepoSlug, params.Encode())
	response, err := r.c.executeRaw("GET", urlStr, "")
	if err != nil {
		return nil, err
	}
	bodyBytes, err := ioutil.ReadAll(response)
	if err != nil {
		return nil, err
	}
	bodyString := string(bodyBytes)
	return decodeRepositoryRefs(bodyString)
}

func (r *Repository) ListBranches(rbo *RepositoryBranchOptions) (*RepositoryBranches, error) {

	params := url.Values{}
	if rbo.Query != "" {
		params.Add("q", rbo.Query)
	}

	if rbo.Sort != "" {
		params.Add("sort", rbo.Sort)
	}

	if rbo.PageNum > 0 {
		params.Add("page", strconv.Itoa(rbo.PageNum))
	}

	if rbo.Pagelen > 0 {
		params.Add("pagelen", strconv.Itoa(rbo.Pagelen))
	}

	if rbo.MaxDepth > 0 {
		params.Add("max_depth", strconv.Itoa(rbo.MaxDepth))
	}

	urlStr := r.c.requestUrl("/repositories/%s/%s/refs/branches?%s", rbo.Owner, rbo.RepoSlug, params.Encode())
	response, err := r.c.executeRaw("GET", urlStr, "")
	if err != nil {
		return nil, err
	}
	bodyBytes, err := ioutil.ReadAll(response)
	if err != nil {
		return nil, err
	}
	bodyString := string(bodyBytes)
	return decodeRepositoryBranches(bodyString)
}

func (r *Repository) GetBranch(rbo *RepositoryBranchOptions) (*RepositoryBranch, error) {
	if rbo.BranchName == "" {
		return nil, errors.New("Error: Branch Name is empty")
	}
	urlStr := r.c.requestUrl("/repositories/%s/%s/refs/branches/%s", rbo.Owner, rbo.RepoSlug, rbo.BranchName)
	response, err := r.c.executeRaw("GET", urlStr, "")
	if err != nil {
		return nil, err
	}
	bodyBytes, err := ioutil.ReadAll(response)
	if err != nil {
		return nil, err
	}
	bodyString := string(bodyBytes)
	return decodeRepositoryBranch(bodyString)
}

func (r *Repository) CreateBranch(rbo *RepositoryBranchCreationOptions) (*RepositoryBranch, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s/refs/branches", rbo.Owner, rbo.RepoSlug)
	data := r.buildBranchBody(rbo)

	response, err := r.c.executeRaw("POST", urlStr, data)
	if err != nil {
		return nil, err
	}

	bodyBytes, err := ioutil.ReadAll(response)
	if err != nil {
		return nil, err
	}

	bodyString := string(bodyBytes)
	return decodeRepositoryBranchCreated(bodyString)
}

func (r *Repository) ListTags(rbo *RepositoryTagOptions) (*RepositoryTags, error) {

	params := url.Values{}
	if rbo.Query != "" {
		params.Add("q", rbo.Query)
	}

	if rbo.Sort != "" {
		params.Add("sort", rbo.Sort)
	}

	if rbo.PageNum > 0 {
		params.Add("page", strconv.Itoa(rbo.PageNum))
	}

	if rbo.Pagelen > 0 {
		params.Add("pagelen", strconv.Itoa(rbo.Pagelen))
	}

	if rbo.MaxDepth > 0 {
		params.Add("max_depth", strconv.Itoa(rbo.MaxDepth))
	}

	urlStr := r.c.requestUrl("/repositories/%s/%s/refs/tags?%s", rbo.Owner, rbo.RepoSlug, params.Encode())
	response, err := r.c.executeRaw("GET", urlStr, "")
	if err != nil {
		return nil, err
	}
	bodyBytes, err := ioutil.ReadAll(response)
	if err != nil {
		return nil, err
	}
	bodyString := string(bodyBytes)
	return decodeRepositoryTags(bodyString)
}

func (r *Repository) CreateTag(rbo *RepositoryTagCreationOptions) (*RepositoryTag, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s/refs/tags", rbo.Owner, rbo.RepoSlug)
	data := r.buildTagBody(rbo)

	response, err := r.c.executeRaw("POST", urlStr, data)
	if err != nil {
		return nil, err
	}

	bodyBytes, err := ioutil.ReadAll(response)
	if err != nil {
		return nil, err
	}

	bodyString := string(bodyBytes)
	return decodeRepositoryTagCreated(bodyString)
}

func (r *Repository) Update(ro *RepositoryOptions) (*Repository, error) {
	data := r.buildRepositoryBody(ro)
	key := ro.RepoSlug
	if ro.Uuid != "" {
		key = ro.Uuid
	}
	urlStr := r.c.requestUrl("/repositories/%s/%s", ro.Owner, key)
	response, err := r.c.execute("PUT", urlStr, data)
	if err != nil {
		return nil, err
	}
	return decodeRepository(response)
}

func (r *Repository) Delete(ro *RepositoryOptions) (interface{}, error) {
	key := ro.RepoSlug
	if ro.Uuid != "" {
		key = ro.Uuid
	}
	urlStr := r.c.requestUrl("/repositories/%s/%s", ro.Owner, key)
	return r.c.execute("DELETE", urlStr, "")
}

func (r *Repository) ListWatchers(ro *RepositoryOptions) (interface{}, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s/watchers", ro.Owner, ro.RepoSlug)
	return r.c.execute("GET", urlStr, "")
}

func (r *Repository) ListForks(ro *RepositoryOptions) (interface{}, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s/forks", ro.Owner, ro.RepoSlug)
	return r.c.execute("GET", urlStr, "")
}

func (r *Repository) ListDefaultReviewers(ro *RepositoryOptions) (*DefaultReviewers, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s/default-reviewers?pagelen=1", ro.Owner, ro.RepoSlug)

	res, err := r.c.execute("GET", urlStr, "")
	if err != nil {
		return nil, err
	}
	return decodeDefaultReviewers(res)
}

func (r *Repository) GetDefaultReviewer(rdro *RepositoryDefaultReviewerOptions) (*DefaultReviewer, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s/default-reviewers/%s", rdro.Owner, rdro.RepoSlug, rdro.Username)
	res, err := r.c.execute("GET", urlStr, "")
	if err != nil {
		return nil, fmt.Errorf("unable to get default reviewer: %w", err)
	}
	return decodeDefaultReviewer(res)
}

func (r *Repository) AddDefaultReviewer(rdro *RepositoryDefaultReviewerOptions) (*DefaultReviewer, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s/default-reviewers/%s", rdro.Owner, rdro.RepoSlug, rdro.Username)
	res, err := r.c.execute("PUT", urlStr, "")
	if err != nil {
		return nil, err
	}
	return decodeDefaultReviewer(res)
}

func (r *Repository) DeleteDefaultReviewer(rdro *RepositoryDefaultReviewerOptions) (interface{}, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s/default-reviewers/%s", rdro.Owner, rdro.RepoSlug, rdro.Username)
	return r.c.execute("DELETE", urlStr, "")
}

func (r *Repository) UpdatePipelineConfig(rpo *RepositoryPipelineOptions) (*Pipeline, error) {
	data := r.buildPipelineBody(rpo)
	urlStr := r.c.requestUrl("/repositories/%s/%s/pipelines_config", rpo.Owner, rpo.RepoSlug)
	response, err := r.c.execute("PUT", urlStr, data)
	if err != nil {
		return nil, err
	}

	return decodePipelineRepository(response)
}

func (r *Repository) ListPipelineVariables(opt *RepositoryPipelineVariablesOptions) (*PipelineVariables, error) {

	params := url.Values{}
	if opt.Query != "" {
		params.Add("q", opt.Query)
	}

	if opt.Sort != "" {
		params.Add("sort", opt.Sort)
	}

	if opt.PageNum > 0 {
		params.Add("page", strconv.Itoa(opt.PageNum))
	}

	if opt.Pagelen > 0 {
		params.Add("pagelen", strconv.Itoa(opt.Pagelen))
	}

	if opt.MaxDepth > 0 {
		params.Add("max_depth", strconv.Itoa(opt.MaxDepth))
	}

	urlStr := r.c.requestUrl("/repositories/%s/%s/pipelines_config/variables/?%s", opt.Owner, opt.RepoSlug, params.Encode())
	response, err := r.c.executeRaw("GET", urlStr, "")
	if err != nil {
		return nil, err
	}
	bodyBytes, err := ioutil.ReadAll(response)
	if err != nil {
		return nil, err
	}
	bodyString := string(bodyBytes)
	return decodePipelineVariables(bodyString)
}

func (r *Repository) AddPipelineVariable(rpvo *RepositoryPipelineVariableOptions) (*PipelineVariable, error) {
	data := r.buildPipelineVariableBody(rpvo)
	urlStr := r.c.requestUrl("/repositories/%s/%s/pipelines_config/variables/", rpvo.Owner, rpvo.RepoSlug)

	response, err := r.c.execute("POST", urlStr, data)
	if err != nil {
		return nil, err
	}

	return decodePipelineVariableRepository(response)
}

func (r *Repository) DeletePipelineVariable(opt *RepositoryPipelineVariableDeleteOptions) (interface{}, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s/pipelines_config/variables/%s", opt.Owner, opt.RepoSlug, opt.Uuid)
	return r.c.execute("DELETE", urlStr, "")
}

func (r *Repository) GetPipelineVariable(opt *RepositoryPipelineVariableOptions) (*PipelineVariable, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s/pipelines_config/variables/%s", opt.Owner, opt.RepoSlug, opt.Uuid)
	response, err := r.c.execute("GET", urlStr, "")
	if err != nil {
		return nil, err
	}
	return decodePipelineVariableRepository(response)
}

func (r *Repository) UpdatePipelineVariable(opt *RepositoryPipelineVariableOptions) (*PipelineVariable, error) {
	data := r.buildPipelineVariableBody(opt)
	urlStr := r.c.requestUrl("/repositories/%s/%s/pipelines_config/variables/%s", opt.Owner, opt.RepoSlug, opt.Uuid)
	response, err := r.c.execute("PUT", urlStr, data)
	if err != nil {
		return nil, err
	}
	return decodePipelineVariableRepository(response)
}

func (r *Repository) AddPipelineKeyPair(rpkpo *RepositoryPipelineKeyPairOptions) (*PipelineKeyPair, error) {
	data := r.buildPipelineKeyPairBody(rpkpo)
	urlStr := r.c.requestUrl("/repositories/%s/%s/pipelines_config/ssh/key_pair", rpkpo.Owner, rpkpo.RepoSlug)

	response, err := r.c.execute("PUT", urlStr, data)
	if err != nil {
		return nil, err
	}

	return decodePipelineKeyPairRepository(response)
}

func (r *Repository) UpdatePipelineBuildNumber(rpbno *RepositoryPipelineBuildNumberOptions) (*PipelineBuildNumber, error) {
	data := r.buildPipelineBuildNumberBody(rpbno)
	urlStr := r.c.requestUrl("/repositories/%s/%s/pipelines_config/build_number", rpbno.Owner, rpbno.RepoSlug)

	response, err := r.c.execute("PUT", urlStr, data)
	if err != nil {
		return nil, err
	}

	return decodePipelineBuildNumberRepository(response)
}

func (r *Repository) BranchingModel(rbmo *RepositoryBranchingModelOptions) (*BranchingModel, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s/branching-model", rbmo.Owner, rbmo.RepoSlug)
	response, err := r.c.execute("GET", urlStr, "")
	if err != nil {
		return nil, err
	}
	return decodeBranchingModel(response)
}

func (r *Repository) ListEnvironments(opt *RepositoryEnvironmentsOptions) (*Environments, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s/environments/", opt.Owner, opt.RepoSlug)
	res, err := r.c.executeRaw("GET", urlStr, "")
	if err != nil {
		return nil, err
	}

	bodyBytes, err := ioutil.ReadAll(res)
	if err != nil {
		return nil, err
	}

	bodyString := string(bodyBytes)
	return decodeEnvironments(bodyString)
}

func (r *Repository) AddEnvironment(opt *RepositoryEnvironmentOptions) (*Environment, error) {
	body := r.buildEnvironmentBody(opt)
	urlStr := r.c.requestUrl("/repositories/%s/%s/environments/", opt.Owner, opt.RepoSlug)
	res, err := r.c.execute("POST", urlStr, body)
	if err != nil {
		return nil, err
	}

	return decodeEnvironment(res)
}

func (r *Repository) DeleteEnvironment(opt *RepositoryEnvironmentDeleteOptions) (interface{}, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s/environments/%s", opt.Owner, opt.RepoSlug, opt.Uuid)
	return r.c.execute("DELETE", urlStr, "")
}

func (r *Repository) GetEnvironment(opt *RepositoryEnvironmentOptions) (*Environment, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s/environments/%s", opt.Owner, opt.RepoSlug, opt.Uuid)
	res, err := r.c.execute("GET", urlStr, "")
	if err != nil {
		return nil, err
	}

	return decodeEnvironment(res)
}

func (r *Repository) ListDeploymentVariables(opt *RepositoryDeploymentVariablesOptions) (*DeploymentVariables, error) {
	params := url.Values{}
	if opt.Query != "" {
		params.Add("q", opt.Query)
	}

	if opt.Sort != "" {
		params.Add("sort", opt.Sort)
	}

	if opt.PageNum > 0 {
		params.Add("page", strconv.Itoa(opt.PageNum))
	}

	if opt.Pagelen > 0 {
		params.Add("pagelen", strconv.Itoa(opt.Pagelen))
	}

	if opt.MaxDepth > 0 {
		params.Add("max_depth", strconv.Itoa(opt.MaxDepth))
	}

	urlStr := r.c.requestUrl("/repositories/%s/%s/deployments_config/environments/%s/variables?%s", opt.Owner, opt.RepoSlug, opt.Environment.Uuid, params.Encode())
	response, err := r.c.executeRaw("GET", urlStr, "")
	if err != nil {
		return nil, err
	}
	bodyBytes, err := ioutil.ReadAll(response)
	if err != nil {
		return nil, err
	}
	bodyString := string(bodyBytes)
	return decodeDeploymentVariables(bodyString)
}

func (r *Repository) AddDeploymentVariable(opt *RepositoryDeploymentVariableOptions) (*DeploymentVariable, error) {
	body := r.buildDeploymentVariableBody(opt)
	urlStr := r.c.requestUrl("/repositories/%s/%s/deployments_config/environments/%s/variables", opt.Owner, opt.RepoSlug, opt.Environment.Uuid)

	response, err := r.c.execute("POST", urlStr, body)
	if err != nil {
		return nil, err
	}

	return decodeDeploymentVariable(response)
}

func (r *Repository) DeleteDeploymentVariable(opt *RepositoryDeploymentVariableDeleteOptions) (interface{}, error) {
	urlStr := r.c.requestUrl("/repositories/%s/%s/deployments_config/environments/%s/variables/%s", opt.Owner, opt.RepoSlug, opt.Environment.Uuid, opt.Uuid)
	return r.c.execute("DELETE", urlStr, "")
}

func (r *Repository) UpdateDeploymentVariable(opt *RepositoryDeploymentVariableOptions) (*DeploymentVariable, error) {
	body := r.buildDeploymentVariableBody(opt)
	urlStr := r.c.requestUrl("/repositories/%s/%s/deployments_config/environments/%s/variables/%s", opt.Owner, opt.RepoSlug, opt.Environment.Uuid, opt.Uuid)

	response, err := r.c.execute("PUT", urlStr, body)
	if err != nil {
		return nil, err
	}

	return decodeDeploymentVariable(response)
}

func (r *Repository) buildRepositoryBody(ro *RepositoryOptions) string {

	body := map[string]interface{}{}

	if ro.Uuid != "" {
		body["uuid"] = ro.Uuid
	}
	if ro.RepoSlug != "" {
		body["name"] = ro.RepoSlug
	}
	if ro.Scm != "" {
		body["scm"] = ro.Scm
	}
	if ro.IsPrivate != "" {
		body["is_private"] = strings.ToLower(strings.TrimSpace(ro.IsPrivate)) != "false"
	}
	if ro.Description != "" {
		body["description"] = ro.Description
	}
	if ro.ForkPolicy != "" {
		body["fork_policy"] = ro.ForkPolicy
	}
	if ro.Language != "" {
		body["language"] = ro.Language
	}
	if ro.HasIssues != "" {
		body["has_issues"] = ro.HasIssues
	}
	if ro.HasWiki != "" {
		body["has_wiki"] = ro.HasWiki
	}
	if ro.Project != "" {
		body["project"] = map[string]string{
			"key": ro.Project,
		}
	}

	return r.buildJsonBody(body)
}

func (r *Repository) buildForkBody(fo *RepositoryForkOptions) string {

	body := map[string]interface{}{}

	if fo.Owner != "" {
		body["workspace"] = map[string]string{
			"slug": fo.Owner,
		}
	}
	if fo.Name != "" {
		body["name"] = fo.Name
	}
	if fo.IsPrivate != "" {
		body["is_private"] = strings.ToLower(strings.TrimSpace(fo.IsPrivate)) != "false"
	}
	if fo.Description != "" {
		body["description"] = fo.Description
	}
	if fo.ForkPolicy != "" {
		body["fork_policy"] = fo.ForkPolicy
	}
	if fo.Language != "" {
		body["language"] = fo.Language
	}
	if fo.HasIssues != "" {
		body["has_issues"] = fo.HasIssues
	}
	if fo.HasWiki != "" {
		body["has_wiki"] = fo.HasWiki
	}
	if fo.Project != "" {
		body["project"] = map[string]string{
			"key": fo.Project,
		}
	}

	return r.buildJsonBody(body)
}

func (r *Repository) buildPipelineBody(rpo *RepositoryPipelineOptions) string {

	body := map[string]interface{}{}

	body["enabled"] = rpo.Enabled

	return r.buildJsonBody(body)
}

func (r *Repository) buildPipelineVariableBody(rpvo *RepositoryPipelineVariableOptions) string {

	body := map[string]interface{}{}

	if rpvo.Uuid != "" {
		body["uuid"] = rpvo.Uuid
	}
	body["key"] = rpvo.Key
	body["value"] = rpvo.Value
	body["secured"] = rpvo.Secured

	return r.buildJsonBody(body)
}

func (r *Repository) buildPipelineKeyPairBody(rpkpo *RepositoryPipelineKeyPairOptions) string {

	body := map[string]interface{}{}

	if rpkpo.PrivateKey != "" {
		body["private_key"] = rpkpo.PrivateKey
	}
	if rpkpo.PublicKey != "" {
		body["public_key"] = rpkpo.PublicKey
	}

	return r.buildJsonBody(body)
}

func (r *Repository) buildPipelineBuildNumberBody(rpbno *RepositoryPipelineBuildNumberOptions) string {

	body := map[string]interface{}{}

	body["next"] = rpbno.Next

	return r.buildJsonBody(body)
}

func (r *Repository) buildBranchBody(rbo *RepositoryBranchCreationOptions) string {
	body := map[string]interface{}{
		"name": rbo.Name,
		"target": map[string]string{
			"hash": rbo.Target.Hash,
		},
	}

	return r.buildJsonBody(body)
}

func (r *Repository) buildTagBody(rbo *RepositoryTagCreationOptions) string {
	body := map[string]interface{}{
		"name": rbo.Name,
		"target": map[string]string{
			"hash": rbo.Target.Hash,
		},
	}

	return r.buildJsonBody(body)
}

func (r *Repository) buildEnvironmentBody(opt *RepositoryEnvironmentOptions) string {
	body := map[string]interface{}{}

	body["environment_type"] = map[string]interface{}{
		"name": opt.EnvironmentType.String(),
		"rank": opt.Rank,
		"type": "deployment_environment_type",
	}
	if opt.Uuid != "" {
		body["uuid"] = opt.Uuid
	}
	body["name"] = opt.Name
	body["rank"] = opt.Rank

	return r.buildJsonBody(body)
}

func (r *Repository) buildDeploymentVariableBody(opt *RepositoryDeploymentVariableOptions) string {
	body := map[string]interface{}{}

	if opt.Uuid != "" {
		body["uuid"] = opt.Uuid
	}
	body["key"] = opt.Key
	body["value"] = opt.Value
	body["secured"] = opt.Secured

	return r.buildJsonBody(body)
}

func (r *Repository) buildJsonBody(body map[string]interface{}) string {

	data, err := json.Marshal(body)
	if err != nil {
		pp.Println(err)
		os.Exit(9)
	}

	return string(data)
}

func decodeRepository(repoResponse interface{}) (*Repository, error) {
	repoMap := repoResponse.(map[string]interface{})

	if repoMap["type"] == "error" {
		return nil, DecodeError(repoMap)
	}

	var repository = new(Repository)
	err := mapstructure.Decode(repoMap, repository)
	if err != nil {
		return nil, err
	}

	return repository, nil
}

func decodeRepositoryFiles(repoResponse interface{}) ([]RepositoryFile, error) {
	repoFileMap := repoResponse.(map[string]interface{})

	if repoFileMap["type"] == "error" {
		return nil, DecodeError(repoFileMap)
	}

	var repositoryFiles = new([]RepositoryFile)
	err := mapstructure.Decode(repoFileMap["values"], repositoryFiles)
	if err != nil {
		return nil, err
	}

	return *repositoryFiles, nil
}

func decodeRepositoryRefs(refResponseStr string) (*RepositoryRefs, error) {

	var refResponseMap map[string]interface{}
	err := json.Unmarshal([]byte(refResponseStr), &refResponseMap)
	if err != nil {
		return nil, err
	}

	refArray := refResponseMap["values"].([]interface{})
	var refs []map[string]interface{}
	for _, refEntry := range refArray {
		var ref map[string]interface{}
		err = mapstructure.Decode(refEntry, &ref)
		if err == nil {
			refs = append(refs, ref)
		}
	}

	page, ok := refResponseMap["page"].(float64)
	if !ok {
		page = 0
	}

	pagelen, ok := refResponseMap["pagelen"].(float64)
	if !ok {
		pagelen = 0
	}
	max_depth, ok := refResponseMap["max_depth"].(float64)
	if !ok {
		max_depth = 0
	}
	size, ok := refResponseMap["size"].(float64)
	if !ok {
		size = 0
	}

	next, ok := refResponseMap["next"].(string)
	if !ok {
		next = ""
	}

	repositoryBranches := RepositoryRefs{
		Page:     int(page),
		Pagelen:  int(pagelen),
		MaxDepth: int(max_depth),
		Size:     int(size),
		Next:     next,
		Refs:     refs,
	}
	return &repositoryBranches, nil
}

func decodeRepositoryBranches(branchResponseStr string) (*RepositoryBranches, error) {

	var branchResponseMap map[string]interface{}
	err := json.Unmarshal([]byte(branchResponseStr), &branchResponseMap)
	if err != nil {
		return nil, err
	}

	branchArray := branchResponseMap["values"].([]interface{})
	var branches []RepositoryBranch
	for _, branchEntry := range branchArray {
		var branch RepositoryBranch
		err = mapstructure.Decode(branchEntry, &branch)
		if err == nil {
			branches = append(branches, branch)
		}
	}

	page, ok := branchResponseMap["page"].(float64)
	if !ok {
		page = 0
	}

	pagelen, ok := branchResponseMap["pagelen"].(float64)
	if !ok {
		pagelen = 0
	}
	max_depth, ok := branchResponseMap["max_depth"].(float64)
	if !ok {
		max_depth = 0
	}
	size, ok := branchResponseMap["size"].(float64)
	if !ok {
		size = 0
	}

	next, ok := branchResponseMap["next"].(string)
	if !ok {
		next = ""
	}

	repositoryBranches := RepositoryBranches{
		Page:     int(page),
		Pagelen:  int(pagelen),
		MaxDepth: int(max_depth),
		Size:     int(size),
		Next:     next,
		Branches: branches,
	}
	return &repositoryBranches, nil
}

func decodeRepositoryBranch(branchResponseStr string) (*RepositoryBranch, error) {

	var branchResponseMap map[string]interface{}
	err := json.Unmarshal([]byte(branchResponseStr), &branchResponseMap)
	if err != nil {
		return nil, err
	}
	var repositoryBranch RepositoryBranch
	err = mapstructure.Decode(branchResponseMap, &repositoryBranch)
	if err != nil {
		return nil, err
	}
	return &repositoryBranch, nil
}

func decodeRepositoryBranchCreated(branchResponseStr string) (*RepositoryBranch, error) {
	var responseBranchCreated RepositoryBranch
	err := json.Unmarshal([]byte(branchResponseStr), &responseBranchCreated)
	if err != nil {
		return nil, err
	}
	return &responseBranchCreated, nil
}

func decodeRepositoryTagCreated(tagResponseStr string) (*RepositoryTag, error) {
	var responseTagCreated RepositoryTag
	err := json.Unmarshal([]byte(tagResponseStr), &responseTagCreated)
	if err != nil {
		return nil, err
	}
	return &responseTagCreated, nil
}

func decodeRepositoryTags(tagResponseStr string) (*RepositoryTags, error) {

	var tagResponseMap map[string]interface{}
	err := json.Unmarshal([]byte(tagResponseStr), &tagResponseMap)
	if err != nil {
		return nil, err
	}

	tagArray := tagResponseMap["values"].([]interface{})
	var tags []RepositoryTag
	for _, tagEntry := range tagArray {
		var tag RepositoryTag
		err = mapstructure.Decode(tagEntry, &tag)
		if err == nil {
			tags = append(tags, tag)
		}
	}

	page, ok := tagResponseMap["page"].(float64)
	if !ok {
		page = 0
	}

	pagelen, ok := tagResponseMap["pagelen"].(float64)
	if !ok {
		pagelen = 0
	}
	max_depth, ok := tagResponseMap["max_depth"].(float64)
	if !ok {
		max_depth = 0
	}
	size, ok := tagResponseMap["size"].(float64)
	if !ok {
		size = 0
	}

	next, ok := tagResponseMap["next"].(string)
	if !ok {
		next = ""
	}

	repositoryTags := RepositoryTags{
		Page:     int(page),
		Pagelen:  int(pagelen),
		MaxDepth: int(max_depth),
		Size:     int(size),
		Next:     next,
		Tags:     tags,
	}
	return &repositoryTags, nil
}

func decodePipelineRepository(repoResponse interface{}) (*Pipeline, error) {
	repoMap := repoResponse.(map[string]interface{})

	if repoMap["type"] == "error" {
		return nil, DecodeError(repoMap)
	}

	var pipeline = new(Pipeline)
	err := mapstructure.Decode(repoMap, pipeline)
	if err != nil {
		return nil, err
	}

	return pipeline, nil
}

func decodePipelineVariables(responseStr string) (*PipelineVariables, error) {

	var responseMap map[string]interface{}
	err := json.Unmarshal([]byte(responseStr), &responseMap)
	if err != nil {
		return nil, err
	}

	values := responseMap["values"].([]interface{})
	var variables []PipelineVariable
	for _, variable := range values {
		var pipelineVariable PipelineVariable
		err = mapstructure.Decode(variable, &pipelineVariable)
		if err == nil {
			variables = append(variables, pipelineVariable)
		}
	}

	page, ok := responseMap["page"].(float64)
	if !ok {
		page = 0
	}

	pagelen, ok := responseMap["pagelen"].(float64)
	if !ok {
		pagelen = 0
	}
	max_depth, ok := responseMap["max_depth"].(float64)
	if !ok {
		max_depth = 0
	}
	size, ok := responseMap["size"].(float64)
	if !ok {
		size = 0
	}

	next, ok := responseMap["next"].(string)
	if !ok {
		next = ""
	}

	pipelineVariables := PipelineVariables{
		Page:      int(page),
		Pagelen:   int(pagelen),
		MaxDepth:  int(max_depth),
		Size:      int(size),
		Next:      next,
		Variables: variables,
	}
	return &pipelineVariables, nil
}

func decodePipelineVariableRepository(repoResponse interface{}) (*PipelineVariable, error) {
	repoMap := repoResponse.(map[string]interface{})

	if repoMap["type"] == "error" {
		return nil, DecodeError(repoMap)
	}

	var pipelineVariable = new(PipelineVariable)
	err := mapstructure.Decode(repoMap, pipelineVariable)
	if err != nil {
		return nil, err
	}

	return pipelineVariable, nil
}

func decodePipelineKeyPairRepository(repoResponse interface{}) (*PipelineKeyPair, error) {
	repoMap := repoResponse.(map[string]interface{})

	if repoMap["type"] == "error" {
		return nil, DecodeError(repoMap)
	}

	var pipelineKeyPair = new(PipelineKeyPair)
	err := mapstructure.Decode(repoMap, pipelineKeyPair)
	if err != nil {
		return nil, err
	}

	return pipelineKeyPair, nil
}

func decodePipelineBuildNumberRepository(repoResponse interface{}) (*PipelineBuildNumber, error) {
	repoMap := repoResponse.(map[string]interface{})

	if repoMap["type"] == "error" {
		return nil, DecodeError(repoMap)
	}

	var pipelineBuildNumber = new(PipelineBuildNumber)
	err := mapstructure.Decode(repoMap, pipelineBuildNumber)
	if err != nil {
		return nil, err
	}

	return pipelineBuildNumber, nil
}

func decodeBranchingModel(branchingModelResponse interface{}) (*BranchingModel, error) {
	branchingModelMap := branchingModelResponse.(map[string]interface{})

	if branchingModelMap["type"] == "error" {
		return nil, DecodeError(branchingModelMap)
	}

	var branchingModel = new(BranchingModel)
	err := mapstructure.Decode(branchingModelMap, branchingModel)
	if err != nil {
		return nil, err
	}

	return branchingModel, nil
}

func decodeEnvironments(response string) (*Environments, error) {
	var responseMap map[string]interface{}
	err := json.Unmarshal([]byte(response), &responseMap)
	if err != nil {
		return nil, err
	}

	values := responseMap["values"].([]interface{})
	var environmentsArray []Environment
	var errs error = nil
	for idx, value := range values {
		var environment Environment
		err = mapstructure.Decode(value, &environment)
		if err != nil {
			if errs == nil {
				errs = err
			} else {
				errs = fmt.Errorf("%w; environment %d: %w", errs, idx, err)
			}
		} else {
			environmentsArray = append(environmentsArray, environment)
		}
	}

	page, ok := responseMap["page"].(float64)
	if !ok {
		page = 0
	}

	pagelen, ok := responseMap["pagelen"].(float64)
	if !ok {
		pagelen = 0
	}

	max_depth, ok := responseMap["max_depth"].(float64)
	if !ok {
		max_depth = 0
	}

	size, ok := responseMap["size"].(float64)
	if !ok {
		size = 0
	}

	next, ok := responseMap["next"].(string)
	if !ok {
		next = ""
	}

	environments := Environments{
		Page:         int(page),
		Pagelen:      int(pagelen),
		MaxDepth:     int(max_depth),
		Size:         int(size),
		Next:         next,
		Environments: environmentsArray,
	}

	return &environments, nil
}

func decodeEnvironment(response interface{}) (*Environment, error) {
	responseMap := response.(map[string]interface{})

	if responseMap["type"] == "error" {
		return nil, DecodeError(responseMap)
	}

	var environment = new(Environment)
	err := mapstructure.Decode(responseMap, &environment)
	if err != nil {
		return nil, err
	}

	return environment, nil
}

func decodeDeploymentVariables(response string) (*DeploymentVariables, error) {
	var responseMap map[string]interface{}
	err := json.Unmarshal([]byte(response), &responseMap)
	if err != nil {
		return nil, err
	}

	values := responseMap["values"].([]interface{})
	var variablesArray []DeploymentVariable
	var errs error = nil
	for idx, value := range values {
		var variable DeploymentVariable
		err = mapstructure.Decode(value, &variable)
		if err != nil {
			if errs == nil {
				errs = err
			} else {
				errs = fmt.Errorf("%w; deployment variable %d: %w", errs, idx, err)
			}
		} else {
			variablesArray = append(variablesArray, variable)
		}
	}

	page, ok := responseMap["page"].(float64)
	if !ok {
		page = 0
	}

	pagelen, ok := responseMap["pagelen"].(float64)
	if !ok {
		pagelen = 0
	}

	max_depth, ok := responseMap["max_depth"].(float64)
	if !ok {
		max_depth = 0
	}

	size, ok := responseMap["size"].(float64)
	if !ok {
		size = 0
	}

	next, ok := responseMap["next"].(string)
	if !ok {
		next = ""
	}

	deploymentVariables := DeploymentVariables{
		Page:      int(page),
		Pagelen:   int(pagelen),
		MaxDepth:  int(max_depth),
		Size:      int(size),
		Next:      next,
		Variables: variablesArray,
	}

	return &deploymentVariables, nil
}

func decodeDeploymentVariable(response interface{}) (*DeploymentVariable, error) {
	responseMap := response.(map[string]interface{})

	if responseMap["type"] == "error" {
		return nil, DecodeError(responseMap)
	}

	var variable = new(DeploymentVariable)
	err := mapstructure.Decode(responseMap, &variable)
	if err != nil {
		return nil, err
	}

	return variable, nil
}

func (rf RepositoryFile) String() string {
	return rf.Path
}

func (rb RepositoryBlob) String() string {
	return string(rb.Content)
}

func decodeDefaultReviewer(response interface{}) (*DefaultReviewer, error) {
	var defaultReviewerVariable DefaultReviewer
	err := mapstructure.Decode(response, &defaultReviewerVariable)
	if err != nil {
		return nil, err
	}
	return &defaultReviewerVariable, nil
}

func decodeDefaultReviewers(response interface{}) (*DefaultReviewers, error) {
	responseMap := response.(map[string]interface{})
	values := responseMap["values"].([]interface{})
	var variables []DefaultReviewer
	for _, variable := range values {
		var defaultReviewerVariable DefaultReviewer
		err := mapstructure.Decode(variable, &defaultReviewerVariable)
		if err == nil {
			variables = append(variables, defaultReviewerVariable)
		}
	}

	page, ok := responseMap["page"].(float64)
	if !ok {
		page = 0
	}

	pagelen, ok := responseMap["pagelen"].(float64)
	if !ok {
		pagelen = 0
	}
	max_depth, ok := responseMap["max_depth"].(float64)
	if !ok {
		max_depth = 0
	}
	size, ok := responseMap["size"].(float64)
	if !ok {
		size = 0
	}

	next, ok := responseMap["next"].(string)
	if !ok {
		next = ""
	}

	defaultReviewerVariables := DefaultReviewers{
		Page:             int(page),
		Pagelen:          int(pagelen),
		MaxDepth:         int(max_depth),
		Size:             int(size),
		Next:             next,
		DefaultReviewers: variables,
	}
	return &defaultReviewerVariables, nil
}

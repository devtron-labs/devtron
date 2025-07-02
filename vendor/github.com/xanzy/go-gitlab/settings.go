//
// Copyright 2021, Sander van Harmelen
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package gitlab

import (
	"encoding/json"
	"net/http"
	"time"
)

// SettingsService handles communication with the application SettingsService
// related methods of the GitLab API.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/settings.html
type SettingsService struct {
	client *Client
}

// Settings represents the GitLab application settings.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/settings.html
//
// The available parameters have been modeled directly after the code, as the
// documentation seems to be inaccurate.
//
// https://gitlab.com/gitlab-org/gitlab/-/blob/v14.9.3-ee/lib/api/settings.rb
// https://gitlab.com/gitlab-org/gitlab/-/blob/v14.9.3-ee/lib/api/entities/application_setting.rb#L5
// https://gitlab.com/gitlab-org/gitlab/-/blob/v14.9.3-ee/app/helpers/application_settings_helper.rb#L192
// https://gitlab.com/gitlab-org/gitlab/-/blob/v14.9.3-ee/ee/lib/ee/api/helpers/settings_helpers.rb#L10
// https://gitlab.com/gitlab-org/gitlab/-/blob/v14.9.3-ee/ee/app/helpers/ee/application_settings_helper.rb#L20
type Settings struct {
	ID                                                    int                       `json:"id"`
	AbuseNotificationEmail                                string                    `json:"abuse_notification_email"`
	AdminMode                                             bool                      `json:"admin_mode"`
	AfterSignOutPath                                      string                    `json:"after_sign_out_path"`
	AfterSignUpText                                       string                    `json:"after_sign_up_text"`
	AkismetAPIKey                                         string                    `json:"akismet_api_key"`
	AkismetEnabled                                        bool                      `json:"akismet_enabled"`
	AllowAccountDeletion                                  bool                      `json:"allow_account_deletion"`
	AllowGroupOwnersToManageLDAP                          bool                      `json:"allow_group_owners_to_manage_ldap"`
	AllowLocalRequestsFromSystemHooks                     bool                      `json:"allow_local_requests_from_system_hooks"`
	AllowLocalRequestsFromWebHooksAndServices             bool                      `json:"allow_local_requests_from_web_hooks_and_services"`
	AllowProjectCreationForGuestAndBelow                  bool                      `json:"allow_project_creation_for_guest_and_below"`
	AllowRunnerRegistrationToken                          bool                      `json:"allow_runner_registration_token"`
	ArchiveBuildsInHumanReadable                          string                    `json:"archive_builds_in_human_readable"`
	ASCIIDocMaxIncludes                                   int                       `json:"asciidoc_max_includes"`
	AssetProxyAllowlist                                   []string                  `json:"asset_proxy_allowlist"`
	AssetProxyEnabled                                     bool                      `json:"asset_proxy_enabled"`
	AssetProxyURL                                         string                    `json:"asset_proxy_url"`
	AssetProxySecretKey                                   string                    `json:"asset_proxy_secret_key"`
	AuthorizedKeysEnabled                                 bool                      `json:"authorized_keys_enabled"`
	AutoBanUserOnExcessiveProjectsDownload                bool                      `json:"auto_ban_user_on_excessive_projects_download"`
	AutoDevOpsDomain                                      string                    `json:"auto_devops_domain"`
	AutoDevOpsEnabled                                     bool                      `json:"auto_devops_enabled"`
	AutomaticPurchasedStorageAllocation                   bool                      `json:"automatic_purchased_storage_allocation"`
	BulkImportConcurrentPipelineBatchLimit                int                       `json:"bulk_import_concurrent_pipeline_batch_limit"`
	BulkImportEnabled                                     bool                      `json:"bulk_import_enabled"`
	BulkImportMaxDownloadFileSize                         int                       `json:"bulk_import_max_download_file_size"`
	CanCreateGroup                                        bool                      `json:"can_create_group"`
	CheckNamespacePlan                                    bool                      `json:"check_namespace_plan"`
	CIMaxIncludes                                         int                       `json:"ci_max_includes"`
	CIMaxTotalYAMLSizeBytes                               int                       `json:"ci_max_total_yaml_size_bytes"`
	CommitEmailHostname                                   string                    `json:"commit_email_hostname"`
	ConcurrentBitbucketImportJobsLimit                    int                       `json:"concurrent_bitbucket_import_jobs_limit"`
	ConcurrentBitbucketServerImportJobsLimit              int                       `json:"concurrent_bitbucket_server_import_jobs_limit"`
	ConcurrentGitHubImportJobsLimit                       int                       `json:"concurrent_github_import_jobs_limit"`
	ContainerExpirationPoliciesEnableHistoricEntries      bool                      `json:"container_expiration_policies_enable_historic_entries"`
	ContainerRegistryCleanupTagsServiceMaxListSize        int                       `json:"container_registry_cleanup_tags_service_max_list_size"`
	ContainerRegistryDeleteTagsServiceTimeout             int                       `json:"container_registry_delete_tags_service_timeout"`
	ContainerRegistryExpirationPoliciesCaching            bool                      `json:"container_registry_expiration_policies_caching"`
	ContainerRegistryExpirationPoliciesWorkerCapacity     int                       `json:"container_registry_expiration_policies_worker_capacity"`
	ContainerRegistryImportCreatedBefore                  *time.Time                `json:"container_registry_import_created_before"`
	ContainerRegistryImportMaxRetries                     int                       `json:"container_registry_import_max_retries"`
	ContainerRegistryImportMaxStepDuration                int                       `json:"container_registry_import_max_step_duration"`
	ContainerRegistryImportMaxTagsCount                   int                       `json:"container_registry_import_max_tags_count"`
	ContainerRegistryImportStartMaxRetries                int                       `json:"container_registry_import_start_max_retries"`
	ContainerRegistryImportTargetPlan                     string                    `json:"container_registry_import_target_plan"`
	ContainerRegistryTokenExpireDelay                     int                       `json:"container_registry_token_expire_delay"`
	CreatedAt                                             *time.Time                `json:"created_at"`
	CustomHTTPCloneURLRoot                                string                    `json:"custom_http_clone_url_root"`
	DNSRebindingProtectionEnabled                         bool                      `json:"dns_rebinding_protection_enabled"`
	DSAKeyRestriction                                     int                       `json:"dsa_key_restriction"`
	DeactivateDormantUsers                                bool                      `json:"deactivate_dormant_users"`
	DeactivateDormantUsersPeriod                          int                       `json:"deactivate_dormant_users_period"`
	DecompressArchiveFileTimeout                          int                       `json:"decompress_archive_file_timeout"`
	DefaultArtifactsExpireIn                              string                    `json:"default_artifacts_expire_in"`
	DefaultBranchName                                     string                    `json:"default_branch_name"`
	DefaultBranchProtection                               int                       `json:"default_branch_protection"`
	DefaultBranchProtectionDefaults                       *BranchProtectionDefaults `json:"default_branch_protection_defaults,omitempty"`
	DefaultCiConfigPath                                   string                    `json:"default_ci_config_path"`
	DefaultGroupVisibility                                VisibilityValue           `json:"default_group_visibility"`
	DefaultPreferredLanguage                              string                    `json:"default_preferred_language"`
	DefaultProjectCreation                                int                       `json:"default_project_creation"`
	DefaultProjectDeletionProtection                      bool                      `json:"default_project_deletion_protection"`
	DefaultProjectVisibility                              VisibilityValue           `json:"default_project_visibility"`
	DefaultProjectsLimit                                  int                       `json:"default_projects_limit"`
	DefaultSnippetVisibility                              VisibilityValue           `json:"default_snippet_visibility"`
	DefaultSyntaxHighlightingTheme                        int                       `json:"default_syntax_highlighting_theme"`
	DelayedGroupDeletion                                  bool                      `json:"delayed_group_deletion"`
	DelayedProjectDeletion                                bool                      `json:"delayed_project_deletion"`
	DeleteInactiveProjects                                bool                      `json:"delete_inactive_projects"`
	DeleteUnconfirmedUsers                                bool                      `json:"delete_unconfirmed_users"`
	DeletionAdjournedPeriod                               int                       `json:"deletion_adjourned_period"`
	DiagramsnetEnabled                                    bool                      `json:"diagramsnet_enabled"`
	DiagramsnetURL                                        string                    `json:"diagramsnet_url"`
	DiffMaxFiles                                          int                       `json:"diff_max_files"`
	DiffMaxLines                                          int                       `json:"diff_max_lines"`
	DiffMaxPatchBytes                                     int                       `json:"diff_max_patch_bytes"`
	DisableAdminOAuthScopes                               bool                      `json:"disable_admin_oauth_scopes"`
	DisableFeedToken                                      bool                      `json:"disable_feed_token"`
	DisableOverridingApproversPerMergeRequest             bool                      `json:"disable_overriding_approvers_per_merge_request"`
	DisablePersonalAccessTokens                           bool                      `json:"disable_personal_access_tokens"`
	DisabledOauthSignInSources                            []string                  `json:"disabled_oauth_sign_in_sources"`
	DomainAllowlist                                       []string                  `json:"domain_allowlist"`
	DomainDenylist                                        []string                  `json:"domain_denylist"`
	DomainDenylistEnabled                                 bool                      `json:"domain_denylist_enabled"`
	DownstreamPipelineTriggerLimitPerProjectUserSHA       int                       `json:"downstream_pipeline_trigger_limit_per_project_user_sha"`
	DuoFeaturesEnabled                                    bool                      `json:"duo_features_enabled"`
	ECDSAKeyRestriction                                   int                       `json:"ecdsa_key_restriction"`
	ECDSASKKeyRestriction                                 int                       `json:"ecdsa_sk_key_restriction"`
	EKSAccessKeyID                                        string                    `json:"eks_access_key_id"`
	EKSAccountID                                          string                    `json:"eks_account_id"`
	EKSIntegrationEnabled                                 bool                      `json:"eks_integration_enabled"`
	EKSSecretAccessKey                                    string                    `json:"eks_secret_access_key"`
	Ed25519KeyRestriction                                 int                       `json:"ed25519_key_restriction"`
	Ed25519SKKeyRestriction                               int                       `json:"ed25519_sk_key_restriction"`
	ElasticsearchAWS                                      bool                      `json:"elasticsearch_aws"`
	ElasticsearchAWSAccessKey                             string                    `json:"elasticsearch_aws_access_key"`
	ElasticsearchAWSRegion                                string                    `json:"elasticsearch_aws_region"`
	ElasticsearchAWSSecretAccessKey                       string                    `json:"elasticsearch_aws_secret_access_key"`
	ElasticsearchAnalyzersKuromojiEnabled                 bool                      `json:"elasticsearch_analyzers_kuromoji_enabled"`
	ElasticsearchAnalyzersKuromojiSearch                  bool                      `json:"elasticsearch_analyzers_kuromoji_search"`
	ElasticsearchAnalyzersSmartCNEnabled                  bool                      `json:"elasticsearch_analyzers_smartcn_enabled"`
	ElasticsearchAnalyzersSmartCNSearch                   bool                      `json:"elasticsearch_analyzers_smartcn_search"`
	ElasticsearchClientRequestTimeout                     int                       `json:"elasticsearch_client_request_timeout"`
	ElasticsearchIndexedFieldLengthLimit                  int                       `json:"elasticsearch_indexed_field_length_limit"`
	ElasticsearchIndexedFileSizeLimitKB                   int                       `json:"elasticsearch_indexed_file_size_limit_kb"`
	ElasticsearchIndexing                                 bool                      `json:"elasticsearch_indexing"`
	ElasticsearchLimitIndexing                            bool                      `json:"elasticsearch_limit_indexing"`
	ElasticsearchMaxBulkConcurrency                       int                       `json:"elasticsearch_max_bulk_concurrency"`
	ElasticsearchMaxBulkSizeMB                            int                       `json:"elasticsearch_max_bulk_size_mb"`
	ElasticsearchMaxCodeIndexingConcurrency               int                       `json:"elasticsearch_max_code_indexing_concurrency"`
	ElasticsearchNamespaceIDs                             []int                     `json:"elasticsearch_namespace_ids"`
	ElasticsearchPassword                                 string                    `json:"elasticsearch_password"`
	ElasticsearchPauseIndexing                            bool                      `json:"elasticsearch_pause_indexing"`
	ElasticsearchProjectIDs                               []int                     `json:"elasticsearch_project_ids"`
	ElasticsearchReplicas                                 int                       `json:"elasticsearch_replicas"`
	ElasticsearchRequeueWorkers                           bool                      `json:"elasticsearch_requeue_workers"`
	ElasticsearchSearch                                   bool                      `json:"elasticsearch_search"`
	ElasticsearchShards                                   int                       `json:"elasticsearch_shards"`
	ElasticsearchURL                                      []string                  `json:"elasticsearch_url"`
	ElasticsearchUsername                                 string                    `json:"elasticsearch_username"`
	ElasticsearchWorkerNumberOfShards                     int                       `json:"elasticsearch_worker_number_of_shards"`
	EmailAdditionalText                                   string                    `json:"email_additional_text"`
	EmailAuthorInBody                                     bool                      `json:"email_author_in_body"`
	EmailConfirmationSetting                              string                    `json:"email_confirmation_setting"`
	EmailRestrictions                                     string                    `json:"email_restrictions"`
	EmailRestrictionsEnabled                              bool                      `json:"email_restrictions_enabled"`
	EnableArtifactExternalRedirectWarningPage             bool                      `json:"enable_artifact_external_redirect_warning_page"`
	EnabledGitAccessProtocol                              string                    `json:"enabled_git_access_protocol"`
	EnforceNamespaceStorageLimit                          bool                      `json:"enforce_namespace_storage_limit"`
	EnforcePATExpiration                                  bool                      `json:"enforce_pat_expiration"`
	EnforceSSHKeyExpiration                               bool                      `json:"enforce_ssh_key_expiration"`
	EnforceTerms                                          bool                      `json:"enforce_terms"`
	ExternalAuthClientCert                                string                    `json:"external_auth_client_cert"`
	ExternalAuthClientKey                                 string                    `json:"external_auth_client_key"`
	ExternalAuthClientKeyPass                             string                    `json:"external_auth_client_key_pass"`
	ExternalAuthorizationServiceDefaultLabel              string                    `json:"external_authorization_service_default_label"`
	ExternalAuthorizationServiceEnabled                   bool                      `json:"external_authorization_service_enabled"`
	ExternalAuthorizationServiceTimeout                   float64                   `json:"external_authorization_service_timeout"`
	ExternalAuthorizationServiceURL                       string                    `json:"external_authorization_service_url"`
	ExternalPipelineValidationServiceTimeout              int                       `json:"external_pipeline_validation_service_timeout"`
	ExternalPipelineValidationServiceToken                string                    `json:"external_pipeline_validation_service_token"`
	ExternalPipelineValidationServiceURL                  string                    `json:"external_pipeline_validation_service_url"`
	FailedLoginAttemptsUnlockPeriodInMinutes              int                       `json:"failed_login_attempts_unlock_period_in_minutes"`
	FileTemplateProjectID                                 int                       `json:"file_template_project_id"`
	FirstDayOfWeek                                        int                       `json:"first_day_of_week"`
	FlocEnabled                                           bool                      `json:"floc_enabled"`
	GeoNodeAllowedIPs                                     string                    `json:"geo_node_allowed_ips"`
	GeoStatusTimeout                                      int                       `json:"geo_status_timeout"`
	GitRateLimitUsersAlertlist                            []string                  `json:"git_rate_limit_users_alertlist"`
	GitTwoFactorSessionExpiry                             int                       `json:"git_two_factor_session_expiry"`
	GitalyTimeoutDefault                                  int                       `json:"gitaly_timeout_default"`
	GitalyTimeoutFast                                     int                       `json:"gitaly_timeout_fast"`
	GitalyTimeoutMedium                                   int                       `json:"gitaly_timeout_medium"`
	GitlabDedicatedInstance                               bool                      `json:"gitlab_dedicated_instance"`
	GitlabEnvironmentToolkitInstance                      bool                      `json:"gitlab_environment_toolkit_instance"`
	GitlabShellOperationLimit                             int                       `json:"gitlab_shell_operation_limit"`
	GitpodEnabled                                         bool                      `json:"gitpod_enabled"`
	GitpodURL                                             string                    `json:"gitpod_url"`
	GitRateLimitUsersAllowlist                            []string                  `json:"git_rate_limit_users_allowlist"`
	GloballyAllowedIPs                                    string                    `json:"globally_allowed_ips"`
	GrafanaEnabled                                        bool                      `json:"grafana_enabled"`
	GrafanaURL                                            string                    `json:"grafana_url"`
	GravatarEnabled                                       bool                      `json:"gravatar_enabled"`
	GroupDownloadExportLimit                              int                       `json:"group_download_export_limit"`
	GroupExportLimit                                      int                       `json:"group_export_limit"`
	GroupImportLimit                                      int                       `json:"group_import_limit"`
	GroupOwnersCanManageDefaultBranchProtection           bool                      `json:"group_owners_can_manage_default_branch_protection"`
	GroupRunnerTokenExpirationInterval                    int                       `json:"group_runner_token_expiration_interval"`
	HTMLEmailsEnabled                                     bool                      `json:"html_emails_enabled"`
	HashedStorageEnabled                                  bool                      `json:"hashed_storage_enabled"`
	HelpPageDocumentationBaseURL                          string                    `json:"help_page_documentation_base_url"`
	HelpPageHideCommercialContent                         bool                      `json:"help_page_hide_commercial_content"`
	HelpPageSupportURL                                    string                    `json:"help_page_support_url"`
	HelpPageText                                          string                    `json:"help_page_text"`
	HelpText                                              string                    `json:"help_text"`
	HideThirdPartyOffers                                  bool                      `json:"hide_third_party_offers"`
	HomePageURL                                           string                    `json:"home_page_url"`
	HousekeepingBitmapsEnabled                            bool                      `json:"housekeeping_bitmaps_enabled"`
	HousekeepingEnabled                                   bool                      `json:"housekeeping_enabled"`
	HousekeepingFullRepackPeriod                          int                       `json:"housekeeping_full_repack_period"`
	HousekeepingGcPeriod                                  int                       `json:"housekeeping_gc_period"`
	HousekeepingIncrementalRepackPeriod                   int                       `json:"housekeeping_incremental_repack_period"`
	HousekeepingOptimizeRepositoryPeriod                  int                       `json:"housekeeping_optimize_repository_period"`
	ImportSources                                         []string                  `json:"import_sources"`
	InactiveProjectsDeleteAfterMonths                     int                       `json:"inactive_projects_delete_after_months"`
	InactiveProjectsMinSizeMB                             int                       `json:"inactive_projects_min_size_mb"`
	InactiveProjectsSendWarningEmailAfterMonths           int                       `json:"inactive_projects_send_warning_email_after_months"`
	IncludeOptionalMetricsInServicePing                   bool                      `json:"include_optional_metrics_in_service_ping"`
	InProductMarketingEmailsEnabled                       bool                      `json:"in_product_marketing_emails_enabled"`
	InvisibleCaptchaEnabled                               bool                      `json:"invisible_captcha_enabled"`
	IssuesCreateLimit                                     int                       `json:"issues_create_limit"`
	JiraConnectApplicationKey                             string                    `json:"jira_connect_application_key"`
	JiraConnectPublicKeyStorageEnabled                    bool                      `json:"jira_connect_public_key_storage_enabled"`
	JiraConnectProxyURL                                   string                    `json:"jira_connect_proxy_url"`
	KeepLatestArtifact                                    bool                      `json:"keep_latest_artifact"`
	KrokiEnabled                                          bool                      `json:"kroki_enabled"`
	KrokiFormats                                          map[string]bool           `json:"kroki_formats"`
	KrokiURL                                              string                    `json:"kroki_url"`
	LocalMarkdownVersion                                  int                       `json:"local_markdown_version"`
	LockDuoFeaturesEnabled                                bool                      `json:"lock_duo_features_enabled"`
	LockMembershipsToLDAP                                 bool                      `json:"lock_memberships_to_ldap"`
	LoginRecaptchaProtectionEnabled                       bool                      `json:"login_recaptcha_protection_enabled"`
	MailgunEventsEnabled                                  bool                      `json:"mailgun_events_enabled"`
	MailgunSigningKey                                     string                    `json:"mailgun_signing_key"`
	MaintenanceMode                                       bool                      `json:"maintenance_mode"`
	MaintenanceModeMessage                                string                    `json:"maintenance_mode_message"`
	MavenPackageRequestsForwarding                        bool                      `json:"maven_package_requests_forwarding"`
	MaxArtifactsSize                                      int                       `json:"max_artifacts_size"`
	MaxAttachmentSize                                     int                       `json:"max_attachment_size"`
	MaxDecompressedArchiveSize                            int                       `json:"max_decompressed_archive_size"`
	MaxExportSize                                         int                       `json:"max_export_size"`
	MaxImportRemoteFileSize                               int                       `json:"max_import_remote_file_size"`
	MaxImportSize                                         int                       `json:"max_import_size"`
	MaxLoginAttempts                                      int                       `json:"max_login_attempts"`
	MaxNumberOfRepositoryDownloads                        int                       `json:"max_number_of_repository_downloads"`
	MaxNumberOfRepositoryDownloadsWithinTimePeriod        int                       `json:"max_number_of_repository_downloads_within_time_period"`
	MaxPagesSize                                          int                       `json:"max_pages_size"`
	MaxPersonalAccessTokenLifetime                        int                       `json:"max_personal_access_token_lifetime"`
	MaxSSHKeyLifetime                                     int                       `json:"max_ssh_key_lifetime"`
	MaxTerraformStateSizeBytes                            int                       `json:"max_terraform_state_size_bytes"`
	MaxYAMLDepth                                          int                       `json:"max_yaml_depth"`
	MaxYAMLSizeBytes                                      int                       `json:"max_yaml_size_bytes"`
	MetricsMethodCallThreshold                            int                       `json:"metrics_method_call_threshold"`
	MinimumPasswordLength                                 int                       `json:"minimum_password_length"`
	MirrorAvailable                                       bool                      `json:"mirror_available"`
	MirrorCapacityThreshold                               int                       `json:"mirror_capacity_threshold"`
	MirrorMaxCapacity                                     int                       `json:"mirror_max_capacity"`
	MirrorMaxDelay                                        int                       `json:"mirror_max_delay"`
	NPMPackageRequestsForwarding                          bool                      `json:"npm_package_requests_forwarding"`
	NotesCreateLimit                                      int                       `json:"notes_create_limit"`
	NotifyOnUnknownSignIn                                 bool                      `json:"notify_on_unknown_sign_in"`
	NugetSkipMetadataURLValidation                        bool                      `json:"nuget_skip_metadata_url_validation"`
	OutboundLocalRequestsAllowlistRaw                     string                    `json:"outbound_local_requests_allowlist_raw"`
	OutboundLocalRequestsWhitelist                        []string                  `json:"outbound_local_requests_whitelist"`
	PackageMetadataPURLTypes                              []int                     `json:"package_metadata_purl_types"`
	PackageRegistryAllowAnyoneToPullOption                bool                      `json:"package_registry_allow_anyone_to_pull_option"`
	PackageRegistryCleanupPoliciesWorkerCapacity          int                       `json:"package_registry_cleanup_policies_worker_capacity"`
	PagesDomainVerificationEnabled                        bool                      `json:"pages_domain_verification_enabled"`
	PasswordAuthenticationEnabledForGit                   bool                      `json:"password_authentication_enabled_for_git"`
	PasswordAuthenticationEnabledForWeb                   bool                      `json:"password_authentication_enabled_for_web"`
	PasswordNumberRequired                                bool                      `json:"password_number_required"`
	PasswordSymbolRequired                                bool                      `json:"password_symbol_required"`
	PasswordUppercaseRequired                             bool                      `json:"password_uppercase_required"`
	PasswordLowercaseRequired                             bool                      `json:"password_lowercase_required"`
	PerformanceBarAllowedGroupID                          int                       `json:"performance_bar_allowed_group_id"`
	PerformanceBarAllowedGroupPath                        string                    `json:"performance_bar_allowed_group_path"`
	PerformanceBarEnabled                                 bool                      `json:"performance_bar_enabled"`
	PersonalAccessTokenPrefix                             string                    `json:"personal_access_token_prefix"`
	PipelineLimitPerProjectUserSha                        int                       `json:"pipeline_limit_per_project_user_sha"`
	PlantumlEnabled                                       bool                      `json:"plantuml_enabled"`
	PlantumlURL                                           string                    `json:"plantuml_url"`
	PollingIntervalMultiplier                             float64                   `json:"polling_interval_multiplier,string"`
	PreventMergeRequestsAuthorApproval                    bool                      `json:"prevent_merge_request_author_approval"`
	PreventMergeRequestsCommittersApproval                bool                      `json:"prevent_merge_request_committers_approval"`
	ProjectDownloadExportLimit                            int                       `json:"project_download_export_limit"`
	ProjectExportEnabled                                  bool                      `json:"project_export_enabled"`
	ProjectExportLimit                                    int                       `json:"project_export_limit"`
	ProjectImportLimit                                    int                       `json:"project_import_limit"`
	ProjectJobsAPIRateLimit                               int                       `json:"project_jobs_api_rate_limit"`
	ProjectRunnerTokenExpirationInterval                  int                       `json:"project_runner_token_expiration_interval"`
	ProjectsAPIRateLimitUnauthenticated                   int                       `json:"projects_api_rate_limit_unauthenticated"`
	PrometheusMetricsEnabled                              bool                      `json:"prometheus_metrics_enabled"`
	ProtectedCIVariables                                  bool                      `json:"protected_ci_variables"`
	PseudonymizerEnabled                                  bool                      `json:"pseudonymizer_enabled"`
	PushEventActivitiesLimit                              int                       `json:"push_event_activities_limit"`
	PushEventHooksLimit                                   int                       `json:"push_event_hooks_limit"`
	PyPIPackageRequestsForwarding                         bool                      `json:"pypi_package_requests_forwarding"`
	RSAKeyRestriction                                     int                       `json:"rsa_key_restriction"`
	RateLimitingResponseText                              string                    `json:"rate_limiting_response_text"`
	RawBlobRequestLimit                                   int                       `json:"raw_blob_request_limit"`
	RecaptchaEnabled                                      bool                      `json:"recaptcha_enabled"`
	RecaptchaPrivateKey                                   string                    `json:"recaptcha_private_key"`
	RecaptchaSiteKey                                      string                    `json:"recaptcha_site_key"`
	ReceiveMaxInputSize                                   int                       `json:"receive_max_input_size"`
	ReceptiveClusterAgentsEnabled                         bool                      `json:"receptive_cluster_agents_enabled"`
	RememberMeEnabled                                     bool                      `json:"remember_me_enabled"`
	RepositoryChecksEnabled                               bool                      `json:"repository_checks_enabled"`
	RepositorySizeLimit                                   int                       `json:"repository_size_limit"`
	RepositoryStorages                                    []string                  `json:"repository_storages"`
	RepositoryStoragesWeighted                            map[string]int            `json:"repository_storages_weighted"`
	RequireAdminApprovalAfterUserSignup                   bool                      `json:"require_admin_approval_after_user_signup"`
	RequireAdminTwoFactorAuthentication                   bool                      `json:"require_admin_two_factor_authentication"`
	RequirePersonalAccessTokenExpiry                      bool                      `json:"require_personal_access_token_expiry"`
	RequireTwoFactorAuthentication                        bool                      `json:"require_two_factor_authentication"`
	RestrictedVisibilityLevels                            []VisibilityValue         `json:"restricted_visibility_levels"`
	RunnerTokenExpirationInterval                         int                       `json:"runner_token_expiration_interval"`
	SearchRateLimit                                       int                       `json:"search_rate_limit"`
	SearchRateLimitUnauthenticated                        int                       `json:"search_rate_limit_unauthenticated"`
	SecretDetectionRevocationTokenTypesURL                string                    `json:"secret_detection_revocation_token_types_url"`
	SecretDetectionTokenRevocationEnabled                 bool                      `json:"secret_detection_token_revocation_enabled"`
	SecretDetectionTokenRevocationToken                   string                    `json:"secret_detection_token_revocation_token"`
	SecretDetectionTokenRevocationURL                     string                    `json:"secret_detection_token_revocation_url"`
	SecurityApprovalPoliciesLimit                         int                       `json:"security_approval_policies_limit"`
	SecurityPolicyGlobalGroupApproversEnabled             bool                      `json:"security_policy_global_group_approvers_enabled"`
	SecurityTXTContent                                    string                    `json:"security_txt_content"`
	SendUserConfirmationEmail                             bool                      `json:"send_user_confirmation_email"`
	SentryClientsideDSN                                   string                    `json:"sentry_clientside_dsn"`
	SentryDSN                                             string                    `json:"sentry_dsn"`
	SentryEnabled                                         bool                      `json:"sentry_enabled"`
	SentryEnvironment                                     string                    `json:"sentry_environment"`
	ServiceAccessTokensExpirationEnforced                 bool                      `json:"service_access_tokens_expiration_enforced"`
	SessionExpireDelay                                    int                       `json:"session_expire_delay"`
	SharedRunnersEnabled                                  bool                      `json:"shared_runners_enabled"`
	SharedRunnersMinutes                                  int                       `json:"shared_runners_minutes"`
	SharedRunnersText                                     string                    `json:"shared_runners_text"`
	SidekiqJobLimiterCompressionThresholdBytes            int                       `json:"sidekiq_job_limiter_compression_threshold_bytes"`
	SidekiqJobLimiterLimitBytes                           int                       `json:"sidekiq_job_limiter_limit_bytes"`
	SidekiqJobLimiterMode                                 string                    `json:"sidekiq_job_limiter_mode"`
	SignInText                                            string                    `json:"sign_in_text"`
	SignupEnabled                                         bool                      `json:"signup_enabled"`
	SilentAdminExportsEnabled                             bool                      `json:"silent_admin_exports_enabled"`
	SilentModeEnabled                                     bool                      `json:"silent_mode_enabled"`
	SlackAppEnabled                                       bool                      `json:"slack_app_enabled"`
	SlackAppID                                            string                    `json:"slack_app_id"`
	SlackAppSecret                                        string                    `json:"slack_app_secret"`
	SlackAppSigningSecret                                 string                    `json:"slack_app_signing_secret"`
	SlackAppVerificationToken                             string                    `json:"slack_app_verification_token"`
	SnippetSizeLimit                                      int                       `json:"snippet_size_limit"`
	SnowplowAppID                                         string                    `json:"snowplow_app_id"`
	SnowplowCollectorHostname                             string                    `json:"snowplow_collector_hostname"`
	SnowplowCookieDomain                                  string                    `json:"snowplow_cookie_domain"`
	SnowplowDatabaseCollectorHostname                     string                    `json:"snowplow_database_collector_hostname"`
	SnowplowEnabled                                       bool                      `json:"snowplow_enabled"`
	SourcegraphEnabled                                    bool                      `json:"sourcegraph_enabled"`
	SourcegraphPublicOnly                                 bool                      `json:"sourcegraph_public_only"`
	SourcegraphURL                                        string                    `json:"sourcegraph_url"`
	SpamCheckAPIKey                                       string                    `json:"spam_check_api_key"`
	SpamCheckEndpointEnabled                              bool                      `json:"spam_check_endpoint_enabled"`
	SpamCheckEndpointURL                                  string                    `json:"spam_check_endpoint_url"`
	StaticObjectsExternalStorageAuthToken                 string                    `json:"static_objects_external_storage_auth_token"`
	StaticObjectsExternalStorageURL                       string                    `json:"static_objects_external_storage_url"`
	SuggestPipelineEnabled                                bool                      `json:"suggest_pipeline_enabled"`
	TerminalMaxSessionTime                                int                       `json:"terminal_max_session_time"`
	Terms                                                 string                    `json:"terms"`
	ThrottleAuthenticatedAPIEnabled                       bool                      `json:"throttle_authenticated_api_enabled"`
	ThrottleAuthenticatedAPIPeriodInSeconds               int                       `json:"throttle_authenticated_api_period_in_seconds"`
	ThrottleAuthenticatedAPIRequestsPerPeriod             int                       `json:"throttle_authenticated_api_requests_per_period"`
	ThrottleAuthenticatedDeprecatedAPIEnabled             bool                      `json:"throttle_authenticated_deprecated_api_enabled"`
	ThrottleAuthenticatedDeprecatedAPIPeriodInSeconds     int                       `json:"throttle_authenticated_deprecated_api_period_in_seconds"`
	ThrottleAuthenticatedDeprecatedAPIRequestsPerPeriod   int                       `json:"throttle_authenticated_deprecated_api_requests_per_period"`
	ThrottleAuthenticatedFilesAPIEnabled                  bool                      `json:"throttle_authenticated_files_api_enabled"`
	ThrottleAuthenticatedFilesAPIPeriodInSeconds          int                       `json:"throttle_authenticated_files_api_period_in_seconds"`
	ThrottleAuthenticatedFilesAPIRequestsPerPeriod        int                       `json:"throttle_authenticated_files_api_requests_per_period"`
	ThrottleAuthenticatedGitLFSEnabled                    bool                      `json:"throttle_authenticated_git_lfs_enabled"`
	ThrottleAuthenticatedGitLFSPeriodInSeconds            int                       `json:"throttle_authenticated_git_lfs_period_in_seconds"`
	ThrottleAuthenticatedGitLFSRequestsPerPeriod          int                       `json:"throttle_authenticated_git_lfs_requests_per_period"`
	ThrottleAuthenticatedPackagesAPIEnabled               bool                      `json:"throttle_authenticated_packages_api_enabled"`
	ThrottleAuthenticatedPackagesAPIPeriodInSeconds       int                       `json:"throttle_authenticated_packages_api_period_in_seconds"`
	ThrottleAuthenticatedPackagesAPIRequestsPerPeriod     int                       `json:"throttle_authenticated_packages_api_requests_per_period"`
	ThrottleAuthenticatedWebEnabled                       bool                      `json:"throttle_authenticated_web_enabled"`
	ThrottleAuthenticatedWebPeriodInSeconds               int                       `json:"throttle_authenticated_web_period_in_seconds"`
	ThrottleAuthenticatedWebRequestsPerPeriod             int                       `json:"throttle_authenticated_web_requests_per_period"`
	ThrottleIncidentManagementNotificationEnabled         bool                      `json:"throttle_incident_management_notification_enabled"`
	ThrottleIncidentManagementNotificationPerPeriod       int                       `json:"throttle_incident_management_notification_per_period"`
	ThrottleIncidentManagementNotificationPeriodInSeconds int                       `json:"throttle_incident_management_notification_period_in_seconds"`
	ThrottleProtectedPathsEnabled                         bool                      `json:"throttle_protected_paths_enabled"`
	ThrottleProtectedPathsPeriodInSeconds                 int                       `json:"throttle_protected_paths_period_in_seconds"`
	ThrottleProtectedPathsRequestsPerPeriod               int                       `json:"throttle_protected_paths_requests_per_period"`
	ThrottleUnauthenticatedAPIEnabled                     bool                      `json:"throttle_unauthenticated_api_enabled"`
	ThrottleUnauthenticatedAPIPeriodInSeconds             int                       `json:"throttle_unauthenticated_api_period_in_seconds"`
	ThrottleUnauthenticatedAPIRequestsPerPeriod           int                       `json:"throttle_unauthenticated_api_requests_per_period"`
	ThrottleUnauthenticatedDeprecatedAPIEnabled           bool                      `json:"throttle_unauthenticated_deprecated_api_enabled"`
	ThrottleUnauthenticatedDeprecatedAPIPeriodInSeconds   int                       `json:"throttle_unauthenticated_deprecated_api_period_in_seconds"`
	ThrottleUnauthenticatedDeprecatedAPIRequestsPerPeriod int                       `json:"throttle_unauthenticated_deprecated_api_requests_per_period"`
	ThrottleUnauthenticatedFilesAPIEnabled                bool                      `json:"throttle_unauthenticated_files_api_enabled"`
	ThrottleUnauthenticatedFilesAPIPeriodInSeconds        int                       `json:"throttle_unauthenticated_files_api_period_in_seconds"`
	ThrottleUnauthenticatedFilesAPIRequestsPerPeriod      int                       `json:"throttle_unauthenticated_files_api_requests_per_period"`
	ThrottleUnauthenticatedGitLFSEnabled                  bool                      `json:"throttle_unauthenticated_git_lfs_enabled"`
	ThrottleUnauthenticatedGitLFSPeriodInSeconds          int                       `json:"throttle_unauthenticated_git_lfs_period_in_seconds"`
	ThrottleUnauthenticatedGitLFSRequestsPerPeriod        int                       `json:"throttle_unauthenticated_git_lfs_requests_per_period"`
	ThrottleUnauthenticatedPackagesAPIEnabled             bool                      `json:"throttle_unauthenticated_packages_api_enabled"`
	ThrottleUnauthenticatedPackagesAPIPeriodInSeconds     int                       `json:"throttle_unauthenticated_packages_api_period_in_seconds"`
	ThrottleUnauthenticatedPackagesAPIRequestsPerPeriod   int                       `json:"throttle_unauthenticated_packages_api_requests_per_period"`
	ThrottleUnauthenticatedWebEnabled                     bool                      `json:"throttle_unauthenticated_web_enabled"`
	ThrottleUnauthenticatedWebPeriodInSeconds             int                       `json:"throttle_unauthenticated_web_period_in_seconds"`
	ThrottleUnauthenticatedWebRequestsPerPeriod           int                       `json:"throttle_unauthenticated_web_requests_per_period"`
	TimeTrackingLimitToHours                              bool                      `json:"time_tracking_limit_to_hours"`
	TwoFactorGracePeriod                                  int                       `json:"two_factor_grace_period"`
	UnconfirmedUsersDeleteAfterDays                       int                       `json:"unconfirmed_users_delete_after_days"`
	UniqueIPsLimitEnabled                                 bool                      `json:"unique_ips_limit_enabled"`
	UniqueIPsLimitPerUser                                 int                       `json:"unique_ips_limit_per_user"`
	UniqueIPsLimitTimeWindow                              int                       `json:"unique_ips_limit_time_window"`
	UpdateRunnerVersionsEnabled                           bool                      `json:"update_runner_versions_enabled"`
	UpdatedAt                                             *time.Time                `json:"updated_at"`
	UpdatingNameDisabledForUsers                          bool                      `json:"updating_name_disabled_for_users"`
	UsagePingEnabled                                      bool                      `json:"usage_ping_enabled"`
	UsagePingFeaturesEnabled                              bool                      `json:"usage_ping_features_enabled"`
	UseClickhouseForAnalytics                             bool                      `json:"use_clickhouse_for_analytics"`
	UserDeactivationEmailsEnabled                         bool                      `json:"user_deactivation_emails_enabled"`
	UserDefaultExternal                                   bool                      `json:"user_default_external"`
	UserDefaultInternalRegex                              string                    `json:"user_default_internal_regex"`
	UserDefaultsToPrivateProfile                          bool                      `json:"user_defaults_to_private_profile"`
	UserOauthApplications                                 bool                      `json:"user_oauth_applications"`
	UserShowAddSSHKeyMessage                              bool                      `json:"user_show_add_ssh_key_message"`
	UsersGetByIDLimit                                     int                       `json:"users_get_by_id_limit"`
	UsersGetByIDLimitAllowlistRaw                         string                    `json:"users_get_by_id_limit_allowlist_raw"`
	ValidRunnerRegistrars                                 []string                  `json:"valid_runner_registrars"`
	VersionCheckEnabled                                   bool                      `json:"version_check_enabled"`
	WebIDEClientsidePreviewEnabled                        bool                      `json:"web_ide_clientside_preview_enabled"`
	WhatsNewVariant                                       string                    `json:"whats_new_variant"`
	WikiPageMaxContentBytes                               int                       `json:"wiki_page_max_content_bytes"`

	// Deprecated: Use AbuseNotificationEmail instead.
	AdminNotificationEmail string `json:"admin_notification_email"`
	// Deprecated: Use AllowLocalRequestsFromWebHooksAndServices instead.
	AllowLocalRequestsFromHooksAndServices bool `json:"allow_local_requests_from_hooks_and_services"`
	// Deprecated: Use AssetProxyAllowlist instead.
	AssetProxyWhitelist []string `json:"asset_proxy_whitelist"`
	// Deprecated: Use ThrottleUnauthenticatedWebEnabled or ThrottleUnauthenticatedAPIEnabled instead. (Deprecated in GitLab 14.3)
	ThrottleUnauthenticatedEnabled bool `json:"throttle_unauthenticated_enabled"`
	// Deprecated: Use ThrottleUnauthenticatedWebPeriodInSeconds or ThrottleUnauthenticatedAPIPeriodInSeconds instead. (Deprecated in GitLab 14.3)
	ThrottleUnauthenticatedPeriodInSeconds int `json:"throttle_unauthenticated_period_in_seconds"`
	// Deprecated: Use ThrottleUnauthenticatedWebRequestsPerPeriod or ThrottleUnauthenticatedAPIRequestsPerPeriod instead. (Deprecated in GitLab 14.3)
	ThrottleUnauthenticatedRequestsPerPeriod int `json:"throttle_unauthenticated_requests_per_period"`
	// Deprecated: Replaced by SearchRateLimit in GitLab 14.9 (removed in 15.0).
	UserEmailLookupLimit int `json:"user_email_lookup_limit"`
}

// Settings requires a custom unmarshaller in order to properly unmarshal
// `container_registry_import_created_before` which is either a time.Time or
// an empty string if no value is set.
func (s *Settings) UnmarshalJSON(data []byte) error {
	type Alias Settings

	raw := make(map[string]interface{})
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}

	// If empty string, remove the value to leave it nil in the response.
	if v, ok := raw["container_registry_import_created_before"]; ok && v == "" {
		delete(raw, "container_registry_import_created_before")

		data, err = json.Marshal(raw)
		if err != nil {
			return err
		}
	}

	return json.Unmarshal(data, (*Alias)(s))
}

func (s Settings) String() string {
	return Stringify(s)
}

// GetSettings gets the current application settings.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/settings.html#get-current-application-settings
func (s *SettingsService) GetSettings(options ...RequestOptionFunc) (*Settings, *Response, error) {
	req, err := s.client.NewRequest(http.MethodGet, "application/settings", nil, options)
	if err != nil {
		return nil, nil, err
	}

	as := new(Settings)
	resp, err := s.client.Do(req, as)
	if err != nil {
		return nil, resp, err
	}

	return as, resp, nil
}

// UpdateSettingsOptions represents the available UpdateSettings() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/settings.html#change-application-settings
type UpdateSettingsOptions struct {
	AbuseNotificationEmail                                *string                          `url:"abuse_notification_email,omitempty" json:"abuse_notification_email,omitempty"`
	AdminMode                                             *bool                            `url:"admin_mode,omitempty" json:"admin_mode,omitempty"`
	AdminNotificationEmail                                *string                          `url:"admin_notification_email,omitempty" json:"admin_notification_email,omitempty"`
	AfterSignOutPath                                      *string                          `url:"after_sign_out_path,omitempty" json:"after_sign_out_path,omitempty"`
	AfterSignUpText                                       *string                          `url:"after_sign_up_text,omitempty" json:"after_sign_up_text,omitempty"`
	AkismetAPIKey                                         *string                          `url:"akismet_api_key,omitempty" json:"akismet_api_key,omitempty"`
	AkismetEnabled                                        *bool                            `url:"akismet_enabled,omitempty" json:"akismet_enabled,omitempty"`
	AllowAccountDeletion                                  *bool                            `url:"allow_account_deletion,omitempty" json:"allow_account_deletion,omitempty"`
	AllowGroupOwnersToManageLDAP                          *bool                            `url:"allow_group_owners_to_manage_ldap,omitempty" json:"allow_group_owners_to_manage_ldap,omitempty"`
	AllowLocalRequestsFromHooksAndServices                *bool                            `url:"allow_local_requests_from_hooks_and_services,omitempty" json:"allow_local_requests_from_hooks_and_services,omitempty"`
	AllowLocalRequestsFromSystemHooks                     *bool                            `url:"allow_local_requests_from_system_hooks,omitempty" json:"allow_local_requests_from_system_hooks,omitempty"`
	AllowLocalRequestsFromWebHooksAndServices             *bool                            `url:"allow_local_requests_from_web_hooks_and_services,omitempty" json:"allow_local_requests_from_web_hooks_and_services,omitempty"`
	AllowProjectCreationForGuestAndBelow                  *bool                            `url:"allow_project_creation_for_guest_and_below,omitempty" json:"allow_project_creation_for_guest_and_below,omitempty"`
	AllowRunnerRegistrationToken                          *bool                            `url:"allow_runner_registration_token,omitempty" json:"allow_runner_registration_token,omitempty"`
	ArchiveBuildsInHumanReadable                          *string                          `url:"archive_builds_in_human_readable,omitempty" json:"archive_builds_in_human_readable,omitempty"`
	ASCIIDocMaxIncludes                                   *int                             `url:"asciidoc_max_includes,omitempty" json:"asciidoc_max_includes,omitempty"`
	AssetProxyAllowlist                                   *[]string                        `url:"asset_proxy_allowlist,omitempty" json:"asset_proxy_allowlist,omitempty"`
	AssetProxyEnabled                                     *bool                            `url:"asset_proxy_enabled,omitempty" json:"asset_proxy_enabled,omitempty"`
	AssetProxySecretKey                                   *string                          `url:"asset_proxy_secret_key,omitempty" json:"asset_proxy_secret_key,omitempty"`
	AssetProxyURL                                         *string                          `url:"asset_proxy_url,omitempty" json:"asset_proxy_url,omitempty"`
	AssetProxyWhitelist                                   *[]string                        `url:"asset_proxy_whitelist,omitempty" json:"asset_proxy_whitelist,omitempty"`
	AuthorizedKeysEnabled                                 *bool                            `url:"authorized_keys_enabled,omitempty" json:"authorized_keys_enabled,omitempty"`
	AutoBanUserOnExcessiveProjectsDownload                *bool                            `url:"auto_ban_user_on_excessive_projects_download,omitempty" json:"auto_ban_user_on_excessive_projects_download,omitempty"`
	AutoDevOpsDomain                                      *string                          `url:"auto_devops_domain,omitempty" json:"auto_devops_domain,omitempty"`
	AutoDevOpsEnabled                                     *bool                            `url:"auto_devops_enabled,omitempty" json:"auto_devops_enabled,omitempty"`
	AutomaticPurchasedStorageAllocation                   *bool                            `url:"automatic_purchased_storage_allocation,omitempty" json:"automatic_purchased_storage_allocation,omitempty"`
	BulkImportConcurrentPipelineBatchLimit                *int                             `url:"bulk_import_concurrent_pipeline_batch_limit,omitempty" json:"bulk_import_concurrent_pipeline_batch_limit,omitempty"`
	BulkImportEnabled                                     *bool                            `url:"bulk_import_enabled,omitempty" json:"bulk_import_enabled,omitempty"`
	BulkImportMaxDownloadFileSize                         *int                             `url:"bulk_import_max_download_file_size,omitempty" json:"bulk_import_max_download_file_size,omitempty"`
	CanCreateGroup                                        *bool                            `url:"can_create_group,omitempty" json:"can_create_group,omitempty"`
	CheckNamespacePlan                                    *bool                            `url:"check_namespace_plan,omitempty" json:"check_namespace_plan,omitempty"`
	CIMaxIncludes                                         *int                             `url:"ci_max_includes,omitempty" json:"ci_max_includes,omitempty"`
	CIMaxTotalYAMLSizeBytes                               *int                             `url:"ci_max_total_yaml_size_bytes,omitempty" json:"ci_max_total_yaml_size_bytes,omitempty"`
	CommitEmailHostname                                   *string                          `url:"commit_email_hostname,omitempty" json:"commit_email_hostname,omitempty"`
	ConcurrentBitbucketImportJobsLimit                    *int                             `url:"concurrent_bitbucket_import_jobs_limit,omitempty" json:"concurrent_bitbucket_import_jobs_limit,omitempty"`
	ConcurrentBitbucketServerImportJobsLimit              *int                             `url:"concurrent_bitbucket_server_import_jobs_limit,omitempty" json:"concurrent_bitbucket_server_import_jobs_limit,omitempty"`
	ConcurrentGitHubImportJobsLimit                       *int                             `url:"concurrent_github_import_jobs_limit,omitempty" json:"concurrent_github_import_jobs_limit,omitempty"`
	ContainerExpirationPoliciesEnableHistoricEntries      *bool                            `url:"container_expiration_policies_enable_historic_entries,omitempty" json:"container_expiration_policies_enable_historic_entries,omitempty"`
	ContainerRegistryCleanupTagsServiceMaxListSize        *int                             `url:"container_registry_cleanup_tags_service_max_list_size,omitempty" json:"container_registry_cleanup_tags_service_max_list_size,omitempty"`
	ContainerRegistryDeleteTagsServiceTimeout             *int                             `url:"container_registry_delete_tags_service_timeout,omitempty" json:"container_registry_delete_tags_service_timeout,omitempty"`
	ContainerRegistryExpirationPoliciesCaching            *bool                            `url:"container_registry_expiration_policies_caching,omitempty" json:"container_registry_expiration_policies_caching,omitempty"`
	ContainerRegistryExpirationPoliciesWorkerCapacity     *int                             `url:"container_registry_expiration_policies_worker_capacity,omitempty" json:"container_registry_expiration_policies_worker_capacity,omitempty"`
	ContainerRegistryImportCreatedBefore                  *time.Time                       `url:"container_registry_import_created_before,omitempty" json:"container_registry_import_created_before,omitempty"`
	ContainerRegistryImportMaxRetries                     *int                             `url:"container_registry_import_max_retries,omitempty" json:"container_registry_import_max_retries,omitempty"`
	ContainerRegistryImportMaxStepDuration                *int                             `url:"container_registry_import_max_step_duration,omitempty" json:"container_registry_import_max_step_duration,omitempty"`
	ContainerRegistryImportMaxTagsCount                   *int                             `url:"container_registry_import_max_tags_count,omitempty" json:"container_registry_import_max_tags_count,omitempty"`
	ContainerRegistryImportStartMaxRetries                *int                             `url:"container_registry_import_start_max_retries,omitempty" json:"container_registry_import_start_max_retries,omitempty"`
	ContainerRegistryImportTargetPlan                     *string                          `url:"container_registry_import_target_plan,omitempty" json:"container_registry_import_target_plan,omitempty"`
	ContainerRegistryTokenExpireDelay                     *int                             `url:"container_registry_token_expire_delay,omitempty" json:"container_registry_token_expire_delay,omitempty"`
	CustomHTTPCloneURLRoot                                *string                          `url:"custom_http_clone_url_root,omitempty" json:"custom_http_clone_url_root,omitempty"`
	DNSRebindingProtectionEnabled                         *bool                            `url:"dns_rebinding_protection_enabled,omitempty" json:"dns_rebinding_protection_enabled,omitempty"`
	DSAKeyRestriction                                     *int                             `url:"dsa_key_restriction,omitempty" json:"dsa_key_restriction,omitempty"`
	DeactivateDormantUsers                                *bool                            `url:"deactivate_dormant_users,omitempty" json:"deactivate_dormant_users,omitempty"`
	DeactivateDormantUsersPeriod                          *int                             `url:"deactivate_dormant_users_period,omitempty" json:"deactivate_dormant_users_period,omitempty"`
	DecompressArchiveFileTimeout                          *int                             `url:"decompress_archive_file_timeout,omitempty" json:"decompress_archive_file_timeout,omitempty"`
	DefaultArtifactsExpireIn                              *string                          `url:"default_artifacts_expire_in,omitempty" json:"default_artifacts_expire_in,omitempty"`
	DefaultBranchName                                     *string                          `url:"default_branch_name,omitempty" json:"default_branch_name,omitempty"`
	DefaultBranchProtection                               *int                             `url:"default_branch_protection,omitempty" json:"default_branch_protection,omitempty"`
	DefaultBranchProtectionDefaults                       *BranchProtectionDefaultsOptions `url:"default_branch_protection_defaults,omitempty" json:"default_branch_protection_defaults,omitempty"`
	DefaultCiConfigPath                                   *string                          `url:"default_ci_config_path,omitempty" json:"default_ci_config_path,omitempty"`
	DefaultGroupVisibility                                *VisibilityValue                 `url:"default_group_visibility,omitempty" json:"default_group_visibility,omitempty"`
	DefaultPreferredLanguage                              *string                          `url:"default_preferred_language,omitempty" json:"default_preferred_language,omitempty"`
	DefaultProjectCreation                                *int                             `url:"default_project_creation,omitempty" json:"default_project_creation,omitempty"`
	DefaultProjectDeletionProtection                      *bool                            `url:"default_project_deletion_protection,omitempty" json:"default_project_deletion_protection,omitempty"`
	DefaultProjectVisibility                              *VisibilityValue                 `url:"default_project_visibility,omitempty" json:"default_project_visibility,omitempty"`
	DefaultProjectsLimit                                  *int                             `url:"default_projects_limit,omitempty" json:"default_projects_limit,omitempty"`
	DefaultSnippetVisibility                              *VisibilityValue                 `url:"default_snippet_visibility,omitempty" json:"default_snippet_visibility,omitempty"`
	DefaultSyntaxHighlightingTheme                        *int                             `url:"default_syntax_highlighting_theme,omitempty" json:"default_syntax_highlighting_theme,omitempty"`
	DelayedGroupDeletion                                  *bool                            `url:"delayed_group_deletion,omitempty" json:"delayed_group_deletion,omitempty"`
	DelayedProjectDeletion                                *bool                            `url:"delayed_project_deletion,omitempty" json:"delayed_project_deletion,omitempty"`
	DeleteInactiveProjects                                *bool                            `url:"delete_inactive_projects,omitempty" json:"delete_inactive_projects,omitempty"`
	DeleteUnconfirmedUsers                                *bool                            `url:"delete_unconfirmed_users,omitempty" json:"delete_unconfirmed_users,omitempty"`
	DeletionAdjournedPeriod                               *int                             `url:"deletion_adjourned_period,omitempty" json:"deletion_adjourned_period,omitempty"`
	DiagramsnetEnabled                                    *bool                            `url:"diagramsnet_enabled,omitempty" json:"diagramsnet_enabled,omitempty"`
	DiagramsnetURL                                        *string                          `url:"diagramsnet_url,omitempty" json:"diagramsnet_url,omitempty"`
	DiffMaxFiles                                          *int                             `url:"diff_max_files,omitempty" json:"diff_max_files,omitempty"`
	DiffMaxLines                                          *int                             `url:"diff_max_lines,omitempty" json:"diff_max_lines,omitempty"`
	DiffMaxPatchBytes                                     *int                             `url:"diff_max_patch_bytes,omitempty" json:"diff_max_patch_bytes,omitempty"`
	DisableFeedToken                                      *bool                            `url:"disable_feed_token,omitempty" json:"disable_feed_token,omitempty"`
	DisableAdminOAuthScopes                               *bool                            `url:"disable_admin_oauth_scopes,omitempty" json:"disable_admin_oauth_scopes,omitempty"`
	DisableOverridingApproversPerMergeRequest             *bool                            `url:"disable_overriding_approvers_per_merge_request,omitempty" json:"disable_overriding_approvers_per_merge_request,omitempty"`
	DisablePersonalAccessTokens                           *bool                            `url:"disable_personal_access_tokens,omitempty" json:"disable_personal_access_tokens,omitempty"`
	DisabledOauthSignInSources                            *[]string                        `url:"disabled_oauth_sign_in_sources,omitempty" json:"disabled_oauth_sign_in_sources,omitempty"`
	DomainAllowlist                                       *[]string                        `url:"domain_allowlist,omitempty" json:"domain_allowlist,omitempty"`
	DomainDenylist                                        *[]string                        `url:"domain_denylist,omitempty" json:"domain_denylist,omitempty"`
	DomainDenylistEnabled                                 *bool                            `url:"domain_denylist_enabled,omitempty" json:"domain_denylist_enabled,omitempty"`
	DownstreamPipelineTriggerLimitPerProjectUserSHA       *int                             `url:"downstream_pipeline_trigger_limit_per_project_user_sha,omitempty" json:"downstream_pipeline_trigger_limit_per_project_user_sha,omitempty"`
	DuoFeaturesEnabled                                    *bool                            `url:"duo_features_enabled,omitempty" json:"duo_features_enabled,omitempty"`
	ECDSAKeyRestriction                                   *int                             `url:"ecdsa_key_restriction,omitempty" json:"ecdsa_key_restriction,omitempty"`
	ECDSASKKeyRestriction                                 *int                             `url:"ecdsa_sk_key_restriction,omitempty" json:"ecdsa_sk_key_restriction,omitempty"`
	EKSAccessKeyID                                        *string                          `url:"eks_access_key_id,omitempty" json:"eks_access_key_id,omitempty"`
	EKSAccountID                                          *string                          `url:"eks_account_id,omitempty" json:"eks_account_id,omitempty"`
	EKSIntegrationEnabled                                 *bool                            `url:"eks_integration_enabled,omitempty" json:"eks_integration_enabled,omitempty"`
	EKSSecretAccessKey                                    *string                          `url:"eks_secret_access_key,omitempty" json:"eks_secret_access_key,omitempty"`
	Ed25519KeyRestriction                                 *int                             `url:"ed25519_key_restriction,omitempty" json:"ed25519_key_restriction,omitempty"`
	Ed25519SKKeyRestriction                               *int                             `url:"ed25519_sk_key_restriction,omitempty" json:"ed25519_sk_key_restriction,omitempty"`
	ElasticsearchAWS                                      *bool                            `url:"elasticsearch_aws,omitempty" json:"elasticsearch_aws,omitempty"`
	ElasticsearchAWSAccessKey                             *string                          `url:"elasticsearch_aws_access_key,omitempty" json:"elasticsearch_aws_access_key,omitempty"`
	ElasticsearchAWSRegion                                *string                          `url:"elasticsearch_aws_region,omitempty" json:"elasticsearch_aws_region,omitempty"`
	ElasticsearchAWSSecretAccessKey                       *string                          `url:"elasticsearch_aws_secret_access_key,omitempty" json:"elasticsearch_aws_secret_access_key,omitempty"`
	ElasticsearchAnalyzersKuromojiEnabled                 *bool                            `url:"elasticsearch_analyzers_kuromoji_enabled,omitempty" json:"elasticsearch_analyzers_kuromoji_enabled,omitempty"`
	ElasticsearchAnalyzersKuromojiSearch                  *int                             `url:"elasticsearch_analyzers_kuromoji_search,omitempty" json:"elasticsearch_analyzers_kuromoji_search,omitempty"`
	ElasticsearchAnalyzersSmartCNEnabled                  *bool                            `url:"elasticsearch_analyzers_smartcn_enabled,omitempty" json:"elasticsearch_analyzers_smartcn_enabled,omitempty"`
	ElasticsearchAnalyzersSmartCNSearch                   *int                             `url:"elasticsearch_analyzers_smartcn_search,omitempty" json:"elasticsearch_analyzers_smartcn_search,omitempty"`
	ElasticsearchClientRequestTimeout                     *int                             `url:"elasticsearch_client_request_timeout,omitempty" json:"elasticsearch_client_request_timeout,omitempty"`
	ElasticsearchIndexedFieldLengthLimit                  *int                             `url:"elasticsearch_indexed_field_length_limit,omitempty" json:"elasticsearch_indexed_field_length_limit,omitempty"`
	ElasticsearchIndexedFileSizeLimitKB                   *int                             `url:"elasticsearch_indexed_file_size_limit_kb,omitempty" json:"elasticsearch_indexed_file_size_limit_kb,omitempty"`
	ElasticsearchIndexing                                 *bool                            `url:"elasticsearch_indexing,omitempty" json:"elasticsearch_indexing,omitempty"`
	ElasticsearchLimitIndexing                            *bool                            `url:"elasticsearch_limit_indexing,omitempty" json:"elasticsearch_limit_indexing,omitempty"`
	ElasticsearchMaxBulkConcurrency                       *int                             `url:"elasticsearch_max_bulk_concurrency,omitempty" json:"elasticsearch_max_bulk_concurrency,omitempty"`
	ElasticsearchMaxBulkSizeMB                            *int                             `url:"elasticsearch_max_bulk_size_mb,omitempty" json:"elasticsearch_max_bulk_size_mb,omitempty"`
	ElasticsearchMaxCodeIndexingConcurrency               *int                             `url:"elasticsearch_max_code_indexing_concurrency,omitempty" json:"elasticsearch_max_code_indexing_concurrency,omitempty"`
	ElasticsearchNamespaceIDs                             *[]int                           `url:"elasticsearch_namespace_ids,omitempty" json:"elasticsearch_namespace_ids,omitempty"`
	ElasticsearchPassword                                 *string                          `url:"elasticsearch_password,omitempty" json:"elasticsearch_password,omitempty"`
	ElasticsearchPauseIndexing                            *bool                            `url:"elasticsearch_pause_indexing,omitempty" json:"elasticsearch_pause_indexing,omitempty"`
	ElasticsearchProjectIDs                               *[]int                           `url:"elasticsearch_project_ids,omitempty" json:"elasticsearch_project_ids,omitempty"`
	ElasticsearchReplicas                                 *int                             `url:"elasticsearch_replicas,omitempty" json:"elasticsearch_replicas,omitempty"`
	ElasticsearchRequeueWorkers                           *bool                            `url:"elasticsearch_requeue_workers,omitempty" json:"elasticsearch_requeue_workers,omitempty"`
	ElasticsearchSearch                                   *bool                            `url:"elasticsearch_search,omitempty" json:"elasticsearch_search,omitempty"`
	ElasticsearchShards                                   *int                             `url:"elasticsearch_shards,omitempty" json:"elasticsearch_shards,omitempty"`
	ElasticsearchURL                                      *string                          `url:"elasticsearch_url,omitempty" json:"elasticsearch_url,omitempty"`
	ElasticsearchUsername                                 *string                          `url:"elasticsearch_username,omitempty" json:"elasticsearch_username,omitempty"`
	ElasticsearchWorkerNumberOfShards                     *int                             `url:"elasticsearch_worker_number_of_shards,omitempty" json:"elasticsearch_worker_number_of_shards,omitempty"`
	EmailAdditionalText                                   *string                          `url:"email_additional_text,omitempty" json:"email_additional_text,omitempty"`
	EmailAuthorInBody                                     *bool                            `url:"email_author_in_body,omitempty" json:"email_author_in_body,omitempty"`
	EmailConfirmationSetting                              *string                          `url:"email_confirmation_setting,omitempty" json:"email_confirmation_setting,omitempty"`
	EmailRestrictions                                     *string                          `url:"email_restrictions,omitempty" json:"email_restrictions,omitempty"`
	EmailRestrictionsEnabled                              *bool                            `url:"email_restrictions_enabled,omitempty" json:"email_restrictions_enabled,omitempty"`
	EnableArtifactExternalRedirectWarningPage             *bool                            `url:"enable_artifact_external_redirect_warning_page,omitempty" json:"enable_artifact_external_redirect_warning_page,omitempty"`
	EnabledGitAccessProtocol                              *string                          `url:"enabled_git_access_protocol,omitempty" json:"enabled_git_access_protocol,omitempty"`
	EnforceNamespaceStorageLimit                          *bool                            `url:"enforce_namespace_storage_limit,omitempty" json:"enforce_namespace_storage_limit,omitempty"`
	EnforcePATExpiration                                  *bool                            `url:"enforce_pat_expiration,omitempty" json:"enforce_pat_expiration,omitempty"`
	EnforceSSHKeyExpiration                               *bool                            `url:"enforce_ssh_key_expiration,omitempty" json:"enforce_ssh_key_expiration,omitempty"`
	EnforceTerms                                          *bool                            `url:"enforce_terms,omitempty" json:"enforce_terms,omitempty"`
	ExternalAuthClientCert                                *string                          `url:"external_auth_client_cert,omitempty" json:"external_auth_client_cert,omitempty"`
	ExternalAuthClientKey                                 *string                          `url:"external_auth_client_key,omitempty" json:"external_auth_client_key,omitempty"`
	ExternalAuthClientKeyPass                             *string                          `url:"external_auth_client_key_pass,omitempty" json:"external_auth_client_key_pass,omitempty"`
	ExternalAuthorizationServiceDefaultLabel              *string                          `url:"external_authorization_service_default_label,omitempty" json:"external_authorization_service_default_label,omitempty"`
	ExternalAuthorizationServiceEnabled                   *bool                            `url:"external_authorization_service_enabled,omitempty" json:"external_authorization_service_enabled,omitempty"`
	ExternalAuthorizationServiceTimeout                   *float64                         `url:"external_authorization_service_timeout,omitempty" json:"external_authorization_service_timeout,omitempty"`
	ExternalAuthorizationServiceURL                       *string                          `url:"external_authorization_service_url,omitempty" json:"external_authorization_service_url,omitempty"`
	ExternalPipelineValidationServiceTimeout              *int                             `url:"external_pipeline_validation_service_timeout,omitempty" json:"external_pipeline_validation_service_timeout,omitempty"`
	ExternalPipelineValidationServiceToken                *string                          `url:"external_pipeline_validation_service_token,omitempty" json:"external_pipeline_validation_service_token,omitempty"`
	ExternalPipelineValidationServiceURL                  *string                          `url:"external_pipeline_validation_service_url,omitempty" json:"external_pipeline_validation_service_url,omitempty"`
	FailedLoginAttemptsUnlockPeriodInMinutes              *int                             `url:"failed_login_attempts_unlock_period_in_minutes,omitempty" json:"failed_login_attempts_unlock_period_in_minutes,omitempty"`
	FileTemplateProjectID                                 *int                             `url:"file_template_project_id,omitempty" json:"file_template_project_id,omitempty"`
	FirstDayOfWeek                                        *int                             `url:"first_day_of_week,omitempty" json:"first_day_of_week,omitempty"`
	FlocEnabled                                           *bool                            `url:"floc_enabled,omitempty" json:"floc_enabled,omitempty"`
	GeoNodeAllowedIPs                                     *string                          `url:"geo_node_allowed_ips,omitempty" json:"geo_node_allowed_ips,omitempty"`
	GeoStatusTimeout                                      *int                             `url:"geo_status_timeout,omitempty" json:"geo_status_timeout,omitempty"`
	GitRateLimitUsersAlertlist                            *[]string                        `url:"git_rate_limit_users_alertlist,omitempty" json:"git_rate_limit_users_alertlist,omitempty"`
	GitTwoFactorSessionExpiry                             *int                             `url:"git_two_factor_session_expiry,omitempty" json:"git_two_factor_session_expiry,omitempty"`
	GitalyTimeoutDefault                                  *int                             `url:"gitaly_timeout_default,omitempty" json:"gitaly_timeout_default,omitempty"`
	GitalyTimeoutFast                                     *int                             `url:"gitaly_timeout_fast,omitempty" json:"gitaly_timeout_fast,omitempty"`
	GitalyTimeoutMedium                                   *int                             `url:"gitaly_timeout_medium,omitempty" json:"gitaly_timeout_medium,omitempty"`
	GitlabDedicatedInstance                               *bool                            `url:"gitlab_dedicated_instance,omitempty" json:"gitlab_dedicated_instance,omitempty"`
	GitlabEnvironmentToolkitInstance                      *bool                            `url:"gitlab_environment_toolkit_instance,omitempty" json:"gitlab_environment_toolkit_instance,omitempty"`
	GitlabShellOperationLimit                             *int                             `url:"gitlab_shell_operation_limit,omitempty" json:"gitlab_shell_operation_limit,omitempty"`
	GitpodEnabled                                         *bool                            `url:"gitpod_enabled,omitempty" json:"gitpod_enabled,omitempty"`
	GitpodURL                                             *string                          `url:"gitpod_url,omitempty" json:"gitpod_url,omitempty"`
	GitRateLimitUsersAllowlist                            *[]string                        `url:"git_rate_limit_users_allowlist,omitempty" json:"git_rate_limit_users_allowlist,omitempty"`
	GloballyAllowedIPs                                    *string                          `url:"globally_allowed_ips,omitempty" json:"globally_allowed_ips,omitempty"`
	GrafanaEnabled                                        *bool                            `url:"grafana_enabled,omitempty" json:"grafana_enabled,omitempty"`
	GrafanaURL                                            *string                          `url:"grafana_url,omitempty" json:"grafana_url,omitempty"`
	GravatarEnabled                                       *bool                            `url:"gravatar_enabled,omitempty" json:"gravatar_enabled,omitempty"`
	GroupDownloadExportLimit                              *int                             `url:"group_download_export_limit,omitempty" json:"group_download_export_limit,omitempty"`
	GroupExportLimit                                      *int                             `url:"group_export_limit,omitempty" json:"group_export_limit,omitempty"`
	GroupImportLimit                                      *int                             `url:"group_import_limit,omitempty" json:"group_import_limit,omitempty"`
	GroupOwnersCanManageDefaultBranchProtection           *bool                            `url:"group_owners_can_manage_default_branch_protection,omitempty" json:"group_owners_can_manage_default_branch_protection,omitempty"`
	GroupRunnerTokenExpirationInterval                    *int                             `url:"group_runner_token_expiration_interval,omitempty" json:"group_runner_token_expiration_interval,omitempty"`
	HTMLEmailsEnabled                                     *bool                            `url:"html_emails_enabled,omitempty" json:"html_emails_enabled,omitempty"`
	HashedStorageEnabled                                  *bool                            `url:"hashed_storage_enabled,omitempty" json:"hashed_storage_enabled,omitempty"`
	HelpPageDocumentationBaseURL                          *string                          `url:"help_page_documentation_base_url,omitempty" json:"help_page_documentation_base_url,omitempty"`
	HelpPageHideCommercialContent                         *bool                            `url:"help_page_hide_commercial_content,omitempty" json:"help_page_hide_commercial_content,omitempty"`
	HelpPageSupportURL                                    *string                          `url:"help_page_support_url,omitempty" json:"help_page_support_url,omitempty"`
	HelpPageText                                          *string                          `url:"help_page_text,omitempty" json:"help_page_text,omitempty"`
	HelpText                                              *string                          `url:"help_text,omitempty" json:"help_text,omitempty"`
	HideThirdPartyOffers                                  *bool                            `url:"hide_third_party_offers,omitempty" json:"hide_third_party_offers,omitempty"`
	HomePageURL                                           *string                          `url:"home_page_url,omitempty" json:"home_page_url,omitempty"`
	HousekeepingBitmapsEnabled                            *bool                            `url:"housekeeping_bitmaps_enabled,omitempty" json:"housekeeping_bitmaps_enabled,omitempty"`
	HousekeepingEnabled                                   *bool                            `url:"housekeeping_enabled,omitempty" json:"housekeeping_enabled,omitempty"`
	HousekeepingFullRepackPeriod                          *int                             `url:"housekeeping_full_repack_period,omitempty" json:"housekeeping_full_repack_period,omitempty"`
	HousekeepingGcPeriod                                  *int                             `url:"housekeeping_gc_period,omitempty" json:"housekeeping_gc_period,omitempty"`
	HousekeepingIncrementalRepackPeriod                   *int                             `url:"housekeeping_incremental_repack_period,omitempty" json:"housekeeping_incremental_repack_period,omitempty"`
	HousekeepingOptimizeRepositoryPeriod                  *int                             `url:"housekeeping_optimize_repository_period,omitempty" json:"housekeeping_optimize_repository_period,omitempty"`
	ImportSources                                         *[]string                        `url:"import_sources,omitempty" json:"import_sources,omitempty"`
	InactiveProjectsDeleteAfterMonths                     *int                             `url:"inactive_projects_delete_after_months,omitempty" json:"inactive_projects_delete_after_months,omitempty"`
	InactiveProjectsMinSizeMB                             *int                             `url:"inactive_projects_min_size_mb,omitempty" json:"inactive_projects_min_size_mb,omitempty"`
	InactiveProjectsSendWarningEmailAfterMonths           *int                             `url:"inactive_projects_send_warning_email_after_months,omitempty" json:"inactive_projects_send_warning_email_after_months,omitempty"`
	IncludeOptionalMetricsInServicePing                   *bool                            `url:"include_optional_metrics_in_service_ping,omitempty" json:"include_optional_metrics_in_service_ping,omitempty"`
	InProductMarketingEmailsEnabled                       *bool                            `url:"in_product_marketing_emails_enabled,omitempty" json:"in_product_marketing_emails_enabled,omitempty"`
	InvisibleCaptchaEnabled                               *bool                            `url:"invisible_captcha_enabled,omitempty" json:"invisible_captcha_enabled,omitempty"`
	IssuesCreateLimit                                     *int                             `url:"issues_create_limit,omitempty" json:"issues_create_limit,omitempty"`
	JiraConnectApplicationKey                             *string                          `url:"jira_connect_application_key,omitempty" json:"jira_connect_application_key,omitempty"`
	JiraConnectPublicKeyStorageEnabled                    *bool                            `url:"jira_connect_public_key_storage_enabled,omitempty" json:"jira_connect_public_key_storage_enabled,omitempty"`
	JiraConnectProxyURL                                   *string                          `url:"jira_connect_proxy_url,omitempty" json:"jira_connect_proxy_url,omitempty"`
	KeepLatestArtifact                                    *bool                            `url:"keep_latest_artifact,omitempty" json:"keep_latest_artifact,omitempty"`
	KrokiEnabled                                          *bool                            `url:"kroki_enabled,omitempty" json:"kroki_enabled,omitempty"`
	KrokiFormats                                          *map[string]bool                 `url:"kroki_formats,omitempty" json:"kroki_formats,omitempty"`
	KrokiURL                                              *string                          `url:"kroki_url,omitempty" json:"kroki_url,omitempty"`
	LocalMarkdownVersion                                  *int                             `url:"local_markdown_version,omitempty" json:"local_markdown_version,omitempty"`
	LockDuoFeaturesEnabled                                *bool                            `url:"lock_duo_features_enabled,omitempty" json:"lock_duo_features_enabled,omitempty"`
	LockMembershipsToLDAP                                 *bool                            `url:"lock_memberships_to_ldap,omitempty" json:"lock_memberships_to_ldap,omitempty"`
	LoginRecaptchaProtectionEnabled                       *bool                            `url:"login_recaptcha_protection_enabled,omitempty" json:"login_recaptcha_protection_enabled,omitempty"`
	MailgunEventsEnabled                                  *bool                            `url:"mailgun_events_enabled,omitempty" json:"mailgun_events_enabled,omitempty"`
	MailgunSigningKey                                     *string                          `url:"mailgun_signing_key,omitempty" json:"mailgun_signing_key,omitempty"`
	MaintenanceMode                                       *bool                            `url:"maintenance_mode,omitempty" json:"maintenance_mode,omitempty"`
	MaintenanceModeMessage                                *string                          `url:"maintenance_mode_message,omitempty" json:"maintenance_mode_message,omitempty"`
	MavenPackageRequestsForwarding                        *bool                            `url:"maven_package_requests_forwarding,omitempty" json:"maven_package_requests_forwarding,omitempty"`
	MaxArtifactsSize                                      *int                             `url:"max_artifacts_size,omitempty" json:"max_artifacts_size,omitempty"`
	MaxAttachmentSize                                     *int                             `url:"max_attachment_size,omitempty" json:"max_attachment_size,omitempty"`
	MaxDecompressedArchiveSize                            *int                             `url:"max_decompressed_archive_size,omitempty" json:"max_decompressed_archive_size,omitempty"`
	MaxExportSize                                         *int                             `url:"max_export_size,omitempty" json:"max_export_size,omitempty"`
	MaxImportRemoteFileSize                               *int                             `url:"max_import_remote_file_size,omitempty" json:"max_import_remote_file_size,omitempty"`
	MaxImportSize                                         *int                             `url:"max_import_size,omitempty" json:"max_import_size,omitempty"`
	MaxLoginAttempts                                      *int                             `url:"max_login_attempts,omitempty" json:"max_login_attempts,omitempty"`
	MaxNumberOfRepositoryDownloads                        *int                             `url:"max_number_of_repository_downloads,omitempty" json:"max_number_of_repository_downloads,omitempty"`
	MaxNumberOfRepositoryDownloadsWithinTimePeriod        *int                             `url:"max_number_of_repository_downloads_within_time_period,omitempty" json:"max_number_of_repository_downloads_within_time_period,omitempty"`
	MaxPagesSize                                          *int                             `url:"max_pages_size,omitempty" json:"max_pages_size,omitempty"`
	MaxPersonalAccessTokenLifetime                        *int                             `url:"max_personal_access_token_lifetime,omitempty" json:"max_personal_access_token_lifetime,omitempty"`
	MaxSSHKeyLifetime                                     *int                             `url:"max_ssh_key_lifetime,omitempty" json:"max_ssh_key_lifetime,omitempty"`
	MaxTerraformStateSizeBytes                            *int                             `url:"max_terraform_state_size_bytes,omitempty" json:"max_terraform_state_size_bytes,omitempty"`
	MaxYAMLDepth                                          *int                             `url:"max_yaml_depth,omitempty" json:"max_yaml_depth,omitempty"`
	MaxYAMLSizeBytes                                      *int                             `url:"max_yaml_size_bytes,omitempty" json:"max_yaml_size_bytes,omitempty"`
	MetricsMethodCallThreshold                            *int                             `url:"metrics_method_call_threshold,omitempty" json:"metrics_method_call_threshold,omitempty"`
	MinimumPasswordLength                                 *int                             `url:"minimum_password_length,omitempty" json:"minimum_password_length,omitempty"`
	MirrorAvailable                                       *bool                            `url:"mirror_available,omitempty" json:"mirror_available,omitempty"`
	MirrorCapacityThreshold                               *int                             `url:"mirror_capacity_threshold,omitempty" json:"mirror_capacity_threshold,omitempty"`
	MirrorMaxCapacity                                     *int                             `url:"mirror_max_capacity,omitempty" json:"mirror_max_capacity,omitempty"`
	MirrorMaxDelay                                        *int                             `url:"mirror_max_delay,omitempty" json:"mirror_max_delay,omitempty"`
	NPMPackageRequestsForwarding                          *bool                            `url:"npm_package_requests_forwarding,omitempty" json:"npm_package_requests_forwarding,omitempty"`
	NotesCreateLimit                                      *int                             `url:"notes_create_limit,omitempty" json:"notes_create_limit,omitempty"`
	NotifyOnUnknownSignIn                                 *bool                            `url:"notify_on_unknown_sign_in,omitempty" json:"notify_on_unknown_sign_in,omitempty"`
	NugetSkipMetadataURLValidation                        *bool                            `url:"nuget_skip_metadata_url_validation,omitempty" json:"nuget_skip_metadata_url_validation,omitempty"`
	OutboundLocalRequestsAllowlistRaw                     *string                          `url:"outbound_local_requests_allowlist_raw,omitempty" json:"outbound_local_requests_allowlist_raw,omitempty"`
	OutboundLocalRequestsWhitelist                        *[]string                        `url:"outbound_local_requests_whitelist,omitempty" json:"outbound_local_requests_whitelist,omitempty"`
	PackageMetadataPURLTypes                              *[]int                           `url:"package_metadata_purl_types,omitempty" json:"package_metadata_purl_types,omitempty"`
	PackageRegistryAllowAnyoneToPullOption                *bool                            `url:"package_registry_allow_anyone_to_pull_option,omitempty" json:"package_registry_allow_anyone_to_pull_option,omitempty"`
	PackageRegistryCleanupPoliciesWorkerCapacity          *int                             `url:"package_registry_cleanup_policies_worker_capacity,omitempty" json:"package_registry_cleanup_policies_worker_capacity,omitempty"`
	PagesDomainVerificationEnabled                        *bool                            `url:"pages_domain_verification_enabled,omitempty" json:"pages_domain_verification_enabled,omitempty"`
	PasswordAuthenticationEnabledForGit                   *bool                            `url:"password_authentication_enabled_for_git,omitempty" json:"password_authentication_enabled_for_git,omitempty"`
	PasswordAuthenticationEnabledForWeb                   *bool                            `url:"password_authentication_enabled_for_web,omitempty" json:"password_authentication_enabled_for_web,omitempty"`
	PasswordNumberRequired                                *bool                            `url:"password_number_required,omitempty" json:"password_number_required,omitempty"`
	PasswordSymbolRequired                                *bool                            `url:"password_symbol_required,omitempty" json:"password_symbol_required,omitempty"`
	PasswordUppercaseRequired                             *bool                            `url:"password_uppercase_required,omitempty" json:"password_uppercase_required,omitempty"`
	PasswordLowercaseRequired                             *bool                            `url:"password_lowercase_required,omitempty" json:"password_lowercase_required,omitempty"`
	PerformanceBarAllowedGroupID                          *int                             `url:"performance_bar_allowed_group_id,omitempty" json:"performance_bar_allowed_group_id,omitempty"`
	PerformanceBarAllowedGroupPath                        *string                          `url:"performance_bar_allowed_group_path,omitempty" json:"performance_bar_allowed_group_path,omitempty"`
	PerformanceBarEnabled                                 *bool                            `url:"performance_bar_enabled,omitempty" json:"performance_bar_enabled,omitempty"`
	PersonalAccessTokenPrefix                             *string                          `url:"personal_access_token_prefix,omitempty" json:"personal_access_token_prefix,omitempty"`
	PlantumlEnabled                                       *bool                            `url:"plantuml_enabled,omitempty" json:"plantuml_enabled,omitempty"`
	PlantumlURL                                           *string                          `url:"plantuml_url,omitempty" json:"plantuml_url,omitempty"`
	PipelineLimitPerProjectUserSha                        *int                             `url:"pipeline_limit_per_project_user_sha,omitempty" json:"pipeline_limit_per_project_user_sha,omitempty"`
	PollingIntervalMultiplier                             *float64                         `url:"polling_interval_multiplier,omitempty" json:"polling_interval_multiplier,omitempty"`
	PreventMergeRequestsAuthorApproval                    *bool                            `url:"prevent_merge_requests_author_approval,omitempty" json:"prevent_merge_requests_author_approval,omitempty"`
	PreventMergeRequestsCommittersApproval                *bool                            `url:"prevent_merge_requests_committers_approval,omitempty" json:"prevent_merge_requests_committers_approval,omitempty"`
	ProjectDownloadExportLimit                            *int                             `url:"project_download_export_limit,omitempty" json:"project_download_export_limit,omitempty"`
	ProjectExportEnabled                                  *bool                            `url:"project_export_enabled,omitempty" json:"project_export_enabled,omitempty"`
	ProjectExportLimit                                    *int                             `url:"project_export_limit,omitempty" json:"project_export_limit,omitempty"`
	ProjectImportLimit                                    *int                             `url:"project_import_limit,omitempty" json:"project_import_limit,omitempty"`
	ProjectJobsAPIRateLimit                               *int                             `url:"project_jobs_api_rate_limit,omitempty" json:"project_jobs_api_rate_limit,omitempty"`
	ProjectRunnerTokenExpirationInterval                  *int                             `url:"project_runner_token_expiration_interval,omitempty" json:"project_runner_token_expiration_interval,omitempty"`
	ProjectsAPIRateLimitUnauthenticated                   *int                             `url:"projects_api_rate_limit_unauthenticated,omitempty" json:"projects_api_rate_limit_unauthenticated,omitempty"`
	PrometheusMetricsEnabled                              *bool                            `url:"prometheus_metrics_enabled,omitempty" json:"prometheus_metrics_enabled,omitempty"`
	ProtectedCIVariables                                  *bool                            `url:"protected_ci_variables,omitempty" json:"protected_ci_variables,omitempty"`
	PseudonymizerEnabled                                  *bool                            `url:"pseudonymizer_enabled,omitempty" json:"pseudonymizer_enabled,omitempty"`
	PushEventActivitiesLimit                              *int                             `url:"push_event_activities_limit,omitempty" json:"push_event_activities_limit,omitempty"`
	PushEventHooksLimit                                   *int                             `url:"push_event_hooks_limit,omitempty" json:"push_event_hooks_limit,omitempty"`
	PyPIPackageRequestsForwarding                         *bool                            `url:"pypi_package_requests_forwarding,omitempty" json:"pypi_package_requests_forwarding,omitempty"`
	RSAKeyRestriction                                     *int                             `url:"rsa_key_restriction,omitempty" json:"rsa_key_restriction,omitempty"`
	RateLimitingResponseText                              *string                          `url:"rate_limiting_response_text,omitempty" json:"rate_limiting_response_text,omitempty"`
	RawBlobRequestLimit                                   *int                             `url:"raw_blob_request_limit,omitempty" json:"raw_blob_request_limit,omitempty"`
	RecaptchaEnabled                                      *bool                            `url:"recaptcha_enabled,omitempty" json:"recaptcha_enabled,omitempty"`
	RecaptchaPrivateKey                                   *string                          `url:"recaptcha_private_key,omitempty" json:"recaptcha_private_key,omitempty"`
	RecaptchaSiteKey                                      *string                          `url:"recaptcha_site_key,omitempty" json:"recaptcha_site_key,omitempty"`
	ReceiveMaxInputSize                                   *int                             `url:"receive_max_input_size,omitempty" json:"receive_max_input_size,omitempty"`
	ReceptiveClusterAgentsEnabled                         *bool                            `url:"receptive_cluster_agents_enabled,omitempty" json:"receptive_cluster_agents_enabled,omitempty"`
	RememberMeEnabled                                     *bool                            `url:"remember_me_enabled,omitempty" json:"remember_me_enabled,omitempty"`
	RepositoryChecksEnabled                               *bool                            `url:"repository_checks_enabled,omitempty" json:"repository_checks_enabled,omitempty"`
	RepositorySizeLimit                                   *int                             `url:"repository_size_limit,omitempty" json:"repository_size_limit,omitempty"`
	RepositoryStorages                                    *[]string                        `url:"repository_storages,omitempty" json:"repository_storages,omitempty"`
	RepositoryStoragesWeighted                            *map[string]int                  `url:"repository_storages_weighted,omitempty" json:"repository_storages_weighted,omitempty"`
	RequireAdminApprovalAfterUserSignup                   *bool                            `url:"require_admin_approval_after_user_signup,omitempty" json:"require_admin_approval_after_user_signup,omitempty"`
	RequireAdminTwoFactorAuthentication                   *bool                            `url:"require_admin_two_factor_authentication,omitempty" json:"require_admin_two_factor_authentication,omitempty"`
	RequirePersonalAccessTokenExpiry                      *bool                            `url:"require_personal_access_token_expiry,omitempty" json:"require_personal_access_token_expiry,omitempty"`
	RequireTwoFactorAuthentication                        *bool                            `url:"require_two_factor_authentication,omitempty" json:"require_two_factor_authentication,omitempty"`
	RestrictedVisibilityLevels                            *[]VisibilityValue               `url:"restricted_visibility_levels,omitempty" json:"restricted_visibility_levels,omitempty"`
	RunnerTokenExpirationInterval                         *int                             `url:"runner_token_expiration_interval,omitempty" json:"runner_token_expiration_interval,omitempty"`
	SearchRateLimit                                       *int                             `url:"search_rate_limit,omitempty" json:"search_rate_limit,omitempty"`
	SearchRateLimitUnauthenticated                        *int                             `url:"search_rate_limit_unauthenticated,omitempty" json:"search_rate_limit_unauthenticated,omitempty"`
	SecretDetectionRevocationTokenTypesURL                *string                          `url:"secret_detection_revocation_token_types_url,omitempty" json:"secret_detection_revocation_token_types_url,omitempty"`
	SecretDetectionTokenRevocationEnabled                 *bool                            `url:"secret_detection_token_revocation_enabled,omitempty" json:"secret_detection_token_revocation_enabled,omitempty"`
	SecretDetectionTokenRevocationToken                   *string                          `url:"secret_detection_token_revocation_token,omitempty" json:"secret_detection_token_revocation_token,omitempty"`
	SecretDetectionTokenRevocationURL                     *string                          `url:"secret_detection_token_revocation_url,omitempty" json:"secret_detection_token_revocation_url,omitempty"`
	SecurityApprovalPoliciesLimit                         *int                             `url:"security_approval_policies_limit,omitempty" json:"security_approval_policies_limit,omitempty"`
	SecurityPolicyGlobalGroupApproversEnabled             *bool                            `url:"security_policy_global_group_approvers_enabled,omitempty" json:"security_policy_global_group_approvers_enabled,omitempty"`
	SecurityTXTContent                                    *string                          `url:"security_txt_content,omitempty" json:"security_txt_content,omitempty"`
	SendUserConfirmationEmail                             *bool                            `url:"send_user_confirmation_email,omitempty" json:"send_user_confirmation_email,omitempty"`
	SentryClientsideDSN                                   *string                          `url:"sentry_clientside_dsn,omitempty" json:"sentry_clientside_dsn,omitempty"`
	SentryDSN                                             *string                          `url:"sentry_dsn,omitempty" json:"sentry_dsn,omitempty"`
	SentryEnabled                                         *string                          `url:"sentry_enabled,omitempty" json:"sentry_enabled,omitempty"`
	SentryEnvironment                                     *string                          `url:"sentry_environment,omitempty" json:"sentry_environment,omitempty"`
	ServiceAccessTokensExpirationEnforced                 *bool                            `url:"service_access_tokens_expiration_enforced,omitempty" json:"service_access_tokens_expiration_enforced,omitempty"`
	SessionExpireDelay                                    *int                             `url:"session_expire_delay,omitempty" json:"session_expire_delay,omitempty"`
	SharedRunnersEnabled                                  *bool                            `url:"shared_runners_enabled,omitempty" json:"shared_runners_enabled,omitempty"`
	SharedRunnersMinutes                                  *int                             `url:"shared_runners_minutes,omitempty" json:"shared_runners_minutes,omitempty"`
	SharedRunnersText                                     *string                          `url:"shared_runners_text,omitempty" json:"shared_runners_text,omitempty"`
	SidekiqJobLimiterCompressionThresholdBytes            *int                             `url:"sidekiq_job_limiter_compression_threshold_bytes,omitempty" json:"sidekiq_job_limiter_compression_threshold_bytes,omitempty"`
	SidekiqJobLimiterLimitBytes                           *int                             `url:"sidekiq_job_limiter_limit_bytes,omitempty" json:"sidekiq_job_limiter_limit_bytes,omitempty"`
	SidekiqJobLimiterMode                                 *string                          `url:"sidekiq_job_limiter_mode,omitempty" json:"sidekiq_job_limiter_mode,omitempty"`
	SignInText                                            *string                          `url:"sign_in_text,omitempty" json:"sign_in_text,omitempty"`
	SignupEnabled                                         *bool                            `url:"signup_enabled,omitempty" json:"signup_enabled,omitempty"`
	SilentAdminExportsEnabled                             *bool                            `url:"silent_admin_exports_enabled,omitempty" json:"silent_admin_exports_enabled,omitempty"`
	SilentModeEnabled                                     *bool                            `url:"silent_mode_enabled,omitempty" json:"silent_mode_enabled,omitempty"`
	SlackAppEnabled                                       *bool                            `url:"slack_app_enabled,omitempty" json:"slack_app_enabled,omitempty"`
	SlackAppID                                            *string                          `url:"slack_app_id,omitempty" json:"slack_app_id,omitempty"`
	SlackAppSecret                                        *string                          `url:"slack_app_secret,omitempty" json:"slack_app_secret,omitempty"`
	SlackAppSigningSecret                                 *string                          `url:"slack_app_signing_secret,omitempty" json:"slack_app_signing_secret,omitempty"`
	SlackAppVerificationToken                             *string                          `url:"slack_app_verification_token,omitempty" json:"slack_app_verification_token,omitempty"`
	SnippetSizeLimit                                      *int                             `url:"snippet_size_limit,omitempty" json:"snippet_size_limit,omitempty"`
	SnowplowAppID                                         *string                          `url:"snowplow_app_id,omitempty" json:"snowplow_app_id,omitempty"`
	SnowplowCollectorHostname                             *string                          `url:"snowplow_collector_hostname,omitempty" json:"snowplow_collector_hostname,omitempty"`
	SnowplowCookieDomain                                  *string                          `url:"snowplow_cookie_domain,omitempty" json:"snowplow_cookie_domain,omitempty"`
	SnowplowDatabaseCollectorHostname                     *string                          `url:"snowplow_database_collector_hostname,omitempty" json:"snowplow_database_collector_hostname,omitempty"`
	SnowplowEnabled                                       *bool                            `url:"snowplow_enabled,omitempty" json:"snowplow_enabled,omitempty"`
	SourcegraphEnabled                                    *bool                            `url:"sourcegraph_enabled,omitempty" json:"sourcegraph_enabled,omitempty"`
	SourcegraphPublicOnly                                 *bool                            `url:"sourcegraph_public_only,omitempty" json:"sourcegraph_public_only,omitempty"`
	SourcegraphURL                                        *string                          `url:"sourcegraph_url,omitempty" json:"sourcegraph_url,omitempty"`
	SpamCheckAPIKey                                       *string                          `url:"spam_check_api_key,omitempty" json:"spam_check_api_key,omitempty"`
	SpamCheckEndpointEnabled                              *bool                            `url:"spam_check_endpoint_enabled,omitempty" json:"spam_check_endpoint_enabled,omitempty"`
	SpamCheckEndpointURL                                  *string                          `url:"spam_check_endpoint_url,omitempty" json:"spam_check_endpoint_url,omitempty"`
	StaticObjectsExternalStorageAuthToken                 *string                          `url:"static_objects_external_storage_auth_token,omitempty" json:"static_objects_external_storage_auth_token,omitempty"`
	StaticObjectsExternalStorageURL                       *string                          `url:"static_objects_external_storage_url,omitempty" json:"static_objects_external_storage_url,omitempty"`
	SuggestPipelineEnabled                                *bool                            `url:"suggest_pipeline_enabled,omitempty" json:"suggest_pipeline_enabled,omitempty"`
	TerminalMaxSessionTime                                *int                             `url:"terminal_max_session_time,omitempty" json:"terminal_max_session_time,omitempty"`
	Terms                                                 *string                          `url:"terms,omitempty" json:"terms,omitempty"`
	ThrottleAuthenticatedAPIEnabled                       *bool                            `url:"throttle_authenticated_api_enabled,omitempty" json:"throttle_authenticated_api_enabled,omitempty"`
	ThrottleAuthenticatedAPIPeriodInSeconds               *int                             `url:"throttle_authenticated_api_period_in_seconds,omitempty" json:"throttle_authenticated_api_period_in_seconds,omitempty"`
	ThrottleAuthenticatedAPIRequestsPerPeriod             *int                             `url:"throttle_authenticated_api_requests_per_period,omitempty" json:"throttle_authenticated_api_requests_per_period,omitempty"`
	ThrottleAuthenticatedDeprecatedAPIEnabled             *bool                            `url:"throttle_authenticated_deprecated_api_enabled,omitempty" json:"throttle_authenticated_deprecated_api_enabled,omitempty"`
	ThrottleAuthenticatedDeprecatedAPIPeriodInSeconds     *int                             `url:"throttle_authenticated_deprecated_api_period_in_seconds,omitempty" json:"throttle_authenticated_deprecated_api_period_in_seconds,omitempty"`
	ThrottleAuthenticatedDeprecatedAPIRequestsPerPeriod   *int                             `url:"throttle_authenticated_deprecated_api_requests_per_period,omitempty" json:"throttle_authenticated_deprecated_api_requests_per_period,omitempty"`
	ThrottleAuthenticatedFilesAPIEnabled                  *bool                            `url:"throttle_authenticated_files_api_enabled,omitempty" json:"throttle_authenticated_files_api_enabled,omitempty"`
	ThrottleAuthenticatedFilesAPIPeriodInSeconds          *int                             `url:"throttle_authenticated_files_api_period_in_seconds,omitempty" json:"throttle_authenticated_files_api_period_in_seconds,omitempty"`
	ThrottleAuthenticatedFilesAPIRequestsPerPeriod        *int                             `url:"throttle_authenticated_files_api_requests_per_period,omitempty" json:"throttle_authenticated_files_api_requests_per_period,omitempty"`
	ThrottleAuthenticatedGitLFSEnabled                    *bool                            `url:"throttle_authenticated_git_lfs_enabled,omitempty" json:"throttle_authenticated_git_lfs_enabled,omitempty"`
	ThrottleAuthenticatedGitLFSPeriodInSeconds            *int                             `url:"throttle_authenticated_git_lfs_period_in_seconds,omitempty" json:"throttle_authenticated_git_lfs_period_in_seconds,omitempty"`
	ThrottleAuthenticatedGitLFSRequestsPerPeriod          *int                             `url:"throttle_authenticated_git_lfs_requests_per_period,omitempty" json:"throttle_authenticated_git_lfs_requests_per_period,omitempty"`
	ThrottleAuthenticatedPackagesAPIEnabled               *bool                            `url:"throttle_authenticated_packages_api_enabled,omitempty" json:"throttle_authenticated_packages_api_enabled,omitempty"`
	ThrottleAuthenticatedPackagesAPIPeriodInSeconds       *int                             `url:"throttle_authenticated_packages_api_period_in_seconds,omitempty" json:"throttle_authenticated_packages_api_period_in_seconds,omitempty"`
	ThrottleAuthenticatedPackagesAPIRequestsPerPeriod     *int                             `url:"throttle_authenticated_packages_api_requests_per_period,omitempty" json:"throttle_authenticated_packages_api_requests_per_period,omitempty"`
	ThrottleAuthenticatedWebEnabled                       *bool                            `url:"throttle_authenticated_web_enabled,omitempty" json:"throttle_authenticated_web_enabled,omitempty"`
	ThrottleAuthenticatedWebPeriodInSeconds               *int                             `url:"throttle_authenticated_web_period_in_seconds,omitempty" json:"throttle_authenticated_web_period_in_seconds,omitempty"`
	ThrottleAuthenticatedWebRequestsPerPeriod             *int                             `url:"throttle_authenticated_web_requests_per_period,omitempty" json:"throttle_authenticated_web_requests_per_period,omitempty"`
	ThrottleIncidentManagementNotificationEnabled         *bool                            `url:"throttle_incident_management_notification_enabled,omitempty" json:"throttle_incident_management_notification_enabled,omitempty"`
	ThrottleIncidentManagementNotificationPerPeriod       *int                             `url:"throttle_incident_management_notification_per_period,omitempty" json:"throttle_incident_management_notification_per_period,omitempty"`
	ThrottleIncidentManagementNotificationPeriodInSeconds *int                             `url:"throttle_incident_management_notification_period_in_seconds,omitempty" json:"throttle_incident_management_notification_period_in_seconds,omitempty"`
	ThrottleProtectedPathsEnabled                         *bool                            `url:"throttle_protected_paths_enabled_enabled,omitempty" json:"throttle_protected_paths_enabled,omitempty"`
	ThrottleProtectedPathsPeriodInSeconds                 *int                             `url:"throttle_protected_paths_enabled_period_in_seconds,omitempty" json:"throttle_protected_paths_period_in_seconds,omitempty"`
	ThrottleProtectedPathsRequestsPerPeriod               *int                             `url:"throttle_protected_paths_enabled_requests_per_period,omitempty" json:"throttle_protected_paths_per_period,omitempty"`
	ThrottleUnauthenticatedAPIEnabled                     *bool                            `url:"throttle_unauthenticated_api_enabled,omitempty" json:"throttle_unauthenticated_api_enabled,omitempty"`
	ThrottleUnauthenticatedAPIPeriodInSeconds             *int                             `url:"throttle_unauthenticated_api_period_in_seconds,omitempty" json:"throttle_unauthenticated_api_period_in_seconds,omitempty"`
	ThrottleUnauthenticatedAPIRequestsPerPeriod           *int                             `url:"throttle_unauthenticated_api_requests_per_period,omitempty" json:"throttle_unauthenticated_api_requests_per_period,omitempty"`
	ThrottleUnauthenticatedDeprecatedAPIEnabled           *bool                            `url:"throttle_unauthenticated_deprecated_api_enabled,omitempty" json:"throttle_unauthenticated_deprecated_api_enabled,omitempty"`
	ThrottleUnauthenticatedDeprecatedAPIPeriodInSeconds   *int                             `url:"throttle_unauthenticated_deprecated_api_period_in_seconds,omitempty" json:"throttle_unauthenticated_deprecated_api_period_in_seconds,omitempty"`
	ThrottleUnauthenticatedDeprecatedAPIRequestsPerPeriod *int                             `url:"throttle_unauthenticated_deprecated_api_requests_per_period,omitempty" json:"throttle_unauthenticated_deprecated_api_requests_per_period,omitempty"`
	ThrottleUnauthenticatedEnabled                        *bool                            `url:"throttle_unauthenticated_enabled,omitempty" json:"throttle_unauthenticated_enabled,omitempty"`
	ThrottleUnauthenticatedFilesAPIEnabled                *bool                            `url:"throttle_unauthenticated_files_api_enabled,omitempty" json:"throttle_unauthenticated_files_api_enabled,omitempty"`
	ThrottleUnauthenticatedFilesAPIPeriodInSeconds        *int                             `url:"throttle_unauthenticated_files_api_period_in_seconds,omitempty" json:"throttle_unauthenticated_files_api_period_in_seconds,omitempty"`
	ThrottleUnauthenticatedFilesAPIRequestsPerPeriod      *int                             `url:"throttle_unauthenticated_files_api_requests_per_period,omitempty" json:"throttle_unauthenticated_files_api_requests_per_period,omitempty"`
	ThrottleUnauthenticatedGitLFSEnabled                  *bool                            `url:"throttle_unauthenticated_git_lfs_enabled,omitempty" json:"throttle_unauthenticated_git_lfs_enabled,omitempty"`
	ThrottleUnauthenticatedGitLFSPeriodInSeconds          *int                             `url:"throttle_unauthenticated_git_lfs_period_in_seconds,omitempty" json:"throttle_unauthenticated_git_lfs_period_in_seconds,omitempty"`
	ThrottleUnauthenticatedGitLFSRequestsPerPeriod        *int                             `url:"throttle_unauthenticated_git_lfs_requests_per_period,omitempty" json:"throttle_unauthenticated_git_lfs_requests_per_period,omitempty"`
	ThrottleUnauthenticatedPackagesAPIEnabled             *bool                            `url:"throttle_unauthenticated_packages_api_enabled,omitempty" json:"throttle_unauthenticated_packages_api_enabled,omitempty"`
	ThrottleUnauthenticatedPackagesAPIPeriodInSeconds     *int                             `url:"throttle_unauthenticated_packages_api_period_in_seconds,omitempty" json:"throttle_unauthenticated_packages_api_period_in_seconds,omitempty"`
	ThrottleUnauthenticatedPackagesAPIRequestsPerPeriod   *int                             `url:"throttle_unauthenticated_packages_api_requests_per_period,omitempty" json:"throttle_unauthenticated_packages_api_requests_per_period,omitempty"`
	ThrottleUnauthenticatedPeriodInSeconds                *int                             `url:"throttle_unauthenticated_period_in_seconds,omitempty" json:"throttle_unauthenticated_period_in_seconds,omitempty"`
	ThrottleUnauthenticatedRequestsPerPeriod              *int                             `url:"throttle_unauthenticated_requests_per_period,omitempty" json:"throttle_unauthenticated_requests_per_period,omitempty"`
	ThrottleUnauthenticatedWebEnabled                     *bool                            `url:"throttle_unauthenticated_web_enabled,omitempty" json:"throttle_unauthenticated_web_enabled,omitempty"`
	ThrottleUnauthenticatedWebPeriodInSeconds             *int                             `url:"throttle_unauthenticated_web_period_in_seconds,omitempty" json:"throttle_unauthenticated_web_period_in_seconds,omitempty"`
	ThrottleUnauthenticatedWebRequestsPerPeriod           *int                             `url:"throttle_unauthenticated_web_requests_per_period,omitempty" json:"throttle_unauthenticated_web_requests_per_period,omitempty"`
	TimeTrackingLimitToHours                              *bool                            `url:"time_tracking_limit_to_hours,omitempty" json:"time_tracking_limit_to_hours,omitempty"`
	TwoFactorGracePeriod                                  *int                             `url:"two_factor_grace_period,omitempty" json:"two_factor_grace_period,omitempty"`
	UnconfirmedUsersDeleteAfterDays                       *int                             `url:"unconfirmed_users_delete_after_days,omitempty" json:"unconfirmed_users_delete_after_days,omitempty"`
	UniqueIPsLimitEnabled                                 *bool                            `url:"unique_ips_limit_enabled,omitempty" json:"unique_ips_limit_enabled,omitempty"`
	UniqueIPsLimitPerUser                                 *int                             `url:"unique_ips_limit_per_user,omitempty" json:"unique_ips_limit_per_user,omitempty"`
	UniqueIPsLimitTimeWindow                              *int                             `url:"unique_ips_limit_time_window,omitempty" json:"unique_ips_limit_time_window,omitempty"`
	UpdateRunnerVersionsEnabled                           *bool                            `url:"update_runner_versions_enabled,omitempty" json:"update_runner_versions_enabled,omitempty"`
	UpdatingNameDisabledForUsers                          *bool                            `url:"updating_name_disabled_for_users,omitempty" json:"updating_name_disabled_for_users,omitempty"`
	UsagePingEnabled                                      *bool                            `url:"usage_ping_enabled,omitempty" json:"usage_ping_enabled,omitempty"`
	UsagePingFeaturesEnabled                              *bool                            `url:"usage_ping_features_enabled,omitempty" json:"usage_ping_features_enabled,omitempty"`
	UseClickhouseForAnalytics                             *bool                            `url:"use_clickhouse_for_analytics,omitempty" json:"use_clickhouse_for_analytics,omitempty"`
	UserDeactivationEmailsEnabled                         *bool                            `url:"user_deactivation_emails_enabled,omitempty" json:"user_deactivation_emails_enabled,omitempty"`
	UserDefaultExternal                                   *bool                            `url:"user_default_external,omitempty" json:"user_default_external,omitempty"`
	UserDefaultInternalRegex                              *string                          `url:"user_default_internal_regex,omitempty" json:"user_default_internal_regex,omitempty"`
	UserDefaultsToPrivateProfile                          *bool                            `url:"user_defaults_to_private_profile,omitempty" json:"user_defaults_to_private_profile,omitempty"`
	UserEmailLookupLimit                                  *int                             `url:"user_email_lookup_limit,omitempty" json:"user_email_lookup_limit,omitempty"`
	UserOauthApplications                                 *bool                            `url:"user_oauth_applications,omitempty" json:"user_oauth_applications,omitempty"`
	UserShowAddSSHKeyMessage                              *bool                            `url:"user_show_add_ssh_key_message,omitempty" json:"user_show_add_ssh_key_message,omitempty"`
	UsersGetByIDLimit                                     *int                             `url:"users_get_by_id_limit,omitempty" json:"users_get_by_id_limit,omitempty"`
	UsersGetByIDLimitAllowlistRaw                         *string                          `url:"users_get_by_id_limit_allowlist_raw,omitempty" json:"users_get_by_id_limit_allowlist_raw,omitempty"`
	ValidRunnerRegistrars                                 *[]string                        `url:"valid_runner_registrars,omitempty" json:"valid_runner_registrars,omitempty"`
	VersionCheckEnabled                                   *bool                            `url:"version_check_enabled,omitempty" json:"version_check_enabled,omitempty"`
	WebIDEClientsidePreviewEnabled                        *bool                            `url:"web_ide_clientside_preview_enabled,omitempty" json:"web_ide_clientside_preview_enabled,omitempty"`
	WhatsNewVariant                                       *string                          `url:"whats_new_variant,omitempty" json:"whats_new_variant,omitempty"`
	WikiPageMaxContentBytes                               *int                             `url:"wiki_page_max_content_bytes,omitempty" json:"wiki_page_max_content_bytes,omitempty"`
}

// BranchProtectionDefaultsOptions represents default Git protected branch permissions options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#options-for-default_branch_protection_defaults
type BranchProtectionDefaultsOptions struct {
	AllowedToPush           *[]int `url:"allowed_to_push,omitempty" json:"allowed_to_push,omitempty"`
	AllowForcePush          *bool  `url:"allow_force_push,omitempty" json:"allow_force_push,omitempty"`
	AllowedToMerge          *[]int `url:"allowed_to_merge,omitempty" json:"allowed_to_merge,omitempty"`
	DeveloperCanInitialPush *bool  `url:"developer_can_initial_push,omitempty" json:"developer_can_initial_push,omitempty"`
}

// UpdateSettings updates the application settings.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/settings.html#change-application-settings
func (s *SettingsService) UpdateSettings(opt *UpdateSettingsOptions, options ...RequestOptionFunc) (*Settings, *Response, error) {
	req, err := s.client.NewRequest(http.MethodPut, "application/settings", opt, options)
	if err != nil {
		return nil, nil, err
	}

	as := new(Settings)
	resp, err := s.client.Do(req, as)
	if err != nil {
		return nil, resp, err
	}

	return as, resp, nil
}

package scanningResultsParser

import (
	"github.com/tidwall/gjson"
)

type JsonKey string
type JsonVal string

func (jp JsonKey) string() string {
	return string(jp)
}

func (jv JsonVal) string() string {
	return string(jv)
}

const Results JsonKey = "Results"

// License parameters json path
const (
	ClassificationKey JsonKey = "Category"
	SeverityKey       JsonKey = "Severity"
	LicenseKey        JsonKey = "Name"
	PackageKey        JsonKey = "PkgName"
	SourceKey         JsonKey = "FilePath"
	ClassKey          JsonKey = "Class"
)

func parseLicense(scanResult string) *Licenses {
	var licenseRes *Licenses
	if results := gjson.Get(scanResult, Results.string()); results.IsArray() {
		results.ForEach(func(_, val gjson.Result) bool {
			if val.Get(ClassKey.string()).String() == "license-file" {
				licenseRes = &Licenses{}
				if licenses := val.Get("Licenses"); licenses.IsArray() {
					licenses.ForEach(func(_, licenseVal gjson.Result) bool {
						license := License{
							Classification: licenseVal.Get(ClassificationKey.string()).String(),
							Severity:       Severity(licenseVal.Get(SeverityKey.string()).String()),
							License:        licenseVal.Get(LicenseKey.string()).String(),
							Package:        licenseVal.Get(PackageKey.string()).String(),
							Source:         licenseVal.Get(SourceKey.string()).String(),
						}
						licenseRes.Licenses = append(licenseRes.Licenses, license)
						return true
					})
				}
			}
			return true
		})
	}
	if licenseRes != nil {
		licenseRes.Summary = buildLicenseSummary(*licenseRes)
	}
	return licenseRes
}

// Vulnerabilities paths
const (
	VulnerabilitiesKey JsonKey = "Vulnerabilities"
	CVEIdKey           JsonKey = "VulnerabilityID"
	CurrentVersionKey  JsonKey = "InstalledVersion"
	FixedInVersionKey  JsonKey = "FixedVersion"
)

func parseVulnerabilities(scanResult string) *Vulnerabilities {
	var vulnerabilitiesRes *Vulnerabilities
	if results := gjson.Get(scanResult, Results.string()); results.IsArray() {
		results.ForEach(func(_, val gjson.Result) bool {
			if vulnerabilities := val.Get(VulnerabilitiesKey.string()); vulnerabilities.IsArray() {
				vulnerabilitiesRes = &Vulnerabilities{}
				vulnerabilities.ForEach(func(_, vulnerability gjson.Result) bool {
					license := Vulnerability{
						CVEId:          vulnerability.Get(CVEIdKey.string()).String(),
						Severity:       Severity(vulnerability.Get(SeverityKey.string()).String()),
						CurrentVersion: vulnerability.Get(CurrentVersionKey.string()).String(),
						Package:        vulnerability.Get(PackageKey.string()).String(),
						FixedInVersion: vulnerability.Get(FixedInVersionKey.string()).String(),
					}
					vulnerabilitiesRes.Vulnerabilities = append(vulnerabilitiesRes.Vulnerabilities, license)
					return true
				})
			}

			return true
		})
	}
	if vulnerabilitiesRes != nil {
		vulnerabilitiesRes.Summary = buildVulnerabilitySummary(*vulnerabilitiesRes)
	}
	return vulnerabilitiesRes
}

const (
	TypeKey              JsonKey = "Type"
	FilePathKey          JsonKey = "Target"
	SuccessesKey         JsonKey = "MisconfSummary.Successes"
	FailuresKey          JsonKey = "MisconfSummary.Failures"
	ExceptionsKey        JsonKey = "MisconfSummary.Exceptions"
	MisConfigurationsKey JsonKey = "Misconfigurations"
)

const (
	ConfigVal JsonVal = "config"
)

func parseMisConfigurations(scanResult string) []*MisConfiguration {
	MisConfRes := make([]*MisConfiguration, 0)
	if results := gjson.Get(scanResult, Results.string()); results.IsArray() {
		results.ForEach(func(_, result gjson.Result) bool {
			if result.Get(ClassKey.string()).String() == ConfigVal.string() {
				misConfigurationRes := &MisConfiguration{}
				misConfigurationRes.Type = result.Get(TypeKey.string()).String()
				misConfigurationRes.FilePath = result.Get(FilePathKey.string()).String()
				misconf := MisConfigurationSummary{
					success:    result.Get(SuccessesKey.string()).Int(),
					fail:       result.Get(FailuresKey.string()).Int(),
					exceptions: result.Get(ExceptionsKey.string()).Int(),
				}
				misconf.load()
				misConfigurationRes.MisConfSummary = misconf
				// compute misConfiguration
				configurations := make([]Configuration, 0)
				if misconfigurations := result.Get(MisConfigurationsKey.string()); misconfigurations.IsArray() {
					misconfigurations.ForEach(func(_, misconfiguration gjson.Result) bool {
						configuration := Configuration{
							Id:         misconfiguration.Get("ID").String(),
							Title:      misconfiguration.Get("Title").String(),
							Message:    misconfiguration.Get("Message").String(),
							Resolution: misconfiguration.Get("Resolution").String(),
							Status:     misconfiguration.Get("Status").String(),
							Severity:   Severity(misconfiguration.Get("Severity").String()),
							CauseMetadata: CauseMetadata{
								StartLine: misconfiguration.Get("CauseMetadata.StartLine").Int(),
								EndLine:   misconfiguration.Get("CauseMetadata.EndLine").Int(),
							},
						}

						if codeLines := misconfiguration.Get("CauseMetadata.Code.Lines"); codeLines.IsArray() {
							lines := make([]Line, 0)
							codeLines.ForEach(func(_, line gjson.Result) bool {
								lines = append(lines, Line{
									Number:    line.Get("Number").Int(),
									Content:   line.Get("Content").String(),
									IsCause:   line.Get("IsCause").Bool(),
									Truncated: line.Get("Truncated").Bool(),
								})
								return true
							})
							configuration.CauseMetadata.Lines = lines

						}
						configurations = append(configurations, configuration)
						return true
					})
				}
				misConfigurationRes.Configurations = configurations
				MisConfRes = append(MisConfRes, misConfigurationRes)
			}

			return true
		})

	}

	for _, misConfigurations := range MisConfRes {
		if misConfigurations != nil {
			misConfigurations.Summary = buildConfigSummary(*misConfigurations)
		}
	}
	return MisConfRes
}

const (
	SecretVal JsonVal = "secret"
)

func parseExposedSecrets(scanResult string) []*ExposedSecret {
	var exposedSecretsRes []*ExposedSecret
	if results := gjson.Get(scanResult, Results.string()); results.IsArray() {
		results.ForEach(func(_, result gjson.Result) bool {
			if result.Get(ClassKey.string()).String() == SecretVal.string() {
				exposedSecrets := &ExposedSecret{}
				exposedSecrets.FilePath = result.Get(FilePathKey.string()).String()
				secrets := make([]Secret, 0)
				if secretObjs := result.Get("Secrets"); secretObjs.IsArray() {
					secretObjs.ForEach(func(_, secretObj gjson.Result) bool {
						secret := Secret{
							RuleId:   secretObj.Get("RuleID").String(),
							Category: secretObj.Get("Category").String(),
							Severity: Severity(secretObj.Get(SeverityKey.string()).String()),
							CauseMetadata: CauseMetadata{
								StartLine: secretObj.Get("StartLine").Int(),
								EndLine:   secretObj.Get("EndLine").Int(),
							},
						}

						if codeLines := secretObj.Get("Code.Lines"); codeLines.IsArray() {
							lines := make([]Line, 0)
							codeLines.ForEach(func(_, line gjson.Result) bool {
								lines = append(lines, Line{
									Number:    line.Get("Number").Int(),
									Content:   line.Get("Content").String(),
									IsCause:   line.Get("IsCause").Bool(),
									Truncated: line.Get("Truncated").Bool(),
								})
								return true
							})
							secret.CauseMetadata.Lines = lines
						}
						secrets = append(secrets, secret)
						return true
					})
				}
				exposedSecrets.Secrets = secrets
				exposedSecretsRes = append(exposedSecretsRes, exposedSecrets)
			}
			return true
		})
	}

	for _, exposedSecretRes := range exposedSecretsRes {
		if exposedSecretRes != nil {
			exposedSecretRes.Summary = buildSecretSummary(*exposedSecretRes)
		}
	}
	return exposedSecretsRes
}

func buildMisConfSummary(configs []Configuration) MisConfigurationSummary {
	statusMap := make(map[string]int64)
	for _, config := range configs {
		statusMap[config.Status] = statusMap[config.Status] + 1
	}
	return MisConfigurationSummary{
		Severities: statusMap,
	}
}

func buildConfigSummary(configs MisConfiguration) Summary {
	summary := make(map[Severity]int)
	for _, config := range configs.Configurations {
		summary[config.Severity] = summary[config.Severity] + 1
	}
	return Summary{
		Severities: summary,
	}

}
func buildLicenseSummary(licenses Licenses) Summary {
	summary := make(map[Severity]int)
	for _, license := range licenses.Licenses {
		summary[license.Severity] = summary[license.Severity] + 1
	}
	return Summary{
		Severities: summary,
	}
}

func buildVulnerabilitySummary(vulnerabilities Vulnerabilities) Summary {
	summary := make(map[Severity]int)
	for _, vulnerability := range vulnerabilities.Vulnerabilities {
		summary[vulnerability.Severity] = summary[vulnerability.Severity] + 1
	}
	return Summary{
		Severities: summary,
	}
}

func buildSecretSummary(exposedSecrets ExposedSecret) Summary {
	summary := make(map[Severity]int)
	for _, secret := range exposedSecrets.Secrets {
		summary[secret.Severity] = summary[secret.Severity] + 1
	}
	return Summary{
		Severities: summary,
	}
}

// ParseImageScanResult will parse the scan results of an image
func ParseImageScanResult(scanResultJson string) *ImageScanResult {
	vulnerabilities := parseVulnerabilities(scanResultJson)
	licenses := parseLicense(scanResultJson)
	return &ImageScanResult{
		License:       licenses,
		Vulnerability: vulnerabilities,
	}
}

// ParseCodeScanResult will parse the scan results of the code
func ParseCodeScanResult(scanResultJson string) *CodeScanResponse {
	vulnerabilities := parseVulnerabilities(scanResultJson)
	licenses := parseLicense(scanResultJson)
	misconfigs := parseMisConfigurations(scanResultJson)
	exposedSecrets := parseExposedSecrets(scanResultJson)
	codeScanResult := &CodeScanResponse{
		Vulnerability: vulnerabilities,
		License:       licenses,
	}
	if misconfigs != nil {
		codeScanResult.MisConfigurations = &MisConfigurations{
			MisConfigurations: misconfigs,
		}
		// update summary
		misconfigSummary := &MisConfigurationSummary{}
		for _, misconfig := range misconfigs {
			misconfigSummary.success = misconfigSummary.success + misconfig.MisConfSummary.success
			misconfigSummary.fail = misconfigSummary.fail + misconfig.MisConfSummary.fail
			misconfigSummary.exceptions = misconfigSummary.exceptions + misconfig.MisConfSummary.exceptions
		}
		misconfigSummary.load()
		codeScanResult.MisConfigurations.Summary = *misconfigSummary
	}

	if exposedSecrets != nil {
		codeScanResult.ExposedSecrets = &ExposedSecrets{
			ExposedSecrets: exposedSecrets,
		}

		// 	update summary
		severities := make(map[Severity]int)
		for _, expoSecret := range exposedSecrets {
			for key, val := range expoSecret.Summary.Severities {
				severities[key] += val
			}
		}

		codeScanResult.ExposedSecrets.Summary = Summary{
			Severities: severities,
		}

	}

	return codeScanResult
}

// ParseK8sConfigScanResult will parse the scan results of manifest
func ParseK8sConfigScanResult(scanResultJson string) *K8sManifestScanResponse {
	misconfigs := parseMisConfigurations(scanResultJson)
	exposedSecrets := parseExposedSecrets(scanResultJson)
	manifestResult := &K8sManifestScanResponse{}
	if misconfigs != nil {
		manifestResult.MisConfigurations = &MisConfigurations{
			MisConfigurations: misconfigs,
		}
		// update summary
		misconfigSummary := &MisConfigurationSummary{}
		for _, misconfig := range misconfigs {
			misconfigSummary.success = misconfigSummary.success + misconfig.MisConfSummary.success
			misconfigSummary.fail = misconfigSummary.fail + misconfig.MisConfSummary.fail
			misconfigSummary.exceptions = misconfigSummary.exceptions + misconfig.MisConfSummary.exceptions
		}
		misconfigSummary.load()
		manifestResult.MisConfigurations.Summary = *misconfigSummary
	}

	if exposedSecrets != nil {
		manifestResult.ExposedSecrets = &ExposedSecrets{
			ExposedSecrets: exposedSecrets,
		}

		// 	update summary
		severities := make(map[Severity]int)
		for _, expoSecret := range exposedSecrets {
			for key, val := range expoSecret.Summary.Severities {
				severities[key] += val
			}
		}

		manifestResult.ExposedSecrets.Summary = Summary{
			Severities: severities,
		}

	}
	return manifestResult
}

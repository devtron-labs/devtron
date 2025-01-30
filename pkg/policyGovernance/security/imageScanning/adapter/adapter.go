package adapter

import (
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/bean"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/helper/parser"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/repository"
	bean3 "github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/repository/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"time"
)

func BuildVulnerabilitiesWrapperWithSummary(allVulnerability []*bean.Vulnerabilities) *parser.Vulnerabilities {
	parsedVul := ConvertBeanVulnerabilityToParserFormat(allVulnerability)
	return &parser.Vulnerabilities{
		Summary:         parser.BuildVulnerabilitySummary(parsedVul),
		Vulnerabilities: parsedVul,
	}
}

func ConvertBeanVulnerabilityToParserFormat(vulnerabilities []*bean.Vulnerabilities) []parser.Vulnerability {
	parsedVulnerabilities := make([]parser.Vulnerability, 0, len(vulnerabilities))
	for _, vulnerability := range vulnerabilities {
		parsedVulnerabilities = append(parsedVulnerabilities, parser.Vulnerability{
			CVEId:          vulnerability.CVEName,
			Severity:       vulnerability.ToSeverity(),
			Package:        vulnerability.Package,
			CurrentVersion: vulnerability.CVersion,
			FixedInVersion: vulnerability.FVersion,
			Target:         vulnerability.Target,
			Class:          vulnerability.Class,
			Type:           vulnerability.Type,
			Permission:     vulnerability.Permission,
		})
	}
	return parsedVulnerabilities

}

func BuildImageVulnerabilityResponse(image string, vulnerabilities parser.Vulnerabilities, metadata *parser.Metadata) *parser.ImageVulnerability {
	return &parser.ImageVulnerability{Image: image, Vulnerabilities: vulnerabilities, Metadata: metadata}
}

func BuildMetadata(status string, startedOn time.Time, scanToolName string, scanToolUrl string) *parser.Metadata {
	return &parser.Metadata{
		Status:       status,
		StartedOn:    startedOn,
		ScanToolName: scanToolName,
		ScanToolUrl:  scanToolUrl,
	}
}

func ExecutionDetailsToResourceScanResponseDto(respFromExecutionDetail *bean.ImageScanExecutionDetail) (resp parser.ResourceScanResponseDto) {
	resp.Scanned = respFromExecutionDetail.Scanned
	resp.IsImageScanEnabled = respFromExecutionDetail.ScanEnabled
	// if not scanned
	if resp.Scanned == false {
		// sanitise response if not scanned
		resp.ImageScan = nil
		return resp
	}
	vulnerabilityResponse := &parser.VulnerabilityResponse{}
	vulnerabilities := BuildVulnerabilitiesWrapperWithSummary(respFromExecutionDetail.Vulnerabilities)
	imageVulResp := BuildImageVulnerabilityResponse(respFromExecutionDetail.Image, *vulnerabilities, BuildMetadata(respFromExecutionDetail.Status.String(), respFromExecutionDetail.ExecutionTime, respFromExecutionDetail.ScanToolName, respFromExecutionDetail.ScanToolUrl))
	vulnerabilityResponse.Append(*imageVulResp)
	resp.ImageScan = &parser.ImageScanResponse{Vulnerability: vulnerabilityResponse}
	return resp
}

func BuildCvePolicy(request *bean2.CreateVulnerabilityPolicyRequest, action bean3.PolicyAction, severity bean3.Severity, time time.Time, userId int32) *repository.CvePolicy {
	cvePolicy := &repository.CvePolicy{
		Action:   action,
		Severity: &severity,
		AuditLog: sql.AuditLog{
			CreatedOn: time,
			CreatedBy: userId,
			UpdatedOn: time,
			UpdatedBy: userId,
		},
	}
	if request != nil {
		cvePolicy.Global = request.IsRequestGlobal()
		cvePolicy.ClusterId = request.ClusterId
		cvePolicy.EnvironmentId = request.EnvId
		cvePolicy.AppId = request.AppId
		cvePolicy.CVEStoreId = request.CveId
	}
	return cvePolicy
}

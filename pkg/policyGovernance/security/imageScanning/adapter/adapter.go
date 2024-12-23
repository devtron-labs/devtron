package adapter

import (
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/bean"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/helper/parser"
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

func BuildImageVulnerabilityResponse(image string, vulnerabilities parser.Vulnerabilities, metadata parser.Metadata) *parser.ImageVulnerability {
	return &parser.ImageVulnerability{Image: image, Vulnerabilities: vulnerabilities, Metadata: metadata}
}

func BuildMetadata(status string, startedOn time.Time, scanToolName string) parser.Metadata {
	return parser.Metadata{
		Status:       status,
		StartedOn:    startedOn,
		ScanToolName: scanToolName,
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
	imageVulResp := BuildImageVulnerabilityResponse(respFromExecutionDetail.Image, *vulnerabilities, BuildMetadata(respFromExecutionDetail.Status.String(), respFromExecutionDetail.ExecutionTime, respFromExecutionDetail.ScanToolName))
	vulnerabilityResponse.Append(*imageVulResp)
	resp.ImageScan = &parser.ImageScanResponse{Vulnerability: vulnerabilityResponse}
	return resp
}

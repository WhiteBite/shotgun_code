package analysis

import (
	"context"
	"fmt"
	"os"
	"shotgun_code/domain"
	"shotgun_code/infrastructure/sbomlicensing"
	"strings"
	"time"
)

// SBOMService provides high-level API for SBOM and licensing within the analysis bounded context
type SBOMService struct {
	log            domain.Logger
	syftGenerator  *sbomlicensing.SyftGenerator
	grypeScanner   *sbomlicensing.GrypeScanner
	licenseScanner *sbomlicensing.LicenseScanner
}

// NewSBOMService creates a new SBOM service
func NewSBOMService(log domain.Logger) *SBOMService {
	return &SBOMService{
		log:            log,
		syftGenerator:  sbomlicensing.NewSyftGenerator(log),
		grypeScanner:   sbomlicensing.NewGrypeScanner(log),
		licenseScanner: sbomlicensing.NewLicenseScanner(log),
	}
}

// GenerateSBOM generates SBOM for a project
func (s *SBOMService) GenerateSBOM(ctx context.Context, projectPath string, format domain.SBOMFormat) (*domain.SBOMResult, error) {
	s.log.Info(fmt.Sprintf("Generating SBOM for project: %s, format: %s", projectPath, format))

	// Check if project exists
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return &domain.SBOMResult{
			Success:     false,
			ProjectPath: projectPath,
			Format:      format,
			Error:       fmt.Sprintf("project path does not exist: %s", projectPath),
		}, nil
	}

	// Check Syft availability
	if !s.syftGenerator.IsAvailable() {
		return &domain.SBOMResult{
			Success:     false,
			ProjectPath: projectPath,
			Format:      format,
			Error:       "Syft is not available",
		}, nil
	}

	// Generate SBOM using Syft
	result, err := s.syftGenerator.GenerateSBOM(ctx, projectPath, format)
	if err != nil {
		return &domain.SBOMResult{
			Success:     false,
			ProjectPath: projectPath,
			Format:      format,
			Error:       err.Error(),
		}, nil
	}

	return result, nil
}

// ScanVulnerabilities scans vulnerabilities in a project
func (s *SBOMService) ScanVulnerabilities(ctx context.Context, projectPath string) (*domain.VulnerabilityScanResult, error) {
	s.log.Info(fmt.Sprintf("Scanning vulnerabilities for project: %s", projectPath))

	// Check if project exists
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return &domain.VulnerabilityScanResult{
			Success:     false,
			ProjectPath: projectPath,
			Error:       fmt.Sprintf("project path does not exist: %s", projectPath),
		}, nil
	}

	// Check Grype availability
	if !s.grypeScanner.IsAvailable() {
		return &domain.VulnerabilityScanResult{
			Success:     false,
			ProjectPath: projectPath,
			Error:       "Grype is not available",
		}, nil
	}

	// Scan vulnerabilities using Grype
	result, err := s.grypeScanner.ScanVulnerabilities(ctx, projectPath)
	if err != nil {
		return &domain.VulnerabilityScanResult{
			Success:     false,
			ProjectPath: projectPath,
			Error:       err.Error(),
		}, nil
	}

	return result, nil
}

// ScanLicenses scans licenses in a project
func (s *SBOMService) ScanLicenses(ctx context.Context, projectPath string) (*domain.LicenseScanResult, error) {
	s.log.Info(fmt.Sprintf("Scanning licenses for project: %s", projectPath))

	// Check if project exists
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return &domain.LicenseScanResult{
			Success:     false,
			ProjectPath: projectPath,
			Error:       fmt.Sprintf("project path does not exist: %s", projectPath),
		}, nil
	}

	// Check license scanning tools availability
	if !s.licenseScanner.IsAvailable() {
		return &domain.LicenseScanResult{
			Success:     false,
			ProjectPath: projectPath,
			Error:       "No license scanning tools available",
		}, nil
	}

	// Scan licenses using LicenseScanner
	result, err := s.licenseScanner.ScanLicenses(ctx, projectPath)
	if err != nil {
		return &domain.LicenseScanResult{
			Success:     false,
			ProjectPath: projectPath,
			Error:       err.Error(),
		}, nil
	}

	return result, nil
}

// GenerateComplianceReport generates compliance report
func (s *SBOMService) GenerateComplianceReport(ctx context.Context, projectPath string, requirements *domain.ComplianceRequirements) (*domain.ComplianceReport, error) {
	s.log.Info(fmt.Sprintf("Generating compliance report for project: %s", projectPath))

	// Check if project exists
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return &domain.ComplianceReport{
			Success:     false,
			ProjectPath: projectPath,
			Compliant:   false,
		}, nil
	}

	// Generate SBOM
	sbomResult, err := s.GenerateSBOM(ctx, projectPath, domain.SBOMFormatSPDX)
	if err != nil {
		return &domain.ComplianceReport{
			Success:     false,
			ProjectPath: projectPath,
			Compliant:   false,
		}, err
	}

	// Scan vulnerabilities
	vulnResult, err := s.ScanVulnerabilities(ctx, projectPath)
	if err != nil {
		return &domain.ComplianceReport{
			Success:     false,
			ProjectPath: projectPath,
			Compliant:   false,
		}, err
	}

	// Scan licenses
	licenseResult, err := s.ScanLicenses(ctx, projectPath)
	if err != nil {
		return &domain.ComplianceReport{
			Success:     false,
			ProjectPath: projectPath,
			Compliant:   false,
		}, err
	}

	// Analyze compliance
	report := s.analyzeCompliance(projectPath, requirements, sbomResult, vulnResult, licenseResult)

	return report, nil
}

// GetSupportedSBOMFormats returns supported SBOM formats
func (s *SBOMService) GetSupportedSBOMFormats() []domain.SBOMFormat {
	return []domain.SBOMFormat{
		domain.SBOMFormatSPDX,
		domain.SBOMFormatCycloneDX,
		domain.SBOMFormatJSON,
	}
}

// ValidateSBOM validates SBOM
func (s *SBOMService) ValidateSBOM(ctx context.Context, sbomPath string, format domain.SBOMFormat) error {
	s.log.Info(fmt.Sprintf("Validating SBOM: %s, format: %s", sbomPath, format))

	// Check if file exists
	if _, err := os.Stat(sbomPath); os.IsNotExist(err) {
		return fmt.Errorf("SBOM file does not exist: %s", sbomPath)
	}

	// Read SBOM content
	content, err := os.ReadFile(sbomPath)
	if err != nil {
		return fmt.Errorf("failed to read SBOM file: %w", err)
	}

	// Basic validation based on format
	switch format {
	case domain.SBOMFormatSPDX:
		return s.validateSPDX(content)
	case domain.SBOMFormatCycloneDX:
		return s.validateCycloneDX(content)
	case domain.SBOMFormatJSON:
		return s.validateJSON(content)
	default:
		return fmt.Errorf("unsupported SBOM format: %s", format)
	}
}

// CheckToolsAvailability checks if SBOM tools are available
func (s *SBOMService) CheckToolsAvailability() *domain.SBOMToolsStatus {
	return &domain.SBOMToolsStatus{
		SyftAvailable:  s.syftGenerator.IsAvailable(),
		GrypeAvailable: s.grypeScanner.IsAvailable(),
	}
}

// GetVulnerabilityStats returns vulnerability statistics
func (s *SBOMService) GetVulnerabilityStats(result *domain.VulnerabilityScanResult) *domain.VulnerabilityStats {
	if result == nil || !result.Success || len(result.Vulnerabilities) == 0 {
		return &domain.VulnerabilityStats{
			Total:    0,
			Critical: 0,
			High:     0,
			Medium:   0,
			Low:      0,
		}
	}

	stats := &domain.VulnerabilityStats{}
	for _, vuln := range result.Vulnerabilities {
		stats.Total++
		switch strings.ToLower(vuln.Severity) {
		case "critical":
			stats.Critical++
		case "high":
			stats.High++
		case "medium":
			stats.Medium++
		case "low":
			stats.Low++
		}
	}

	return stats
}

// GetLicenseStats returns license statistics
func (s *SBOMService) GetLicenseStats(result *domain.LicenseScanResult) *domain.LicenseStats {
	if result == nil || !result.Success || len(result.Licenses) == 0 {
		return &domain.LicenseStats{
			Total:         0,
			ByLicense:     make(map[string]int),
			Permissive:    0,
			Copyleft:      0,
			Proprietary:   0,
			Unknown:       0,
		}
	}

	stats := &domain.LicenseStats{
		ByLicense: make(map[string]int),
	}

	for _, license := range result.Licenses {
		stats.Total++
		stats.ByLicense[license.SPDXID]++
		
		// Categorize license types
		switch s.categorizeLicense(license.SPDXID) {
		case "permissive":
			stats.Permissive++
		case "copyleft":
			stats.Copyleft++
		case "proprietary":
			stats.Proprietary++
		}
	}

	return stats
}

// Private helper methods

func (s *SBOMService) analyzeCompliance(projectPath string, requirements *domain.ComplianceRequirements, sbom *domain.SBOMResult, vuln *domain.VulnerabilityScanResult, license *domain.LicenseScanResult) *domain.ComplianceReport {
	report := &domain.ComplianceReport{
		ProjectPath:       projectPath,
		Success:           true,
		Compliant:         true,
		GeneratedAt:       time.Now(),
		SBOMResult:        sbom,
		VulnerabilityResult: vuln,
		LicenseResult:     license,
		Issues:            make([]*domain.ComplianceIssue, 0),
	}

	// Check SBOM generation success
	if sbom == nil || !sbom.Success {
		report.Issues = append(report.Issues, &domain.ComplianceIssue{
			Type:        "sbom-generation",
			Severity:    "high",
			Description: "Failed to generate SBOM",
		})
		report.Compliant = false
	}

	// Check vulnerability compliance
	if vuln != nil && vuln.Success && requirements != nil {
		vulnStats := s.GetVulnerabilityStats(vuln)
		
		if requirements.MaxCriticalVulnerabilities >= 0 && vulnStats.Critical > requirements.MaxCriticalVulnerabilities {
			report.Issues = append(report.Issues, &domain.ComplianceIssue{
				Type:        "vulnerability-critical",
				Severity:    "critical",
				Description: fmt.Sprintf("Critical vulnerabilities exceed limit: %d > %d", vulnStats.Critical, requirements.MaxCriticalVulnerabilities),
			})
			report.Compliant = false
		}

		if requirements.MaxHighVulnerabilities >= 0 && vulnStats.High > requirements.MaxHighVulnerabilities {
			report.Issues = append(report.Issues, &domain.ComplianceIssue{
				Type:        "vulnerability-high",
				Severity:    "high",
				Description: fmt.Sprintf("High vulnerabilities exceed limit: %d > %d", vulnStats.High, requirements.MaxHighVulnerabilities),
			})
			report.Compliant = false
		}
	}

	// Check license compliance
	if license != nil && license.Success && requirements != nil {
		for _, lic := range license.Licenses {
			if s.isLicenseForbidden(lic.SPDXID, requirements.ForbiddenLicenses) {
				report.Issues = append(report.Issues, &domain.ComplianceIssue{
					Type:        "license-forbidden",
					Severity:    "high",
					Description: fmt.Sprintf("Forbidden license found: %s", lic.SPDXID),
				})
				report.Compliant = false
			}
		}
	}

	return report
}

func (s *SBOMService) validateSPDX(content []byte) error {
	// Basic SPDX validation
	contentStr := string(content)
	if !strings.Contains(contentStr, "SPDXVersion") {
		return fmt.Errorf("invalid SPDX format: missing SPDXVersion")
	}
	return nil
}

func (s *SBOMService) validateCycloneDX(content []byte) error {
	// Basic CycloneDX validation
	contentStr := string(content)
	if !strings.Contains(contentStr, "bomFormat") && !strings.Contains(contentStr, "cyclonedx") {
		return fmt.Errorf("invalid CycloneDX format")
	}
	return nil
}

func (s *SBOMService) validateJSON(content []byte) error {
	// Basic JSON validation - could be enhanced with schema validation
	contentStr := strings.TrimSpace(string(content))
	if !strings.HasPrefix(contentStr, "{") || !strings.HasSuffix(contentStr, "}") {
		return fmt.Errorf("invalid JSON format")
	}
	return nil
}

func (s *SBOMService) categorizeLicense(spdxID string) string {
	permissive := []string{"MIT", "Apache-2.0", "BSD-2-Clause", "BSD-3-Clause", "ISC"}
	copyleft := []string{"GPL-2.0", "GPL-3.0", "LGPL-2.1", "LGPL-3.0", "AGPL-3.0"}

	for _, p := range permissive {
		if strings.Contains(spdxID, p) {
			return "permissive"
		}
	}

	for _, c := range copyleft {
		if strings.Contains(spdxID, c) {
			return "copyleft"
		}
	}

	return "proprietary"
}

func (s *SBOMService) isLicenseForbidden(spdxID string, forbidden []string) bool {
	for _, f := range forbidden {
		if strings.EqualFold(spdxID, f) {
			return true
		}
	}
	return false
}
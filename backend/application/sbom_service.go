package application

import (
	"context"
	"fmt"
	"shotgun_code/domain"
	"strings"
)

// SBOMService предоставляет высокоуровневый API для работы с SBOM и лицензиями
type SBOMService struct {
	log               domain.Logger
	sbomGenerator     domain.SBOMGenerator
	vulnScanner       domain.VulnerabilityScanner
	licenseScanner    domain.LicenseScanner
	fileStatProvider  domain.FileStatProvider
}

// NewSBOMService создает новый сервис SBOM
func NewSBOMService(log domain.Logger, sbomGenerator domain.SBOMGenerator, vulnScanner domain.VulnerabilityScanner, licenseScanner domain.LicenseScanner, fileStatProvider domain.FileStatProvider) *SBOMService {
	return &SBOMService{
		log:              log,
		sbomGenerator:    sbomGenerator,
		vulnScanner:      vulnScanner,
		licenseScanner:   licenseScanner,
		fileStatProvider: fileStatProvider,
	}
}

// GenerateSBOM генерирует SBOM для проекта
func (s *SBOMService) GenerateSBOM(ctx context.Context, projectPath string, format domain.SBOMFormat) (*domain.SBOMResult, error) {
	s.log.Info(fmt.Sprintf("Generating SBOM for project: %s, format: %s", projectPath, format))

	// Проверяем существование проекта
	if _, err := s.fileStatProvider.Stat(projectPath); err != nil {
		return &domain.SBOMResult{
			Success:     false,
			ProjectPath: projectPath,
			Format:      format,
			Error:       fmt.Sprintf("project path does not exist: %s", projectPath),
		}, nil
	}

	// Проверяем доступность Syft
	if !s.sbomGenerator.IsAvailable() {
		return &domain.SBOMResult{
			Success:     false,
			ProjectPath: projectPath,
			Format:      format,
			Error:       "SBOM generator is not available",
		}, nil
	}

	// Генерируем SBOM с помощью Syft
	result, err := s.sbomGenerator.GenerateSBOM(ctx, projectPath, format)
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

// ScanVulnerabilities сканирует уязвимости в проекте
func (s *SBOMService) ScanVulnerabilities(ctx context.Context, projectPath string) (*domain.VulnerabilityScanResult, error) {
	s.log.Info(fmt.Sprintf("Scanning vulnerabilities for project: %s", projectPath))

	// Проверяем существование проекта
	if _, err := s.fileStatProvider.Stat(projectPath); err != nil {
		return &domain.VulnerabilityScanResult{
			Success:     false,
			ProjectPath: projectPath,
			Error:       fmt.Sprintf("project path does not exist: %s", projectPath),
		}, nil
	}

	// Проверяем доступность Grype
	if !s.vulnScanner.IsAvailable() {
		return &domain.VulnerabilityScanResult{
			Success:     false,
			ProjectPath: projectPath,
			Error:       "Vulnerability scanner is not available",
		}, nil
	}

	// Сканируем уязвимости с помощью Grype
	result, err := s.vulnScanner.ScanVulnerabilities(ctx, projectPath)
	if err != nil {
		return &domain.VulnerabilityScanResult{
			Success:     false,
			ProjectPath: projectPath,
			Error:       err.Error(),
		}, nil
	}

	return result, nil
}

// ScanLicenses сканирует лицензии в проекте
func (s *SBOMService) ScanLicenses(ctx context.Context, projectPath string) (*domain.LicenseScanResult, error) {
	s.log.Info(fmt.Sprintf("Scanning licenses for project: %s", projectPath))

	// Проверяем существование проекта
	if _, err := s.fileStatProvider.Stat(projectPath); err != nil {
		return &domain.LicenseScanResult{
			Success:     false,
			ProjectPath: projectPath,
			Error:       fmt.Sprintf("project path does not exist: %s", projectPath),
		}, nil
	}

	// Проверяем доступность инструментов сканирования лицензий
	if !s.licenseScanner.IsAvailable() {
		return &domain.LicenseScanResult{
			Success:     false,
			ProjectPath: projectPath,
			Error:       "No license scanning tools available",
		}, nil
	}

	// Сканируем лицензии с помощью LicenseScanner
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

// GenerateComplianceReport генерирует отчет о соответствии
func (s *SBOMService) GenerateComplianceReport(ctx context.Context, projectPath string, requirements *domain.ComplianceRequirements) (*domain.ComplianceReport, error) {
	s.log.Info(fmt.Sprintf("Generating compliance report for project: %s", projectPath))

	// Проверяем существование проекта
	if _, err := s.fileStatProvider.Stat(projectPath); err != nil {
		return &domain.ComplianceReport{
			Success:     false,
			ProjectPath: projectPath,
			Compliant:   false,
		}, nil
	}

	// Генерируем SBOM
	sbomResult, err := s.GenerateSBOM(ctx, projectPath, domain.SBOMFormatSPDX)
	if err != nil {
		return &domain.ComplianceReport{
			Success:     false,
			ProjectPath: projectPath,
			Compliant:   false,
		}, err
	}

	// Сканируем уязвимости
	vulnResult, err := s.ScanVulnerabilities(ctx, projectPath)
	if err != nil {
		return &domain.ComplianceReport{
			Success:     false,
			ProjectPath: projectPath,
			Compliant:   false,
		}, err
	}

	// Сканируем лицензии
	licenseResult, err := s.ScanLicenses(ctx, projectPath)
	if err != nil {
		return &domain.ComplianceReport{
			Success:     false,
			ProjectPath: projectPath,
			Compliant:   false,
		}, err
	}

	// Анализируем соответствие
	report := s.analyzeCompliance(projectPath, requirements, sbomResult, vulnResult, licenseResult)

	return report, nil
}

// GetSupportedSBOMFormats возвращает поддерживаемые форматы SBOM
func (s *SBOMService) GetSupportedSBOMFormats() []domain.SBOMFormat {
	// TODO: This should be configurable based on the SBOM generator
	return []domain.SBOMFormat{
		domain.SBOMFormatSPDX,
		domain.SBOMFormatCycloneDX,
		domain.SBOMFormatJSON,
	}
}

// ValidateSBOM валидирует SBOM
func (s *SBOMService) ValidateSBOM(ctx context.Context, sbomPath string, format domain.SBOMFormat) error {
	s.log.Info(fmt.Sprintf("Validating SBOM: %s, format: %s", sbomPath, format))

	// Проверяем существование файла
	if _, err := s.fileStatProvider.Stat(sbomPath); err != nil {
		return fmt.Errorf("SBOM file does not exist: %s", sbomPath)
	}

	// Проверяем доступность Syft
	if !s.sbomGenerator.IsAvailable() {
		return fmt.Errorf("SBOM generator is not available for validation")
	}

	// Валидируем SBOM с помощью Syft
	return s.sbomGenerator.ValidateSBOM(ctx, sbomPath, format)
}

// analyzeCompliance анализирует соответствие требованиям
func (s *SBOMService) analyzeCompliance(projectPath string, requirements *domain.ComplianceRequirements, sbomResult *domain.SBOMResult, vulnResult *domain.VulnerabilityScanResult, licenseResult *domain.LicenseScanResult) *domain.ComplianceReport {
	var issues []*domain.ComplianceIssue

	// Проверяем SBOM
	if requirements.RequireSBOM && !sbomResult.Success {
		issues = append(issues, &domain.ComplianceIssue{
			Type:        "sbom",
			Severity:    "high",
			Description: "SBOM generation failed",
		})
	}

	// Проверяем уязвимости
	if vulnResult.Success && vulnResult.Summary != nil {
		if vulnResult.Summary.Total > requirements.MaxVulnerabilities {
			issues = append(issues, &domain.ComplianceIssue{
				Type:        "vulnerability",
				Severity:    "high",
				Description: fmt.Sprintf("Too many vulnerabilities: %d (max: %d)", vulnResult.Summary.Total, requirements.MaxVulnerabilities),
			})
		}

		// Проверяем критические уязвимости
		if vulnResult.Summary.Critical > 0 {
			issues = append(issues, &domain.ComplianceIssue{
				Type:        "vulnerability",
				Severity:    "critical",
				Description: fmt.Sprintf("Critical vulnerabilities found: %d", vulnResult.Summary.Critical),
			})
		}
	}

	// Проверяем лицензии
	if licenseResult.Success && licenseResult.Summary != nil {
		for _, license := range licenseResult.Licenses {
			// Проверяем запрещенные лицензии
			for _, forbidden := range requirements.ForbiddenLicenses {
				if strings.Contains(strings.ToLower(license.Name), strings.ToLower(forbidden)) {
					issues = append(issues, &domain.ComplianceIssue{
						Type:        "license",
						Severity:    "high",
						Description: fmt.Sprintf("Forbidden license found: %s", license.Name),
						Component:   license.Name,
					})
				}
			}
		}
	}

	// Рассчитываем сводку
	summary := &domain.ComplianceSummary{
		TotalIssues: len(issues),
	}

	for _, issue := range issues {
		switch issue.Severity {
		case "critical":
			summary.Critical++
		case "high":
			summary.High++
		case "medium":
			summary.Medium++
		case "low":
			summary.Low++
		}
	}

	// Определяем общее соответствие
	compliant := len(issues) == 0

	return &domain.ComplianceReport{
		Success:     true,
		ProjectPath: projectPath,
		Compliant:   compliant,
		Issues:      issues,
		Summary:     summary,
	}
}

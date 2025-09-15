package sbomlicensing

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"shotgun_code/domain"
	"strings"
)

// LicenseScanner представляет сканер лицензий
type LicenseScanner struct {
	log domain.Logger
}

// NewLicenseScanner создает новый сканер лицензий
func NewLicenseScanner(log domain.Logger) *LicenseScanner {
	return &LicenseScanner{
		log: log,
	}
}

// IsAvailable проверяет доступность инструментов сканирования лицензий
func (l *LicenseScanner) IsAvailable() bool {
	// Проверяем доступность различных инструментов
	tools := []string{"licensecheck", "licensee", "go-licenses"}
	for _, tool := range tools {
		if _, err := exec.LookPath(tool); err == nil {
			return true
		}
	}
	return false
}

// GetAvailableTools возвращает список доступных инструментов
func (l *LicenseScanner) GetAvailableTools() []string {
	var tools []string
	availableTools := []string{"licensecheck", "licensee", "go-licenses"}

	for _, tool := range availableTools {
		if _, err := exec.LookPath(tool); err == nil {
			tools = append(tools, tool)
		}
	}

	return tools
}

// ScanLicenses сканирует лицензии в проекте
func (l *LicenseScanner) ScanLicenses(ctx context.Context, projectPath string) (*domain.LicenseScanResult, error) {
	if !l.IsAvailable() {
		return &domain.LicenseScanResult{
			Success:     false,
			ProjectPath: projectPath,
			Error:       "no license scanning tools available",
		}, nil
	}

	l.log.Info(fmt.Sprintf("Scanning licenses for: %s", projectPath))

	// Определяем язык проекта
	language := l.detectLanguage(projectPath)

	// Выбираем подходящий инструмент
	tool := l.selectTool(language)

	// Сканируем лицензии
	var licenses []*domain.LicenseInfo
	var err error

	switch tool {
	case "licensecheck":
		licenses, err = l.scanWithLicensecheck(ctx, projectPath)
	case "licensee":
		licenses, err = l.scanWithLicensee(ctx, projectPath)
	case "go-licenses":
		licenses, err = l.scanWithGoLicenses(ctx, projectPath)
	default:
		return &domain.LicenseScanResult{
			Success:     false,
			ProjectPath: projectPath,
			Error:       fmt.Sprintf("no suitable tool found for language: %s", language),
		}, nil
	}

	if err != nil {
		return &domain.LicenseScanResult{
			Success:     false,
			ProjectPath: projectPath,
			Error:       fmt.Sprintf("license scan failed: %v", err),
		}, nil
	}

	// Рассчитываем сводку
	summary := l.calculateLicenseSummary(licenses)

	return &domain.LicenseScanResult{
		Success:     true,
		ProjectPath: projectPath,
		Licenses:    licenses,
		Summary:     summary,
		Metadata: map[string]interface{}{
			"tool":     tool,
			"language": language,
		},
	}, nil
}

// detectLanguage определяет язык проекта
func (l *LicenseScanner) detectLanguage(projectPath string) string {
	if _, err := exec.LookPath(filepath.Join(projectPath, "go.mod")); err == nil {
		return "go"
	}
	if _, err := exec.LookPath(filepath.Join(projectPath, "package.json")); err == nil {
		return "node"
	}
	if _, err := exec.LookPath(filepath.Join(projectPath, "pom.xml")); err == nil {
		return "java"
	}
	if _, err := exec.LookPath(filepath.Join(projectPath, "build.gradle")); err == nil {
		return "java"
	}
	return "unknown"
}

// selectTool выбирает подходящий инструмент для языка
func (l *LicenseScanner) selectTool(language string) string {
	switch language {
	case "go":
		if _, err := exec.LookPath("go-licenses"); err == nil {
			return "go-licenses"
		}
	case "node":
		if _, err := exec.LookPath("licensee"); err == nil {
			return "licensee"
		}
	}

	// Fallback к licensecheck
	if _, err := exec.LookPath("licensecheck"); err == nil {
		return "licensecheck"
	}

	return ""
}

// scanWithLicensecheck сканирует с помощью licensecheck
func (l *LicenseScanner) scanWithLicensecheck(ctx context.Context, projectPath string) ([]*domain.LicenseInfo, error) {
	l.log.Info("Scanning licenses with licensecheck")

	cmd := exec.CommandContext(ctx, "licensecheck", "-r", projectPath)
	cmd.Dir = projectPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("licensecheck failed: %v, output: %s", err, string(output))
	}

	return l.parseLicensecheckOutput(output), nil
}

// scanWithLicensee сканирует с помощью licensee
func (l *LicenseScanner) scanWithLicensee(ctx context.Context, projectPath string) ([]*domain.LicenseInfo, error) {
	l.log.Info("Scanning licenses with licensee")

	cmd := exec.CommandContext(ctx, "licensee", "detect", projectPath)
	cmd.Dir = projectPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("licensee failed: %v, output: %s", err, string(output))
	}

	return l.parseLicenseeOutput(output), nil
}

// scanWithGoLicenses сканирует с помощью go-licenses
func (l *LicenseScanner) scanWithGoLicenses(ctx context.Context, projectPath string) ([]*domain.LicenseInfo, error) {
	l.log.Info("Scanning licenses with go-licenses")

	cmd := exec.CommandContext(ctx, "go-licenses", "csv", ".")
	cmd.Dir = projectPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("go-licenses failed: %v, output: %s", err, string(output))
	}

	return l.parseGoLicensesOutput(output), nil
}

// parseLicensecheckOutput парсит вывод licensecheck
func (l *LicenseScanner) parseLicensecheckOutput(output []byte) []*domain.LicenseInfo {
	// Упрощенная реализация
	l.log.Info(fmt.Sprintf("Parsing licensecheck output, size: %d bytes", len(output)))

	// TODO: Реализовать парсинг вывода licensecheck
	// Пример формата: filename: LICENSE_TYPE

	return []*domain.LicenseInfo{}
}

// parseLicenseeOutput парсит вывод licensee
func (l *LicenseScanner) parseLicenseeOutput(output []byte) []*domain.LicenseInfo {
	// Упрощенная реализация
	l.log.Info(fmt.Sprintf("Parsing licensee output, size: %d bytes", len(output)))

	// TODO: Реализовать парсинг вывода licensee
	// Пример формата: filename: LICENSE_TYPE (confidence: XX%)

	return []*domain.LicenseInfo{}
}

// parseGoLicensesOutput парсит вывод go-licenses
func (l *LicenseScanner) parseGoLicensesOutput(output []byte) []*domain.LicenseInfo {
	// Упрощенная реализация
	l.log.Info(fmt.Sprintf("Parsing go-licenses output, size: %d bytes", len(output)))

	// TODO: Реализовать парсинг CSV вывода go-licenses
	// Пример формата: package,license,confidence

	return []*domain.LicenseInfo{}
}

// calculateLicenseSummary рассчитывает сводку лицензий
func (l *LicenseScanner) calculateLicenseSummary(licenses []*domain.LicenseInfo) *domain.LicenseSummary {
	summary := &domain.LicenseSummary{
		TotalLicenses: len(licenses),
		ByType:        make(map[string]int),
		ByLicense:     make(map[string]int),
		Conflicts:     []*domain.LicenseConflict{},
	}

	// Подсчитываем лицензии по типам и названиям
	for _, license := range licenses {
		summary.ByType[license.Type]++
		summary.ByLicense[license.Name]++
	}

	// Проверяем конфликты лицензий
	summary.Conflicts = l.detectLicenseConflicts(licenses)

	return summary
}

// detectLicenseConflicts обнаруживает конфликты лицензий
func (l *LicenseScanner) detectLicenseConflicts(licenses []*domain.LicenseInfo) []*domain.LicenseConflict {
	var conflicts []*domain.LicenseConflict

	// Упрощенная логика обнаружения конфликтов
	// В реальном приложении здесь будет более сложная логика

	// Пример: GPL и проприетарные лицензии
	gplLicenses := make(map[string]bool)
	proprietaryLicenses := make(map[string]bool)

	for _, license := range licenses {
		if strings.Contains(strings.ToLower(license.Name), "gpl") {
			gplLicenses[license.Name] = true
		}
		if strings.Contains(strings.ToLower(license.Type), "proprietary") {
			proprietaryLicenses[license.Name] = true
		}
	}

	// Проверяем конфликты
	if len(gplLicenses) > 0 && len(proprietaryLicenses) > 0 {
		for gplLicense := range gplLicenses {
			for propLicense := range proprietaryLicenses {
				conflicts = append(conflicts, &domain.LicenseConflict{
					License1:    gplLicense,
					License2:    propLicense,
					Description: "GPL license conflicts with proprietary license",
					Severity:    "high",
				})
			}
		}
	}

	return conflicts
}

// ValidateLicense проверяет валидность лицензии
func (l *LicenseScanner) ValidateLicense(licenseName string) (bool, string) {
	// Список известных лицензий
	knownLicenses := map[string]bool{
		"MIT":          true,
		"Apache-2.0":   true,
		"GPL-3.0":      true,
		"GPL-2.0":      true,
		"LGPL-3.0":     true,
		"BSD-3-Clause": true,
		"BSD-2-Clause": true,
		"ISC":          true,
		"MPL-2.0":      true,
		"CC0-1.0":      true,
		"Unlicense":    true,
	}

	if knownLicenses[licenseName] {
		return true, "Valid license"
	}

	return false, "Unknown or invalid license"
}

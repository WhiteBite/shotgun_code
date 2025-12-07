package sbomlicensing

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"shotgun_code/domain"
)

// SyftGenerator представляет генератор SBOM с использованием Syft
type SyftGenerator struct {
	log domain.Logger
}

// NewSyftGenerator создает новый генератор Syft
func NewSyftGenerator(log domain.Logger) *SyftGenerator {
	return &SyftGenerator{
		log: log,
	}
}

// IsAvailable проверяет доступность Syft
func (s *SyftGenerator) IsAvailable() bool {
	_, err := exec.LookPath("syft")
	return err == nil
}

// GenerateSBOM генерирует SBOM для проекта
func (s *SyftGenerator) GenerateSBOM(ctx context.Context, projectPath string, format domain.SBOMFormat) (*domain.SBOMResult, error) {
	if !s.IsAvailable() {
		return &domain.SBOMResult{
			Success:     false,
			ProjectPath: projectPath,
			Format:      format,
			Error:       "syft is not available",
		}, nil
	}

	s.log.Info(fmt.Sprintf("Generating SBOM with Syft for: %s, format: %s", projectPath, format))

	// Определяем расширение файла
	var fileExt string
	switch format {
	case domain.SBOMFormatSPDX:
		fileExt = "spdx"
	case domain.SBOMFormatCycloneDX:
		fileExt = "cyclonedx"
	case domain.SBOMFormatJSON:
		fileExt = "json"
	default:
		fileExt = "json"
	}

	outputPath := filepath.Join(projectPath, fmt.Sprintf("sbom.%s", fileExt))

	// Запускаем Syft
	cmd := exec.CommandContext(ctx, "syft", projectPath, "-o", string(format), "-f", outputPath) //nolint:gosec // External tool command
	cmd.Dir = projectPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return &domain.SBOMResult{
			Success:     false,
			ProjectPath: projectPath,
			Format:      format,
			Error:       fmt.Sprintf("syft failed: %v, output: %s", err, string(output)),
		}, nil
	}

	// Парсим результат
	components, err := s.parseSyftOutput(output, format)
	if err != nil {
		return &domain.SBOMResult{
			Success:     false,
			ProjectPath: projectPath,
			Format:      format,
			Error:       fmt.Sprintf("failed to parse syft output: %v", err),
		}, nil
	}

	return &domain.SBOMResult{
		Success:     true,
		ProjectPath: projectPath,
		Format:      format,
		OutputPath:  outputPath,
		Components:  components,
		Metadata: map[string]interface{}{
			"tool":    "syft",
			"version": s.getVersion(),
			"command": cmd.String(),
		},
	}, nil
}

// ValidateSBOM валидирует SBOM файл
func (s *SyftGenerator) ValidateSBOM(ctx context.Context, sbomPath string, format domain.SBOMFormat) error {
	if !s.IsAvailable() {
		return fmt.Errorf("syft is not available")
	}

	s.log.Info(fmt.Sprintf("Validating SBOM with Syft: %s", sbomPath))

	// Запускаем Syft для валидации
	cmd := exec.CommandContext(ctx, "syft", "validate", sbomPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("syft validation failed: %w, output: %s", err, string(output))
	}

	return nil
}

// GetSupportedFormats возвращает поддерживаемые форматы
func (s *SyftGenerator) GetSupportedFormats() []domain.SBOMFormat {
	return []domain.SBOMFormat{
		domain.SBOMFormatSPDX,
		domain.SBOMFormatCycloneDX,
		domain.SBOMFormatJSON,
	}
}

// getVersion возвращает версию Syft
func (s *SyftGenerator) getVersion() string {
	cmd := exec.Command("syft", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return string(output)
}

// parseSyftOutput парсит вывод Syft
func (s *SyftGenerator) parseSyftOutput(output []byte, format domain.SBOMFormat) ([]*domain.SBOMComponent, error) {
	// Упрощенная реализация - в реальном приложении здесь будет парсинг JSON/XML
	// в зависимости от формата
	s.log.Info(fmt.Sprintf("Parsing Syft output, format: %s, size: %d bytes", format, len(output)))

	// Возвращаем пустой список компонентов
	// TODO: Реализовать парсинг в зависимости от формата
	return []*domain.SBOMComponent{}, nil
}

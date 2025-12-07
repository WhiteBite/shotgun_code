package application

import (
	"encoding/json"
	"fmt"
	"shotgun_code/domain"
	"shotgun_code/infrastructure/projectstructure"
	"strings"
)

// ProjectStructureService provides project structure analysis
type ProjectStructureService struct {
	detector *projectstructure.Detector
	logger   domain.Logger
}

// NewProjectStructureService creates a new ProjectStructureService
func NewProjectStructureService(logger domain.Logger) *ProjectStructureService {
	return &ProjectStructureService{
		detector: projectstructure.NewDetector(),
		logger:   logger,
	}
}

// DetectStructure analyzes project and returns complete structure info
func (s *ProjectStructureService) DetectStructure(projectPath string) (*domain.ProjectStructure, error) {
	s.logger.Info(fmt.Sprintf("Detecting project structure for: %s", projectPath))
	return s.detector.DetectStructure(projectPath)
}

// DetectArchitecture detects architecture pattern
func (s *ProjectStructureService) DetectArchitecture(projectPath string) (*domain.ArchitectureInfo, error) {
	return s.detector.DetectArchitecture(projectPath)
}

// DetectConventions detects naming and code conventions
func (s *ProjectStructureService) DetectConventions(projectPath string) (*domain.ConventionInfo, error) {
	return s.detector.DetectConventions(projectPath)
}

// DetectFrameworks detects frameworks used
func (s *ProjectStructureService) DetectFrameworks(projectPath string) ([]domain.FrameworkInfo, error) {
	return s.detector.DetectFrameworks(projectPath)
}

// GetRelatedLayers returns layers related to a file
func (s *ProjectStructureService) GetRelatedLayers(projectPath, filePath string) ([]domain.LayerInfo, error) {
	return s.detector.GetRelatedLayers(projectPath, filePath)
}

// SuggestRelatedFiles suggests related files based on architecture
func (s *ProjectStructureService) SuggestRelatedFiles(projectPath, filePath string) ([]string, error) {
	return s.detector.SuggestRelatedFiles(projectPath, filePath)
}

// GetStructureJSON returns project structure as JSON string
func (s *ProjectStructureService) GetStructureJSON(projectPath string) (string, error) {
	structure, err := s.DetectStructure(projectPath)
	if err != nil {
		return "", err
	}

	jsonBytes, err := json.MarshalIndent(structure, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// formatArchitectureSummary formats architecture info into summary string
func formatArchitectureSummary(sb *strings.Builder, arch *domain.ArchitectureInfo) {
	sb.WriteString(fmt.Sprintf("Architecture: %s (%.0f%% confidence)\n", arch.Type, arch.Confidence*100))
	sb.WriteString(fmt.Sprintf("Description: %s\n", arch.Description))

	if len(arch.Indicators) > 0 {
		sb.WriteString("Indicators:\n")
		for _, ind := range arch.Indicators {
			sb.WriteString(fmt.Sprintf("  - %s\n", ind))
		}
	}

	if len(arch.Layers) > 0 {
		sb.WriteString("\nArchitectural Layers:\n")
		for _, layer := range arch.Layers {
			sb.WriteString(fmt.Sprintf("  [%s] %s\n", layer.Name, layer.Path))
			if layer.Description != "" {
				sb.WriteString(fmt.Sprintf("    %s\n", layer.Description))
			}
			if len(layer.Dependencies) > 0 {
				sb.WriteString(fmt.Sprintf("    Dependencies: %v\n", layer.Dependencies))
			}
		}
	}
	sb.WriteString("\n")
}

// formatLanguagesSummary formats languages info into summary string
func formatLanguagesSummary(sb *strings.Builder, languages []domain.LanguageInfo) {
	sb.WriteString("Languages:\n")
	for _, lang := range languages {
		primary := ""
		if lang.Primary {
			primary = " (primary)"
		}
		sb.WriteString(fmt.Sprintf("  - %s: %d files (%.1f%%)%s\n", lang.Name, lang.FileCount, lang.Percentage, primary))
	}
	sb.WriteString("\n")
}

// formatFrameworksSummary formats frameworks info into summary string
func formatFrameworksSummary(sb *strings.Builder, frameworks []domain.FrameworkInfo) {
	sb.WriteString("Frameworks:\n")
	for _, fw := range frameworks {
		version := ""
		if fw.Version != "" {
			version = " v" + fw.Version
		}
		sb.WriteString(fmt.Sprintf("  - %s%s (%s, %s)\n", fw.Name, version, fw.Category, fw.Language))
		if len(fw.BestPractices) > 0 {
			sb.WriteString("    Best Practices:\n")
			for _, bp := range fw.BestPractices[:min(3, len(fw.BestPractices))] {
				sb.WriteString(fmt.Sprintf("      • %s\n", bp))
			}
		}
	}
	sb.WriteString("\n")
}

// formatBuildSystemsSummary formats build systems info into summary string
func formatBuildSystemsSummary(sb *strings.Builder, buildSystems []domain.BuildSystemInfo) {
	sb.WriteString("Build Systems:\n")
	for _, bs := range buildSystems {
		sb.WriteString(fmt.Sprintf("  - %s (%s)\n", bs.Name, bs.ConfigFile))
		if len(bs.Scripts) > 0 {
			sb.WriteString(fmt.Sprintf("    Scripts: %v\n", bs.Scripts[:min(5, len(bs.Scripts))]))
		}
	}
	sb.WriteString("\n")
}

// formatConventionsSummary formats conventions info into summary string
func formatConventionsSummary(sb *strings.Builder, conv *domain.ConventionInfo) {
	sb.WriteString("Conventions:\n")
	sb.WriteString(fmt.Sprintf("  Naming Style: %s\n", conv.NamingStyle))
	sb.WriteString(fmt.Sprintf("  Folder Structure: %s\n", conv.FolderStructure))
	if conv.TestConventions.Framework != "" {
		sb.WriteString(fmt.Sprintf("  Test Framework: %s\n", conv.TestConventions.Framework))
		sb.WriteString(fmt.Sprintf("  Test Location: %s\n", conv.TestConventions.Location))
	}
	if conv.CodeStyle.ConfigFile != "" {
		sb.WriteString(fmt.Sprintf("  Code Style Config: %s\n", conv.CodeStyle.ConfigFile))
	}
}

// GetArchitectureSummary returns a human-readable architecture summary
func (s *ProjectStructureService) GetArchitectureSummary(projectPath string) (string, error) {
	structure, err := s.DetectStructure(projectPath)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.Grow(2048) // Pre-allocate for performance

	sb.WriteString("Project Structure Analysis\n==========================\n\n")
	sb.WriteString(fmt.Sprintf("Project Type: %s\nConfidence: %.0f%%\n\n", structure.ProjectType, structure.Confidence*100))

	if structure.Architecture != nil {
		formatArchitectureSummary(&sb, structure.Architecture)
	}
	if len(structure.Languages) > 0 {
		formatLanguagesSummary(&sb, structure.Languages)
	}
	if len(structure.Frameworks) > 0 {
		formatFrameworksSummary(&sb, structure.Frameworks)
	}
	if len(structure.BuildSystems) > 0 {
		formatBuildSystemsSummary(&sb, structure.BuildSystems)
	}
	if structure.Conventions != nil {
		formatConventionsSummary(&sb, structure.Conventions)
	}

	return sb.String(), nil
}

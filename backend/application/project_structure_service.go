package application

import (
	"encoding/json"
	"fmt"
	"shotgun_code/domain"
	"shotgun_code/infrastructure/projectstructure"
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

// GetArchitectureSummary returns a human-readable architecture summary
func (s *ProjectStructureService) GetArchitectureSummary(projectPath string) (string, error) {
	structure, err := s.DetectStructure(projectPath)
	if err != nil {
		return "", err
	}

	summary := fmt.Sprintf("Project Structure Analysis\n")
	summary += fmt.Sprintf("==========================\n\n")

	// Project type
	summary += fmt.Sprintf("Project Type: %s\n", structure.ProjectType)
	summary += fmt.Sprintf("Confidence: %.0f%%\n\n", structure.Confidence*100)

	// Architecture
	if structure.Architecture != nil {
		summary += fmt.Sprintf("Architecture: %s (%.0f%% confidence)\n", 
			structure.Architecture.Type, structure.Architecture.Confidence*100)
		summary += fmt.Sprintf("Description: %s\n", structure.Architecture.Description)
		
		if len(structure.Architecture.Indicators) > 0 {
			summary += "Indicators:\n"
			for _, ind := range structure.Architecture.Indicators {
				summary += fmt.Sprintf("  - %s\n", ind)
			}
		}

		if len(structure.Architecture.Layers) > 0 {
			summary += "\nArchitectural Layers:\n"
			for _, layer := range structure.Architecture.Layers {
				summary += fmt.Sprintf("  [%s] %s\n", layer.Name, layer.Path)
				if layer.Description != "" {
					summary += fmt.Sprintf("    %s\n", layer.Description)
				}
				if len(layer.Dependencies) > 0 {
					summary += fmt.Sprintf("    Dependencies: %v\n", layer.Dependencies)
				}
			}
		}
		summary += "\n"
	}

	// Languages
	if len(structure.Languages) > 0 {
		summary += "Languages:\n"
		for _, lang := range structure.Languages {
			primary := ""
			if lang.Primary {
				primary = " (primary)"
			}
			summary += fmt.Sprintf("  - %s: %d files (%.1f%%)%s\n", 
				lang.Name, lang.FileCount, lang.Percentage, primary)
		}
		summary += "\n"
	}

	// Frameworks
	if len(structure.Frameworks) > 0 {
		summary += "Frameworks:\n"
		for _, fw := range structure.Frameworks {
			version := ""
			if fw.Version != "" {
				version = " v" + fw.Version
			}
			summary += fmt.Sprintf("  - %s%s (%s, %s)\n", fw.Name, version, fw.Category, fw.Language)
			if len(fw.BestPractices) > 0 {
				summary += "    Best Practices:\n"
				for _, bp := range fw.BestPractices[:min(3, len(fw.BestPractices))] {
					summary += fmt.Sprintf("      • %s\n", bp)
				}
			}
		}
		summary += "\n"
	}

	// Build Systems
	if len(structure.BuildSystems) > 0 {
		summary += "Build Systems:\n"
		for _, bs := range structure.BuildSystems {
			summary += fmt.Sprintf("  - %s (%s)\n", bs.Name, bs.ConfigFile)
			if len(bs.Scripts) > 0 {
				summary += fmt.Sprintf("    Scripts: %v\n", bs.Scripts[:min(5, len(bs.Scripts))])
			}
		}
		summary += "\n"
	}

	// Conventions
	if structure.Conventions != nil {
		summary += "Conventions:\n"
		summary += fmt.Sprintf("  Naming Style: %s\n", structure.Conventions.NamingStyle)
		summary += fmt.Sprintf("  Folder Structure: %s\n", structure.Conventions.FolderStructure)
		if structure.Conventions.TestConventions.Framework != "" {
			summary += fmt.Sprintf("  Test Framework: %s\n", structure.Conventions.TestConventions.Framework)
			summary += fmt.Sprintf("  Test Location: %s\n", structure.Conventions.TestConventions.Location)
		}
		if structure.Conventions.CodeStyle.ConfigFile != "" {
			summary += fmt.Sprintf("  Code Style Config: %s\n", structure.Conventions.CodeStyle.ConfigFile)
		}
	}

	return summary, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

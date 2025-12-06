package projectstructure

import (
	"os"
	"path/filepath"
	"shotgun_code/domain"
	"strings"
)

type architectureDetector interface {
	detect(projectPath string) *domain.ArchitectureInfo
}

func (d *Detector) initArchitectureDetectors() {
	d.archDetectors = []architectureDetector{
		&cleanArchDetector{},
		&hexagonalDetector{},
		&mvcDetector{},
		&mvvmDetector{},
		&layeredDetector{},
		&dddDetector{},
	}
}

// Clean Architecture detector
type cleanArchDetector struct{}

func (c *cleanArchDetector) detect(projectPath string) *domain.ArchitectureInfo {
	indicators := []string{}
	score := 0.0

	// Check for Clean Architecture folder structure
	cleanDirs := map[string]float64{
		"domain":         0.2,
		"application":    0.2,
		"infrastructure": 0.2,
		"interfaces":     0.1,
		"usecases":       0.15,
		"use_cases":      0.15,
		"use-cases":      0.15,
		"entities":       0.15,
	}

	entries, err := os.ReadDir(projectPath)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := strings.ToLower(entry.Name())
		if weight, ok := cleanDirs[name]; ok {
			score += weight
			indicators = append(indicators, "Found "+entry.Name()+" directory")
		}
	}

	// Check for nested structure (e.g., backend/domain, backend/application)
	backendPath := filepath.Join(projectPath, "backend")
	if entries, err := os.ReadDir(backendPath); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			name := strings.ToLower(entry.Name())
			if weight, ok := cleanDirs[name]; ok {
				score += weight
				indicators = append(indicators, "Found backend/"+entry.Name()+" directory")
			}
		}
	}

	// Check for internal structure (Go projects)
	internalPath := filepath.Join(projectPath, "internal")
	if entries, err := os.ReadDir(internalPath); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			name := strings.ToLower(entry.Name())
			if weight, ok := cleanDirs[name]; ok {
				score += weight * 0.8
				indicators = append(indicators, "Found internal/"+entry.Name()+" directory")
			}
		}
	}

	if score < 0.3 {
		return nil
	}

	// Cap at 0.95
	if score > 0.95 {
		score = 0.95
	}

	return &domain.ArchitectureInfo{
		Type:        domain.ArchCleanArchitecture,
		Confidence:  score,
		Description: "Clean Architecture with separation of domain, application, and infrastructure layers",
		Indicators:  indicators,
	}
}

// Hexagonal Architecture detector
type hexagonalDetector struct{}

func (h *hexagonalDetector) detect(projectPath string) *domain.ArchitectureInfo {
	indicators := []string{}
	score := 0.0

	// Check for Hexagonal/Ports & Adapters structure
	hexDirs := map[string]float64{
		"ports":    0.25,
		"adapters": 0.25,
		"core":     0.2,
		"inbound":  0.15,
		"outbound": 0.15,
		"driven":   0.15,
		"driving":  0.15,
	}

	entries, err := os.ReadDir(projectPath)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := strings.ToLower(entry.Name())
		if weight, ok := hexDirs[name]; ok {
			score += weight
			indicators = append(indicators, "Found "+entry.Name()+" directory")
		}
	}

	// Check in src directory
	srcPath := filepath.Join(projectPath, "src")
	if entries, err := os.ReadDir(srcPath); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			name := strings.ToLower(entry.Name())
			if weight, ok := hexDirs[name]; ok {
				score += weight
				indicators = append(indicators, "Found src/"+entry.Name()+" directory")
			}
		}
	}

	if score < 0.3 {
		return nil
	}

	if score > 0.95 {
		score = 0.95
	}

	return &domain.ArchitectureInfo{
		Type:        domain.ArchHexagonal,
		Confidence:  score,
		Description: "Hexagonal Architecture (Ports & Adapters) with clear separation of core logic and external interfaces",
		Indicators:  indicators,
	}
}

// MVC detector
type mvcDetector struct{}

func (m *mvcDetector) detect(projectPath string) *domain.ArchitectureInfo {
	indicators := []string{}
	score := 0.0

	// Check for MVC structure
	mvcDirs := map[string]float64{
		"models":      0.3,
		"views":       0.3,
		"controllers": 0.3,
		"model":       0.25,
		"view":        0.25,
		"controller":  0.25,
	}

	searchPaths := []string{projectPath, filepath.Join(projectPath, "app"), filepath.Join(projectPath, "src")}

	for _, searchPath := range searchPaths {
		entries, err := os.ReadDir(searchPath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			name := strings.ToLower(entry.Name())
			if weight, ok := mvcDirs[name]; ok {
				score += weight
				rel, _ := filepath.Rel(projectPath, filepath.Join(searchPath, entry.Name()))
				indicators = append(indicators, "Found "+rel+" directory")
			}
		}
	}

	// Check for Rails-style structure
	if _, err := os.Stat(filepath.Join(projectPath, "app", "models")); err == nil {
		if _, err := os.Stat(filepath.Join(projectPath, "app", "controllers")); err == nil {
			score += 0.3
			indicators = append(indicators, "Rails-style app/models and app/controllers structure")
		}
	}

	if score < 0.5 {
		return nil
	}

	if score > 0.95 {
		score = 0.95
	}

	return &domain.ArchitectureInfo{
		Type:        domain.ArchMVC,
		Confidence:  score,
		Description: "Model-View-Controller architecture with separation of data, presentation, and control logic",
		Indicators:  indicators,
	}
}

// MVVM detector
type mvvmDetector struct{}

func (m *mvvmDetector) detect(projectPath string) *domain.ArchitectureInfo {
	indicators := []string{}
	score := 0.0

	// Check for MVVM structure
	mvvmDirs := map[string]float64{
		"viewmodels":  0.35,
		"viewmodel":   0.35,
		"view-models": 0.35,
		"models":      0.2,
		"views":       0.2,
	}

	searchPaths := []string{projectPath, filepath.Join(projectPath, "src"), filepath.Join(projectPath, "lib")}

	for _, searchPath := range searchPaths {
		entries, err := os.ReadDir(searchPath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			name := strings.ToLower(entry.Name())
			if weight, ok := mvvmDirs[name]; ok {
				score += weight
				rel, _ := filepath.Rel(projectPath, filepath.Join(searchPath, entry.Name()))
				indicators = append(indicators, "Found "+rel+" directory")
			}
		}
	}

	// Check for ViewModel files
	viewModelCount := 0
	filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.Contains(path, "node_modules") || strings.Contains(path, ".git") {
			return filepath.SkipDir
		}
		name := strings.ToLower(info.Name())
		if strings.Contains(name, "viewmodel") || strings.Contains(name, "view_model") {
			viewModelCount++
		}
		return nil
	})

	if viewModelCount > 3 {
		score += 0.2
		indicators = append(indicators, "Found multiple ViewModel files")
	}

	if score < 0.4 {
		return nil
	}

	if score > 0.95 {
		score = 0.95
	}

	return &domain.ArchitectureInfo{
		Type:        domain.ArchMVVM,
		Confidence:  score,
		Description: "Model-View-ViewModel architecture with data binding between View and ViewModel",
		Indicators:  indicators,
	}
}

// Layered Architecture detector
type layeredDetector struct{}

func (l *layeredDetector) detect(projectPath string) *domain.ArchitectureInfo {
	indicators := []string{}
	score := 0.0

	// Check for layered structure
	layeredDirs := map[string]float64{
		"presentation": 0.2,
		"business":     0.2,
		"data":         0.2,
		"services":     0.15,
		"handlers":     0.15,
		"repositories": 0.15,
		"repository":   0.15,
		"api":          0.1,
		"dal":          0.15, // Data Access Layer
		"bll":          0.15, // Business Logic Layer
	}

	entries, err := os.ReadDir(projectPath)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := strings.ToLower(entry.Name())
		if weight, ok := layeredDirs[name]; ok {
			score += weight
			indicators = append(indicators, "Found "+entry.Name()+" directory")
		}
	}

	// Check in backend/src directories
	for _, subdir := range []string{"backend", "src", "app"} {
		subPath := filepath.Join(projectPath, subdir)
		if entries, err := os.ReadDir(subPath); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				name := strings.ToLower(entry.Name())
				if weight, ok := layeredDirs[name]; ok {
					score += weight
					indicators = append(indicators, "Found "+subdir+"/"+entry.Name()+" directory")
				}
			}
		}
	}

	if score < 0.3 {
		return nil
	}

	if score > 0.95 {
		score = 0.95
	}

	return &domain.ArchitectureInfo{
		Type:        domain.ArchLayered,
		Confidence:  score,
		Description: "Layered architecture with separation of presentation, business, and data access layers",
		Indicators:  indicators,
	}
}

// DDD (Domain-Driven Design) detector
type dddDetector struct{}

func (d *dddDetector) detect(projectPath string) *domain.ArchitectureInfo {
	indicators := []string{}
	score := 0.0

	// Check for DDD structure
	dddDirs := map[string]float64{
		"aggregates":    0.25,
		"entities":      0.2,
		"value_objects": 0.2,
		"valueobjects":  0.2,
		"value-objects": 0.2,
		"repositories":  0.15,
		"services":      0.1,
		"events":        0.15,
		"commands":      0.15,
		"queries":       0.15,
	}

	// Search in domain directory first
	domainPath := filepath.Join(projectPath, "domain")
	if entries, err := os.ReadDir(domainPath); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			name := strings.ToLower(entry.Name())
			if weight, ok := dddDirs[name]; ok {
				score += weight
				indicators = append(indicators, "Found domain/"+entry.Name()+" directory")
			}
		}
	}

	// Check root level
	entries, err := os.ReadDir(projectPath)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := strings.ToLower(entry.Name())
		if weight, ok := dddDirs[name]; ok {
			score += weight * 0.8
			indicators = append(indicators, "Found "+entry.Name()+" directory")
		}
	}

	// Check for CQRS pattern (commands/queries separation)
	hasCommands := false
	hasQueries := false
	filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if strings.Contains(path, "node_modules") || strings.Contains(path, ".git") {
			return filepath.SkipDir
		}
		name := strings.ToLower(info.Name())
		if info.IsDir() {
			if name == "commands" {
				hasCommands = true
			}
			if name == "queries" {
				hasQueries = true
			}
		}
		return nil
	})

	if hasCommands && hasQueries {
		score += 0.2
		indicators = append(indicators, "CQRS pattern detected (commands/queries separation)")
	}

	if score < 0.3 {
		return nil
	}

	if score > 0.95 {
		score = 0.95
	}

	return &domain.ArchitectureInfo{
		Type:        domain.ArchDDD,
		Confidence:  score,
		Description: "Domain-Driven Design with rich domain model, aggregates, and bounded contexts",
		Indicators:  indicators,
	}
}

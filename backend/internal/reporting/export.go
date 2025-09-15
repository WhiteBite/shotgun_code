package reporting

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"shotgun_code/domain"
	"shotgun_code/infrastructure/contextbuilder"
	"strings"
)

const (
	maxInMemorySize = 50 * 1024 * 1024 // 50MB
	maxFileSize     = 5 * 1024 * 1024  // 5 MB limit per file
)

// ExportService handles context export operations
type ExportService struct {
	contextSplitter domain.ContextSplitter
	log             domain.Logger
	pdf             domain.PDFGenerator
	archiver        domain.Archiver
}

// NewExportService creates a new export service
func NewExportService(log domain.Logger, splitter domain.ContextSplitter, pdf domain.PDFGenerator, arch domain.Archiver) *ExportService {
	return &ExportService{
		contextSplitter: splitter,
		log:             log,
		pdf:             pdf,
		archiver:        arch,
	}
}

// Export performs context export based on settings
func (s *ExportService) Export(_ context.Context, settings domain.ExportSettings) (domain.ExportResult, error) {
	if settings.Context == "" {
		s.log.Warning("Attempted to export empty context.")
		return domain.ExportResult{}, fmt.Errorf("context is empty, nothing to export")
	}

	switch settings.Mode {
	case domain.ExportModeClipboard:
		return s.exportToClipboard(settings)
	case domain.ExportModeAI:
		return s.exportForAI(settings)
	case domain.ExportModeHuman:
		return s.exportForHuman(settings)
	default:
		return domain.ExportResult{}, fmt.Errorf("unknown export mode: %s", settings.Mode)
	}
}

// exportToClipboard exports context to clipboard format
func (s *ExportService) exportToClipboard(settings domain.ExportSettings) (domain.ExportResult, error) {
	format := settings.ExportFormat
	if format == "" {
		format = "manifest"
	}
	
	out, err := contextbuilder.BuildFromContext(format, settings.Context, contextbuilder.BuildOptions{
		StripComments:   settings.StripComments,
		IncludeManifest: settings.IncludeManifest,
	})
	if err != nil {
		return domain.ExportResult{}, fmt.Errorf("failed to build clipboard context: %w", err)
	}
	
	return domain.ExportResult{Mode: settings.Mode, Text: out}, nil
}

// exportForAI exports context optimized for AI consumption
func (s *ExportService) exportForAI(settings domain.ExportSettings) (domain.ExportResult, error) {
	var chunks []string
	
	if settings.EnableAutoSplit {
		splitSettings := domain.SplitSettings{
			MaxTokensPerChunk: settings.MaxTokensPerChunk,
			OverlapTokens:     settings.OverlapTokens,
			SplitStrategy:     settings.SplitStrategy,
		}
		var err error
		chunks, err = s.contextSplitter.SplitContext(settings.Context, splitSettings)
		if err != nil {
			return domain.ExportResult{}, fmt.Errorf("failed to split context for AI export: %w", err)
		}
	} else {
		totalTokens := s.approxTokens(settings.Context)
		if totalTokens > settings.TokenLimit && settings.TokenLimit > 0 {
			s.log.Warning(fmt.Sprintf("Context (%d tokens) exceeds specified token limit (%d) for AI export, but auto-split is disabled. Exporting as single large PDF.", totalTokens, settings.TokenLimit))
		}
		chunks = []string{settings.Context}
	}

	estimatedSize := int64(len(settings.Context) * 2)

	if len(chunks) == 1 && estimatedSize < maxInMemorySize {
		// Small PDF in memory
		pdfBytes, err := s.pdf.Generate(chunks[0], domain.PDFOptions{})
		if err != nil {
			return domain.ExportResult{}, fmt.Errorf("failed to generate AI PDF: %w", err)
		}
		return domain.ExportResult{
			Mode:       settings.Mode,
			FileName:   "context-ai.pdf",
			DataBase64: base64.StdEncoding.EncodeToString(pdfBytes),
			SizeBytes:  int64(len(pdfBytes)),
		}, nil
	}

	// Large PDF or multiple chunks -> file
	return s.createLargePDFExport(chunks, "ai")
}

// exportForHuman exports context optimized for human reading
func (s *ExportService) exportForHuman(settings domain.ExportSettings) (domain.ExportResult, error) {
	dark := strings.EqualFold(settings.Theme, "Dark")
	opts := domain.PDFOptions{
		Dark:        dark,
		LineNumbers: settings.IncludeLineNumbers,
		PageNumbers: settings.IncludePageNumbers,
	}
	
	estimatedSize := int64(len(settings.Context) * 2)

	if estimatedSize < maxInMemorySize {
		pdfBytes, err := s.pdf.Generate(settings.Context, opts)
		if err != nil {
			return domain.ExportResult{}, fmt.Errorf("failed to generate human-readable PDF: %w", err)
		}
		return domain.ExportResult{
			Mode:       settings.Mode,
			FileName:   "context-human.pdf",
			DataBase64: base64.StdEncoding.EncodeToString(pdfBytes),
			SizeBytes:  int64(len(pdfBytes)),
		}, nil
	}

	tempDir, err := os.MkdirTemp("", "shotgun-export-*")
	if err != nil {
		return domain.ExportResult{}, fmt.Errorf("failed to create temp dir: %w", err)
	}
	
	fileName := "context-human.pdf"
	outputPath := filepath.Join(tempDir, fileName)

	if err := s.pdf.WriteAtomic(settings.Context, opts, outputPath); err != nil {
		os.RemoveAll(tempDir)
		return domain.ExportResult{}, fmt.Errorf("failed to generate human-readable PDF: %w", err)
	}

	fi, err := os.Stat(outputPath)
	if err != nil {
		os.RemoveAll(tempDir)
		return domain.ExportResult{}, fmt.Errorf("failed to stat output file: %w", err)
	}

	return domain.ExportResult{
		Mode:      settings.Mode,
		FileName:  fileName,
		FilePath:  outputPath,
		IsLarge:   true,
		SizeBytes: fi.Size(),
	}, nil
}

// createLargePDFExport creates large PDF export with multiple chunks
func (s *ExportService) createLargePDFExport(chunks []string, prefix string) (domain.ExportResult, error) {
	tempDir, err := os.MkdirTemp("", "shotgun-export-*")
	if err != nil {
		return domain.ExportResult{}, fmt.Errorf("failed to create temp dir: %w", err)
	}

	var outputPath string
	var fileName string

	if len(chunks) == 1 {
		fileName = fmt.Sprintf("context-%s.pdf", prefix)
		outputPath = filepath.Join(tempDir, fileName)
		if err := s.pdf.WriteAtomic(chunks[0], domain.PDFOptions{}, outputPath); err != nil {
			os.RemoveAll(tempDir)
			return domain.ExportResult{}, fmt.Errorf("failed to generate %s PDF: %w", prefix, err)
		}
	} else {
		files := make(map[string][]byte, len(chunks))
		for i, chunk := range chunks {
			b, err := s.pdf.Generate(chunk, domain.PDFOptions{})
			if err != nil {
				os.RemoveAll(tempDir)
				return domain.ExportResult{}, fmt.Errorf("failed to generate PDF chunk %d: %w", i+1, err)
			}
			files[fmt.Sprintf("context-%s-part-%02d.pdf", prefix, i+1)] = b
		}
		fileName = fmt.Sprintf("context-%s.zip", prefix)
		outputPath = filepath.Join(tempDir, fileName)
		if err := s.archiver.ZipFilesAtomic(files, outputPath); err != nil {
			os.RemoveAll(tempDir)
			return domain.ExportResult{}, fmt.Errorf("failed to create ZIP: %w", err)
		}
	}

	fi, err := os.Stat(outputPath)
	if err != nil {
		os.RemoveAll(tempDir)
		return domain.ExportResult{}, fmt.Errorf("failed to stat output file: %w", err)
	}
	
	return domain.ExportResult{
		Mode:      domain.ExportModeAI,
		FileName:  fileName,
		FilePath:  outputPath,
		IsLarge:   true,
		SizeBytes: fi.Size(),
	}, nil
}

// approxTokens provides rough token estimation
func (s *ExportService) approxTokens(text string) int {
	return len([]rune(text)) / 4
}
package application

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"regexp"
	"shotgun_code/domain"
	"shotgun_code/infrastructure/contextbuilder"
	"strings"
)

const (
	maxInMemorySize = 50 * 1024 * 1024 // 50MB
)

type ExportService struct {
	contextSplitter  domain.ContextSplitter
	log              domain.Logger
	pdf              domain.PDFGenerator
	archiver         domain.Archiver
	tempFileProvider domain.TempFileProvider
	pathProvider     domain.PathProvider
	fileSystemWriter domain.FileSystemWriter
	fileStatProvider domain.FileStatProvider
}

func NewExportService(log domain.Logger, splitter domain.ContextSplitter, pdf domain.PDFGenerator, arch domain.Archiver, tempFileProvider domain.TempFileProvider, pathProvider domain.PathProvider, fileSystemWriter domain.FileSystemWriter, fileStatProvider domain.FileStatProvider) *ExportService {
	return &ExportService{
		contextSplitter:  splitter,
		log:              log,
		pdf:              pdf,
		archiver:         arch,
		tempFileProvider: tempFileProvider,
		pathProvider:     pathProvider,
		fileSystemWriter: fileSystemWriter,
		fileStatProvider: fileStatProvider,
	}
}

// Грубая оценка числа токенов (~ четверть от количества рун)
func approxTokens(s string) int { return len([]rune(s)) / 4 }

// Разделение по заголовкам --- File: ... ---
func splitByFiles(text string) []string {
	re := regexp.MustCompile(`(?m)^--- File: .*? ---\s*`)
	idxs := re.FindAllStringIndex(text, -1)
	if len(idxs) == 0 {
		return []string{text}
	}
	var parts []string
	for i := 0; i < len(idxs); i++ {
		start := idxs[i][0]
		end := len(text)
		if i+1 < len(idxs) {
			end = idxs[i+1][0]
		}
		parts = append(parts, text[start:end])
	}
	return parts
}

func (s *ExportService) Export(_ context.Context, settings domain.ExportSettings) (domain.ExportResult, error) {
	if settings.Context == "" {
		s.log.Warning("Attempted to export empty context.")
		return domain.ExportResult{}, fmt.Errorf("context is empty, nothing to export")
	}

	switch settings.Mode {
	case domain.ExportModeClipboard:
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

	case domain.ExportModeAI:
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
			totalTokens := approxTokens(settings.Context)
			if totalTokens > settings.TokenLimit && settings.TokenLimit > 0 {
				s.log.Warning(fmt.Sprintf("Context (%d tokens) exceeds specified token limit (%d) for AI export, but auto-split is disabled. Exporting as single large PDF.", totalTokens, settings.TokenLimit))
			}
			chunks = []string{settings.Context}
		}

		estimatedSize := int64(len(settings.Context) * 2)

		if len(chunks) == 1 && estimatedSize < maxInMemorySize {
			// Маленький PDF в память
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

		// Большой PDF или много чанков -> файл
		tempDir, err := s.tempFileProvider.MkdirTemp("", "shotgun-export-*")
		if err != nil {
			return domain.ExportResult{}, fmt.Errorf("failed to create temp dir: %w", err)
		}

		var outputPath string
		var fileName string

		if len(chunks) == 1 {
			fileName = "context-ai.pdf"
			outputPath = s.pathProvider.Join(tempDir, fileName)
			if err := s.pdf.WriteAtomic(chunks[0], domain.PDFOptions{}, outputPath); err != nil {
				s.fileSystemWriter.RemoveAll(tempDir)
				return domain.ExportResult{}, fmt.Errorf("failed to generate AI PDF: %w", err)
			}
		} else {
			files := make(map[string][]byte, len(chunks))
			for i, chunk := range chunks {
				b, err := s.pdf.Generate(chunk, domain.PDFOptions{})
				if err != nil {
					s.fileSystemWriter.RemoveAll(tempDir)
					return domain.ExportResult{}, fmt.Errorf("failed to generate PDF chunk %d: %w", i+1, err)
				}
				files[fmt.Sprintf("context-ai-part-%02d.pdf", i+1)] = b
			}
			fileName = "context-ai.zip"
			outputPath = s.pathProvider.Join(tempDir, fileName)
			if err := s.archiver.ZipFilesAtomic(files, outputPath); err != nil {
				s.fileSystemWriter.RemoveAll(tempDir)
				return domain.ExportResult{}, fmt.Errorf("failed to create ZIP: %w", err)
			}
		}

		fi, err := s.fileStatProvider.Stat(outputPath)
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

	case domain.ExportModeHuman:
		dark := strings.EqualFold(settings.Theme, "Dark")
		opts := domain.PDFOptions{Dark: dark, LineNumbers: settings.IncludeLineNumbers, PageNumbers: settings.IncludePageNumbers}
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

		tempDir, err := s.tempFileProvider.MkdirTemp("", "shotgun-export-*")
		if err != nil {
			return domain.ExportResult{}, fmt.Errorf("failed to create temp dir: %w", err)
		}
		fileName := "context-human.pdf"
		outputPath := s.pathProvider.Join(tempDir, fileName)

		if err := s.pdf.WriteAtomic(settings.Context, opts, outputPath); err != nil {
			s.fileSystemWriter.RemoveAll(tempDir)
			return domain.ExportResult{}, fmt.Errorf("failed to generate human-readable PDF: %w", err)
		}

		fi, err := s.fileStatProvider.Stat(outputPath)
		if err != nil {
			s.fileSystemWriter.RemoveAll(tempDir)
			return domain.ExportResult{}, fmt.Errorf("failed to stat output file: %w", err)
		}

		return domain.ExportResult{
			Mode:      settings.Mode,
			FileName:  fileName,
			FilePath:  outputPath,
			IsLarge:   true,
			SizeBytes: fi.Size(),
		}, nil

	default:
		return domain.ExportResult{}, fmt.Errorf("unknown export mode: %s", settings.Mode)
	}
}

// GetExportHistory returns export history for a project
func (s *ExportService) GetExportHistory(ctx context.Context, projectPath string) ([]domain.ExportHistoryItem, error) {
	s.log.Info(fmt.Sprintf("Getting export history for project: %s", projectPath))

	// For now, return an empty history as this is a new feature
	// In the future, this could read from a history file or database
	var history []domain.ExportHistoryItem

	// TODO: Implement actual history tracking
	// This could involve:
	// 1. Reading from a history file in the project directory
	// 2. Querying a database of export operations
	// 3. Scanning export directories for previous exports

	s.log.Debug(fmt.Sprintf("Found %d export history items for project %s", len(history), projectPath))
	return history, nil
}

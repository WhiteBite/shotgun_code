package application

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"regexp"
	"shotgun_code/domain"
	"shotgun_code/infrastructure/contextbuilder"
	"shotgun_code/infrastructure/fonts"
	"strings"
	"unicode"

	"github.com/jung-kurt/gofpdf"
)

type ExportService struct{}

func NewExportService() *ExportService { return &ExportService{} }

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

// мягкий перенос длинных строк
func softWrapLongLines(text string, widthCols int) string {
	if widthCols <= 0 {
		return text
	}
	lines := strings.Split(text, "\n")
	var out strings.Builder
	for _, ln := range lines {
		runes := []rune(ln)
		for i := 0; i < len(runes); i += widthCols {
			j := i + widthCols
			if j > len(runes) {
				j = len(runes)
			}
			out.WriteString(string(runes[i:j]))
			out.WriteByte('\n')
		}
	}
	return out.String()
}

// заменяем экзотические руны на ASCII-маркер <U+XXXX>
func replaceUnsupported(text string) string {
	var b strings.Builder
	for _, r := range []rune(text) {
		if r == '\n' || r == '\r' || r == '\t' {
			b.WriteRune(r)
			continue
		}
		switch {
		case r >= 0x20 && r <= 0x7E:
			b.WriteRune(r)
		case (r >= 0x00A0 && r <= 0x024F) || (r >= 0x1E00 && r <= 0x1EFF):
			b.WriteRune(r)
		case (r >= 0x0400 && r <= 0x052F) || (r >= 0x2DE0 && r <= 0x2DFF) || (r >= 0xA640 && r <= 0xA69F):
			b.WriteRune(r)
		case r >= 0x0370 && r <= 0x03FF:
			b.WriteRune(r)
		case (r >= 0x2000 && r <= 0x206F) || (r >= 0x20A0 && r <= 0x20CF) ||
			(r >= 0x2100 && r <= 0x214F) || (r >= 0x2190 && r <= 0x21FF) ||
			(r >= 0x2200 && r <= 0x22FF) || (r >= 0x2300 && r <= 0x23FF) ||
			(r >= 0x2460 && r <= 0x24FF) || (r >= 0x2500 && r <= 0x257F) ||
			(r >= 0x2580 && r <= 0x259F) || (r >= 0x25A0 && r <= 0x25FF) ||
			(r >= 0x2600 && r <= 0x26FF):
			b.WriteRune(r)
		default:
			if !unicode.IsSpace(r) {
				b.WriteString(fmt.Sprintf("<U+%04X>", r))
			} else {
				b.WriteRune(r)
			}
		}
	}
	s := strings.ReplaceAll(b.String(), "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s = strings.ReplaceAll(s, "\t", "    ")
	return s
}

func registerUTF8Mono(pdf *gofpdf.Fpdf) (string, error) {
	tmp, err := os.CreateTemp("", "dejavu-mono-*.ttf")
	if err != nil {
		return "", err
	}
	defer func() { tmp.Close(); os.Remove(tmp.Name()) }()
	if _, err = tmp.Write(fonts.DejaVuSansMonoTTF); err != nil {
		return "", err
	}
	font := "DejaVuMono"
	pdf.AddUTF8Font(font, "", tmp.Name())
	return font, nil
}

func makeMonoPDF(text string, dark bool, lineNums bool, pageNums bool) ([]byte, error) {
	text = replaceUnsupported(text)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(12, 12, 12)
	pdf.SetAutoPageBreak(true, 12)

	bgR, bgG, bgB := 255, 255, 255
	fgR, fgG, fgB := 20, 22, 28
	if dark {
		bgR, bgG, bgB = 24, 26, 32
		fgR, fgG, fgB = 235, 235, 235
	}
	if pageNums {
		pdf.AliasNbPages("{nb}")
		pdf.SetFooterFunc(func() {
			pdf.SetY(-10)
			pdf.SetTextColor(fgR, fgG, fgB)
			pdf.SetFont("DejaVuMono", "", 9)
			pdf.CellFormat(0, 6, fmt.Sprintf("%d/{nb}", pdf.PageNo()), "", 0, "C", false, 0, "")
		})
	}

	font, err := registerUTF8Mono(pdf)
	if err != nil {
		return nil, fmt.Errorf("register font: %w", err)
	}

	pdf.AddPage()
	pdf.SetFillColor(bgR, bgG, bgB)
	pdf.Rect(0, 0, 210, 297, "F")

	pdf.SetTextColor(fgR, fgG, fgB)
	pdf.SetFont(font, "", 9)

	const maxCols = 160
	var out strings.Builder
	i := 1
	for _, line := range strings.Split(text, "\n") {
		if lineNums {
			out.WriteString(fmt.Sprintf("%6d  %s\n", i, line))
		} else {
			out.WriteString(line + "\n")
		}
		i++
	}
	wrapped := softWrapLongLines(out.String(), maxCols)

	pdf.SetX(12)
	pdf.MultiCell(0, 4.5, wrapped, "", "L", false)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *ExportService) Export(_ context.Context, settings domain.ExportSettings) (domain.ExportResult, error) {
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
			return domain.ExportResult{}, err
		}
		return domain.ExportResult{Mode: settings.Mode, Text: out}, nil

	case domain.ExportModeAI:
		parts := splitByFiles(settings.Context)
		limit := settings.TokenLimit
		if limit <= 0 {
			limit = 180000
		}
		var chunks []string
		cur := ""
		for _, p := range parts {
			if approxTokens(cur)+approxTokens(p) > limit && strings.TrimSpace(cur) != "" {
				chunks = append(chunks, cur)
				cur = ""
			}
			cur += p
		}
		if strings.TrimSpace(cur) != "" {
			chunks = append(chunks, cur)
		}

		if len(chunks) == 1 {
			pdfBytes, err := makeMonoPDF(chunks[0], false, false, false)
			if err != nil {
				return domain.ExportResult{}, err
			}
			return domain.ExportResult{Mode: settings.Mode, FileName: "context-ai.pdf",
				DataBase64: base64.StdEncoding.EncodeToString(pdfBytes)}, nil
		}

		var zb bytes.Buffer
		zw := zip.NewWriter(&zb)
		for i, c := range chunks {
			pdfBytes, err := makeMonoPDF(c, false, false, false)
			if err != nil {
				return domain.ExportResult{}, err
			}
			f, _ := zw.Create(fmt.Sprintf("context-%02d.pdf", i+1))
			_, _ = f.Write(pdfBytes)
		}
		_ = zw.Close()
		return domain.ExportResult{Mode: settings.Mode, FileName: "context-ai.zip",
			DataBase64: base64.StdEncoding.EncodeToString(zb.Bytes())}, nil

	case domain.ExportModeHuman:
		dark := strings.EqualFold(settings.Theme, "Dark")
		pdfBytes, err := makeMonoPDF(settings.Context, dark, settings.IncludeLineNumbers, settings.IncludePageNumbers)
		if err != nil {
			return domain.ExportResult{}, err
		}
		return domain.ExportResult{Mode: settings.Mode, FileName: "context-human.pdf",
			DataBase64: base64.StdEncoding.EncodeToString(pdfBytes)}, nil

	default:
		return domain.ExportResult{}, fmt.Errorf("unknown export mode")
	}
}

package pdfgen

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"shotgun_code/domain"
	"shotgun_code/infrastructure/fonts"

	"github.com/jung-kurt/gofpdf"
)

// GofpdfGenerator реализует domain.PDFGenerator.
type GofpdfGenerator struct {
	log domain.Logger
}

// NewGofpdfGenerator создаёт новый генератор PDF.
func NewGofpdfGenerator(log domain.Logger) domain.PDFGenerator {
	return &GofpdfGenerator{log: log}
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

func registerUTF8Mono(pdf *gofpdf.Fpdf) (string, error) {
	tmp, err := os.CreateTemp("", "dejavu-mono-*.ttf")
	if err != nil {
		return "", err
	}
	defer func() {
		tmp.Close()
		os.Remove(tmp.Name())
	}()
	if _, err = tmp.Write(fonts.DejaVuSansMonoTTF); err != nil {
		return "", err
	}
	font := "DejaVuMono"
	pdf.AddUTF8Font(font, "", tmp.Name())
	return font, nil
}

func (g *GofpdfGenerator) setupPDF(opts domain.PDFOptions) (*gofpdf.Fpdf, string, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(12, 12, 12)
	pdf.SetAutoPageBreak(true, 12)

	bgR, bgG, bgB := 255, 255, 255
	fgR, fgG, fgB := 20, 22, 28
	if opts.Dark {
		bgR, bgG, bgB = 24, 26, 32
		fgR, fgG, fgB = 235, 235, 235
	}
	if opts.PageNumbers {
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
		return nil, "", fmt.Errorf("register font: %w", err)
	}

	pdf.AddPage()
	pdf.SetFillColor(bgR, bgG, bgB)
	pdf.Rect(0, 0, 210, 297, "F")

	pdf.SetTextColor(fgR, fgG, fgB)
	pdf.SetFont(font, "", 9)

	return pdf, font, nil
}

// Generate создаёт PDF и возвращает байты.
func (g *GofpdfGenerator) Generate(text string, opts domain.PDFOptions) ([]byte, error) {
	text = replaceUnsupported(text)
	pdf, _, err := g.setupPDF(opts)
	if err != nil {
		return nil, err
	}

	const maxCols = 160
	var out strings.Builder
	i := 1
	for _, line := range strings.Split(text, "\n") {
		if opts.LineNumbers {
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

// WriteAtomic создаёт PDF и атомарно записывает в файл.
func (g *GofpdfGenerator) WriteAtomic(text string, opts domain.PDFOptions, outputPath string) error {
	text = replaceUnsupported(text)
	pdf, _, err := g.setupPDF(opts)
	if err != nil {
		return err
	}

	const maxCols = 160
	var out strings.Builder
	i := 1
	for _, line := range strings.Split(text, "\n") {
		if opts.LineNumbers {
			out.WriteString(fmt.Sprintf("%6d  %s\n", i, line))
		} else {
			out.WriteString(line + "\n")
		}
		i++
	}
	wrapped := softWrapLongLines(out.String(), maxCols)

	pdf.SetX(12)
	pdf.MultiCell(0, 4.5, wrapped, "", "L", false)

	dir := filepath.Dir(outputPath)
	tmp, err := os.CreateTemp(dir, "pdf-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	if err := pdf.OutputFileAndClose(tmpPath); err != nil {
		if err := os.Remove(tmpPath); err != nil {
			g.log.Warning(fmt.Sprintf("Failed to remove temporary PDF file: %v", err))
		}
		return fmt.Errorf("failed to generate PDF: %w", err)
	}
	if err := os.Rename(tmpPath, outputPath); err != nil {
		if err := os.Remove(tmpPath); err != nil {
			g.log.Warning(fmt.Sprintf("Failed to remove temporary PDF file after rename failure: %v", err))
		}
		return fmt.Errorf("failed to move PDF to final location: %w", err)
	}
	return nil
}

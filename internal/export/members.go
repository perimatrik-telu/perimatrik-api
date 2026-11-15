// Package export produces the member spreadsheet (xlsx / csv) with the same
// layout as the Laravel MemberExport.
package export

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"perimatrik-api-go/internal/models"

	"github.com/xuri/excelize/v2"
)

var headings = []string{
	"No Anggota", "Nama Lengkap", "Nama Angkatan", "Tahun Angkatan",
	"Nama Lapangan", "Domisili", "Kontak",
}

var columnWidths = []float64{18, 30, 24, 16, 24, 22, 20}

func MemberRows(members []models.Member) [][]string {
	rows := make([][]string, 0, len(members))
	for _, m := range members {
		var namaAngkatan, year string
		if m.Angkatan != nil {
			namaAngkatan = m.Angkatan.NamaAngkatan
			year = deref(m.Angkatan.Year)
		}
		rows = append(rows, []string{
			m.NoAnggota, m.NamaLengkap, namaAngkatan, year,
			deref(m.NamaLapangan), deref(m.Domisili), deref(m.Kontak),
		})
	}
	return rows
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// MembersCSV mirrors PhpSpreadsheet's CSV writer with excel_compatibility:
// UTF-8 BOM, "sep=;" line, ';' delimiter, every field quoted, CRLF.
func MembersCSV(rows [][]string) []byte {
	var b bytes.Buffer
	b.WriteString("\xEF\xBB\xBF")
	b.WriteString("sep=;\r\n")
	write := func(fields []string) {
		for i, f := range fields {
			if i > 0 {
				b.WriteByte(';')
			}
			b.WriteByte('"')
			b.WriteString(strings.ReplaceAll(f, `"`, `""`))
			b.WriteByte('"')
		}
		b.WriteString("\r\n")
	}
	write(headings)
	for _, r := range rows {
		write(r)
	}
	return b.Bytes()
}

// numericRe matches values PhpSpreadsheet's DefaultValueBinder stores as numbers
// (no leading zero, so phone numbers / member numbers like "0812" stay text).
var numericRe = regexp.MustCompile(`^-?(0|[1-9][0-9]{0,14})(\.[0-9]+)?$`)

func MembersXLSX(rows [][]string) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	const sheet = "Data Anggota"
	if err := f.SetSheetName("Sheet1", sheet); err != nil {
		return nil, err
	}

	set := func(row int, values []string) error {
		for i, v := range values {
			cell, _ := excelize.CoordinatesToCellName(i+1, row)
			var val any = v
			if row > 1 && numericRe.MatchString(v) {
				if n, err := strconv.ParseFloat(v, 64); err == nil {
					val = n
				}
			}
			if err := f.SetCellValue(sheet, cell, val); err != nil {
				return err
			}
		}
		return nil
	}
	if err := set(1, headings); err != nil {
		return nil, err
	}
	for i, r := range rows {
		if err := set(i+2, r); err != nil {
			return nil, err
		}
	}

	lastCol, _ := excelize.ColumnNumberToName(len(headings))
	lastRow := len(rows) + 1
	border := func(color string) []excelize.Border {
		return []excelize.Border{
			{Type: "left", Color: color, Style: 1}, {Type: "right", Color: color, Style: 1},
			{Type: "top", Color: color, Style: 1}, {Type: "bottom", Color: color, Style: 1},
		}
	}

	headerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF", Size: 12},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"1F4E78"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border:    border("E2E8F0"),
	})
	if err != nil {
		return nil, err
	}
	dataStyle, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
		Border:    border("E2E8F0"),
	})
	if err != nil {
		return nil, err
	}

	if err := f.SetCellStyle(sheet, "A1", lastCol+"1", headerStyle); err != nil {
		return nil, err
	}
	if lastRow > 1 {
		if err := f.SetCellStyle(sheet, "A2", fmt.Sprintf("%s%d", lastCol, lastRow), dataStyle); err != nil {
			return nil, err
		}
	}
	for i, w := range columnWidths {
		col, _ := excelize.ColumnNumberToName(i + 1)
		if err := f.SetColWidth(sheet, col, col, w); err != nil {
			return nil, err
		}
	}
	if err := f.SetRowHeight(sheet, 1, 24); err != nil {
		return nil, err
	}
	if err := f.AutoFilter(sheet, fmt.Sprintf("A1:%s%d", lastCol, lastRow), nil); err != nil {
		return nil, err
	}
	zoom := 90.0
	if err := f.SetSheetView(sheet, 0, &excelize.ViewOptions{ZoomScale: &zoom}); err != nil {
		return nil, err
	}
	if err := f.SetPanes(sheet, &excelize.Panes{
		Freeze: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft",
	}); err != nil {
		return nil, err
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

package export

import (
	"bytes"
	"testing"

	"perimatrik-api-go/internal/models"

	"github.com/xuri/excelize/v2"
)

func rows() [][]string {
	y := "2020"
	k := "0812345"
	return MemberRows([]models.Member{{NoAnggota: "001", NamaLengkap: `Budi "B"`, Kontak: &k, Angkatan: &models.Angkatan{NamaAngkatan: "XII", Year: &y}}})
}

func TestCSV(t *testing.T) {
	want := "\xEF\xBB\xBFsep=;\r\n" +
		`"No Anggota";"Nama Lengkap";"Nama Angkatan";"Tahun Angkatan";"Nama Lapangan";"Domisili";"Kontak"` + "\r\n" +
		`"001";"Budi ""B""";"XII";"2020";"";"";"0812345"` + "\r\n"
	if got := string(MembersCSV(rows())); got != want {
		t.Fatalf("got\n%q\nwant\n%q", got, want)
	}
}

func TestXLSX(t *testing.T) {
	data, err := MembersXLSX(rows())
	if err != nil {
		t.Fatal(err)
	}
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	got, _ := f.GetRows("Data Anggota")
	if len(got) != 2 || got[0][0] != "No Anggota" || got[1][0] != "001" || got[1][3] != "2020" || got[1][6] != "0812345" {
		t.Fatalf("rows: %v", got)
	}
	if typ, _ := f.GetCellType("Data Anggota", "D2"); typ != excelize.CellTypeNumber && typ != excelize.CellTypeUnset {
		t.Fatalf("year should be numeric, got %v", typ)
	}
}

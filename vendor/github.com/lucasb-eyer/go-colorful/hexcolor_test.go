package colorful

import (
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestHexColor(t *testing.T) {
	for _, tc := range []struct {
		hc HexColor
		s  string
	}{
		{HexColor{R: 0, G: 0, B: 0}, "#000000"},
		{HexColor{R: 1, G: 0, B: 0}, "#ff0000"},
		{HexColor{R: 0, G: 1, B: 0}, "#00ff00"},
		{HexColor{R: 0, G: 0, B: 1}, "#0000ff"},
		{HexColor{R: 1, G: 1, B: 1}, "#ffffff"},
	} {
		var gotHC HexColor
		if err := gotHC.Scan(tc.s); err != nil {
			t.Errorf("_.Scan(%q) == %v, want <nil>", tc.s, err)
		}
		if !reflect.DeepEqual(gotHC, tc.hc) {
			t.Errorf("_.Scan(%q) wrote %v, want %v", tc.s, gotHC, tc.hc)
		}
		if gotValue, err := tc.hc.Value(); err != nil || !reflect.DeepEqual(gotValue, tc.s) {
			t.Errorf("%v.Value() == %v, %v, want %v, <nil>", tc.hc, gotValue, err, tc.s)
		}
	}
}

func Example_HexColor_Scan() {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT '#ff0000' AS color;").
		WillReturnRows(
			sqlmock.NewRows([]string{"color"}).
				AddRow("#ff0000"),
		)

	var hc HexColor
	if err := db.QueryRow("SELECT '#ff0000' AS color;").Scan(&hc); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("hc = %+v\n", hc)

	// Output:
	// hc = {R:1 G:0 B:0}
}

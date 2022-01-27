package util

import (
	"reflect"
	"strings"
	"testing"
)

func Test_uniqueInts(t *testing.T) {
	tests := []struct {
		name string
		args []int
		want []int
	}{
		{`a`, []int{1}, []int{1}},
		{`b`, []int{1, 2, 2, 3}, []int{1, 2, 3}},
		{`c`, []int{}, []int{}},
		{`d`, []int{1, 2, 1, 2}, []int{1, 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UniqueInts(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("uniqueInts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIntSliceToCSV(t *testing.T) {
	tests := []struct {
		name string
		args []int
		want string
	}{
		{`a`, []int{1, 2, 3}, `1,2,3`},
		{`b`, []int{1}, `1`},
		{`c`, []int{}, ``},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IntSliceToCSV(tt.args); got != tt.want {
				t.Errorf("IntSliceToCSV() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSONMarshalNoEscape(t *testing.T) {

	type ts struct {
		Content string
	}
	tests := []struct {
		name    string
		args    ts
		want    string
		wantErr bool
	}{
		{`a`, ts{`hello`}, `{"Content":"hello"}` + "\n", false},
		{`b`, ts{`Sanford & Son`}, `{"Content":"Sanford & Son"}` + "\n", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := JSONMarshalNoEscape(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("JSONMarshalNoEscape() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gs := string(got)
			if gs != tt.want {
				t.Errorf(`JSONMarshalNoEscape() got = "%s", want "%s"`, gs, tt.want)
			}
		})
	}
}

func Test_EncodeInteger(t *testing.T) {
	tests := []struct {
		name string
		drid int
	}{
		{`A`, -1},
		{`B`, 0},
		{`C1`, 100},
		{`C2`, 101},
		{`C3`, 102},
		{`C4`, 103},
		{`C5`, 104},
		{`C6`, 105},
		{`C7`, 106},
		{`D`, 1000},
		{`E`, 10000},
		{`F`, 100000},
		{`G`, 1000000},
		{`H`, 10000000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//time.Sleep(2 * time.Second)
			got := EncodeInteger(tt.drid)
			got2 := DecodeInteger(got)

			if got2 != tt.drid {
				t.Errorf("%s: EncodeInteger() = %s, %d => %d", tt.name, got, tt.drid, got2)
			}
		})
	}
	got := DecodeInteger(`A7`)
	if got != 0 {
		t.Errorf("A7: DecodeInteger() = %d expected 0", got)
	}
	g2 := DecodeInteger(`1608CC`)
	if g2 != 0 {
		t.Errorf("1608CC: DecodeInteger() = %d expected 0", g2)
	}

}

func TestFilter(t *testing.T) {
	tfa := func(v string) bool {
		return strings.Contains(v, `a`)
	}
	tests := []struct {
		name    string
		strings []string
		want    []string
	}{
		{`a`, []string{`ab`, `bb`, `bc`, `da`}, []string{`ab`, `da`}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Filter(tt.strings, tfa); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Filter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringsDiff(t *testing.T) {
	tests := []struct {
		name     string
		bstring  []string
		astring  []string
		wantDiff []string
	}{
		{`a`, []string{}, []string{}, []string{}},
		{`b`, []string{`a`}, []string{`b`}, []string{`b`}},
		{`c`, []string{`a`, `b`}, []string{`b`}, []string{}},
		{`d`, []string{`a`, `d`, `c`}, []string{`b`, `d`}, []string{`b`}},
		{`e`, []string{}, []string{`b`, `d`}, []string{`b`, `d`}},
		{`f`, []string{`a`}, []string{}, []string{}},
		{`g`, []string{`a`}, []string{`a`, `b`, `c`}, []string{`b`, `c`}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDiff := StringsDiff(tt.astring, tt.bstring)
			if len(gotDiff) == 0 && len(tt.wantDiff) == 0 { // we are good
			} else {
				if !reflect.DeepEqual(gotDiff, tt.wantDiff) {
					t.Errorf("StringsDiff() = %v, want %v", gotDiff, tt.wantDiff)
				}
			}
		})
	}
}

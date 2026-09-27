package convert

import "testing"

func TestConvertDateWithTime(t *testing.T) {
	got := ConvertDate("2026-09-20 22:08")
	want := "2026-09-20T22:08:00Z"
	if got != want {
		t.Errorf("ConvertDate = %q, want %q", got, want)
	}
}

func TestConvertDateWithoutTime(t *testing.T) {
	got := ConvertDate("2026-09-20")
	want := "2026-09-20T00:00:00Z"
	if got != want {
		t.Errorf("ConvertDate = %q, want %q", got, want)
	}
}

func TestConvertDateOfEmptyStringIsEmpty(t *testing.T) {
	if got := ConvertDate(""); got != "" {
		t.Errorf("ConvertDate(\"\") = %q, want empty", got)
	}
}

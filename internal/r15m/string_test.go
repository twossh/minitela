package r15m

import (
	"bytes"
	"testing"
)

func TestCleanDisplayString(t *testing.T) {
	raw := []byte{
		'H', 'o', 'n', 'd', 'a', ' ',
		'N', 'i', 'c', '.', '.', '.',
		0x00,
		0x00,
		0x04,
		0x00,
		0xFF,
	}

	got := cleanDisplayString(raw)
	want := []byte("Honda Nic...")

	if !bytes.Equal(got, want) {
		t.Fatalf(
			"resultado=%q esperado=%q",
			got,
			want,
		)
	}
}

func TestCleanDisplayStringWithoutNUL(t *testing.T) {
	raw := []byte("Pebble M350s")

	got := cleanDisplayString(raw)

	if !bytes.Equal(got, raw) {
		t.Fatalf(
			"resultado=%q esperado=%q",
			got,
			raw,
		)
	}
}

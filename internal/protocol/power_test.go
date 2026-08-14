package protocol

import (
	"bytes"
	"testing"
)

func TestBuildRebootRequest(t *testing.T) {
	want := []byte{
		0x41, 0x48,
		0x00, 0x02,
		0x00, 0x70,
		0x00, 0x00,
		0x4D, 0x49,
	}

	got := BuildRebootRequest()

	if !bytes.Equal(got, want) {
		t.Fatalf(
			"reboot:\n got: % X\nwant: % X",
			got,
			want,
		)
	}
}

func TestValidateRebootResponse(t *testing.T) {
	frame := []byte{
		0x41, 0x48,
		0x00, 0x02,
		0x00, 0xB0,
		0x00, 0x00,
		0x4D, 0x49,
	}

	if err := ValidateRebootResponse(frame); err != nil {
		t.Fatal(err)
	}
}

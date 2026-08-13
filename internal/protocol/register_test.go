package protocol

import (
	"bytes"
	"testing"
)

func TestBuildReadBrightnessRequest(t *testing.T) {
	got := BuildReadNumRegisterRequest(7)

	want := []byte{
		0x41, 0x48,
		0x00, 0x05,
		0x00, 0x90,
		0xC0,
		0x00, 0x07,
		0x00, 0x00,
		0x4D, 0x49,
	}

	if !bytes.Equal(got, want) {
		t.Fatalf(
			"frame:\n got: % X\nwant: % X",
			got,
			want,
		)
	}
}

func TestParseBrightnessResponse(t *testing.T) {
	// Captura real do Positivo R15M:
	//
	// reg 7 = 75.
	frame := []byte{
		0x41, 0x48,
		0x80, 0x09,
		0x00, 0xD0,
		0x00,
		0x00, 0x07,
		0x00, 0x00, 0x00, 0x4B,
		0x00, 0x00,
		0x4D, 0x49,
	}

	regID, value, err := ParseReadNumRegisterResponse(frame)
	if err != nil {
		t.Fatal(err)
	}

	if regID != 7 {
		t.Fatalf(
			"regID = %d, esperado 7",
			regID,
		)
	}

	if value != 75 {
		t.Fatalf(
			"value = %d, esperado 75",
			value,
		)
	}
}

func TestRejectInvalidRegisterResponse(t *testing.T) {
	frame := make([]byte, ReadRegisterResponseSize)

	if _, _, err := ParseReadNumRegisterResponse(frame); err == nil {
		t.Fatal("frame inválido foi aceito")
	}
}

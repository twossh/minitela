package protocol

import (
	"bytes"
	"testing"
)

func TestBuildReadStringRegisterRequest(t *testing.T) {
	got := BuildReadStringRegisterRequest(
		1083,
		64,
	)

	want := []byte{
		0x41, 0x48,
		0x00, 0x07,
		0x00, 0x90,
		0xE0,
		0x04, 0x3B,
		0x00, 0x40,
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

func TestBuildWriteStringRegisterRequest(t *testing.T) {
	got, err := BuildWriteStringRegisterRequest(
		1083,
		[]byte("Test"),
	)
	if err != nil {
		t.Fatal(err)
	}

	want := []byte{
		0x41, 0x48,
		0x00, 0x0B,
		0x00, 0x90,
		0xD0,
		0x04, 0x3B,
		0x00, 0x04,
		'T', 'e', 's', 't',
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

func TestParseReadStringRegisterResponse(t *testing.T) {
	text := []byte("Honda Nic...")

	// command type (2) +
	// content: header(1)+reg(2)+len(2)+text
	commandLength := 7 + len(text)

	frame := make(
		[]byte,
		2+2+commandLength+2+2,
	)

	frame[0] = 0x41
	frame[1] = 0x48
	frame[2] = 0x00
	frame[3] = byte(commandLength)

	frame[4] = 0x00
	frame[5] = 0xD0

	frame[6] = 0x10
	frame[7] = 0x04
	frame[8] = 0x3B
	frame[9] = 0x00
	frame[10] = byte(len(text))

	copy(frame[11:], text)

	crcOffset := 11 + len(text)

	frame[crcOffset] = 0x00
	frame[crcOffset+1] = 0x00
	frame[crcOffset+2] = 0x4D
	frame[crcOffset+3] = 0x49

	regID, got, err :=
		ParseReadStringRegisterResponse(frame)

	if err != nil {
		t.Fatal(err)
	}

	if regID != 1083 {
		t.Fatalf(
			"regID=%d esperado=1083",
			regID,
		)
	}

	if !bytes.Equal(got, text) {
		t.Fatalf(
			"texto=%q esperado=%q",
			got,
			text,
		)
	}
}

package stc

import (
	"encoding/hex"
	"testing"
)

func TestEncodeMode0VendorBlue(t *testing.T) {
	var pixels [16]RGBA

	for i := range pixels {
		pixels[i] = RGBA{
			R: 0,
			G: 0,
			B: 80,
			A: 255,
		}
	}

	got, errValue, err :=
		EncodeMode0Block(pixels)

	if err != nil {
		t.Fatalf(
			"EncodeMode0Block: %v",
			err,
		)
	}

	if errValue != 0 {
		t.Fatalf(
			"erro esperado=0 obtido=%d",
			errValue,
		)
	}

	const expectedHex = "00000000a0a0feff0100000000000000"

	expected, err := hex.DecodeString(
		expectedHex,
	)

	if err != nil {
		t.Fatal(err)
	}

	if string(got[:]) != string(expected) {
		t.Fatalf(
			"\nbloco diferente\nesperado: %s\nobtido  : %x",
			expectedHex,
			got,
		)
	}
}

func TestEncodeMode0SolidRed(t *testing.T) {
	var pixels [16]RGBA

	for i := range pixels {
		pixels[i] = RGBA{
			R: 180,
			G: 0,
			B: 0,
			A: 255,
		}
	}

	got, errValue, err :=
		EncodeMode0Block(pixels)

	if err != nil {
		t.Fatal(err)
	}

	if errValue != 0 {
		t.Fatalf(
			"erro esperado=0 obtido=%d",
			errValue,
		)
	}

	const expected = "686901000000feff0100000000000000"

	if hex.EncodeToString(
		got[:],
	) != expected {
		t.Fatalf(
			"esperado=%s obtido=%x",
			expected,
			got,
		)
	}
}

func TestEncodeMode0SolidGreen(t *testing.T) {
	var pixels [16]RGBA

	for i := range pixels {
		pixels[i] = RGBA{
			R: 0,
			G: 180,
			B: 0,
			A: 255,
		}
	}

	got, errValue, err :=
		EncodeMode0Block(pixels)

	if err != nil {
		t.Fatal(err)
	}

	if errValue != 0 {
		t.Fatalf(
			"erro esperado=0 obtido=%d",
			errValue,
		)
	}

	const expected = "000068690100feff0100000000000000"

	if hex.EncodeToString(
		got[:],
	) != expected {
		t.Fatalf(
			"esperado=%s obtido=%x",
			expected,
			got,
		)
	}
}

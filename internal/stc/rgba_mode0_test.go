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
func TestEncodeMode0OptimizesCorrelatedGradient(t *testing.T) {
	pixels := [16]RGBA{
		{R: 48, G: 60, B: 187, A: 255},
		{R: 114, G: 75, B: 147, A: 255},
		{R: 228, G: 95, B: 18, A: 255},
		{R: 188, G: 88, B: 67, A: 255},
		{R: 70, G: 52, B: 179, A: 255},
		{R: 230, G: 88, B: 28, A: 255},
		{R: 131, G: 73, B: 110, A: 255},
		{R: 163, G: 70, B: 86, A: 255},
		{R: 204, G: 77, B: 51, A: 255},
		{R: 199, G: 71, B: 62, A: 255},
		{R: 79, G: 55, B: 162, A: 255},
		{R: 34, G: 46, B: 214, A: 255},
		{R: 22, G: 49, B: 211, A: 255},
		{R: 206, G: 84, B: 43, A: 255},
		{R: 158, G: 82, B: 102, A: 255},
		{R: 156, G: 82, B: 92, A: 255},
	}

	baselineEP0, baselineEP1 :=
		minMaxMode0Endpoints(pixels)

	_, baselineError :=
		chooseIndices(
			pixels,
			baselineEP0,
			baselineEP1,
		)

	_, optimizedError, err :=
		EncodeMode0Block(pixels)

	if err != nil {
		t.Fatal(err)
	}

	if optimizedError >= baselineError {
		t.Fatalf(
			"otimização não reduziu erro: baseline=%d otimizado=%d",
			baselineError,
			optimizedError,
		)
	}

	if optimizedError*10 >= baselineError {
		t.Fatalf(
			"ganho menor que o esperado: baseline=%d otimizado=%d",
			baselineError,
			optimizedError,
		)
	}
}

func TestEncodeMode0OptimizationNeverWorseThanMinMax(t *testing.T) {
	var pixels [16]RGBA

	for i := range pixels {
		pixels[i] = RGBA{
			R: uint8((i * 37) % 256),
			G: uint8((i * 83) % 256),
			B: uint8((255 - i*29) & 0xFF),
			A: 255,
		}
	}

	baselineEP0, baselineEP1 :=
		minMaxMode0Endpoints(pixels)

	_, baselineError :=
		chooseIndices(
			pixels,
			baselineEP0,
			baselineEP1,
		)

	_, optimizedError, err :=
		EncodeMode0Block(pixels)

	if err != nil {
		t.Fatal(err)
	}

	if optimizedError > baselineError {
		t.Fatalf(
			"otimização piorou erro: baseline=%d otimizado=%d",
			baselineError,
			optimizedError,
		)
	}
}

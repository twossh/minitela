package protocol

import "testing"

func TestCRC16IBMKnownVector(t *testing.T) {
	got := CRC16IBM([]byte("123456789"))

	const want uint16 = 0xBB3D

	if got != want {
		t.Fatalf(
			"CRC16IBM = 0x%04X, esperado 0x%04X",
			got,
			want,
		)
	}
}

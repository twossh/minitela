package protocol

import "testing"

func TestHandshakeRequest(t *testing.T) {
	expected := []byte{
		0x41, 0x48,
		0x00, 0x02,
		0x00, 0x80,
		0x00, 0x00,
		0x4D, 0x49,
	}

	got := HandshakeRequest()

	if len(got) != len(expected) {
		t.Fatalf(
			"tamanho = %d, esperado = %d",
			len(got),
			len(expected),
		)
	}

	for i := range expected {
		if got[i] != expected[i] {
			t.Fatalf(
				"byte %d = 0x%02X, esperado 0x%02X",
				i,
				got[i],
				expected[i],
			)
		}
	}
}

func TestValidateHandshakeResponse(t *testing.T) {
	valid := []byte{
		0x41, 0x48,
		0x00, 0x06,
		0x00, 0xC0,
		0x00, 0x00,
		0x04, 0x00, 0x00, 0x00,
		0x4D, 0x49,
	}

	if err := ValidateHandshakeResponse(valid); err != nil {
		t.Fatalf("resposta válida rejeitada: %v", err)
	}
}

func TestRejectInvalidHandshake(t *testing.T) {
	invalid := []byte{
		0x41, 0x48,
		0x00, 0x06,
		0x00, 0xC0,
		0x00, 0x00,
		0xFF, 0x00, 0x00, 0x00,
		0x4D, 0x49,
	}

	if err := ValidateHandshakeResponse(invalid); err == nil {
		t.Fatal("resposta inválida foi aceita")
	}
}

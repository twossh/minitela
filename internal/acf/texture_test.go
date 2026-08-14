package acf

import (
	"bytes"
	"testing"

	"github.com/twossh/minitela/internal/stc"
)

func TestTextureConstants(
	t *testing.T,
) {
	if PayloadOffset != 0xA77B0 {
		t.Fatalf(
			"PayloadOffset=0x%X esperado=0xA77B0",
			PayloadOffset,
		)
	}

	if TextureFrameSize != 0x9000 {
		t.Fatalf(
			"TextureFrameSize=0x%X esperado=0x9000",
			TextureFrameSize,
		)
	}

	if TextureFrameCount != 30 {
		t.Fatalf(
			"TextureFrameCount=%d esperado=30",
			TextureFrameCount,
		)
	}

	if TexturePayloadEnd != 0x1B57B0 {
		t.Fatalf(
			"TexturePayloadEnd=0x%X esperado=0x1B57B0",
			TexturePayloadEnd,
		)
	}

	if TextureTailSize != 0xA858 {
		t.Fatalf(
			"TextureTailSize=0x%X esperado=0xA858",
			TextureTailSize,
		)
	}

	if TextureTemplateSize != 0x1C0008 {
		t.Fatalf(
			"TextureTemplateSize=0x%X esperado=0x1C0008",
			TextureTemplateSize,
		)
	}

	if TextureFrameSize != stc.FrameSize {
		t.Fatalf(
			"frame ACF=%d frame STC=%d",
			TextureFrameSize,
			stc.FrameSize,
		)
	}
}

func TestBuildStaticTexture(
	t *testing.T,
) {
	template := makeValidTextureTemplate(
		t,
	)

	original := append(
		[]byte(nil),
		template...,
	)

	frame := make(
		[]byte,
		TextureFrameSize,
	)

	for i := range frame {
		frame[i] = byte(
			(i*37 + 11) & 0xFF,
		)
	}

	frameOriginal := append(
		[]byte(nil),
		frame...,
	)

	result, err := BuildStaticTexture(
		template,
		frame,
	)

	if err != nil {
		t.Fatal(err)
	}

	if len(result) != TextureTemplateSize {
		t.Fatalf(
			"tamanho=0x%X esperado=0x%X",
			len(result),
			TextureTemplateSize,
		)
	}

	if err := ValidateChecksum(
		result,
	); err != nil {
		t.Fatalf(
			"ACF gerado inválido: %v",
			err,
		)
	}

	// Bytes anteriores ao checksum.
	if !bytes.Equal(
		result[:4],
		original[:4],
	) {
		t.Fatal(
			"header 0..3 foi modificado",
		)
	}

	// Header após o checksum até o payload.
	if !bytes.Equal(
		result[8:PayloadOffset],
		original[8:PayloadOffset],
	) {
		t.Fatal(
			"header 8..PayloadOffset foi modificado",
		)
	}

	// Os 30 frames precisam ser idênticos
	// ao frame fornecido.
	for i := 0; i < TextureFrameCount; i++ {

		start := PayloadOffset +
			i*TextureFrameSize

		end := start +
			TextureFrameSize

		if !bytes.Equal(
			result[start:end],
			frame,
		) {
			t.Fatalf(
				"frame %d difere do frame de entrada",
				i,
			)
		}
	}

	// Tail precisa permanecer byte por byte igual.
	if !bytes.Equal(
		result[TexturePayloadEnd:],
		original[TexturePayloadEnd:],
	) {
		t.Fatal(
			"tail do template foi modificada",
		)
	}

	// A função não pode modificar os argumentos.
	if !bytes.Equal(
		template,
		original,
	) {
		t.Fatal(
			"template de entrada foi modificado",
		)
	}

	if !bytes.Equal(
		frame,
		frameOriginal,
	) {
		t.Fatal(
			"frame de entrada foi modificado",
		)
	}
}

func TestBuildStaticTextureRejectsFrameSize(
	t *testing.T,
) {
	template := makeValidTextureTemplate(
		t,
	)

	frame := make(
		[]byte,
		TextureFrameSize-1,
	)

	_, err := BuildStaticTexture(
		template,
		frame,
	)

	if err == nil {
		t.Fatal(
			"esperava erro para frame inválido",
		)
	}
}

func TestValidateTextureTemplateRejectsSize(
	t *testing.T,
) {
	template := make(
		[]byte,
		TextureTemplateSize-1,
	)

	err := ValidateTextureTemplate(
		template,
	)

	if err == nil {
		t.Fatal(
			"esperava erro para tamanho inválido",
		)
	}
}

func TestValidateTextureTemplateRejectsChecksum(
	t *testing.T,
) {
	template := makeValidTextureTemplate(
		t,
	)

	template[0x100] ^= 0xFF

	err := ValidateTextureTemplate(
		template,
	)

	if err == nil {
		t.Fatal(
			"esperava erro para checksum inválido",
		)
	}
}

func makeValidTextureTemplate(
	t *testing.T,
) []byte {
	t.Helper()

	data := make(
		[]byte,
		TextureTemplateSize,
	)

	// Preencher com conteúdo determinístico
	// para detectar alterações indevidas.
	for i := range data {
		data[i] = byte(
			(i*13 + 7) & 0xFF,
		)
	}

	// Footer oficial.
	data[len(data)-4] = 0xA5
	data[len(data)-3] = 0x5A
	data[len(data)-2] = 0x5A
	data[len(data)-1] = 0xA5

	if err := SetChecksum(
		data,
	); err != nil {
		t.Fatalf(
			"SetChecksum: %v",
			err,
		)
	}

	if err := ValidateChecksum(
		data,
	); err != nil {
		t.Fatalf(
			"template de teste inválido: %v",
			err,
		)
	}

	return data
}

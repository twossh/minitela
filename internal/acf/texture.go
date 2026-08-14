package acf

import (
	"fmt"

	"github.com/twossh/minitela/internal/stc"
)

const (
	// TextureTemplateSize é o tamanho do template ACF
	// de 30 frames validado fisicamente no R15M.
	TextureTemplateSize = 0x1C0008

	// TextureFrameCount corresponde à animação/template
	// validada fisicamente: 30 frames.
	TextureFrameCount = 30

	// Cada frame STCRGBA possui 48x48 blocos,
	// 16 bytes por bloco.
	TextureFrameSize = stc.FrameSize

	// Região ocupada pelos 30 frames.
	TexturePayloadEnd = PayloadOffset +
		TextureFrameCount*TextureFrameSize

	// Região após os frames, que deve permanecer
	// byte por byte intacta.
	TextureTailSize = TextureTemplateSize -
		TexturePayloadEnd
)

// ValidateTextureTemplate verifica se o ACF possui a
// estrutura esperada para servir como template.
//
// Não exige um SHA específico: qualquer template com
// este layout, checksum e footer válidos pode ser usado.
func ValidateTextureTemplate(
	template []byte,
) error {
	if len(template) != TextureTemplateSize {
		return fmt.Errorf(
			"template ACF possui 0x%X bytes; esperado 0x%X",
			len(template),
			TextureTemplateSize,
		)
	}

	if TextureFrameSize != 0x9000 {
		return fmt.Errorf(
			"frame STCRGBA possui 0x%X bytes; esperado 0x9000",
			TextureFrameSize,
		)
	}

	if TexturePayloadEnd != 0x1B57B0 {
		return fmt.Errorf(
			"fim do payload calculado=0x%X; esperado=0x1B57B0",
			TexturePayloadEnd,
		)
	}

	if TextureTailSize != 0xA858 {
		return fmt.Errorf(
			"tail calculada=0x%X; esperado=0xA858",
			TextureTailSize,
		)
	}

	if err := ValidateChecksum(
		template,
	); err != nil {
		return fmt.Errorf(
			"template ACF inválido: %w",
			err,
		)
	}

	return nil
}

// BuildStaticTexture cria um ACF completo a partir de:
//
//   - um template ACF válido;
//   - um frame STCRGBA de 0x9000 bytes.
//
// O mesmo frame é repetido 30 vezes, mantendo intactos
// todo o header e a tail do template.
//
// Somente estas regiões são alteradas:
//
//   - checksum em bytes 4..7;
//   - payload 0xA77B0..0x1B57B0.
func BuildStaticTexture(
	template []byte,
	frame []byte,
) ([]byte, error) {
	if len(frame) != TextureFrameSize {
		return nil, fmt.Errorf(
			"frame STCRGBA possui 0x%X bytes; esperado 0x%X",
			len(frame),
			TextureFrameSize,
		)
	}

	frames := make(
		[][]byte,
		TextureFrameCount,
	)

	for i := range frames {
		frames[i] = frame
	}

	return BuildTextureFrames(
		template,
		frames,
	)
}

// BuildTextureFrames cria um ACF completo usando os 30 frames
// STCRGBA fornecidos na ordem em que serão reproduzidos.
//
// Header e tail permanecem byte por byte iguais ao template;
// somente payload e checksum são atualizados.
func BuildTextureFrames(
	template []byte,
	frames [][]byte,
) ([]byte, error) {
	if err := ValidateTextureTemplate(
		template,
	); err != nil {
		return nil, err
	}

	if len(frames) != TextureFrameCount {
		return nil, fmt.Errorf(
			"quantidade de frames STCRGBA=%d; esperado=%d",
			len(frames),
			TextureFrameCount,
		)
	}

	result := make(
		[]byte,
		len(template),
	)

	copy(
		result,
		template,
	)

	for frameIndex, frame := range frames {
		if len(frame) != TextureFrameSize {
			return nil, fmt.Errorf(
				"frame STCRGBA %d possui 0x%X bytes; esperado 0x%X",
				frameIndex,
				len(frame),
				TextureFrameSize,
			)
		}

		start := PayloadOffset +
			frameIndex*TextureFrameSize

		end := start +
			TextureFrameSize

		copy(
			result[start:end],
			frame,
		)
	}

	if err := SetChecksum(
		result,
	); err != nil {
		return nil, fmt.Errorf(
			"recalcular checksum ACF: %w",
			err,
		)
	}

	if err := ValidateChecksum(
		result,
	); err != nil {
		return nil, fmt.Errorf(
			"ACF gerado inválido: %w",
			err,
		)
	}

	return result, nil
}

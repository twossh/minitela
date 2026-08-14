package acf

import (
	"fmt"
)

const PayloadOffset = 0xA77B0

func BuildPatternACF(
	template []byte,
	frameCount int,
	packing VendorPacking,
	swapRB bool,
) ([]byte, error) {
	if frameCount < 1 {
		return nil, fmt.Errorf(
			"frameCount inválido: %d",
			frameCount,
		)
	}

	end :=
		PayloadOffset +
			frameCount*
				BC3FrameSize

	if len(template) < end {
		return nil, fmt.Errorf(
			"template curto: tamanho=%d necessário=%d",
			len(template),
			end,
		)
	}

	result :=
		make(
			[]byte,
			len(template),
		)

	copy(
		result,
		template,
	)

	frame :=
		QuadrantPattern(
			packing,
			swapRB,
		)

	for index := 0; index < frameCount; index++ {

		start :=
			PayloadOffset +
				index*
					BC3FrameSize

		stop :=
			start +
				BC3FrameSize

		copy(
			result[start:stop],
			frame,
		)
	}

	if err :=
		SetChecksum(
			result,
		); err != nil {

		return nil, err
	}

	return result, nil
}

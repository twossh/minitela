package customimage

import (
	"fmt"
	"image"
	"image/draw"
	"image/gif"

	"github.com/twossh/minitela/internal/acf"
	"github.com/twossh/minitela/internal/stc"
)

const defaultGIFDelay = 10

// encodeGIFTextureFrames converte um ciclo do GIF em exatamente
// 30 frames STCRGBA, preservando a proporção temporal dos delays.
func encodeGIFTextureFrames(
	animation *gif.GIF,
) ([][]byte, stc.ImageStats, error) {
	var totalStats stc.ImageStats

	if animation == nil {
		return nil, totalStats, fmt.Errorf("GIF nil")
	}

	if len(animation.Image) == 0 {
		return nil, totalStats, fmt.Errorf("GIF sem frames")
	}

	if animation.Config.Width <= 0 ||
		animation.Config.Height <= 0 {
		return nil, totalStats, fmt.Errorf(
			"GIF possui dimensões inválidas: %dx%d",
			animation.Config.Width,
			animation.Config.Height,
		)
	}

	plan := gifFramePlan(
		animation.Delay,
		len(animation.Image),
		acf.TextureFrameCount,
	)

	slotsBySource := make(
		map[int][]int,
		len(animation.Image),
	)

	for slot, sourceIndex := range plan {
		slotsBySource[sourceIndex] = append(
			slotsBySource[sourceIndex],
			slot,
		)
	}

	frames := make(
		[][]byte,
		acf.TextureFrameCount,
	)

	canvas := image.NewRGBA(
		image.Rect(
			0,
			0,
			animation.Config.Width,
			animation.Config.Height,
		),
	)

	for sourceIndex, source := range animation.Image {
		if source == nil {
			return nil, totalStats, fmt.Errorf(
				"GIF frame %d é nil",
				sourceIndex,
			)
		}

		disposal := byte(0)
		if sourceIndex < len(animation.Disposal) {
			disposal = animation.Disposal[sourceIndex]
		}

		var previous *image.RGBA
		if disposal == gif.DisposalPrevious {
			previous = cloneRGBA(canvas)
		}

		draw.Draw(
			canvas,
			source.Bounds(),
			source,
			source.Bounds().Min,
			draw.Over,
		)

		if slots := slotsBySource[sourceIndex]; len(slots) > 0 {
			encoded, frameStats, err :=
				stc.EncodeImageOpaque(canvas)
			if err != nil {
				return nil, totalStats, fmt.Errorf(
					"frame %d: %w",
					sourceIndex,
					err,
				)
			}

			for _, slot := range slots {
				frames[slot] = encoded

				totalStats.Blocks += frameStats.Blocks
				totalStats.Pixels += frameStats.Pixels
				totalStats.TotalError += frameStats.TotalError
			}
		}

		switch disposal {
		case gif.DisposalBackground:
			clearGIFRect(
				canvas,
				source.Bounds(),
			)

		case gif.DisposalPrevious:
			if previous != nil {
				copy(
					canvas.Pix,
					previous.Pix,
				)
			}
		}
	}

	for i, frame := range frames {
		if len(frame) != stc.FrameSize {
			return nil, totalStats, fmt.Errorf(
				"GIF não gerou frame físico %d",
				i,
			)
		}
	}

	return frames, totalStats, nil
}

// gifFramePlan amostra um ciclo completo do GIF em targetCount
// instantes uniformes, usando Delay (centésimos de segundo) como peso.
func gifFramePlan(
	delays []int,
	sourceCount int,
	targetCount int,
) []int {
	if sourceCount <= 0 || targetCount <= 0 {
		return nil
	}

	normalized := make(
		[]int,
		sourceCount,
	)

	totalDelay := 0

	for i := 0; i < sourceCount; i++ {
		delay := defaultGIFDelay

		if i < len(delays) && delays[i] > 0 {
			delay = delays[i]
		}

		normalized[i] = delay
		totalDelay += delay
	}

	plan := make(
		[]int,
		targetCount,
	)

	denominator := 2 * targetCount

	for slot := 0; slot < targetCount; slot++ {
		// Centro temporal de cada slot, evitando viés para o
		// primeiro ou último frame do ciclo.
		sampleNumerator :=
			(2*slot + 1) * totalDelay

		cumulative := 0
		selected := sourceCount - 1

		for sourceIndex, delay := range normalized {
			cumulative += delay

			if sampleNumerator <
				denominator*cumulative {
				selected = sourceIndex
				break
			}
		}

		plan[slot] = selected
	}

	return plan
}

func cloneRGBA(
	src *image.RGBA,
) *image.RGBA {
	if src == nil {
		return nil
	}

	dst := image.NewRGBA(
		src.Bounds(),
	)

	copy(
		dst.Pix,
		src.Pix,
	)

	return dst
}

func clearGIFRect(
	canvas *image.RGBA,
	rect image.Rectangle,
) {
	if canvas == nil {
		return
	}

	rect = rect.Intersect(
		canvas.Bounds(),
	)

	if rect.Empty() {
		return
	}

	draw.Draw(
		canvas,
		rect,
		image.Transparent,
		image.Point{},
		draw.Src,
	)
}

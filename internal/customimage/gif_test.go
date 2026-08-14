package customimage

import "testing"

func TestGIFFramePlanUsesAllThirtySlots(t *testing.T) {
	plan := gifFramePlan(
		[]int{10, 10, 10},
		3,
		30,
	)

	if len(plan) != 30 {
		t.Fatalf(
			"slots=%d esperado=30",
			len(plan),
		)
	}

	counts := [3]int{}

	for _, source := range plan {
		if source < 0 || source >= len(counts) {
			t.Fatalf(
				"frame fonte inválido=%d",
				source,
			)
		}

		counts[source]++
	}

	for source, count := range counts {
		if count != 10 {
			t.Fatalf(
				"frame %d recebeu %d slots; esperado=10",
				source,
				count,
			)
		}
	}
}

func TestGIFFramePlanPreservesRelativeDelay(t *testing.T) {
	plan := gifFramePlan(
		[]int{10, 20},
		2,
		30,
	)

	counts := [2]int{}

	for _, source := range plan {
		counts[source]++
	}

	if counts[0] != 10 ||
		counts[1] != 20 {
		t.Fatalf(
			"distribuição=%v esperado=[10 20]",
			counts,
		)
	}
}

func TestGIFFramePlanUsesFallbackForZeroDelay(t *testing.T) {
	plan := gifFramePlan(
		[]int{0, 0},
		2,
		30,
	)

	counts := [2]int{}

	for _, source := range plan {
		counts[source]++
	}

	if counts[0] != 15 ||
		counts[1] != 15 {
		t.Fatalf(
			"distribuição=%v esperado=[15 15]",
			counts,
		)
	}
}

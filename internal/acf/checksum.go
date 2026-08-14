package acf

import (
	"encoding/binary"
	"fmt"
)

const FooterMagic uint32 = 0xA55A5AA5

func StoredChecksum(
	data []byte,
) (uint32, error) {
	if len(data) < 8 {
		return 0, fmt.Errorf(
			"ACF muito pequeno: %d bytes",
			len(data),
		)
	}

	return binary.LittleEndian.Uint32(
		data[4:8],
	), nil
}

func ComputeChecksum(
	data []byte,
) uint32 {
	if len(data) < 8 {
		return 0
	}

	var xorSum uint32

	for offset := 0; offset+4 <= len(data); offset += 4 {

		if offset == 4 {
			continue
		}

		if offset == len(data)-4 {
			continue
		}

		xorSum ^=
			binary.LittleEndian.Uint32(
				data[offset : offset+4],
			)
	}

	return xorSum
}

func SetChecksum(
	data []byte,
) error {
	if len(data) < 8 {
		return fmt.Errorf(
			"ACF muito pequeno: %d bytes",
			len(data),
		)
	}

	binary.LittleEndian.PutUint32(
		data[4:8],
		0,
	)

	checksum :=
		ComputeChecksum(data)

	binary.LittleEndian.PutUint32(
		data[4:8],
		checksum,
	)

	return nil
}

func ValidateChecksum(
	data []byte,
) error {
	stored, err :=
		StoredChecksum(data)

	if err != nil {
		return err
	}

	calculated :=
		ComputeChecksum(data)

	if stored != calculated {
		return fmt.Errorf(
			"checksum inválido: armazenado=0x%08X calculado=0x%08X",
			stored,
			calculated,
		)
	}

	if len(data) < 4 {
		return fmt.Errorf(
			"footer ausente",
		)
	}

	footer :=
		binary.LittleEndian.Uint32(
			data[len(data)-4:],
		)

	if footer != FooterMagic {
		return fmt.Errorf(
			"footer inválido: 0x%08X",
			footer,
		)
	}

	return nil
}

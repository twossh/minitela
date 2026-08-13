package protocol

// CRC16IBM implements CRC-16/IBM (ARC).
//
// Polynomial reflected: 0xA001
// Initial value:        0x0000
// Known vector:
// "123456789" -> 0xBB3D
func CRC16IBM(data []byte) uint16 {
	var crc uint16

	for _, b := range data {
		crc ^= uint16(b)

		for i := 0; i < 8; i++ {
			if crc&0x0001 != 0 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}

	return crc
}

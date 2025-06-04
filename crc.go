package modbus

// CRC16 calculates CRC-16 for Modbus RTU
func CRC16(data []byte) uint16 {
	crc := uint16(0xFFFF)

	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if crc&0x0001 != 0 {
				crc >>= 1
				crc ^= 0xA001
			} else {
				crc >>= 1
			}
		}
	}

	return crc
}

// AppendCRC appends CRC to data
func AppendCRC(data []byte) []byte {
	crc := CRC16(data)
	result := make([]byte, len(data)+2)
	copy(result, data)
	result[len(data)] = byte(crc & 0xFF)
	result[len(data)+1] = byte((crc >> 8) & 0xFF)
	return result
}

// CheckCRC verifies CRC of received data
func CheckCRC(data []byte) bool {
	if len(data) < 3 {
		return false
	}

	payload := data[:len(data)-2]
	receivedCRC := uint16(data[len(data)-2]) | (uint16(data[len(data)-1]) << 8)
	calculatedCRC := CRC16(payload)

	return receivedCRC == calculatedCRC
}

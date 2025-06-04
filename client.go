package modbus

import (
	"time"
)

// Client interface defines the Modbus client operations
type Client interface {
	Connect() error
	Close() error
	ReadCoils(slaveID byte, address uint16, quantity uint16) ([]bool, error)
	ReadDiscreteInputs(slaveID byte, address uint16, quantity uint16) ([]bool, error)
	ReadHoldingRegisters(slaveID byte, address uint16, quantity uint16) ([]uint16, error)
	ReadInputRegisters(slaveID byte, address uint16, quantity uint16) ([]uint16, error)
	WriteSingleCoil(slaveID byte, address uint16, value bool) error
	WriteSingleRegister(slaveID byte, address uint16, value uint16) error
	WriteMultipleCoils(slaveID byte, address uint16, values []bool) error
	WriteMultipleRegisters(slaveID byte, address uint16, values []uint16) error
	SetTimeout(timeout time.Duration)
}

// ClientConfig holds common configuration
type ClientConfig struct {
	Timeout time.Duration
}

// Helper functions for data conversion
func bytesToBools(data []byte, quantity uint16) []bool {
	result := make([]bool, quantity)
	for i := uint16(0); i < quantity; i++ {
		byteIndex := i / 8
		bitIndex := i % 8
		if byteIndex < uint16(len(data)) {
			result[i] = (data[byteIndex] & (1 << bitIndex)) != 0
		}
	}
	return result
}

func boolsToBytes(values []bool) []byte {
	byteCount := (len(values) + 7) / 8
	result := make([]byte, byteCount)

	for i, val := range values {
		if val {
			byteIndex := i / 8
			bitIndex := i % 8
			result[byteIndex] |= 1 << bitIndex
		}
	}
	return result
}

func bytesToUint16s(data []byte) []uint16 {
	result := make([]uint16, len(data)/2)
	for i := 0; i < len(result); i++ {
		result[i] = uint16(data[i*2])<<8 | uint16(data[i*2+1])
	}
	return result
}

func uint16sToBytes(values []uint16) []byte {
	result := make([]byte, len(values)*2)
	for i, val := range values {
		result[i*2] = byte(val >> 8)
		result[i*2+1] = byte(val & 0xFF)
	}
	return result
}

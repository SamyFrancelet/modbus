package modbus

import (
	"errors"
	"fmt"
)

// Function codes
const (
	FuncCodeReadCoils              = 0x01
	FuncCodeReadDiscreteInputs     = 0x02
	FuncCodeReadHoldingRegisters   = 0x03
	FuncCodeReadInputRegisters     = 0x04
	FuncCodeWriteSingleCoil        = 0x05
	FuncCodeWriteSingleRegister    = 0x06
	FuncCodeWriteMultipleCoils     = 0x0F
	FuncCodeWriteMultipleRegisters = 0x10
)

// Exception codes
const (
	ExceptionIllegalFunction                    = 0x01
	ExceptionIllegalDataAddress                 = 0x02
	ExceptionIllegalDataValue                   = 0x03
	ExceptionSlaveDeviceFailure                 = 0x04
	ExceptionAcknowledge                        = 0x05
	ExceptionSlaveDeviceBusy                    = 0x06
	ExceptionMemoryParityError                  = 0x08
	ExceptionGatewayPathUnavailable             = 0x0A
	ExceptionGatewayTargetDeviceFailedToRespond = 0x0B
)

var (
	ErrInvalidResponse = errors.New("invalid response")
	ErrInvalidCRC      = errors.New("invalid CRC")
	ErrTimeout         = errors.New("timeout")
	ErrInvalidSlaveID  = errors.New("invalid slave ID")
	ErrInvalidAddress  = errors.New("invalid address")
	ErrInvalidQuantity = errors.New("invalid quantity")
)

// ModbusError represents a Modbus exception
type ModbusError struct {
	FunctionCode  byte
	ExceptionCode byte
}

func (e *ModbusError) Error() string {
	return fmt.Sprintf("modbus exception: function=0x%02X, exception=0x%02X",
		e.FunctionCode, e.ExceptionCode)
}

// PDU represents a Protocol Data Unit
type PDU struct {
	FunctionCode byte
	Data         []byte
}

// ADU represents an Application Data Unit
type ADU struct {
	SlaveID byte
	PDU     *PDU
}

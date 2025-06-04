package modbus

import (
	"encoding/binary"
	"fmt"
	"time"

	"go.bug.st/serial"
)

// RTUClient implements Modbus RTU client
type RTUClient struct {
	config *RTUConfig
	port   serial.Port
}

// RTUConfig holds RTU-specific configuration
type RTUConfig struct {
	Device       string
	Baud         int
	DataBits     int
	Parity       serial.Parity
	StopBits     serial.StopBits
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// NewRTUClient creates a new Modbus RTU client
func NewRTUClient(config *RTUConfig) *RTUClient {
	return &RTUClient{
		config: config,
	}
}

// Connect opens the serial port
func (c *RTUClient) Connect() error {
	mode := &serial.Mode{
		BaudRate: c.config.Baud,
		DataBits: c.config.DataBits,
		Parity:   c.config.Parity,
		StopBits: c.config.StopBits,
	}

	port, err := serial.Open(c.config.Device, mode)
	if err != nil {
		return fmt.Errorf("failed to open serial port: %w", err)
	}

	// Set read timeout if specified
	if c.config.ReadTimeout > 0 {
		err = port.SetReadTimeout(c.config.ReadTimeout)
		if err != nil {
			port.Close()
			return fmt.Errorf("failed to set read timeout: %w", err)
		}
	}

	c.port = port
	return nil
}

// Close closes the serial port
func (c *RTUClient) Close() error {
	if c.port != nil {
		return c.port.Close()
	}
	return nil
}

// SetTimeout sets the communication timeout
func (c *RTUClient) SetTimeout(timeout time.Duration) {
	c.config.ReadTimeout = timeout
	if c.port != nil {
		c.port.SetReadTimeout(timeout)
	}
}

// sendRequest sends a Modbus RTU request
func (c *RTUClient) sendRequest(slaveID byte, pdu *PDU) ([]byte, error) {
	if c.port == nil {
		return nil, fmt.Errorf("port not open")
	}

	// Build ADU
	adu := []byte{slaveID, pdu.FunctionCode}
	adu = append(adu, pdu.Data...)
	adu = AppendCRC(adu)

	// Send request
	_, err := c.port.Write(adu)
	if err != nil {
		return nil, fmt.Errorf("write failed: %w", err)
	}

	// Wait for response (RTU inter-frame delay)
	time.Sleep(time.Millisecond * 10)

	// Read response - timeout handled by port
	response := make([]byte, 260) // Max RTU frame size
	n, err := c.port.Read(response)
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}

	if n < 4 {
		return nil, ErrTimeout
	}

	// Validate CRC
	if !CheckCRC(response[:n]) {
		return nil, ErrInvalidCRC
	}

	// Remove CRC and validate slave ID
	frame := response[:n-2]
	if frame[0] != slaveID {
		return nil, ErrInvalidSlaveID
	}

	// Check for exception
	if frame[1] == (pdu.FunctionCode | 0x80) {
		if len(frame) >= 3 {
			return nil, &ModbusError{
				FunctionCode:  pdu.FunctionCode,
				ExceptionCode: frame[2],
			}
		}
		return nil, ErrInvalidResponse
	}

	return frame[2:], nil // Return data without slave ID and function code
}

// Implement the same methods as TCP client but using RTU protocol
// ReadCoils, ReadDiscreteInputs, ReadHoldingRegisters, etc.
// The implementation is identical to TCP except using sendRequest method above

func (c *RTUClient) ReadCoils(slaveID byte, address uint16, quantity uint16) ([]bool, error) {
	if quantity == 0 || quantity > 2000 {
		return nil, ErrInvalidQuantity
	}

	data := make([]byte, 4)
	binary.BigEndian.PutUint16(data[0:2], address)
	binary.BigEndian.PutUint16(data[2:4], quantity)

	pdu := &PDU{
		FunctionCode: FuncCodeReadCoils,
		Data:         data,
	}

	response, err := c.sendRequest(slaveID, pdu)
	if err != nil {
		return nil, err
	}

	if len(response) < 1 {
		return nil, ErrInvalidResponse
	}

	return bytesToBools(response[1:], quantity), nil
}

func (c *RTUClient) ReadDiscreteInputs(slaveID byte, address uint16, quantity uint16) ([]bool, error) {
	if quantity == 0 || quantity > 2000 {
		return nil, ErrInvalidQuantity
	}

	data := make([]byte, 4)
	binary.BigEndian.PutUint16(data[0:2], address)
	binary.BigEndian.PutUint16(data[2:4], quantity)

	pdu := &PDU{
		FunctionCode: FuncCodeReadDiscreteInputs,
		Data:         data,
	}

	response, err := c.sendRequest(slaveID, pdu)
	if err != nil {
		return nil, err
	}

	if len(response) < 1 {
		return nil, ErrInvalidResponse
	}

	return bytesToBools(response[1:], quantity), nil
}

func (c *RTUClient) ReadHoldingRegisters(slaveID byte, address uint16, quantity uint16) ([]uint16, error) {
	if quantity == 0 || quantity > 125 {
		return nil, ErrInvalidQuantity
	}

	data := make([]byte, 4)
	binary.BigEndian.PutUint16(data[0:2], address)
	binary.BigEndian.PutUint16(data[2:4], quantity)

	pdu := &PDU{
		FunctionCode: FuncCodeReadHoldingRegisters,
		Data:         data,
	}

	response, err := c.sendRequest(slaveID, pdu)
	if err != nil {
		return nil, err
	}

	if len(response) < 1 {
		return nil, ErrInvalidResponse
	}

	return bytesToUint16s(response[1:]), nil
}

func (c *RTUClient) ReadInputRegisters(slaveID byte, address uint16, quantity uint16) ([]uint16, error) {
	if quantity == 0 || quantity > 125 {
		return nil, ErrInvalidQuantity
	}

	data := make([]byte, 4)
	binary.BigEndian.PutUint16(data[0:2], address)
	binary.BigEndian.PutUint16(data[2:4], quantity)

	pdu := &PDU{
		FunctionCode: FuncCodeReadInputRegisters,
		Data:         data,
	}

	response, err := c.sendRequest(slaveID, pdu)
	if err != nil {
		return nil, err
	}

	if len(response) < 1 {
		return nil, ErrInvalidResponse
	}

	return bytesToUint16s(response[1:]), nil
}

func (c *RTUClient) WriteSingleCoil(slaveID byte, address uint16, value bool) error {
	data := make([]byte, 4)
	binary.BigEndian.PutUint16(data[0:2], address)
	if value {
		binary.BigEndian.PutUint16(data[2:4], 0xFF00)
	} else {
		binary.BigEndian.PutUint16(data[2:4], 0x0000)
	}

	pdu := &PDU{
		FunctionCode: FuncCodeWriteSingleCoil,
		Data:         data,
	}

	_, err := c.sendRequest(slaveID, pdu)
	return err
}

func (c *RTUClient) WriteSingleRegister(slaveID byte, address uint16, value uint16) error {
	data := make([]byte, 4)
	binary.BigEndian.PutUint16(data[0:2], address)
	binary.BigEndian.PutUint16(data[2:4], value)

	pdu := &PDU{
		FunctionCode: FuncCodeWriteSingleRegister,
		Data:         data,
	}

	_, err := c.sendRequest(slaveID, pdu)
	return err
}

func (c *RTUClient) WriteMultipleCoils(slaveID byte, address uint16, values []bool) error {
	if len(values) == 0 || len(values) > 1968 {
		return ErrInvalidQuantity
	}

	byteCount := (len(values) + 7) / 8
	data := make([]byte, 5+byteCount)
	binary.BigEndian.PutUint16(data[0:2], address)
	binary.BigEndian.PutUint16(data[2:4], uint16(len(values)))
	data[4] = byte(byteCount)

	coilBytes := boolsToBytes(values)
	copy(data[5:], coilBytes)

	pdu := &PDU{
		FunctionCode: FuncCodeWriteMultipleCoils,
		Data:         data,
	}

	_, err := c.sendRequest(slaveID, pdu)
	return err
}

func (c *RTUClient) WriteMultipleRegisters(slaveID byte, address uint16, values []uint16) error {
	if len(values) == 0 || len(values) > 123 {
		return ErrInvalidQuantity
	}

	data := make([]byte, 5+len(values)*2)
	binary.BigEndian.PutUint16(data[0:2], address)
	binary.BigEndian.PutUint16(data[2:4], uint16(len(values)))
	data[4] = byte(len(values) * 2)

	regBytes := uint16sToBytes(values)
	copy(data[5:], regBytes)

	pdu := &PDU{
		FunctionCode: FuncCodeWriteMultipleRegisters,
		Data:         data,
	}

	_, err := c.sendRequest(slaveID, pdu)
	return err
}

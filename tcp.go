package modbus

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync/atomic"
	"time"
)

// TCPClient implements Modbus TCP client
type TCPClient struct {
	address       string
	conn          net.Conn
	timeout       time.Duration
	transactionID uint32
}

// NewTCPClient creates a new Modbus TCP client
func NewTCPClient(address string) *TCPClient {
	return &TCPClient{
		address: address,
		timeout: 5 * time.Second,
	}
}

// Connect establishes TCP connection
func (c *TCPClient) Connect() error {
	conn, err := net.DialTimeout("tcp", c.address, c.timeout)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	c.conn = conn
	return nil
}

// Close closes the TCP connection
func (c *TCPClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// SetTimeout sets the communication timeout
func (c *TCPClient) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// sendRequest sends a Modbus TCP request
func (c *TCPClient) sendRequest(slaveID byte, pdu *PDU) ([]byte, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	// Generate transaction ID
	transID := uint16(atomic.AddUint32(&c.transactionID, 1))

	// Build MBAP header
	mbap := make([]byte, 7)
	binary.BigEndian.PutUint16(mbap[0:2], transID)                 // Transaction ID
	binary.BigEndian.PutUint16(mbap[2:4], 0)                       // Protocol ID
	binary.BigEndian.PutUint16(mbap[4:6], uint16(2+len(pdu.Data))) // Length
	mbap[6] = slaveID                                              // Unit ID

	// Build complete request
	request := append(mbap, pdu.FunctionCode)
	request = append(request, pdu.Data...)

	// Set write timeout
	c.conn.SetWriteDeadline(time.Now().Add(c.timeout))
	_, err := c.conn.Write(request)
	if err != nil {
		return nil, fmt.Errorf("write failed: %w", err)
	}

	// Read response
	c.conn.SetReadDeadline(time.Now().Add(c.timeout))
	header := make([]byte, 7)
	_, err = c.conn.Read(header)
	if err != nil {
		return nil, fmt.Errorf("read header failed: %w", err)
	}

	// Parse MBAP header
	respTransID := binary.BigEndian.Uint16(header[0:2])
	length := binary.BigEndian.Uint16(header[4:6])

	if respTransID != transID {
		return nil, ErrInvalidResponse
	}

	// Read PDU
	pduData := make([]byte, length-1) // -1 for unit ID already read
	_, err = c.conn.Read(pduData)
	if err != nil {
		return nil, fmt.Errorf("read PDU failed: %w", err)
	}

	// Check for exception
	if pduData[0] == (pdu.FunctionCode | 0x80) {
		if len(pduData) >= 2 {
			return nil, &ModbusError{
				FunctionCode:  pdu.FunctionCode,
				ExceptionCode: pduData[1],
			}
		}
		return nil, ErrInvalidResponse
	}

	return pduData[1:], nil // Return data without function code
}

// ReadCoils reads coil status
func (c *TCPClient) ReadCoils(slaveID byte, address uint16, quantity uint16) ([]bool, error) {
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

// ReadDiscreteInputs reads discrete input status
func (c *TCPClient) ReadDiscreteInputs(slaveID byte, address uint16, quantity uint16) ([]bool, error) {
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

// ReadHoldingRegisters reads holding registers
func (c *TCPClient) ReadHoldingRegisters(slaveID byte, address uint16, quantity uint16) ([]uint16, error) {
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

// ReadInputRegisters reads input registers
func (c *TCPClient) ReadInputRegisters(slaveID byte, address uint16, quantity uint16) ([]uint16, error) {
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

// WriteSingleCoil writes a single coil
func (c *TCPClient) WriteSingleCoil(slaveID byte, address uint16, value bool) error {
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

// WriteSingleRegister writes a single register
func (c *TCPClient) WriteSingleRegister(slaveID byte, address uint16, value uint16) error {
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

// WriteMultipleCoils writes multiple coils
func (c *TCPClient) WriteMultipleCoils(slaveID byte, address uint16, values []bool) error {
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

// WriteMultipleRegisters writes multiple registers
func (c *TCPClient) WriteMultipleRegisters(slaveID byte, address uint16, values []uint16) error {
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

package main

import (
	"fmt"
	"log"

	"github.com/SamyFrancelet/modbus"
)

func main() {
	// Create TCP client
	client := modbus.NewTCPClient("localhost:502")

	// Connect
	err := client.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	slaveID := byte(1)

	// Read coils at address 0
	coils, err := client.ReadCoils(slaveID, 0, 10)
	if err != nil {
		log.Printf("Error reading coils: %v", err)
	} else {
		fmt.Printf("Coils (0-9): %v\n", coils)
	}

	// Read discrete inputs at address 0
	inputs, err := client.ReadDiscreteInputs(slaveID, 0, 10)
	if err != nil {
		log.Printf("Error reading discrete inputs: %v", err)
	} else {
		fmt.Printf("Discrete inputs (0-9): %v\n", inputs)
	}

	// Read holding registers at address 0
	holdingRegs, err := client.ReadHoldingRegisters(slaveID, 0, 10)
	if err != nil {
		log.Printf("Error reading holding registers: %v", err)
	} else {
		fmt.Printf("Holding registers (0-9): %v\n", holdingRegs)
	}

	// Read input registers at address 0
	inputRegs, err := client.ReadInputRegisters(slaveID, 0, 10)
	if err != nil {
		log.Printf("Error reading input registers: %v", err)
	} else {
		fmt.Printf("Input registers (0-9): %v\n", inputRegs)
	}

	// Write to coil 1
	err = client.WriteSingleCoil(slaveID, 1, true)
	if err != nil {
		log.Printf("Error writing coil: %v", err)
	} else {
		fmt.Println("Successfully wrote true to coil 1")
	}

	// Write to holding register 1
	err = client.WriteSingleRegister(slaveID, 1, 1234)
	if err != nil {
		log.Printf("Error writing register: %v", err)
	} else {
		fmt.Println("Successfully wrote 1234 to holding register 1")
	}
}

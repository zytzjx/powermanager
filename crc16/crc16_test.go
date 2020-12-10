package crc16

import "testing"

func TestCrcModbus(t *testing.T) {
	// var data = make([]byte, 8)
	// data[0] = 0x01
	// data[1] = 0x06
	// data[2] = 0x00
	// data[3] = 0x01
	// data[4] = 0x00
	// data[5] = 0x00
	testData := []byte{0x01, 0x06, 0x00, 0x01, 0x00, 0x01}
	table := MakeTable(Crc16MODBUS)
	if table == nil {
		t.Errorf("Failed to create %q computer", Crc16MODBUS.Name)
	}
	// 0x534B
	crc := Checksum(testData, table)

	if crc != table.params.Check {
		t.Errorf("Invalid %q sample calculation, expected: %X, actual: %X", table.params.Name, table.params.Check, crc)
	}
}

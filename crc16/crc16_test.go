package crc16

import "testing"

func TestCrcModbus(t *testing.T) {
	testData := []byte{0x01, 0x06, 0x00, 0x30, 0x04, 0xC9}
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

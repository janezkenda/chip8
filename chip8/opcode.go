package chip8

import "fmt"

type OpCode struct {
	B0  byte
	B00 byte
	B01 byte
	B1  byte
	B10 byte
	B11 byte
}

func (op OpCode) String() string {
	return fmt.Sprintf("0x%02x%02x", op.B0, op.B1)
}

func (op OpCode) Addr() uint16 {
	return uint16(op.B01)<<8 | uint16(op.B1)
}

func NewOpCode(op [2]byte) OpCode {
	return OpCode{
		B0:  op[0],
		B00: op[0] >> 4,
		B01: op[0] & 0x0f,
		B1:  op[1],
		B10: op[1] >> 4,
		B11: op[1] & 0x0f,
	}
}

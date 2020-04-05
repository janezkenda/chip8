package chip8

import (
	"fmt"
)

func DisassembleChip8Op(op OpCode) {
	fmt.Printf("0x%02x%02x ", op.B0, op.B1)

	switch op.B00 {
	case 0x00:
		switch op.B1 {
		case 0xe0:
			fmt.Println("Clear the screen")
		case 0xee:
			fmt.Println("Return from a subroutine")
		default:
			fmt.Println("Unknown 0")
		}
	case 0x01:
		fmt.Printf("Jump to address 0x%1x%02x\n", op.B01, op.B1)
	case 0x02:
		fmt.Printf("Call subroutine at 0x%1x%02x\n", op.B01, op.B1)
	case 0x03:
		fmt.Printf("Skip next instruction if V%1x == 0x%02x\n", op.B01, op.B1)
	case 0x04:
		fmt.Printf("Skip next instruction if V%1x != 0x%02x\n", op.B01, op.B1)
	case 0x05:
		fmt.Printf("Skip next instruction if V%1x == V%1x\n", op.B01, op.B10)
	case 0x06:
		fmt.Printf("V%1x = 0x%02x\n", op.B01, op.B1)
	case 0x07:
		fmt.Printf("V%1x += 0x%02x \n", op.B01, op.B1)
	case 0x08:
		switch op.B11 {
		case 0x00:
			fmt.Printf("V%1x = V%1x\n", op.B01, op.B10)
		case 0x01:
			fmt.Printf("V%1x = V%1x | V%1x\n", op.B01, op.B01, op.B10)
		case 0x02:
			fmt.Printf("V%1x = V%1x & V%1x\n", op.B01, op.B01, op.B10)
		case 0x03:
			fmt.Printf("V%1x = V%1x ^ V%1x\n", op.B01, op.B01, op.B10)
		case 0x04:
			fmt.Printf("V%1x += V%1x. VF is set to 1 when there's a carry, and to 0 when there isn't\n", op.B01, op.B10)
		case 0x05:
			fmt.Printf("V%1x -= V%1x. VF is set to 0 when there's a borrow, and 1 when there isn't\n", op.B01, op.B10)
		case 0x06:
			fmt.Printf("Stores the least significant bit of V%1x in VF and then shifts VX to the right by 1\n", op.B01)
		case 0x07:
			fmt.Printf("V%1x=V%1x-V%1x. VF is set to 0 when there's a borrow, and 1 when there isn't\n", op.B01, op.B10, op.B01)
		case 0x0e:
			fmt.Printf("Stores the most significant bit of V%1x in VF and then shifts VX to the left by 1\n", op.B01)
		default:
			fmt.Printf("%1x%1x not implemented\n", op.B00, op.B11)
		}
	case 0x09:
		fmt.Printf("Skip next instruction if V%1x != V%1x\n", op.B01, op.B10)
	case 0x0a:
		fmt.Printf("I = 0x%1x%02x\n", op.B01, op.B1)
	case 0x0b:
		fmt.Printf("PC = V0 + 0x%1x%02x\n", op.B01, op.B1)
	case 0x0c:
		fmt.Printf("V%1x=random(0,255) & 0x%02x\n", op.B01, op.B1)
	case 0x0d:
		fmt.Printf("Draw 8x%1x sprite at (V%1x, V%1x)\n", op.B11, op.B01, op.B10)
	case 0x0e:
		switch op.B1 {
		case 0x9e:
			fmt.Printf("Skip next instruction if key stored in V%1x is pressed\n", op.B01)
		case 0xa1:
			fmt.Printf("Skip next instruction if key stored in V%1x is not pressed\n", op.B01)
		default:
			fmt.Printf("%02x%02x not implemented\n", op.B0, op.B1)
		}
	case 0x0f:
		switch op.B1 {
		case 0x07:
			fmt.Printf("Set V%1x to the value of delay timer\n", op.B01)
		case 0x0a:
			fmt.Printf("A key press is awaited, and then stored in V%1x. (Blocking Operation. All instruction halted until next key event)\n", op.B01)
		case 0x15:
			fmt.Printf("Set delay timer to V%1x\n", op.B01)
		case 0x18:
			fmt.Printf("Set sound timer to V%1x\n", op.B01)
		case 0x1e:
			fmt.Printf("Adds V%1x to I. VF is set to 1 when there is a range overflow (I+V%1x>0xFFF), and to 0 when there isn't\n", op.B01, op.B01)
		case 0x29:
			fmt.Printf("Sets I to the location of the sprite for the character in V%1x. Characters 0-F (in hexadecimal) are represented by a 4x5 font\n", op.B01)
		case 0x33:
			fmt.Printf("Take the decimal representation of V%1x, place the hundreds digit in memory at location in I, the tens digit at location I+1, and the ones digit at location I+2\n", op.B01)
		case 0x55:
			fmt.Printf("Stores V0 to (including) V%1x in memory starting at address I. The offset from I is increased by 1 for each value written, but I itself is left unmodified\n", op.B01)
		case 0x65:
			fmt.Printf("Fills V0 to (including) V%1x with values from memory starting at address I. The offset from I is increased by 1 for each value written, but I itself is left unmodified\n", op.B01)
		default:
			fmt.Printf("%02x%02x not implemented\n", op.B0, op.B1)
		}
	default:
		fmt.Printf("%x not implemented yet\n", op.B00)
	}
}

func DisassembleProgram(program []byte) {
	p := append(make([]byte, 0x200), program...)
	for pc := 0x200; pc < len(p); pc += 2 {
		op := NewOpCode([2]byte{p[pc], p[pc+1]})
		fmt.Printf("0x%04x 0x%02x 0x%02x\t", pc, op.B0, op.B1)
		DisassembleChip8Op(op)
	}
}

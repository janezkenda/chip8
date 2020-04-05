package chip8

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChip8State_Run_op0(t *testing.T) {
	t.Run("00E0", func(t *testing.T) {
		op := NewOpCode([2]byte{0x00, 0xe0}) // Clear the screen

		c8 := Init(nil)
		for i := range c8.memory[0xf00:] {
			c8.memory[0xf00+i] = 0x01
		}

		c8.RunOp(op)

		var sum byte
		for _, b := range c8.memory[0xf00:] {
			sum += b
		}

		assert.Equal(t, byte(0), sum, "expected screen to be cleared")
	})

	t.Run("00EE", func(t *testing.T) {
		op := NewOpCode([2]byte{0x00, 0xee}) // Return from a subroutine

		c8 := Init(nil)
		c8.SP = 0xf9e
		// Put address 0x300 on stack
		c8.memory[c8.SP] = 0x03
		c8.memory[c8.SP+1] = 0x00
		c8.RunOp(op)

		assert.Equal(t, uint16(0xfa0), c8.SP, "expected SP to be set to 0xfa0")
		assert.Equal(t, uint16(0x302), c8.PC, "expected PC to be set to 0x302")
	})
}

func TestChip8State_Run_op1(t *testing.T) {
	op := NewOpCode([2]byte{0x13, 0x00}) // Jump to address 0x300

	c8 := Init(nil)
	c8.RunOp(op)

	assert.Equal(t, uint16(0x300), c8.PC, "expected PC to be set to 0x300")

	op = NewOpCode([2]byte{0x12, 0x00}) // Jump to address 0x200, causing an infinite loop

	c8 = Init(nil)
	c8.RunOp(op)

	assert.True(t, c8.halt, "expected the halt flag to be set because of infinite loop")
}

func TestChip8State_Run_op2(t *testing.T) {
	op := NewOpCode([2]byte{0x23, 0x00}) // Call subroutine at 0x300

	c8 := Init(nil)
	c8.RunOp(op)

	assert.Equal(t, uint16(0xf9e), c8.SP, "expected SP to be set to 0xf9e")
	assert.Equal(t, uint16(0x300), c8.PC, "expected PC to be set to 0x300")
	assert.Equal(t, []byte{0x2, 0x00}, []byte{c8.memory[c8.SP], c8.memory[c8.SP+1]}, "expected address to be on stack")
}

func TestChip8State_Run_op3(t *testing.T) {
	op := NewOpCode([2]byte{0x30, 0x01}) // Skip next instruction if V0 == 0x01

	c8 := Init(nil)
	c8.V[0x0] = 0x01

	c8.RunOp(op)

	assert.Equal(t, uint16(0x204), c8.PC, "expected instruction skip - PC to be set to 0x204")

	c8 = Init(nil)
	c8.V[0x0] = 0x00

	c8.RunOp(op)

	assert.Equal(t, uint16(0x202), c8.PC, "expected no skip - PC to be set to 0x202")
}

func TestChip8State_Run_op4(t *testing.T) {
	op := NewOpCode([2]byte{0x40, 0x01}) // Skip next instruction if V0 != 0x01

	c8 := Init(nil)
	c8.V[0x0] = 0x02

	c8.RunOp(op)

	assert.Equal(t, uint16(0x204), c8.PC, "expected instruction skip - PC to be set to 0x204")

	c8 = Init(nil)
	c8.V[0x0] = 0x01

	c8.RunOp(op)

	assert.Equal(t, uint16(0x202), c8.PC, "expected no skip - PC to be set to 0x202")
}

func TestChip8State_Run_op5(t *testing.T) {
	op := NewOpCode([2]byte{0x50, 0x10}) // Skip next instruction if V0 == V1

	c8 := Init(nil)
	c8.V[0x0] = 0x01
	c8.V[0x1] = 0x01

	c8.RunOp(op)

	assert.Equal(t, uint16(0x204), c8.PC, "expected instruction skip - PC to be set to 0x204")

	c8 = Init(nil)
	c8.V[0x0] = 0x01
	c8.V[0x1] = 0x02

	c8.RunOp(op)

	assert.Equal(t, uint16(0x202), c8.PC, "expected no skip - PC to be set to 0x202")
}

func TestChip8State_Run_op6(t *testing.T) {
	op := NewOpCode([2]byte{0x60, 0x22}) // V0 = 0x22

	c8 := Init(nil)
	c8.RunOp(op)

	assert.Equal(t, byte(0x22), c8.V[0x0], "expected V0 to be set to 0x22")
	assert.Equal(t, uint16(0x202), c8.PC, "expected PC to be set to 0x202")
}

func TestChip8State_Run_op7(t *testing.T) {
	op := NewOpCode([2]byte{0x70, 0x22}) // V0 += 0x22

	c8 := Init(nil)
	c8.V[0x0] = 0x22

	c8.RunOp(op)

	assert.Equal(t, byte(0x44), c8.V[0x0], "expected V0 to be set to 0x44")
	assert.Equal(t, uint16(0x202), c8.PC, "expected PC to be set to 0x202")
}

func TestChip8State_Run_op8(t *testing.T) {
	t.Run("8XY0", func(t *testing.T) {
		op := NewOpCode([2]byte{0x80, 0x10}) // V0 = V1

		c8 := Init(nil)
		c8.V[0x0] = 0x01
		c8.V[0x1] = 0x02

		c8.RunOp(op)

		assert.Equal(t, byte(0x02), c8.V[0x00], "expected V0 to contain 0x02")
	})

	t.Run("8XY1", func(t *testing.T) {
		op := NewOpCode([2]byte{0x80, 0x11}) // V0 |= V1

		c8 := Init(nil)
		c8.V[0x0] = 0x01
		c8.V[0x1] = 0x02

		c8.RunOp(op)

		assert.Equal(t, byte(0x03), c8.V[0x00], "expected V0 to contain 0x03")
	})

	t.Run("8XY2", func(t *testing.T) {
		op := NewOpCode([2]byte{0x80, 0x12}) // V0 &= V1

		c8 := Init(nil)
		c8.V[0x0] = 0x01
		c8.V[0x1] = 0x03

		c8.RunOp(op)

		assert.Equal(t, byte(0x01), c8.V[0x00], "expected V0 to contain 0x01")
	})

	t.Run("8XY3", func(t *testing.T) {
		op := NewOpCode([2]byte{0x80, 0x13}) // V0 ^= V1

		c8 := Init(nil)
		c8.V[0x0] = 0x01
		c8.V[0x1] = 0x03

		c8.RunOp(op)

		assert.Equal(t, byte(0x02), c8.V[0x00], "expected V0 to contain 0x02")
	})

	t.Run("8XY4", func(t *testing.T) {
		op := NewOpCode([2]byte{0x80, 0x14}) // V0 += V1

		// No carry
		c8 := Init(nil)
		c8.V[0x0] = 0x01
		c8.V[0x1] = 0x03

		c8.RunOp(op)

		assert.Equal(t, byte(0x04), c8.V[0x00], "expected V0 to contain 0x04")
		assert.Equal(t, byte(0x00), c8.V[0x0f], "expected Vf to contain 0x00 (no carry)")

		// Carry
		c8 = Init(nil)
		c8.V[0x0] = 0xff
		c8.V[0x1] = 0x03

		c8.RunOp(op)

		assert.Equal(t, byte(0x02), c8.V[0x00], "expected V0 to contain 0x02")
		assert.Equal(t, byte(0x01), c8.V[0x0f], "expected Vf to contain 0x01 (carry)")
	})

	t.Run("8XY5", func(t *testing.T) {
		op := NewOpCode([2]byte{0x80, 0x15}) // V0 -= V1

		// No borrow
		c8 := Init(nil)
		c8.V[0x0] = 0x03
		c8.V[0x1] = 0x01

		c8.RunOp(op)

		assert.Equal(t, byte(0x02), c8.V[0x00], "expected V0 to contain 0x02")
		assert.Equal(t, byte(0x01), c8.V[0x0f], "expected Vf to contain 0x01 (no borrow)")

		// Borrow
		c8 = Init(nil)
		c8.V[0x0] = 0x01
		c8.V[0x1] = 0x03

		c8.RunOp(op)

		assert.Equal(t, byte(0xfe), c8.V[0x00], "expected V0 to contain 0xfe")
		assert.Equal(t, byte(0x00), c8.V[0x0f], "expected Vf to contain 0x00 (borrow)")
	})

	t.Run("8XY6", func(t *testing.T) {
		op := NewOpCode([2]byte{0x80, 0x06}) // V0 >>= 1

		c8 := Init(nil)
		c8.V[0x0] = 0x8b

		c8.RunOp(op)

		assert.Equal(t, byte(0x45), c8.V[0x00], "expected V0 to contain 0x45")
		assert.Equal(t, byte(0x01), c8.V[0x0f], "expected Vf to contain 0x01")
	})

	t.Run("8XY7", func(t *testing.T) {
		op := NewOpCode([2]byte{0x80, 0x17}) // V0 = V1 - V0

		// No borrow
		c8 := Init(nil)
		c8.V[0x0] = 0x01
		c8.V[0x1] = 0x03

		c8.RunOp(op)

		assert.Equal(t, byte(0x02), c8.V[0x00], "expected V0 to contain 0x02")
		assert.Equal(t, byte(0x01), c8.V[0x0f], "expected Vf to contain 0x01 (no borrow)")

		// Borrow
		c8 = Init(nil)
		c8.V[0x0] = 0x03
		c8.V[0x1] = 0x01

		c8.RunOp(op)

		assert.Equal(t, byte(0xfe), c8.V[0x00], "expected V0 to contain 0xfe")
		assert.Equal(t, byte(0x00), c8.V[0x0f], "expected Vf to contain 0x00 (borrow)")
	})

	t.Run("8XYE", func(t *testing.T) {
		op := NewOpCode([2]byte{0x80, 0x0e}) // V0 <<= 1

		c8 := Init(nil)
		c8.V[0x0] = 0x8b

		c8.RunOp(op)

		assert.Equal(t, byte(0x16), c8.V[0x00], "expected V0 to contain 0x16")
		assert.Equal(t, byte(0x01), c8.V[0x0f], "expected Vf to contain 0x01")
	})
}

func TestChip8State_Run_op9(t *testing.T) {
	op := NewOpCode([2]byte{0x90, 0x10}) // Skip next instruction if V0 != V1

	c8 := Init(nil)
	c8.V[0x0] = 0x01
	c8.V[0x1] = 0x02

	c8.RunOp(op)

	assert.Equal(t, uint16(0x204), c8.PC, "expected instruction skip - PC to be set to 0x204")

	c8 = Init(nil)
	c8.V[0x0] = 0x01
	c8.V[0x1] = 0x01

	c8.RunOp(op)

	assert.Equal(t, uint16(0x202), c8.PC, "expected no skip - PC to be set to 0x202")
}

func TestChip8State_Run_opA(t *testing.T) {
	op := NewOpCode([2]byte{0xa1, 0x00}) // Set i to 0x100

	c8 := Init(nil)
	c8.RunOp(op)

	assert.Equal(t, uint16(0x100), c8.I, "expected I to be set to 0x100")
}

func TestChip8State_Run_opB(t *testing.T) {
	op := NewOpCode([2]byte{0xb1, 0x00}) // Jump to 0x100 + V0

	c8 := Init(nil)
	c8.V[0x0] = 0x11

	c8.RunOp(op)

	assert.Equal(t, uint16(0x111), c8.PC, "expected PC to be set to 0x111")
}

func TestChip8State_Run_opC(t *testing.T) {
	op := NewOpCode([2]byte{0xc0, 0x13}) // // VX = random(0,255) & 0x13

	c8 := Init(nil)
	c8.RunOp(op)

	assert.Equal(t, byte(0x12), c8.V[0x0], "expected V0 to be set to 0x12")
}

func TestChip8State_Run_opD(t *testing.T) {
	t.Run("no collision", func(t *testing.T) {
		program := []byte{
			0x60, 0x0e, // Set V0 to 0x0e
			0xf0, 0x29, // Load the address of sprite in V0 (E) into I
			0x61, 0x08, // Set V1 to 0x08
			0x62, 0x08, // Set V2 to 0x08
			0xd1, 0x25, // Draw sprite from address I to (V1, V2), with the height of 5
		}

		c8 := Init(nil)
		c8.RunProgram(program)

		// TODO improve the test
		sprite := make([]byte, 0)
		for _, b := range c8.memory[0xf00:] {
			if b != byte(0x0) {
				sprite = append(sprite, b)
			}
		}

		spriteE := font[0x0e*5 : (0x0e*5)+5]
		for i, b := range spriteE {
			assert.Equal(t, b, sprite[i], "expected bytes in display buffer to match")
		}
		assert.Equal(t, byte(0x00), c8.V[0x0f], "expected no collision")
	})

	t.Run("collision", func(t *testing.T) {
		program := []byte{
			0x60, 0x0e, // Set V0 to 0x0e
			0xf0, 0x29, // Load the address of sprite in V0 (E) into I
			0x61, 0x08, // Set V1 to 0x08
			0x62, 0x08, // Set V2 to 0x08
			0xd1, 0x25, // Draw sprite from address I to (V1, V2), with the height of 5
			0x60, 0x00, // Set V0 t0 0x00
			0xf0, 0x29, // Load the address of sprite in V0 (E) into I
			0xd1, 0x25, // Draw sprite from address I to (V1, V2), with the height of 5
		}

		c8 := Init(nil)
		c8.RunProgram(program)

		// TODO improve the test
		sprite := make([]byte, 0)
		for _, b := range c8.memory[0xf00:] {
			if b != byte(0x0) {
				sprite = append(sprite, b)
			}
		}

		overlaidSprite := []byte{
			0b10010000,
			0b00010000,
			0b01110000,
			0b00010000,
			0b10010000,
		}

		for i, b := range overlaidSprite {
			assert.Equal(t, b, sprite[i], "expected bytes in display buffer to match")
		}
		assert.Equal(t, byte(0x01), c8.V[0x0f], "expected collision")
	})
}

func TestChip8State_Run_opE(t *testing.T) {
	t.Run("EX9E", func(t *testing.T) {
		op := NewOpCode([2]byte{0xe0, 0x9e}) // Skip next instruction if key stored in V0 is pressed

		c8 := Init(nil)
		c8.V[0x0] = 0x0a
		c8.keyboard[c8.V[0x0]] = true

		c8.RunOp(op)

		assert.Equal(t, uint16(0x204), c8.PC, "expected instruction skip - PC to be set to 0x204")

		c8 = Init(nil)
		c8.V[0x0] = 0x0a
		c8.keyboard[c8.V[0x0]] = false

		c8.RunOp(op)

		assert.Equal(t, uint16(0x202), c8.PC, "expected no skip - PC to be set to 0x202")
	})

	t.Run("EXA1", func(t *testing.T) {
		op := NewOpCode([2]byte{0xe0, 0xa1}) // Skip next instruction if key stored in V0 isn't pressed

		c8 := Init(nil)
		c8.V[0x0] = 0x0a
		c8.keyboard[c8.V[0x0]] = false

		c8.RunOp(op)

		assert.Equal(t, uint16(0x204), c8.PC, "expected instruction skip - PC to be set to 0x204")

		c8 = Init(nil)
		c8.V[0x0] = 0x0a
		c8.keyboard[c8.V[0x0]] = true

		c8.RunOp(op)

		assert.Equal(t, uint16(0x202), c8.PC, "expected no skip - PC to be set to 0x202")
	})
}

func TestChip8State_Run_opF(t *testing.T) {
	t.Run("FX07", func(t *testing.T) {
		op := NewOpCode([2]byte{0xf0, 0x07}) // Set V0 to the value of delay timer

		c8 := Init(nil)
		c8.delayTimer = 0xaa

		c8.RunOp(op)

		assert.Equal(t, byte(0xaa), c8.V[0x0], "expected V0 to be set to 0xaa")
	})

	t.Run("FX0A", func(t *testing.T) {

	})

	t.Run("FX15", func(t *testing.T) {
		op := NewOpCode([2]byte{0xf0, 0x15}) // Set delay timer to the value of V0

		c8 := Init(nil)
		c8.V[0x0] = 0xaa

		c8.RunOp(op)

		assert.Equal(t, byte(0xaa), c8.delayTimer, "expected delay timer to be set to 0xaa")
	})

	t.Run("FX18", func(t *testing.T) {
		op := NewOpCode([2]byte{0xf0, 0x18}) // Set delay timer to the value of V0

		c8 := Init(nil)
		c8.V[0x0] = 0xaa

		c8.RunOp(op)

		assert.Equal(t, byte(0xaa), c8.soundTimer, "expected sound timer to be set to 0xaa")
	})

	t.Run("FX1E", func(t *testing.T) {
		op := NewOpCode([2]byte{0xf0, 0x1e}) // Add the value of V0 to I

		// No range overflow
		c8 := Init(nil)
		c8.V[0x0] = 0x05

		c8.RunOp(op)

		assert.Equal(t, uint16(0x05), c8.I, "expected I to be set to 0x05")
		assert.Equal(t, byte(0x00), c8.V[0x0f], "expected no range overflow")

		// Range overflow
		c8 = Init(nil)
		c8.I = 0xfff
		c8.V[0x0] = 0x05

		c8.RunOp(op)

		assert.Equal(t, uint16(0x04), c8.I, "expected I to be set to 0x04")
		assert.Equal(t, byte(0x01), c8.V[0x0f], "expected range overflow")
	})

	t.Run("FX29", func(t *testing.T) {
		op := NewOpCode([2]byte{0xf0, 0x29}) // Set I to the location of sprite of the value in V0

		c8 := Init(nil)
		c8.V[0x0] = 0x0e

		c8.RunOp(op)

		assert.Equal(t, uint16(5*0x0e), c8.I, "expected I to be set to 0x46")
	})

	t.Run("FX33", func(t *testing.T) {
		op := NewOpCode([2]byte{0xf0, 0x33}) // Store a decimal representation of the value in V0 to the addresses i, i+1 and i+2

		c8 := Init(nil)
		c8.I = 0x300
		c8.V[0x0] = 0x7b // 0x7b = decimal 123

		c8.RunOp(op)

		assert.Equal(t, byte(0x01), c8.memory[c8.I], "expected mem[I] to be set to 0x01")
		assert.Equal(t, byte(0x02), c8.memory[c8.I+1], "expected mem[I+1] to be set to 0x02")
		assert.Equal(t, byte(0x03), c8.memory[c8.I+2], "expected mem[I+2] to be set to 0x03'")
	})

	t.Run("FX55", func(t *testing.T) {
		op := NewOpCode([2]byte{0xf3, 0x55}) // Store values from V0:V3 to mem[I:I+3]

		c8 := Init(nil)
		c8.I = 0x300
		c8.V[0x0] = 0x0a
		c8.V[0x1] = 0x0b
		c8.V[0x2] = 0x0c
		c8.V[0x3] = 0x0d
		c8.V[0x4] = 0x0e

		c8.RunOp(op)

		assert.Equal(t, byte(0x0a), c8.memory[c8.I], "expected mem[I] to be set to 0x0a")
		assert.Equal(t, byte(0x0b), c8.memory[c8.I+1], "expected mem[I+1] to be set to 0x0b")
		assert.Equal(t, byte(0x0c), c8.memory[c8.I+2], "expected mem[I+2] to be set to 0x0c")
		assert.Equal(t, byte(0x0d), c8.memory[c8.I+3], "expected mem[I+3] to be set to 0x0d")
		assert.Equal(t, byte(0x00), c8.memory[c8.I+4], "expected mem[I+4] to be set to 0x00")
	})

	t.Run("FX65", func(t *testing.T) {
		op := NewOpCode([2]byte{0xf3, 0x65}) // Store values from mem[I:I+3] to V0:V3

		c8 := Init(nil)
		c8.I = 0x300
		c8.memory[c8.I] = 0x0a
		c8.memory[c8.I+1] = 0x0b
		c8.memory[c8.I+2] = 0x0c
		c8.memory[c8.I+3] = 0x0d
		c8.memory[c8.I+4] = 0x0e

		c8.RunOp(op)

		assert.Equal(t, byte(0x0a), c8.V[0x0], "expected V0 to be set to 0x0a")
		assert.Equal(t, byte(0x0b), c8.V[0x1], "expected V1 to be set to 0x0b")
		assert.Equal(t, byte(0x0c), c8.V[0x2], "expected V2 to be set to 0x0c")
		assert.Equal(t, byte(0x0d), c8.V[0x3], "expected V3 to be set to 0x0d")
		assert.Equal(t, byte(0x00), c8.V[0x4], "expected V4 to be set to 0x00")
	})
}

package chip8

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"time"

	"golang.org/x/image/draw"
)

const screenAddress = 0xf00

type Machine interface {
	GetKeyAt(key byte) bool
	WaitForKeyPress() byte
}

type keyEvent struct {
	key     byte
	pressed bool
}

type State struct {
	memory [4096]byte

	// Registers
	V  [16]byte
	I  uint16
	SP uint16
	PC uint16

	// Timers
	delayTimer byte
	soundTimer byte

	// Keyboard
	keyChannel chan keyEvent
	keyboard   [16]bool

	halt bool

	clock *time.Ticker
	timer *time.Ticker

	machine Machine
}

func Init(machine Machine) State {
	var m [4096]byte
	copy(m[0:16*5], font)

	return State{
		SP: 0xefe,
		PC: 0x200,

		memory: m,

		machine: machine,

		// 500Hz clock
		clock: time.NewTicker(time.Second / 500),
		// 60Hz clock for timers
		timer: time.NewTicker(time.Second / 60),

		// Channel for keypresses
		keyChannel: make(chan keyEvent, 1),
	}
}

func (c *State) next() {
	c.PC += 2
}

func (c *State) RunProgram(program []byte) {
	c.LoadProgram(program)

	for {
		select {
		case <-c.clock.C:
			if c.PC >= 0xFFF || c.halt {
				c.halt = true
				c.timer.Stop()
				c.clock.Stop()
				return
			}

			op := NewOpCode([2]byte{c.memory[c.PC], c.memory[c.PC+1]})
			c.RunOp(op)
		case <-c.timer.C:
			if c.delayTimer > 0 {
				c.delayTimer--
			}

			if c.soundTimer > 0 {
				c.soundTimer--
			}
		case e := <-c.keyChannel:
			c.keyboard[e.key] = e.pressed
		}
	}
}

func (c *State) LoadProgram(program []byte) {
	copy(c.memory[0x200:0x200+len(program)], program)
}

type Frame struct {
	*image.RGBA
}

func (c *State) GetFrame(X, Y int) *Frame {
	img := image.NewRGBA(image.Rectangle{
		Max: image.Pt(64, 32),
	})

	x, y := 0, 0
	for _, b := range c.memory[0xF00:] {
		for i := 0; i < 8; i++ {
			c := color.Black

			bit := (b >> byte(7-i)) & 1
			if bit == 1 {
				c = color.White
			}

			img.Set(x*8+i, y, c)
		}
		x++
		if x == 8 {
			x = 0
			y++
		}
	}

	dst := image.NewRGBA(image.Rectangle{Max: image.Pt(X, Y)})
	draw.NearestNeighbor.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)
	return &Frame{dst}
}

func (c *State) SendKey(key byte, pressed bool) {
	c.keyChannel <- keyEvent{
		key:     key,
		pressed: pressed,
	}
}

func (c *State) Beep() bool {
	return c.soundTimer > 0
}

func (c *State) op0(op OpCode) {
	if op.B01 != 0x0 {
		fmt.Printf("Call RCA 1802 program at 0x%03x (not implemented)\n", op.Addr())
		c.halt = true
		return
	}

	switch op.B1 {
	case 0xe0:
		// Clear the screen.
		copy(c.memory[0xF00:], make([]byte, 256))
		c.next()
	case 0xee:
		// Stack pop
		c.PC = uint16(c.memory[c.SP])<<8 | uint16(c.memory[c.SP+1])
		c.SP += 2
	default:
		c.notImplemented(op)
	}
	c.next()
}

func (c *State) op1(op OpCode) {
	if op.Addr() == c.PC {
		fmt.Println("Infinite loop found, exiting.")
		c.halt = true
	}

	c.PC = op.Addr()
}

func (c *State) op2(op OpCode) {
	// Stack push
	c.SP -= 2
	c.memory[c.SP] = byte((c.PC & 0xff00) >> 8)
	c.memory[c.SP+1] = byte(c.PC & 0xff)
	c.PC = op.Addr()
}

func (c *State) op3(op OpCode) {
	if c.V[op.B01] == op.B1 {
		c.next()
	}
	c.next()
}

func (c *State) op4(op OpCode) {
	if c.V[op.B01] != op.B1 {
		c.next()
	}
	c.next()
}

func (c *State) op5(op OpCode) {
	if c.V[op.B01] == c.V[op.B10] {
		c.next()
	}
	c.next()
}

func (c *State) op6(op OpCode) {
	c.V[op.B01] = op.B1
	c.next()
}

func (c *State) op7(op OpCode) {
	c.V[op.B01] += op.B1
	c.next()
}

func (c *State) op8(op OpCode) {
	x := op.B01
	y := op.B10

	switch op.B11 {
	case 0x00:
		c.V[x] = c.V[y]
	case 0x01:
		c.V[x] |= c.V[y]
	case 0x02:
		c.V[x] &= c.V[y]
	case 0x03:
		c.V[x] ^= c.V[y]
	case 0x04:
		res := uint16(c.V[x]) + uint16(c.V[y])

		carry := res&0xff00 > 0
		if carry {
			c.V[0x0f] = 1
		} else {
			c.V[0x0f] = 0
		}

		c.V[x] = byte(res & 0xff)
	case 0x05:
		borrow := c.V[x] < c.V[y]
		if borrow {
			c.V[0x0f] = 0
		} else {
			c.V[0x0f] = 1
		}

		c.V[x] -= c.V[y]
	case 0x06:
		c.V[0x0f] = c.V[x] & 0x01
		c.V[x] >>= 1
	case 0x07:
		borrow := c.V[x] > c.V[y]
		if borrow {
			c.V[0x0f] = 0
		} else {
			c.V[0x0f] = 1
		}

		c.V[x] = c.V[y] - c.V[x]
	case 0x0e:
		c.V[0x0f] = c.V[x] >> 7
		c.V[x] <<= 1
	default:
		c.notImplemented(op)
	}
	c.next()
}

func (c *State) op9(op OpCode) {
	if c.V[op.B01] != c.V[op.B10] {
		c.next()
	}
	c.next()
}

func (c *State) opA(op OpCode) {
	c.I = op.Addr()
	c.next()
}

func (c *State) opB(op OpCode) {
	c.PC = uint16(c.V[0x00]) + op.Addr()
}

func (c *State) opC(op OpCode) {
	// VX = random(0,255) & NN
	c.V[op.B01] = byte(rand.Intn(0xff)) & op.B1
	c.next()
}

func (c *State) opD(op OpCode) {
	<-c.timer.C
	x := int(c.V[op.B01])
	y := int(c.V[op.B10])
	n := int(op.B11)

	// Set collision to 0
	c.V[0x0F] = 0

	for i := 0; i < n && (i+y) < 32; i++ {
		spriteAddr := int(c.I) + i
		sprite := c.memory[spriteAddr]

		for j := x; j < (x+8) && j < 64; j++ {
			spritePixel := (sprite >> (x + 7 - j)) & 0x01
			if spritePixel == 0 {
				continue
			}

			pixelAddr := screenAddress + ((i+y)*8 + (j / 8))

			pixelByte := c.memory[pixelAddr]
			pixel := pixelByte & (0x80 >> (j % 8))
			sp := spritePixel << (7 - (j % 8))

			// Collision
			if (pixel & sp) > 0 {
				c.V[0xF] = 1
			}

			pixelByte ^= sp

			c.memory[pixelAddr] = pixelByte
		}
	}

	c.next()
}

func (c *State) opE(op OpCode) {
	switch op.B1 {
	case 0x9e:
		// Skip next instruction if key stored in VX is down
		if c.keyboard[c.V[op.B01]] {
			c.next()
		}
	case 0xa1:
		// Skip next instruction if key stored in VX is up
		if !c.keyboard[c.V[op.B01]] {
			c.next()
		}
	default:
		c.notImplemented(op)
	}
	c.next()
}

func (c *State) opF(op OpCode) {
	switch op.B1 {
	case 0x07:
		c.V[op.B01] = c.delayTimer
	case 0x0a:
		<-c.keyChannel
		c.V[op.B01] = c.machine.WaitForKeyPress()
	case 0x15:
		// Set delay timer to the value of VX
		c.delayTimer = c.V[op.B01]
	case 0x18:
		// Set sound timer to the value of VX
		c.soundTimer = c.V[op.B01]
	case 0x1e:
		sum := c.I + uint16(c.V[op.B01])

		// Check for range overflow
		if sum > 0xfff {
			c.V[0x0f] = 1
		} else {
			c.V[0x0f] = 0
		}

		c.I = sum & 0x0fff
	case 0x29:
		// Get address of a sprite representation of VX value.
		// Fonts are loaded from the first byte in memory.
		// Since sprites are 5 lines each, we can find the address
		// by multiplying the value by 5.
		c.I = uint16(c.V[op.B01]) * 5
	case 0x33:
		// Set I, I+1 and I+2 to the decimal representation of
		// the value stored in VX
		value := c.V[op.B01]
		ones := value % 10
		value /= 10
		tens := value % 10
		hundreds := value / 10

		c.memory[c.I] = hundreds
		c.memory[c.I+1] = tens
		c.memory[c.I+2] = ones
	case 0x55:
		for i, v := range c.V[0 : op.B01+1] {
			c.memory[c.I+uint16(i)] = v
		}
	case 0x65:
		for i := range c.V[0 : op.B01+1] {
			c.V[i] = c.memory[c.I+uint16(i)]
		}
	default:
		c.notImplemented(op)
	}

	c.next()
}

func (c *State) RunOp(op OpCode) {
	// DisassembleChip8Op(op)

	switch op.B00 {
	case 0x00:
		c.op0(op)
	case 0x01:
		c.op1(op)
	case 0x02:
		c.op2(op)
	case 0x03:
		c.op3(op)
	case 0x04:
		c.op4(op)
	case 0x05:
		c.op5(op)
	case 0x06:
		c.op6(op)
	case 0x07:
		c.op7(op)
	case 0x08:
		c.op8(op)
	case 0x09:
		c.op9(op)
	case 0x0a:
		c.opA(op)
	case 0x0b:
		c.opB(op)
	case 0x0c:
		c.opC(op)
	case 0x0d:
		c.opD(op)
	case 0x0e:
		c.opE(op)
	case 0x0f:
		c.opF(op)
	}
}

func (c *State) notImplemented(op OpCode) {
	fmt.Printf("Opcode not implemented: %s\n", op)
	c.halt = true
}

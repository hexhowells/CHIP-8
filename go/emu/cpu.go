package emu

import (
	"fmt"
	"log"
	"os"
	"strings"
	"math/rand/v2"
	"github.com/gen2brain/beeep"
)

/*
CHIP-8 Instruction Breakdown
[Each instruction is 2 bytes]

The 2 bytes can be broken down into different sections (layed out below). How the instruction is broken down
depends on the Specific instruction being run.

1010 0010 0110 1011

OP - first nibble, primary Opcode group (1010)
VC (VX/X) - second nibble, usually looking up one of the registers (V0 - VF) (0010)
Y  (VY) - third nibble, usually looking up a second register (0110)
N - fourth nibble, a 4-bit immediate value (1011)
KK (NN) - second byte (last two nibbles), a 8-bit immediate value (0110 1011)
NNN - last three nibbles, a 12-bit immediate memory address (0010 0110 1011)

*/

var fontset = []uint8{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80,  // F
}

type CPU struct {
	Pc uint16  // program counter
	I uint16  // index register
	Vc [16]uint8  // variable registers (16 8-bit)
	stack [16]uint16
	Sp uint16  // stack pointer
	keys [16]uint8
	memory [4096]uint8  // 4kB memory Space
	Screen [32][64]uint8
	Opcode uint8
	Oprand uint16
	lookup []INSTRUCTION
	DelayTimer uint8
	SoundTimer uint8
}


type INSTRUCTION struct {
	name string
	Operate func(*CPU)
}


func NewCPU() *CPU {
	cpu := CPU{}
	cpu.Pc = 0x0200

	// load instruction lookup table
	cpu.lookup = []INSTRUCTION{
		INSTRUCTION{"MAP0", (*CPU)._MAP0}, INSTRUCTION{"1NNN", (*CPU)._1NNN}, INSTRUCTION{"2NNN", (*CPU)._2NNN}, INSTRUCTION{"3XNN", (*CPU)._3XNN},
		INSTRUCTION{"4XNN", (*CPU)._4XNN}, INSTRUCTION{"5XY0", (*CPU)._5XY0}, INSTRUCTION{"6XNN", (*CPU)._6XNN}, INSTRUCTION{"7XNN", (*CPU)._7XNN},
		INSTRUCTION{"MAP8", (*CPU)._MAP8}, INSTRUCTION{"9XY0", (*CPU)._9XY0}, INSTRUCTION{"ANNN", (*CPU)._ANNN}, INSTRUCTION{"BNNN", (*CPU)._BNNN},
		INSTRUCTION{"CXNN", (*CPU)._CXNN}, INSTRUCTION{"DXYN", (*CPU)._DXYN}, INSTRUCTION{"MAPE", (*CPU)._MAPE}, INSTRUCTION{"MAPF", (*CPU)._MAPF},
	}

	// load fonts into memory
	for i := 0; i < len(fontset); i++ {
		cpu.memory[i] = fontset[i]
	}

	cpu.SoundTimer = 60

	return &cpu
}


func (cpu *CPU) LoadROM(filepath string) {
	file, err := os.Open(filepath)

	if err != nil {
		log.Println("Failed to open ROM file: %v", err)
		return
	}

	defer file.Close()

	info, err := file.Stat()

	if err != nil {
		log.Println("Failed to read metadata on ROM file: %v", err)
	}

	if int64(len(cpu.memory) - 512) < info.Size() {
		fmt.Errorf("ROM is larger than available memory!")
		return
	}

	numBytes, err := file.Read(cpu.memory[512:])
	if err != nil {
		fmt.Errorf("Failed to read ROM into memory: %v", err)
		return
	}

	fmt.Printf("Read %d bytes into memory", numBytes)

}


func (cpu *CPU) Render() {
	const mapping = ".#"

	var sb strings.Builder

	sb.WriteString("\n[SCREEN]\n")

	for _, row := range cpu.Screen {
		for _, val := range row {
			sb.WriteByte(mapping[val])
		}
		sb.WriteByte('\n')
	}

	fmt.Print(sb.String())
}


func (cpu *CPU) PrintCPU() {
	// internal values
	fmt.Println("\n------------------------------------------------------------")
	fmt.Printf("| %-12s | %-13s | %-25s |\n", "Field", "Value", "Description")
	fmt.Println("------------------------------------------------------------")
	fmt.Printf("| %-12s | $%-12.4X | %-25s |\n", "pc", cpu.Pc, "Program counter (8-bit)")
	fmt.Printf("| %-12s | $%-12.4X | %-25s |\n", "i", cpu.I, "Index register (16-bit)")
	fmt.Printf("| %-12s | $%-12.4X | %-25s |\n", "Sp", cpu.Sp, "Stack pointer")
	fmt.Printf("| %-12s | $%-12.2X | %-25s |\n", "Opcode", cpu.Opcode, "Current Opcode byte")
	fmt.Printf("| %-12s | $%-12.4X | %-25s |\n", "Oprand", cpu.Oprand, "Current operand")

	// stack info
	stackVal := "N/A"
	if cpu.Sp < 16 {
		stackVal = fmt.Sprintf("$%.4X", cpu.stack[cpu.Sp])
	}
	fmt.Printf("| %-12s | %-13s | %-25s |\n", "stack[Sp]", stackVal, "Value at top of stack")

	// variables
	fmt.Println("------------------------------------------------------------")
	fmt.Printf("| %-45s |\n", "Variable Registers (V0 - VF)  ")
	fmt.Println("-------------------------------------------------")

	for i := 0; i < 16; i += 4 {
		fmt.Printf("| V%X: $%.2X   | V%X: $%.2X   | V%X: $%.2X   | V%X: $%.2X   |\n",
			i, cpu.Vc[i],
			i+1, cpu.Vc[i+1],
			i+2, cpu.Vc[i+2],
			i+3, cpu.Vc[i+3],
		)
	}
	fmt.Println("-------------------------------------------------")

	// memory dump (shows first 512 bytes of loaded ROM)
	fmt.Println("\n------------------------------------------------------------")
	fmt.Printf("| %-56s |\n", "Memory Dump (0x0200 - 0x03FF)")
	fmt.Println("------------------------------------------------------------")
	
	for addr := 512; addr < 1024; addr += 16 {
		fmt.Printf("| 0x%04X | ", addr)
		for j := 0; j < 16; j++ {
			if j < 15 {
				fmt.Printf("%02X ", cpu.memory[addr+j])
			} else {
				fmt.Printf("%02X", cpu.memory[addr+j])
			}
		}
		fmt.Println(" |")
	}
	fmt.Println("------------------------------------------------------------")
}


func (cpu *CPU) ReadInstruction(addr uint16) uint16 {
	return uint16(cpu.memory[addr]) << 8 | uint16(cpu.memory[addr + 1])
}


func (cpu *CPU) RunInstruction(ins uint16) {
	cpu.Opcode = uint8((ins & 0xF000) >> 12)
	cpu.Oprand = ins & 0x0FFF

	inst := cpu.lookup[cpu.Opcode]
	inst.Operate(cpu)
}


func (cpu *CPU) Clock() {
	ins := cpu.ReadInstruction(cpu.Pc)
	cpu.Pc += 2

	cpu.RunInstruction(ins)

	if cpu.DelayTimer > 0 {
		cpu.DelayTimer -= 1
	}

	if cpu.SoundTimer > 0 {
		cpu.Beep()
		cpu.SoundTimer -= 1
	}
}


func (cpu *CPU) Beep() {
	err := beeep.Beep(beeep.DefaultFreq, 300)
	if err != nil {
		log.Println("Failed to beep: %v", err)
	}
}


func (cpu *CPU) PressKey(num uint8, down bool) {
	if down {
		cpu.keys[num] = 1
	} else {
		cpu.keys[num] = 0
	}
}


func (cpu *CPU) _MAP0() {
	n := uint8(cpu.Oprand & 0x000F)
	switch n {
	case 0x00:
		cpu._00E0()
	case 0x0E:
		cpu._00EE()
	default:
		fmt.Errorf("Invalid instruction nibble for 0: %d", n)
	}
}


func (cpu *CPU) _MAP8() {
	n := uint8(cpu.Oprand & 0x000F)
	switch n {
	case 0x00:
		cpu._8XY0()
	case 0x01:
		cpu._8XY1()
	case 0x02:
		cpu._8XY2()
	case 0x03:
		cpu._8XY3()
	case 0x04:
		cpu._8XY4()
	case 0x05:
		cpu._8XY5()
	case 0x06:
		cpu._8XY6()
	case 0x07:
		cpu._8XY7()
	case 0x0E:
		cpu._8XYE()
	default:
		fmt.Errorf("Invalid instruction nibble for 0: %d", n)
	}
}


func (cpu *CPU) _MAPE() {
	n := uint8(cpu.Oprand & 0x00FF)
	switch n {
	case 0x9E:
		cpu._EX9E()
	case 0xA1:
		cpu._EXA1()
	default:
		fmt.Errorf("Invalid instruction nibble for 0: %d", n)
	}
}


func (cpu *CPU) _MAPF() {
	n := uint8(cpu.Oprand & 0x00FF)
	switch n {
	case 0x07:
		cpu._FX07()
	case 0x15:
		cpu._FX15()
	case 0x18:
		cpu._FX18()
	case 0x1E:
		cpu._FX1E()
	case 0x0A:
		cpu._FX0A()
	case 0x29:
		cpu._FX29()
	case 0x33:
		cpu._FX33()
	case 0x55:
		cpu._FX55()
	case 0x65:
		cpu._FX65()
	default:
		fmt.Errorf("Invalid instruction nibble for 0: %d", n)
	}
}


// clear screen
func (cpu *CPU) _00E0() {
	var blank [32][64]uint8
	cpu.Screen = blank
}


// jump
func (cpu *CPU) _1NNN() {
	cpu.Pc = cpu.Oprand & 0x0FFF
}


// return from subroutine
func (cpu *CPU) _00EE() {
	cpu.Pc = cpu.stack[cpu.Sp-1]
	cpu.stack[cpu.Sp] = 0x0000
	cpu.Sp -= 1
}


// call subroutine
func (cpu *CPU) _2NNN() {
	cpu.stack[cpu.Sp] = cpu.Pc
	cpu.Sp += 1
	cpu.Pc = cpu.Oprand
}


// skip one instruction if VX == NN
func (cpu *CPU) _3XNN() {
	vx := uint8((cpu.Oprand & 0x0F00) >> 8)
	if cpu.Vc[vx] == uint8(cpu.Oprand & 0x00FF) {
		cpu.Pc += 2
	}
}


// skip one instruction if VX != NN
func (cpu *CPU) _4XNN() {
	vx := uint8((cpu.Oprand & 0x0F00) >> 8)
	if cpu.Vc[vx] != uint8(cpu.Oprand & 0x00FF) {
		cpu.Pc += 2
	}
}


// skip one instruction if VX == VY
func (cpu *CPU) _5XY0() {
	vx := uint8((cpu.Oprand & 0x0F00) >> 8)
	vy := uint8((cpu.Oprand & 0x00F0) >> 4)
	if cpu.Vc[vx] == cpu.Vc[vy] {
		cpu.Pc += 2
	}
}


// skip one instruction if VX != VY
func (cpu *CPU) _9XY0() {
	vx := uint8((cpu.Oprand & 0x0F00) >> 8)
	vy := uint8((cpu.Oprand & 0x00F0) >> 4)
	if cpu.Vc[vx] != cpu.Vc[vy] {
		cpu.Pc += 2
	}
}


// set VX to NN
func (cpu *CPU) _6XNN() {
	vx := uint8((cpu.Oprand & 0x0F00) >> 8)
	cpu.Vc[vx] = uint8(cpu.Oprand & 0x00FF)
}


// add
func (cpu *CPU) _7XNN() {
	vx := uint8((cpu.Oprand & 0x0F00) >> 8)
	cpu.Vc[vx] += uint8(cpu.Oprand & 0x00FF)
}


// set VX to VY
func (cpu *CPU) _8XY0() {
	vx := uint8((cpu.Oprand & 0x0F00) >> 8)
	vy := uint8((cpu.Oprand & 0x00F0) >> 4)
	cpu.Vc[vx] = cpu.Vc[vy]
}


// binary OR
func (cpu *CPU) _8XY1() {
	vx := uint8((cpu.Oprand & 0x0F00) >> 8)
	vy := uint8((cpu.Oprand & 0x00F0) >> 4)
	cpu.Vc[vx] |= cpu.Vc[vy]
}


// binary AND
func (cpu *CPU) _8XY2() {
	vx := uint8((cpu.Oprand & 0x0F00) >> 8)
	vy := uint8((cpu.Oprand & 0x00F0) >> 4)
	cpu.Vc[vx] &= cpu.Vc[vy]
}


// binary XOR
func (cpu *CPU) _8XY3() {
	vx := uint8((cpu.Oprand & 0x0F00) >> 8)
	vy := uint8((cpu.Oprand & 0x00F0) >> 4)
	cpu.Vc[vx] ^= cpu.Vc[vy]
}


// add (VX + VY)
func (cpu *CPU) _8XY4() {
	vx := uint8((cpu.Oprand & 0x0F00) >> 8)
	vy := uint8((cpu.Oprand & 0x00F0) >> 4)
	sum := cpu.Vc[vx] + cpu.Vc[vy]

	if sum < cpu.Vc[vx] {
		cpu.Vc[0x0F] = 0x01
	}
	cpu.Vc[vx] = sum
}


// subtract (VX - VY)
func (cpu *CPU) _8XY5() {
	vx := uint8((cpu.Oprand & 0x0F00) >> 8)
	vy := uint8((cpu.Oprand & 0x00F0) >> 4)

	if cpu.Vc[vx] >= cpu.Vc[vy] {
		cpu.Vc[0x0F] = 0x01
	} else {
		cpu.Vc[0x0F] = 0x00
	}
	cpu.Vc[vx] -= cpu.Vc[vy]
}


// subtract (VY - VX)
func (cpu *CPU) _8XY7() {
	vx := uint8((cpu.Oprand & 0x0F00) >> 8)
	vy := uint8((cpu.Oprand & 0x00F0) >> 4)

	if cpu.Vc[vy] >= cpu.Vc[vx] {
		cpu.Vc[0x0F] = 0x01
	} else {
		cpu.Vc[0x0F] = 0x00
	}
	cpu.Vc[vx] = cpu.Vc[vy] - cpu.Vc[vx]
}


// 1-bit shift (right)
func (cpu *CPU) _8XY6() {
	vx := uint8((cpu.Oprand & 0x0F00) >> 8)
	cpu.Vc[0x0F] = cpu.Vc[vx] & 0x01
	cpu.Vc[vx] >>= 0x01
}


// 1-bit shift (left)
func (cpu *CPU) _8XYE() {
	vx := uint8((cpu.Oprand & 0x0F00) >> 8)
	cpu.Vc[0x0F] = (cpu.Vc[vx] & 0x80) >> 0x07
	cpu.Vc[vx] <<= 0x01
}


// set index
func (cpu *CPU) _ANNN() {
	cpu.I = cpu.Oprand & 0x0FFF
}


// jump with offset
func (cpu *CPU) _BNNN() {
	cpu.Pc = (cpu.Oprand & 0x0FFF) + uint16(cpu.Vc[0x00])
}


// random number generator
func (cpu *CPU) _CXNN() {
	vx := uint8((cpu.Oprand & 0x0F00) >> 8)
	cpu.Vc[vx] = uint8(rand.Uint32()) & uint8(cpu.Oprand & 0x00FF)
}


func (cpu *CPU) _DXYN() {
	
}


// skip if key
func (cpu *CPU) _EX9E() {
	vx := uint8((cpu.Oprand & 0x0F00) >> 8)
	if cpu.keys[vx] == 0x01 {
		cpu.Pc += 2
	}
}


// skip if not key
func (cpu *CPU) _EXA1() {
	vx := uint8((cpu.Oprand & 0x0F00) >> 8)
	if cpu.keys[vx] == 0x00 {
		cpu.Pc += 2
	}
}


// set VX to delay timer
func (cpu *CPU) _FX07() {
	vx := uint8((cpu.Oprand & 0x0F00) >> 8)
	cpu.Vc[vx] = cpu.DelayTimer
}


// set delay timer
func (cpu *CPU) _FX15() {
	vx := uint8((cpu.Oprand & 0x0F00) >> 8)
	cpu.DelayTimer = cpu.Vc[vx]
}


// set sound timer
func (cpu *CPU) _FX18() {
	vx := uint8((cpu.Oprand & 0x0F00) >> 8)
	cpu.SoundTimer = cpu.Vc[vx]
}


// add to index
func (cpu *CPU) _FX1E() {
	vx := uint8((cpu.Oprand & 0x0F00) >> 8)
	cpu.I += uint16(cpu.Vc[vx])
}


// get key
func (cpu *CPU) _FX0A() {
	for i := 0; i < 16; i++ {
		if cpu.keys[i] == 1 {
			vx := uint8((cpu.Oprand & 0x0F00) >> 8)
			cpu.Vc[vx] = uint8(i)
			return
		}
	}
	cpu.Pc -= 2
}


func (cpu *CPU) _FX29() {
	
}


func (cpu *CPU) _FX33() {
	
}


func (cpu *CPU) _FX55() {
	
}


func (cpu *CPU) _FX65() {
	
}

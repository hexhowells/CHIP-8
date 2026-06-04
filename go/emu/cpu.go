package emu

import (
	"fmt"
	"strings"
)

type CPU struct {
	Pc uint8  // program counter
	I uint16  // index register
	Vc [16]uint8  // variable registers (16 8-bit)
	Stack [16]uint16
	sp uint16  // stack pointer
	Memory [4096]uint8  // 4kB memory space
	Screen [32][64]uint8
	opcode uint8
	oprand uint16
	lookup []INSTRUCTION
}


type INSTRUCTION struct {
	name string
	Operate func(*CPU)
}


func NewCPU() *CPU {
	cpu := CPU{}
	//opcode := 0x00

	cpu.lookup = []INSTRUCTION{
		INSTRUCTION{"00E0", (*CPU)._00E0},
	}

	return &cpu
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


func (cpu *CPU) ReadInstruction(addr uint8) uint16 {
	return uint16(cpu.Memory[addr]) << 8 | uint16(cpu.Memory[addr + 1])
}


/*
CHIP-8 Instruction Breakdown
[Each instruction is 2 bytes]

The 2 bytes can be broken down into different sections (layed out below). How the instruction is broken down
depends on the specific instruction being run.

1010 0010 0110 1011

OP - first nibble, primary opcode group (1010)
VC (VX/X) - second nibble, usually looking up one of the registers (V0 - VF) (0010)
Y  (VY) - third nibble, usually looking up a second register (0110)
N - fourth nibble, a 4-bit immediate value (1011)
KK (NN) - second byte (last two nibbles), a 8-bit immediate value (0110 1011)
NNN - last three nibbles, a 12-bit immediate memory address (0010 0110 1011)

*/
func (cpu *CPU) RunInstruction(ins uint16) {
	cpu.opcode = uint8((ins & 0xF000) >> 12)
	cpu.oprand = ins & 0x0FFF

	inst := cpu.lookup[cpu.opcode]
	fmt.Printf("Running instruction: %s", inst.name)
	inst.Operate(cpu)
}


func (cpu *CPU) Clock() {
	// fetch instruction from memory at the PC
	// decode the instruction to find out what the emu should do
	// execute the instruction
	ins := cpu.ReadInstruction(cpu.Pc)
	cpu.Pc += 2

	cpu.RunInstruction(ins)
}


func (cpu *CPU) _0NNN() {

}


// clear screen
func (cpu *CPU) _00E0() {
	var blank [32][64]uint8
	cpu.Screen = blank
}


// jump
func (cpu *CPU) _1NNN() {
	cpu.Pc = cpu.oprand
}


// return from subroutine
func (cpu *CPU) _00EE() {
	cpu.Pc = cpu.stack[cpu.sp-1]
	cpu.stack[cpu.sp] = 0x0000
	cpu.sp -= 1
}


// call subroutine
func (cpu *CPU) _2NNN() {
	cpu.stack[cpu.sp] = cp.Pc
	cpu.sp += 1
	cpu.Pc = cpu.oprand
}


// skip one instruction if VX == NN
func (cpu *CPU) _3XNN() {
	reg := uint8((cpu.oprand & 0x0F00) > 8)
	if cpu.Vc[reg] == uint8(cpu.oprand & 0x00FF) {
		cpu.Pc += 2
	}
}


// skip one instruction if VX != NN
func (cpu *CPU) _4XNN() {
	reg := uint8((cpu.oprand & 0x0F00) > 8)
	if cpu.Vc[reg] != uint8(cpu.oprand & 0x00FF) {
		cpu.Pc += 2
	}
}


// skip one instruction if VX == VY
func (cpu *CPU) _5XY0() {
	vx := uint8((cpu.oprand & 0x0F00) > 8)
	vy := uint8((cpu.oprand & 0x00F0) > 4)
	if cpu.Vc[vx] == cpu.Vc[vx] {
		cpu.Pc += 2
	}
}


// skip one instruction if VX != VY
func (cpu *CPU) _9XY0() {
	vx := uint8((cpu.oprand & 0x0F00) > 8)
	vy := uint8((cpu.oprand & 0x00F0) > 4)
	if cpu.Vc[vx] != cpu.Vc[vx] {
		cpu.Pc += 2
	}
}


// set
func (cpu *CPU) _6XNN() {
	reg := uint8((cpu.oprand & 0x0F00) > 8)
	cpu.Vc[reg] = uint8(cpu.oprand & 0x00FF)
}


func (cpu *CPU) _7XNN() {
	
}


func (cpu *CPU) _8XY0() {
	
}


func (cpu *CPU) _8XY1() {
	
}


func (cpu *CPU) _8XY2() {
	
}


func (cpu *CPU) _8XY3() {
	
}


func (cpu *CPU) _8XY4() {
	
}


func (cpu *CPU) _8XY5() {
	
}


func (cpu *CPU) _8XY7() {
	
}


func (cpu *CPU) _8XY6() {
	
}


func (cpu *CPU) _8XYE() {
	
}


func (cpu *CPU) _ANNN() {
	
}


func (cpu *CPU) _BNNN() {
	
}


func (cpu *CPU) _CXNN() {
	
}


func (cpu *CPU) _DXYN() {
	
}


func (cpu *CPU) _EX9E() {
	
}


func (cpu *CPU) _EXA1() {
	
}


func (cpu *CPU) _FX07() {
	
}


func (cpu *CPU) _FX15() {
	
}


func (cpu *CPU) _FX18() {
	
}


func (cpu *CPU) _FX1E() {
	
}


func (cpu *CPU) _FX0A() {
	
}


func (cpu *CPU) _FX29() {
	
}


func (cpu *CPU) _FX33() {
	
}


func (cpu *CPU) _FX55() {
	
}


func (cpu *CPU) _FX65() {
	
}

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
	Memory [4096]uint8  // 4kB memory space
	Screen [32][64]uint8
	opcode uint16
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


func (cpu *CPU) ReadMemory(addr uint8) uint16 {
	return uint16(cpu.Memory[addr]) << 8 | uint16(cpu.Memory[addr + 1])
}


func (cpu *CPU) RunInstruction(ins_byte uint16) {
	inst := cpu.lookup[ins_byte]
	fmt.Printf("Running instruction: %s", inst.name)
	inst.Operate(cpu)
}


func (cpu *CPU) Clock() {
	// fetch instruction from memory at the PC
	// decode the instruction to find out what the emu should do
	// execute the instruction
	cpu.opcode = cpu.ReadMemory(cpu.Pc)
	cpu.Pc += 2

	cpu.RunInstruction(cpu.opcode)
}


func (cpu *CPU) _0NNN() {

}


// clear screen
func (cpu *CPU) _00E0() {
	var blank [32][64]uint8
	cpu.Screen = blank
}


func (cpu *CPU) _1NNN() {
	
}


func (cpu *CPU) _00EE() {
	
}


func (cpu *CPU) _2NNN() {
	
}


func (cpu *CPU) _3XNN() {
	
}


func (cpu *CPU) _4XNN() {
	
}


func (cpu *CPU) _5XY0() {
	
}


func (cpu *CPU) _9XY0() {
	
}


func (cpu *CPU) _6XNN() {
	
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
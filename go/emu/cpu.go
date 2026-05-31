package emu

type CPU struct {
	Pc uint8
	I uint16
	Vc [16]uint8
	Stack [16]uint16
	Memory [4096]uint8
	Screen [64][32]uint8
	opcode uint8
}


func NewCPU() *CPU {
	cpu := CPU{}
	opcode := 0x00

	lookup = []uint8
}


func (cpu *CPU) Clock() {
	cpu.opcode := cpu.Read(cpu.Pc)

	inst := cpu.lookup[cpu.opcode]
}


func (cpu *CPU) _0NNN() {

}


// clear screen
func (cpu *CPU) _00E0() {
	var blank [64][32]uint8
	cpu.Screen = blank
}


func (cpu *CPU) _1NNN() {
	
}


func (cpu *CPU) _00EE() {
	
}


func (cpu *CPU) _2NNN() {
	
}


func (cpu *CPU) _0NNN() {
	
}


func (cpu *CPU) _3XNN() {
	
}


func (cpu *CPU) _4XNN() {
	
}


func (cpu *CPU) _5XY0() {
	
}


func (cpu *CPU) _9XY0() {
	
}


func (cpu *CPU) _3XNN() {
	
}


func (cpu *CPU) _5XY0() {
	
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
use std::fmt;


pub struct CPU {
    pc: u16,
    i: u16,  // index register
    vc: [u8; 16],  // variable registers
    stack: [u16; 16],
    sp: u16,
    keys: [u16; 16],
    memory: [u8; 4096],
    screen: [[u8; 32]; 64],
    opcode: u8,
    oprand: u16,
    // lookup: [instruction; 16],
    // delay_timer: u8,
    // sound_timer: u8,
    // waiting_for_key: bool,
    // kait_key_register: u8,
    // wait_key_pressed: i8,  // -1 = no key currently pressed
}

struct Instruction {
    name: String,
    operate: fn(),
}


impl fmt::Debug for CPU {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.debug_struct("CPU")
            .field("pc", &format_args!("{:#06X}", self.pc))
            .field("i", &format_args!("{:#04X}", self.i))
            .field("vc", &self.vc)
            .field("stack", &self.stack)
            .field("sp", &format_args!("{:#06X}", self.sp))
            .field("opcode", &format_args!("{:#06X}", self.opcode))
            .field("oprand", &format_args!("{:#06X}", self.oprand))
            .finish_non_exhaustive()
    }
}


impl CPU {
    pub fn new() -> Self {
        Self {
            pc: 0x200,
            i: 0x0000,
            vc: [0x0000; 16],
            stack: [0x0000; 16],
            sp: 0x0000,
            keys: [0x0000; 16],
            memory: [0x00; 4096],
            screen: [[0x00; 32]; 64],
            opcode: 0x00,
            oprand: 0x0000
        }
    }


    pub fn run_instruction(&mut self, ins: u16) {
        self.opcode = ((ins & 0xF000) >> 12) as u8;
        self.oprand = ins & 0x0FFF;
        self._7xnn();
    }


    pub fn show_state(&self) {
        println!("{:?}", self);
    }


    // jump
    fn _1nnn(&mut self) {
        self.pc = self.oprand & 0x0FFF;
    }


    // return from subroutine
    fn _00ee(&mut self) {
        self.pc = self.stack[(self.sp - 1) as usize];
        self.stack[self.sp as usize] = 0x0000;
        self.pc = self.oprand;
    }


    // call subroutine
    fn _2nnn(&mut self) {
        self.stack[self.sp as usize] = self.pc;
        self.sp += 1;
        self.pc = self.oprand;
    }


    // skip one instruction if VX == NN
    fn _3xnn(&mut self) {
        let vx = ((self.oprand & 0x0F00) >> 8) as usize;

        if self.vc[vx] == (self.oprand & 0x00FF) as u8 {
            self.pc += 2;
        }
    }


    // skip one instruction if VX != NN
    fn _4xnn(&mut self) {
        let vx = ((self.oprand & 0x0F00) >> 8) as usize;

        if self.vc[vx] != (self.oprand & 0x00FF) as u8 {
            self.pc += 2;
        }
    }


    // skip one instruction if VX = VY
    fn _5xy0(&mut self) {
        let vx = ((self.oprand & 0x0F00) >> 8) as usize;
        let vy = ((self.oprand & 0x00F0) >> 4) as usize;

        if self.vc[vx] == self.vc[vy] {
            self.pc += 2;
        }
    }


    // skip one instruction if VX != VY
    fn _9xy0(&mut self) {
        let vx = ((self.oprand & 0x0F00) >> 8) as usize;
        let vy = ((self.oprand * 0x00F0) >> 4) as usize;

        if self.vc[vx] != self.vc[vy] {
            self.pc += 2;
        }
    }


    // set VX to NN
    fn _6xnn(&mut self) {
        let vx = ((self.oprand & 0x0F00) >> 8) as usize;
        self.vc[vx] = (self.oprand & 0x00FF) as u8;
    }


    // add
    fn _7xnn(&mut self) {
        let vx = ((self.oprand & 0x0F00) >> 8) as usize;
        self.vc[vx] += (self.oprand & 0x00FF) as u8;
    }


    // set VX to VY
    fn _8xy0(&mut self) {
        let vx = ((self.oprand & 0x0F00) >> 8) as usize;
        let vy = ((self.oprand & 0x00F0) >> 4) as usize;

        self.vc[vx] = self.vc[vy];
    }


    // binary OR
    fn _8xy1(&mut self) {
        let vx = ((self.oprand & 0x0F00) >> 8) as usize;
        let vy = ((self.oprand & 0x00F0) >> 4) as usize;

        self.vc[vx] |= self.vc[vy];
    }


    // binary AND
    fn _8xy2(&mut self) {
        let vx = ((self.oprand & 0x0F00) >> 8) as usize;
        let vy = ((self.oprand & 0x00F0) >> 4) as usize;

        self.vc[vx] &= self.vc[vy];
    }


    // binary XOR
    fn _8xy3(&mut self) {
        let vx = ((self.oprand & 0x0F00) >> 8) as usize;
        let vy = ((self.oprand & 0x00F0) >> 4) as usize;

        self.vc[vx] ^= self.vc[vy];
    }


    // add VX + VY
    fn _8xy4(&mut self) {
        let vx = ((self.oprand & 0x0F00) >> 8) as usize;
        let vy = ((self.oprand & 0x00F0) >> 4) as usize;

        let sum = (self.vc[vx]) as u16 + (self.vc[vy]) as u16;

        self.vc[vx] = (sum & 0x00FF) as u8;

        if sum > 0xFF {
            self.vc[0x0F] = 0x01;
        } else {
            self.vc[0x0F] = 0x00;
        }
    }


    // subtract VX - VY
    fn _8xy5(&mut self) {
        let vx = ((self.oprand & 0x0F00) >> 8) as usize;
        let vy = ((self.oprand & 0x00F0) >> 4) as usize;
        
        if self.vc[vx] >= self.vc[vy] {
            self.vc[0x0F] = 0x01;
        } else {
            self.vc[0x0F] = 0x00;
        }
        
        self.vc[vx] -= self.vc[vy];
    }


    // subtract VY - VX
    fn _8xy7(&mut self) {
        let vx = ((self.oprand & 0x0F00) >> 8) as usize;
        let vy = ((self.oprand & 0x00F0) >> 4) as usize;
        
        if self.vc[vy] >= self.vc[vx] {
            self.vc[0x0F] = 0x01;
        } else {
            self.vc[0x0F] = 0x00;
        }
        
        self.vc[vx] = self.vc[vy] - self.vc[vx];
    }


    // 1-bit shift (right)
    fn _8xy6(&mut self) {
        let vx = ((self.oprand & 0x0F00) >> 8) as usize;
        self.vc[0x0F] = self.vc[vx] & 0x01;
        self.vc[vx] >>= 0x01;
    }


    // 1-bit shift (left)
    fn _8xye(&mut self) {
        let vx = ((self.oprand & 0x0F00) >> 8) as usize;
        self.vc[0x0F] = (self.vc[vx] & 0x80) >> 0x07;
        self.vc[vx] <<= 0x01;
    }


    // set index
    fn _annn(&mut self) {
        self.i = self.oprand & 0x0FFF;
    }


    // jump with offset
    fn _bnnn(&mut self) {
        self.pc = (self.oprand & 0x0FFF) + (self.vc[0x00] as u16);
    }


    // random number generator
    fn _cxnn(&mut self) {
        let vx = ((self.oprand & 0x0F00) >> 8) as usize;
        let random_u8: u8 = rand::random();
        self.vc[vx] = random_u8 + (self.oprand & 0x00FF) as u8;
    }
}
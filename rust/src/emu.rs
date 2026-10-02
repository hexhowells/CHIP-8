use std::fmt;
use std::fs;
use std::io::{self, Error, ErrorKind};


static FONTSET: [u8; 80] = [
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
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
];

pub struct CPU {
    pc: u16,
    i: u16,  // index register
    vc: [u8; 16],  // variable registers
    stack: [u16; 16],
    sp: u16,
    keys: [u16; 16],
    memory: [u8; 4096],
    screen: [[u8; 64]; 32],
    opcode: u8,
    oprand: u16,
    lookup: [Instruction; 16],
    delay_timer: u8,
    sound_timer: u8,
    waiting_for_key: bool,
    wait_key_register: u8,
    wait_key_pressed: i8,  // -1 = no key currently pressed
}

struct Instruction {
    name: &'static str,
    operate: fn(&mut CPU),
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
        let mut cpu = Self {
            pc: 0x0200,
            i: 0x0000,
            vc: [0x0000; 16],
            stack: [0x0000; 16],
            sp: 0x0000,
            keys: [0x0000; 16],
            memory: [0x00; 4096],
            screen: [[0x00; 64]; 32],
            opcode: 0x00,
            oprand: 0x0000,
            lookup: [
                Instruction{name: "MAP0", operate: CPU::_map0},
                Instruction{name: "1NNN", operate: CPU::_1nnn},
                Instruction{name: "2NNN", operate: CPU::_2nnn},
                Instruction{name: "3XNN", operate: CPU::_3xnn},
                Instruction{name: "4XNN", operate: CPU::_4xnn},
                Instruction{name: "5XY0", operate: CPU::_5xy0},
                Instruction{name: "6XNN", operate: CPU::_6xnn},
                Instruction{name: "7XNN", operate: CPU::_7xnn},
                Instruction{name: "MAP8", operate: CPU::_map8},
                Instruction{name: "9XY0", operate: CPU::_9xy0},
                Instruction{name: "ANNN", operate: CPU::_annn},
                Instruction{name: "BNNN", operate: CPU::_bnnn},
                Instruction{name: "CXNN", operate: CPU::_cxnn},
                Instruction{name: "DXYN", operate: CPU::_dxyn},
                Instruction{name: "MAPE", operate: CPU::_mape},
                Instruction{name: "MAPF", operate: CPU::_mapf},
                ],
            delay_timer: 0x00,
            sound_timer: 0x00,
            waiting_for_key: false,
            wait_key_register: 0x00,
            wait_key_pressed: -1
        };

        // load fonts into memory
        for i in 0..FONTSET.len() {
            cpu.memory[i] = FONTSET[i];
        }
        cpu
    }


    pub fn show_state(&self) {
        println!("{:?}", self);
    }


    pub fn display_screen(&self) {
        print!("\x1B[2J\x1B[1;1H");

        for row in self.screen.iter() {
            for &pixel in row.iter() {
                if pixel == 1 {
                    print!("█");
                } else {
                    print!(" ");
                }
            }
            println!();
        }
    }


    pub fn load_rom(&mut self, filepath: &str) -> io::Result<()> {
        let rom_data = fs::read(filepath)?;

        if rom_data.len() > (self.memory.len() - 512) {
            return Err(Error::new(
                ErrorKind::InvalidData,
                "ROM is larger than available memory!",
            ));
        }

        let start = 0x0200;
        let end = start + rom_data.len();

        self.memory[start..end].copy_from_slice(&rom_data);

        Ok(())
    }


    pub fn clock(&mut self) {
        let ins = self.read_instruction(self.pc as usize);
        self.pc += 2;

        self.run_instruction(ins);

        if self.delay_timer > 0 {
            self.delay_timer -= 1;
        }

        if self.sound_timer > 0 {
            self.sound_timer -= 1;
        }
    }


    fn read_instruction(&mut self, addr: usize) -> u16 {
        (self.memory[addr] as u16) << 8 | (self.memory[addr + 1] as u16)
    }


    pub fn run_instruction(&mut self, ins: u16) {
        self.opcode = ((ins & 0xF000) >> 12) as u8;
        self.oprand = ins & 0x0FFF;
        
        let inst_fn = self.lookup[self.opcode as usize].operate;
        inst_fn(self);
    }


    pub fn press_key(&mut self, num: usize, down: bool) {
        if down {
            self.keys[num] = 1;
        } else {
            self.keys[num] = 0;
        }
    }


    fn _map0(&mut self) {
        let n = (self.oprand & 0x000f) as u8;

        match n {
            0x00 => self._00e0(),
            0x0E => self._00ee(),
            _ => println!("Invalid instruction nibble for 0")
        }
    }


    fn _map8(&mut self) {
        let n = (self.oprand & 0x000f) as u8;

        match n {
            0x00 => self._8xy0(),
            0x01 => self._8xy1(),
            0x02 => self._8xy2(),
            0x03 => self._8xy3(),
            0x04 => self._8xy4(),
            0x05 => self._8xy5(),
            0x06 => self._8xy6(),
            0x07 => self._8xy7(),
            0x0E => self._8xye(),
            _ => println!("Invalid instruction nibble for 0")
        }
    }


    fn _mape(&mut self) {
        let n = (self.oprand & 0x00ff) as u8;

        match n {
            0x9E => self._ex9e(),
            0xA1 => self._exa1(),
            _ => println!("Invalid instruction nibble for 0")
        }
    }


    fn _mapf(&mut self) {
        let n = (self.oprand & 0x00ff) as u8;

        match n {
            0x07 => self._fx07(),
            0x15 => self._fx15(),
            0x18 => self._fx18(),
            0x1E => self._fx1e(),
            0x0A => self._fx0a(),
            0x29 => self._fx29(),
            0x33 => self._fx33(),
            0x55 => self._fx55(),
            0x65 => self._fx65(),
            _ => println!("Invalid instruction nibble for 0")
        }
    }


    // clear screen
    fn _00e0(&mut self) {
        self.screen = [[0x00; 64]; 32];
    }


    // jump
    fn _1nnn(&mut self) {
        self.pc = self.oprand & 0x0FFF;
    }


    // return from subroutine
    fn _00ee(&mut self) {
        self.pc = self.stack[(self.sp - 1) as usize];
        self.stack[self.sp as usize] = 0x0000;
        self.sp -= 1;
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
        let vy = ((self.oprand & 0x00F0) >> 4) as usize;

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
        let nn = (self.oprand & 0x00FF) as u8;

        self.vc[vx] = self.vc[vx].wrapping_add(nn);
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

        //let sum = (self.vc[vx]) as u16 + (self.vc[vy]) as u16;
        let (sum, overflowed) = self.vc[vx].overflowing_add(self.vc[vy]);

        self.vc[vx] = sum;

        self.vc[0x0F] = overflowed as u8;
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
        
        self.vc[vx] = self.vc[vx].wrapping_sub(self.vc[vy]);
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


    // display
    fn _dxyn(&mut self) {
        let vx = ((self.oprand & 0x0F00) >> 8) as usize;
        let vy = ((self.oprand & 0x00F0) >> 4) as usize;
        
        let x = (self.vc[vx] as usize) % 64;
        let y = (self.vc[vy] as usize) % 32;
        
        let n = (self.oprand & 0x000F) as usize;
        self.vc[0xF] = 0;

        for r in 0..n {
            let pixel = self.memory[self.i as usize + r];
            for c in 0..8 {
                if (pixel & (0x80 >> c)) != 0 {
                    let row = (y + r) % 32;
                    let col = (x + c) % 64;

                    if self.screen[row][col] == 1 {
                        self.vc[0xF] = 1;
                    }
                    self.screen[row][col] ^= 1;
                }
            }
        }
    }


    // skip if key
    fn _ex9e(&mut self) {
        let vx = ((self.oprand & 0x0F00) >> 8) as usize;
        if self.keys[vx] == 0x01 {
            self.pc += 2;
        }
    }


    // skip if not key
    fn _exa1(&mut self) {
        let vx = ((self.oprand & 0x0F00) >> 8) as usize;
        if self.keys[vx] == 0x00 {
            self.pc += 1;
        }
    }


    // set VX to delay timer
    fn _fx07(&mut self) {
        let vx = ((self.oprand & 0x0F00) >> 8) as usize;
        self.vc[vx] = self.delay_timer;
    }


    // set delay timer
    fn _fx15(&mut self) {
        let vx = ((self.oprand & 0x0F00) >> 8) as usize;
        self.delay_timer = self.vc[vx];
    }


    // set sound timer
    fn _fx18(&mut self) {
        let vx = ((self.oprand & 0x0F00) >> 8) as usize;
        self.sound_timer = self.vc[vx];
    }


    // add to index
    fn _fx1e(&mut self) {
        let vx = ((self.oprand & 0x0F00) >> 8) as usize;
        self.i += self.vc[vx] as u16;
    }


    // get key
    fn _fx0a(&mut self) {
        let vx = ((self.oprand & 0x0F00) >> 8) as usize;

        if !self.waiting_for_key {
            self.waiting_for_key = true;
            self.wait_key_register = vx as u8;
            self.wait_key_pressed = -1;
        }

        if self.wait_key_pressed == -1 {
            for i in 0..16 {
                if self.keys[i as usize] != 0 {
                    self.wait_key_pressed = i as i8;
                    break
                }
            }
        }

        if self.wait_key_pressed != -1 {
            if self.keys[self.wait_key_pressed as usize] == 0 {
                self.vc[self.wait_key_register as usize] = self.wait_key_pressed as u8;
                self.waiting_for_key = false;
                return
            }
        }

        self.pc -= 2;
    }


    // font character
    fn _fx29(&mut self) {
        let vx = ((self.oprand & 0x0F00) >> 8) as usize;
        self.i = self.vc[vx] as u16 * 0x0005;
    }


    // binary-coded decimal conversion
    fn _fx33(&mut self) {
        let vx = ((self.oprand & 0x0F00) >> 8) as usize;
        let val = self.vc[vx];
        
        self.memory[self.i as usize] = val / 100;
        self.memory[(self.i + 1) as usize] = (val / 10) % 10;
        self.memory[(self.i + 2) as usize] = val % 10
    }


    // store registers into memory
    fn _fx55(&mut self) {
        let vx = ((self.oprand & 0x0F00) >> 8) as usize;

        for idx in 0..=vx {
            self.memory[(self.i + idx as u16) as usize] = self.vc[idx]
        }
    }


    // load registers from memory
    fn _fx65(&mut self) {
        let vx = ((self.oprand & 0x0F00) >> 8) as usize;

        for idx in 0..=vx {
            self.vc[idx] = self.memory[(self.i + idx as u16) as usize]
        }
    }
}
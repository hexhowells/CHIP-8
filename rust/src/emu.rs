#[derive(Debug)]
pub struct CPU {
    pc: u16,
    i: u16,  // index register
    vc: [u8; 16],  // variable registers
    stack: [u16; 16],
    sp: u16,
    // keys: [u16; 16],
    // memory: [u8; 4096],
    // screen: [[u8; 32]; 64],
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

impl CPU {
    pub fn new() -> Self {
        Self {
            pc: 0x200,
            i: 0x0000,
            vc: [0x0000; 16],
            stack: [0x0000; 16],
            sp: 0x0000,
            opcode: 0x00,
            oprand: 0x0000
        }
    }


    pub fn run_instruction(&mut self, ins: u16) {
        self.opcode = ((ins & 0xF000) >> 12) as u8;
        self.oprand = ins & 0x0FFF;
        self._7XNN();
    }


    pub fn show_state(&self) {
        println!("{:?}", self);
    }


    // set VX to NN
    fn _6XNN(&mut self) {
        let vx = ((self.oprand & 0x0F00) >> 8) as usize;
        self.vc[vx] = (self.oprand & 0x00FF) as u8;
    }

    
    // add
    fn _7XNN(&mut self) {
        let vx = ((self.oprand & 0x0F00) >> 8) as usize;
        self.vc[vx] += (self.oprand & 0x00FF) as u8;
    }
}
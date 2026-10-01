mod emu;

fn main() {
    let mut cpu = emu::CPU::new();
    cpu.show_state();

    let n = 2;

    for _ in 0..n {
        cpu.run_instruction(0x0204);
        cpu.show_state();
    }

    
}
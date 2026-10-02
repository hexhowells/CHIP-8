mod emu;

fn main() {
    let mut cpu = emu::CPU::new();
    cpu.show_state();

    let _ = cpu.load_rom("../roms/test_opcode.ch8");

    for _ in 0..100 {
        cpu.clock();
        cpu.show_state();
    }
}
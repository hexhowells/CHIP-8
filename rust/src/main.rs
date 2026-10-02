mod emu;

fn main() {
    let mut cpu = emu::CPU::new();

    let _ = cpu.load_rom("../roms/test_opcode.ch8");

    for _ in 0..10_000 {
        cpu.clock();
    }

    cpu.show_state();
    cpu.display_screen();
}
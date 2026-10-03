use crossterm::{
    event::{self, Event, KeyCode, KeyEventKind},
    execute,
    terminal::{disable_raw_mode, enable_raw_mode, EnterAlternateScreen, LeaveAlternateScreen},
};
use ratatui::{
    backend::CrosstermBackend,
    widgets::{Block, Borders, Paragraph},
    Terminal,
};
use std::{
    io,
    time::{Duration, Instant},
};

mod emu;


fn main() -> io::Result<()> {
    enable_raw_mode()?;
    let mut stdout = io::stdout();
    execute!(stdout, EnterAlternateScreen)?;
    let backend = CrosstermBackend::new(stdout);
    let mut terminal = Terminal::new(backend)?;

    let mut cpu = emu::CPU::new();
    let _ = cpu.load_rom("../roms/keypad-test.ch8");

    let frame_rate = Duration::from_micros(16_666); // 60Hz
    let cycles_per_frame = 8; // 8 cycles * 60hz = 480Hz CPU speed

    let mut last_frame = Instant::now();

    loop {
        terminal.draw(|f| {
            let mut screen_text = String::with_capacity(64 * 32);
            for row in cpu.screen.iter() {
                for &pixel in row.iter() {
                    screen_text.push(if pixel == 1 { '█' } else { ' ' });
                }
                screen_text.push('\n');
            }

            let display = Paragraph::new(screen_text)
                .block(Block::default().title(" CHIP-8 ").borders(Borders::ALL));
            
            f.render_widget(display, f.size());
        })?;

        
        if last_frame.elapsed() >= frame_rate {
            if event::poll(Duration::ZERO)? {
                if let Event::Key(key) = event::read()? {

                    if key.kind == KeyEventKind::Press || key.kind == KeyEventKind::Release {

                        let is_down = key.kind == KeyEventKind::Press;

                        match key.code {
                            KeyCode::Esc => break,

                            KeyCode::Char('1') => cpu.press_key(0x1, is_down),
                            KeyCode::Char('2') => cpu.press_key(0x2, is_down),
                            KeyCode::Char('3') => cpu.press_key(0x3, is_down),
                            KeyCode::Char('4') => cpu.press_key(0xC, is_down),

                            KeyCode::Char('q') => cpu.press_key(0x4, is_down),
                            KeyCode::Char('w') => cpu.press_key(0x5, is_down),
                            KeyCode::Char('e') => cpu.press_key(0x6, is_down),
                            KeyCode::Char('r') => cpu.press_key(0xD, is_down),

                            KeyCode::Char('a') => cpu.press_key(0x7, is_down),
                            KeyCode::Char('s') => cpu.press_key(0x8, is_down),
                            KeyCode::Char('d') => cpu.press_key(0x9, is_down),
                            KeyCode::Char('f') => cpu.press_key(0xE, is_down),

                            KeyCode::Char('z') => cpu.press_key(0xA, is_down),
                            KeyCode::Char('x') => cpu.press_key(0x0, is_down),
                            KeyCode::Char('c') => cpu.press_key(0xB, is_down),
                            KeyCode::Char('v') => cpu.press_key(0xF, is_down),
                            _ => {}
                        }
                    }
                }
            }

            for _ in 0..cycles_per_frame {
                cpu.clock();
            }
            
            last_frame = Instant::now();
        }
    }

    disable_raw_mode()?;
    execute!(terminal.backend_mut(), LeaveAlternateScreen)?;
    terminal.show_cursor()?;

    Ok(())
}
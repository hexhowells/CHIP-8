package main

import (
	"CHIP8/emu"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nsf/termbox-go"
	"golang.org/x/sys/windows"
)

var (
	user32 = windows.NewLazySystemDLL("user32.dll")
	procGetAsyncKeyState = user32.NewProc("GetAsyncKeyState")
)

// emulator config variables
var clockHZ = 60
var romPath = "../roms/ibm-logo.ch8"


var keyMap = map[int]uint8{
	0x31: 0x1, 0x32: 0x2, 0x33: 0x3, 0x34: 0xC,
	0x51: 0x4, 0x57: 0x5, 0x45: 0x6, 0x52: 0xD,
	0x41: 0x7, 0x53: 0x8, 0x44: 0x9, 0x46: 0xE,
	0x5A: 0xA, 0x58: 0x0, 0x43: 0xB, 0x56: 0xF,
}

var keypadLayout = [][]uint8{
	{0x1, 0x2, 0x3, 0xC},
	{0x4, 0x5, 0x6, 0xD},
	{0x7, 0x8, 0x9, 0xE},
	{0xA, 0x0, 0xB, 0xF},
}

var hexToVK = make(map[uint8]int)


func init() {
	for vk, hex := range keyMap {
		hexToVK[hex] = vk
	}
}


func isKeyDown(vk int) bool {
	ret, _, _ := procGetAsyncKeyState.Call(uintptr(vk))
	return ret&0x8000 != 0
}


func main() {
	cpu := emu.NewCPU()
	cpu.LoadROM(romPath)

	if err := termbox.Init(); err != nil {
		panic(err)
	}
	defer termbox.Close()

	var (
		isPaused atomic.Bool
		mu       sync.Mutex
		done     = make(chan struct{})
		stepChan = make(chan struct{})
	)

	go handleConsoleEvents(cpu, done, stepChan, &isPaused)

	// gorountine for sending keyboard inputs to emulator
	go func() {
		ticker := time.NewTicker(16 * time.Millisecond)
		defer ticker.Stop()
		
		keyStates := make(map[int]bool)

		for {
			select {
			case <-ticker.C:
				mu.Lock()
				for vk, hexVal := range keyMap {
					currentlyDown := isKeyDown(vk)
					if currentlyDown != keyStates[vk] {
						cpu.PressKey(hexVal, currentlyDown)
						keyStates[vk] = currentlyDown
					}
				}
				renderUI(cpu, keyStates, isPaused.Load())
				mu.Unlock()
			case <-done:
				return
			}
		}
	}()

	clockTicker := time.NewTicker(time.Second / time.Duration(clockHZ))
	defer clockTicker.Stop()
	
	// run emulator
	for {
		select {
		case <-stepChan:
			mu.Lock()
			cpu.Clock()
			mu.Unlock()

		case <-clockTicker.C:
			if !isPaused.Load() {
				mu.Lock()
				cpu.Clock()
				mu.Unlock()
			}

		case <-done:
			return
		}
	}
}


func handleConsoleEvents(cpu *emu.CPU, done chan struct{}, stepChan chan struct{}, isPaused *atomic.Bool) {
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				close(done)
				return
			case termbox.KeySpace:
				isPaused.Store(!isPaused.Load())
			case termbox.KeyEnter:
				if isPaused.Load() {
					select {
					case stepChan <- struct{}{}:
					default:
					}
				}
			case termbox.KeyTab:
				cpu.Reset()
				cpu.LoadROM(romPath)
			}
		case termbox.EventError:
			close(done)
			return
		}
	}
}


func renderUI(cpu *emu.CPU, keyStates map[int]bool, paused bool) {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	for y := 0; y < 32; y++ {
		for x := 0; x < 64; x++ {
			if cpu.Screen[y][x] != 0 {
				termbox.SetCell(x*2, y, ' ', termbox.ColorWhite, termbox.ColorWhite)
				termbox.SetCell(x*2+1, y, ' ', termbox.ColorWhite, termbox.ColorWhite)
			} else {
				termbox.SetCell(x*2, y, '.', termbox.ColorBlack, termbox.ColorDefault)
				termbox.SetCell(x*2+1, y, ' ', termbox.ColorBlack, termbox.ColorDefault)
			}
		}
	}

	const (
		separatorX = 130
		offsetX    = 134
	)

	for y := 0; y < 32; y++ {
		termbox.SetCell(separatorX, y, '│', termbox.ColorWhite, termbox.ColorDefault)
	}

	if paused {
		drawString(offsetX, 1, "STATE: PAUSED [Enter to Step]", termbox.ColorRed)
	} else {
		drawString(offsetX, 1, "STATE: RUNNING [Space to Pause]", termbox.ColorGreen)
	}

	drawString(offsetX, 3, "CHIP-8 CPU INTERNALS", termbox.ColorYellow)
	drawString(offsetX, 4, "====================", termbox.ColorYellow)
	drawString(offsetX, 6, fmt.Sprintf("PC:          $%.4X", cpu.Pc), termbox.ColorDefault)
	drawString(offsetX, 7, fmt.Sprintf("I:           $%.4X", cpu.I), termbox.ColorDefault)
	drawString(offsetX, 8, fmt.Sprintf("SP:          $%.4X", cpu.Sp), termbox.ColorDefault)
	drawString(offsetX, 9, fmt.Sprintf("Opcode:      $%.4X", cpu.Opcode), termbox.ColorDefault)
	drawString(offsetX, 10, fmt.Sprintf("Oprand:      $%.4X", cpu.Oprand), termbox.ColorDefault)
	drawString(offsetX, 11, fmt.Sprintf("Delay Timer: $%.4X", cpu.DelayTimer), termbox.ColorDefault)
	drawString(offsetX, 12, fmt.Sprintf("Sound Timer: $%.4X", cpu.SoundTimer), termbox.ColorDefault)

	drawString(offsetX, 14, "REGISTERS (V0-VF)", termbox.ColorCyan)
	drawString(offsetX, 15, "-----------------", termbox.ColorCyan)

	for i := 0; i < 8; i++ {
		drawString(offsetX, 16+i, fmt.Sprintf("V%X: $%.2X", i, cpu.Vc[i]), termbox.ColorDefault)
		drawString(offsetX+14, 16+i, fmt.Sprintf("V%X: $%.2X", i+8, cpu.Vc[i+8]), termbox.ColorDefault)
	}

	drawString(offsetX, 26, "HEX KEYPAD MATRIX", termbox.ColorMagenta)
	drawString(offsetX, 27, "-----------------", termbox.ColorMagenta)

	for rowY, rowValues := range keypadLayout {
		for colX, hexVal := range rowValues {
			renderX := offsetX + (colX * 5)
			renderY := 28 + rowY

			vkCode := hexToVK[hexVal]
			fgColor := termbox.ColorWhite
			if keyStates[vkCode] {
				fgColor = termbox.ColorGreen
			}

			drawString(renderX, renderY, fmt.Sprintf("[%X]", hexVal), fgColor)
		}
	}

	termbox.Flush()
}


func drawString(x, y int, msg string, fg termbox.Attribute) {
	for i, ch := range msg {
		termbox.SetCell(x+i, y, ch, fg, termbox.ColorDefault)
	}
}
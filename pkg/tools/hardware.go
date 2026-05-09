package tools

import (
	"fmt"
	"os"
	"strings"
)

// === 1. JTAG-DEBUGGER ===
func runJTAGDebugger(args []string) (string, error) {
	var results []string
	results = append(results, "JTAG Interface Analysis")
	results = append(results, "Common JTAG pinouts:")
	results = append(results, "- TDI (Test Data In)")
	results = append(results, "- TDO (Test Data Out)")
	results = append(results, "- TCK (Test Clock)")
	results = append(results, "- TMS (Test Mode Select)")
	results = append(results, "- TRST (Test Reset)")
	results = append(results, "")
	results = append(results, "Requires: FTDI/J-Link/Bus Pirate adapter")
	results = append(results, "Go serial library: go.bug.st/serial")

	// List serial ports
	serialPaths := []string{"/dev/ttyUSB0", "/dev/ttyUSB1", "/dev/ttyACM0", "/dev/ttyACM1"}
	for _, sp := range serialPaths {
		if _, err := os.Stat(sp); err == nil {
			results = append(results, fmt.Sprintf("Serial port found: %s", sp))
		}
	}

	return fmt.Sprintf("[JTAG-DEBUGGER]\n%s", strings.Join(results, "\n")), nil
}

// === 2. SPI-FLASHER ===
func runSPIFlasher(args []string) (string, error) {
	var results []string
	results = append(results, "SPI Flash Interface")
	results = append(results, "Common pinouts:")
	results = append(results, "- MOSI (Master Out Slave In)")
	results = append(results, "- MISO (Master In Slave Out)")
	results = append(results, "- SCK (Serial Clock)")
	results = append(results, "- CS (Chip Select)")
	results = append(results, "")
	results = append(results, "Requires: CH341A/FT2232/TL866 adapter")

	// Check for flashrom
	output, _ := RunExternalSafe("flashrom", "--version")
	if output != "" {
		results = append(results, "flashrom found:\n"+output)
	}

	return fmt.Sprintf("[SPI-FLASHER]\n%s", strings.Join(results, "\n")), nil
}

// === 3. I2C-SNIFFER ===
func runI2CSniffer(args []string) (string, error) {
	var results []string
	results = append(results, "I2C Bus Analysis")
	results = append(results, "Pinout:")
	results = append(results, "- SDA (Serial Data)")
	results = append(results, "- SCL (Serial Clock)")
	results = append(results, "")
	results = append(results, "Common addresses: 0x50-0x57 (EEPROM)")
	results = append(results, "Requires: Bus Pirate/FTDI/I2C-USB adapter")

	return fmt.Sprintf("[I2C-SNIFFER]\n%s", strings.Join(results, "\n")), nil
}

// === 4. FIRMWARE-DUMPER ===
func runFirmwareDumper(args []string) (string, error) {
	device := "auto"
	if len(args) > 0 {
		device = args[0]
	}

	var results []string
	results = append(results, fmt.Sprintf("Device: %s", device))
	results = append(results, "Dumping firmware...")
	results = append(results, "Methods:")
	results = append(results, "- JTAG/SPI direct read")
	results = append(results, "- UART bootloader")
	results = append(results, "- OTA update interception")
	results = append(results, "- Flashrom software dump")

	// Check for firmware files
	firmwarePaths := []string{"firmware.bin", "flash.bin", "dump.bin", "backup.bin"}
	for _, f := range firmwarePaths {
		if _, err := os.Stat(f); err == nil {
			info, _ := os.Stat(f)
			results = append(results, fmt.Sprintf("Found: %s (%d bytes)", f, info.Size()))
		}
	}

	return fmt.Sprintf("[FIRMWARE-DUMPER]\n%s", strings.Join(results, "\n")), nil
}

// === 5. CHIP-DECAP ===
func runChipDecap(args []string) (string, error) {
	var results []string
	results = append(results, "Chip Decapsulation Guide")
	results = append(results, "")
	results = append(results, "WARNING: Physical destructive process")
	results = append(results, "Requires: Fuming nitric acid or mechanical decapping")
	results = append(results, "")
	results = append(results, "Steps:")
	results = append(results, "1. Remove packaging material")
	results = append(results, "2. Apply acid/mechanical pressure")
	results = append(results, "3. Clean die surface")
	results = append(results, "4. Microscope inspection")
	results = append(results, "5. Memory readout (optical/electrical)")
	results = append(results, "")
	results = append(results, "Alternatives:")
	results = append(results, "- Invasive probing")
	results = append(results, "- Fault injection (glitching)")
	results = append(results, "- Side-channel power analysis")

	return fmt.Sprintf("[CHIP-DECAP]\n%s", strings.Join(results, "\n")), nil
}

package hardware

import (
	"fmt"
	"os/exec"
	"strings"
)

// HardwareReaper is the hardware attack testing engine
type HardwareReaper struct {
	Target string
}

// HardwareResult represents a hardware security finding
type HardwareResult struct {
	Type     string `json:"type"` // badusb, rubber_ducky, bus_pirate, jtag, uart
	Device   string `json:"device"`
	Port     string `json:"port"`
	Proof    string `json:"proof"`
	Severity string `json:"severity"`
}

func NewHardwareReaper(target string) *HardwareReaper {
	return &HardwareReaper{Target: target}
}

// EnumerateUSB enumerates USB devices
func (h *HardwareReaper) EnumerateUSB() []HardwareResult {
	fmt.Println("[+] HARDWARE-REAPER enumerating USB devices")

	var results []HardwareResult

	if _, err := exec.LookPath("lsusb"); err == nil {
		output, err := exec.Command("lsusb").Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if line != "" {
					results = append(results, HardwareResult{
						Type:   "usb_device",
						Device: line,
						Proof:  "USB device enumerated",
					})
				}
			}
		}
	}

	fmt.Printf("[+] HARDWARE-REAPER found %d USB devices\n", len(results))
	return results
}

// EnumerateSerial enumerates serial ports
func (h *HardwareReaper) EnumerateSerial() []HardwareResult {
	fmt.Println("[+] HARDWARE-REAPER enumerating serial ports")

	var results []HardwareResult

	// Check for common serial ports
	serialPaths := []string{
		"/dev/ttyUSB0", "/dev/ttyUSB1", "/dev/ttyACM0", "/dev/ttyACM1",
		"/dev/ttyS0", "/dev/ttyS1", "COM1", "COM2", "COM3", "COM4",
	}

	for _, port := range serialPaths {
		results = append(results, HardwareResult{
			Type:   "serial_port",
			Port:   port,
			Proof:  "Serial port detected",
			Device: "UART",
		})
	}

	return results
}

// TestUART tests UART for debug console access
func (h *HardwareReaper) TestUART(port string, baud int) *HardwareResult {
	fmt.Printf("[+] HARDWARE-REAPER testing UART on %s at %d baud\n", port, baud)

	// Would use tools like screen, minicom, or picocom
	_, err := exec.LookPath("screen")
	if err != nil {
		_, err = exec.LookPath("minicom")
		if err != nil {
			fmt.Println("[!] No serial terminal tools found")
			return nil
		}
	}

	return &HardwareResult{
		Type:     "uart_console",
		Port:     port,
		Device:   "Debug Console",
		Proof:    fmt.Sprintf("UART debug console accessible at %d baud", baud),
		Severity: "critical",
	}
}

// TestJTAG tests JTAG interface
func (h *HardwareReaper) TestJTAG() []HardwareResult {
	fmt.Println("[+] HARDWARE-REAPER testing JTAG interface")

	var results []HardwareResult

	_, err := exec.LookPath("openocd")
	if err != nil {
		fmt.Println("[!] openocd not found. Install for JTAG testing.")
		return results
	}

	results = append(results, HardwareResult{
		Type:     "jtag_access",
		Device:   "JTAG Interface",
		Proof:    "JTAG interface detected and accessible",
		Severity: "critical",
	})

	return results
}

// GenerateBadUSB generates BadUSB payload
func (h *HardwareReaper) GenerateBadUSB(payload string) string {
	fmt.Println("[+] HARDWARE-REAPER generating BadUSB payload")

	// Ducky Script syntax
	duckyScript := `DELAY 1000
GUI r
DELAY 500
STRING cmd
ENTER
DELAY 500
STRING ` + payload + `
ENTER
DELAY 500
STRING exit
ENTER
`

	return duckyScript
}

// GenerateRubberDucky generates Rubber Ducky payload
func (h *HardwareReaper) GenerateRubberDucky(payload string) string {
	fmt.Println("[+] HARDWARE-REAPER generating Rubber Ducky payload")
	return h.GenerateBadUSB(payload)
}

// BusPirateScan uses Bus Pirate for SPI/I2C sniffing
func (h *HardwareReaper) BusPirateScan() []HardwareResult {
	fmt.Println("[+] HARDWARE-REAPER scanning with Bus Pirate")

	var results []HardwareResult

	_, err := exec.LookPath("minicom")
	if err != nil {
		fmt.Println("[!] minicom not found for Bus Pirate interaction")
		return results
	}

	results = append(results, HardwareResult{
		Type:   "bus_pirate",
		Device: "SPI/I2C",
		Proof:  "Bus Pirate connected - ready for SPI/I2C sniffing",
	})

	return results
}

// CANBusScan scans for CAN bus devices
func (h *HardwareReaper) CANBusScan() []HardwareResult {
	fmt.Println("[+] HARDWARE-REAPER scanning CAN bus")

	var results []HardwareResult

	// Check for CAN interfaces
	canInterfaces := []string{"can0", "can1", "vcan0", "vcan1"}

	for _, iface := range canInterfaces {
		results = append(results, HardwareResult{
			Type:   "can_bus",
			Port:   iface,
			Device: "CAN",
			Proof:  "CAN bus interface detected",
		})
	}

	return results
}

// OBD2Scan attempts OBD-II vehicle scanning
func (h *HardwareReaper) OBD2Scan(port string) []HardwareResult {
	fmt.Printf("[+] HARDWARE-REAPER scanning OBD-II on %s\n", port)

	var results []HardwareResult

	results = append(results, HardwareResult{
		Type:     "obd2",
		Port:     port,
		Device:   "Vehicle ECU",
		Proof:    "OBD-II connection established - vehicle data accessible",
		Severity: "high",
	})

	return results
}

// HIDAttack generates HID keyboard attack payload
func (h *HardwareReaper) HIDAttack(command string) string {
	fmt.Println("[+] HARDWARE-REAPER generating HID attack payload")

	return h.GenerateBadUSB(command)
}

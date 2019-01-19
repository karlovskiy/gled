package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/google/gousb"
	"image/color"
	"log"
	"strconv"
	"strings"
)

const (
	vendorId          = gousb.ID(0x046d) // Logitech, Inc.
	productId         = gousb.ID(0xc084) // G102 and G203 Prodigy Gaming Mouse
	format            = "11ff0e%s000000000000"
	defaultRate       = 10000
	defaultBrightness = 100
)

var (
	debug = flag.Int("debug", 0, "libusb debug level (0..3)")
)

func main() {
	flag.Usage = func() {
		fmt.Print(`Logitech G102 and G203 Prodigy Mouse LED control

Usage:
  gled solid <color>                         Solid color mode
  gled cycle <rate> <brightness>             Cycle through all colors
  gled breathe <color> <rate> <brightness>   Single color breathing
  gled intro <toggle>                        Enable/disable startup effect

Arguments:
  color        RRGGBB (RGB hex value)
  rate         100-60000 (Number of milliseconds. Default: 10000ms)
  brightness   0-100 (Percentage. Default: 100%)
  toggle       on|off

Flags:
  gled -debug <0..3> ...                     Debug level for libusb. Default: 0
`)
	}

	flag.Parse()
	mode := flag.Arg(0)
	switch mode {
	case "solid":
		setSolid()
	case "cycle":
		setCycle()
	case "breathe":
		setBreathe()
	case "intro":
		setIntro()
	default:
		flag.Usage()
		log.Fatalf("Unknown mode: %q", mode)
	}
}

func setIntro() {
	toggle := parseToggle(flag.Arg(1))
	sendCommand("5b0001" + toggle + "00000000000000")
}

func setSolid() {
	c := parseColor(flag.Arg(1))
	sendCommand("3b0001" + c + "0000000000")
}

func setCycle() {
	rate := parseRate(flag.Arg(1))
	brightness := parseBrightness(flag.Arg(2))
	sendCommand("3b0002" + "0000000000" + rate + brightness)
}

func setBreathe() {
	c := parseColor(flag.Arg(1))
	rate := parseRate(flag.Arg(2))
	brightness := parseBrightness(flag.Arg(3))
	sendCommand("3b0003" + c + rate + "00" + brightness + "00")
}

func sendCommand(data string) {
	fullData := fmt.Sprintf(format, data)
	log.Printf("Sending command: %s", fullData)
	payload, err := hex.DecodeString(fullData)
	if err != nil {
		log.Fatalf("Error converting data from hex string: %v", err)
	}

	// Only one context should be needed for an application.  It should always be closed.
	ctx := gousb.NewContext()
	defer ctx.Close()
	// Debugging can be turned on; this shows some of the inner workings of the libusb package.
	ctx.Debug(*debug)

	dev, err := ctx.OpenDeviceWithVIDPID(vendorId, productId)
	if err != nil {
		log.Fatalf("Error open device: %v", err)
	}
	defer dev.Close()
	// reset device is very important before send new control command in sequence command executions
	defer dev.Reset()

	if err := dev.SetAutoDetach(true); err != nil {
		log.Fatalf("Error set auto detach kernel for device: %v", err)
	}

	// Claim the default interface using a convenience function.
	// The default interface is always #0 alt #0 in the currently active
	// config.
	_, done, err := dev.DefaultInterface()
	if err != nil {
		log.Fatalf("Error claim default interface: %v", err)
	}
	defer done()

	n, err := dev.Control(0x21, 0x09, 0x0211, 0x01, payload)
	if err != nil {
		log.Fatalf("Error sending control data: %v", err)
	}

	log.Printf("%d bytes transferred to device", n)
}

func parseToggle(toggleArg string) string {
	var toggle string
	switch toggleArg {
	case "on":
		toggle = "01"
	case "off":
		toggle = "02"
	default:
		flag.Usage()
		log.Fatalf("Error parsing toggle argument: %q", toggleArg)
	}
	return toggle
}

func parseColor(colorArg string) string {
	if colorArg == "" {
		flag.Usage()
		log.Fatal("No color argument found")
	}
	if !strings.HasPrefix(colorArg, "#") {
		colorArg = "#" + colorArg
	}
	c, err := parseHexColor(colorArg)
	if err != nil {
		flag.Usage()
		log.Fatalf("Error parsing color argument: %q: %v", colorArg, err)
	}
	return fmt.Sprintf("%02x%02x%02x", c.R, c.G, c.B)
}

func parseRate(rateArg string) string {
	var rate int
	if rateArg == "" {
		rate = defaultRate
	} else {
		var err error
		rate, err = strconv.Atoi(rateArg)
		if err != nil {
			flag.Usage()
			log.Fatalf("Error parsing rate argument: %q: %v", rateArg, err)
		}
		if rate < 100 || rate > 60000 {
			flag.Usage()
			log.Fatalf("Rate argument: %q is out of range", rateArg)
		}
	}
	return fmt.Sprintf("%04x", rate)
}

func parseBrightness(brightnessArg string) string {
	var brightness int
	if brightnessArg == "" {
		brightness = defaultBrightness
	} else {
		var err error
		brightness, err = strconv.Atoi(brightnessArg)
		if err != nil {
			flag.Usage()
			log.Fatalf("Error parsing brightness argument: %q: %v", brightnessArg, err)
		}
		if brightness < 1 || brightness > 100 {
			flag.Usage()
			log.Fatalf("Brightness argument: %q is out of range", brightnessArg)
		}
	}
	return fmt.Sprintf("%02x", brightness)
}

func parseHexColor(s string) (c color.RGBA, err error) {
	c.A = 0xff
	switch len(s) {
	case 7:
		_, err = fmt.Sscanf(s, "#%02x%02x%02x", &c.R, &c.G, &c.B)
	case 4:
		_, err = fmt.Sscanf(s, "#%1x%1x%1x", &c.R, &c.G, &c.B)
		// Double the hex digits:
		c.R *= 17
		c.G *= 17
		c.B *= 17
	default:
		err = fmt.Errorf("error parsing HEX color string")
	}
	return
}

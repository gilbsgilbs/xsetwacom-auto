package main

import (
	"fmt"
	"math"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/gilbsgilbs/xsetwacomauto"
	"github.com/urfave/cli/v2"
	"github.com/vcraescu/go-xrandr"
)

type Size struct {
	Width  int
	Height int
}

func monitorToString(monitor xrandr.Monitor) string {
	s := fmt.Sprintf(
		"%s (%dx%d)",
		monitor.ID,
		int(monitor.Resolution.Width),
		int(monitor.Resolution.Height),
	)

	if monitor.Primary {
		s += " [Primary]"
	}

	return s
}

func monitorsToString(monitors []xrandr.Monitor) []string {
	result := make([]string, len(monitors))
	for i, monitor := range monitors {
		result[i] = monitorToString(monitor)
	}
	return result
}

func getMonitors() ([]xrandr.Monitor, error) {
	screens, err := xrandr.GetScreens()
	if err != nil {
		return nil, err
	}

	result := []xrandr.Monitor{}
	for _, screen := range screens {
		for _, monitor := range screen.Monitors {
			if monitor.Connected {
				result = append(result, monitor)
			}
		}
	}

	return result, nil
}

func findPrimaryMonitor(monitors []xrandr.Monitor) (int, *xrandr.Monitor) {
	for i, monitor := range monitors {
		if monitor.Primary {
			return i, &monitor
		}
	}
	return 0, nil
}

func computeArea(
	screenSize Size,
	originalArea xsetwacomauto.XSetWacomDeviceArea,
) xsetwacomauto.XSetWacomDeviceArea {
	newArea := originalArea
	screenRatio := float64(screenSize.Width) / float64(screenSize.Height)
	tabletRatio := float64(originalArea.X2-originalArea.X1) / float64(originalArea.Y2-originalArea.Y1)

	if tabletRatio < screenRatio {
		newArea.Y2 = int(math.Round(float64(originalArea.X2-originalArea.X1)/screenRatio)) + newArea.Y1
	} else {
		newArea.X2 = int(math.Round(float64(originalArea.Y2-originalArea.Y1)*screenRatio)) + newArea.X1
	}

	return newArea
}

func promptForDevices(allDevices []xsetwacomauto.XSetWacomDevice) ([]*xsetwacomauto.XSetWacomDevice, error) {
	options := make([]string, len(allDevices))
	for i, device := range allDevices {
		options[i] = device.String()
	}

	prompt := survey.MultiSelect{
		Message: "Choose your devices:",
		Options: options,
		Default: options,
	}

	var devicesIdx []int
	if err := survey.AskOne(&prompt, &devicesIdx); err != nil {
		return nil, err
	}

	result := make([]*xsetwacomauto.XSetWacomDevice, len(devicesIdx))
	for i, deviceIdx := range devicesIdx {
		result[i] = &allDevices[deviceIdx]
	}

	return result, nil
}

func promptForMonitor(allMonitors []xrandr.Monitor) (*xrandr.Monitor, error) {
	primaryMonitorPos, _ := findPrimaryMonitor(allMonitors)
	options := monitorsToString(allMonitors)
	prompt := survey.Select{
		Message: "Choose your monitor:",
		Options: options,
		Default: options[primaryMonitorPos],
	}

	var monitorIdx int
	if err := survey.AskOne(&prompt, &monitorIdx); err != nil {
		return nil, err
	}

	return &allMonitors[monitorIdx], nil
}

func promptPreserveAspect(defaultValue bool) (bool, error) {
	prompt := survey.Confirm{
		Message: "Preserve aspect ratio?",
		Default: defaultValue,
	}

	var preserveAspectRatio bool
	err := survey.AskOne(&prompt, &preserveAspectRatio)

	return preserveAspectRatio, err
}

func run(interactive bool, preserveAspect bool) error {
	devices, err := xsetwacomauto.ListDevices()
	if err != nil {
		return err
	}
	monitors, err := getMonitors()
	if err != nil {
		return err
	}

	var selectedDevices []*xsetwacomauto.XSetWacomDevice
	var selectedMonitor *xrandr.Monitor
	if interactive {
		selectedDevices, err = promptForDevices(devices)
		if err != nil {
			return err
		}

		selectedMonitor, err = promptForMonitor(monitors)
		if err != nil {
			return err
		}

		preserveAspect, err = promptPreserveAspect(preserveAspect)
		if err != nil {
			return err
		}
	} else {
		selectedDevices = make([]*xsetwacomauto.XSetWacomDevice, len(devices))
		for i := range devices {
			selectedDevices[i] = &devices[i]
		}

		_, selectedMonitor = findPrimaryMonitor(monitors)
		if selectedMonitor == nil {
			selectedMonitor = &monitors[0]
		}
	}

	screenSize := Size{
		Width:  int(selectedMonitor.Resolution.Width),
		Height: int(selectedMonitor.Resolution.Height),
	}

	for _, device := range selectedDevices {
		fmt.Printf("Mapping device `%s` to monitor `%s`.\n", device, monitorToString(*selectedMonitor))

		if preserveAspect {
			originalArea, err := xsetwacomauto.GetOriginalArea(*device)
			if err != nil {
				return err
			}

			newArea := computeArea(screenSize, originalArea)
			xsetwacomauto.SetArea(*device, newArea)
		} else {
			if err := xsetwacomauto.ResetArea(*device); err != nil {
				return err
			}
		}

		if err := xsetwacomauto.MapToOutput(*device, selectedMonitor.ID); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	app := &cli.App{
		Name: "xsetwacom-auto",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "interactive",
				Value:   false,
				Aliases: []string{"i"},
				Usage:   "Run in interactive mode",
			},
			&cli.BoolFlag{
				Name:    "preserve-aspect-ratio",
				Value:   true,
				Aliases: []string{"p"},
				Usage:   "Make the device area proportional to the screen.",
			},
		},
		Action: func(c *cli.Context) error {
			return run(c.Bool("interactive"), c.Bool("preserve-aspect-ratio"))
		},
	}
	app.Run(os.Args)
}

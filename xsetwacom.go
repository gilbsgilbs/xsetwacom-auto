package xsetwacomauto

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type XSetWacomDevice struct {
	Name string
	ID   int
	Type string
}

func (d XSetWacomDevice) String() string {
	return d.Name
}

type XSetWacomDeviceArea struct {
	X1 int
	Y1 int
	X2 int
	Y2 int
}

func ListDevices() ([]XSetWacomDevice, error) {
	cmd := exec.Command("xsetwacom", "--list", "devices")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("xsetwacom list command failed: %w", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	scanner.Split(bufio.ScanLines)

	result := []XSetWacomDevice{}
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "\t")

		device := XSetWacomDevice{}
		device.Name = strings.TrimSpace(parts[0])
		id, err := strconv.Atoi(
			strings.TrimSpace(strings.TrimPrefix(parts[1], "id: ")),
		)
		if err != nil {
			return nil, fmt.Errorf("wacom device id was not an integer: %w", err)
		}
		device.ID = id
		device.Type = strings.TrimSpace(strings.TrimPrefix(parts[2], "type: "))

		result = append(result, device)
	}

	return result, nil
}

func SetArea(device XSetWacomDevice, area XSetWacomDeviceArea) error {
	cmd := exec.Command(
		"xsetwacom",
		"--set",
		strconv.Itoa(device.ID),
		"Area",
		strconv.Itoa(area.X1),
		strconv.Itoa(area.Y1),
		strconv.Itoa(area.X2),
		strconv.Itoa(area.Y2),
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("couldn't set device area: %w", err)
	}

	return nil
}

func ResetArea(device XSetWacomDevice) error {
	cmd := exec.Command("xsetwacom", "--set", strconv.Itoa(device.ID), "ResetArea")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("couldn't reset device area: %w", err)
	}
	return nil
}

func GetArea(device XSetWacomDevice) (XSetWacomDeviceArea, error) {
	cmd := exec.Command("xsetwacom", "--get", strconv.Itoa(device.ID), "Area")
	output, err := cmd.Output()
	if err != nil {
		return XSetWacomDeviceArea{}, fmt.Errorf("couldn't get device area: %w", err)
	}

	result := XSetWacomDeviceArea{}
	coords := strings.Split(strings.TrimSpace(string(output)), " ")

	for i, coord := range []*int{&result.X1, &result.Y1, &result.X2, &result.Y2} {
		*coord, err = strconv.Atoi(coords[i])
		if err != nil {
			return result, fmt.Errorf("couldn't convert a coordinate to integer: %w", err)
		}
	}

	return result, nil
}

func GetOriginalArea(device XSetWacomDevice) (XSetWacomDeviceArea, error) {
	currentArea, err := GetArea(device)
	if err != nil {
		return XSetWacomDeviceArea{}, err
	}

	if err := ResetArea(device); err != nil {
		return XSetWacomDeviceArea{}, err
	}

	originalArea, err := GetArea(device)
	if err != nil {
		return XSetWacomDeviceArea{}, err
	}

	if err := SetArea(device, currentArea); err != nil {
		return XSetWacomDeviceArea{}, err
	}

	return originalArea, nil
}

func MapToOutput(device XSetWacomDevice, outputName string) error {
	cmd := exec.Command("xsetwacom", "--set", strconv.Itoa(device.ID), "MapToOutput", outputName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("couldn't map device to output: %w", err)
	}
	return nil
}

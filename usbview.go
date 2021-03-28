package main

import (
	"bufio"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

/*
type addchild interface {
	AddChild(cd *ChildDevice)
	LastChild() *ChildDevice
}
*/
// ChildDevice is USB device include USBHUB
type ChildDevice struct {
	Level    int
	Port     int
	Class    string
	Dev      int
	Children []*ChildDevice
}

// AddChild add same level
func (cd *ChildDevice) AddChild(child *ChildDevice) {
	if cd.Level+1 == child.Level {
		cd.Children = append(cd.Children, child)
	} else {
		cd.Children[len(cd.Children)-1].AddChild(child)
	}
}

// LastChild add same level
func (cd *ChildDevice) LastChild() (*ChildDevice, error) {
	if len(cd.Children) == 0 {
		return nil, errors.New("array is empty")
	}
	return cd.Children[len(cd.Children)-1], nil
}

// BusDev is every Bus device tree
type BusDev struct {
	Bus      int
	Port     int
	Dev      int
	Class    string
	Children []*ChildDevice
}

// AddChild add same level
func (bd *BusDev) AddChild(child *ChildDevice) {
	if child.Level == 0 {
		bd.Children = append(bd.Children, child)
	} else {
		bd.Children[len(bd.Children)-1].AddChild(child)
	}
}

// LastChild add same level
func (bd *BusDev) LastChild() (*ChildDevice, error) {
	if len(bd.Children) == 0 {
		return nil, errors.New("array is empty")
	}
	return bd.Children[len(bd.Children)-1], nil
}

// USBVIEW is on system tree
type USBVIEW struct {
	Buses []*BusDev
}

// Add to BUS
func (uv *USBVIEW) Add(db *BusDev) {
	uv.Buses = append(uv.Buses, db)
}

// AddChild to
func (uv *USBVIEW) AddChild(cd *ChildDevice) {
	vb := uv.Buses[len(uv.Buses)-1]
	vb.AddChild(cd)
}

func (uv *USBVIEW) getDev(pathes []string, Children []*ChildDevice) (int, error) {
	index, err := strconv.Atoi(pathes[0])
	if err != nil {
		return 0, err
	}
	if len(pathes) == 1 {
		for _, child := range Children {
			if child.Port == index {
				return child.Dev, nil
			}
		}
	}
	for _, child := range Children {
		if child.Port == index {
			return uv.getDev(pathes[1:], child.Children)
		}
	}
	return 0, errors.New("not found")
}

// GetDevFromPathes get from pathes int array
func (uv *USBVIEW) GetDevFromPathes(pathes []string) (int, error) {
	if len(pathes) < 2 {
		return 0, errors.New("usb device is not BUS. one is BUS")
	}
	var nDev int
	busindex, err := strconv.Atoi(pathes[0])
	if err != nil {
		return 0, err
	}

	for _, bs := range uv.Buses {
		if bs.Bus == busindex {
			return uv.getDev(pathes[1:], bs.Children)
		}
	}
	return nDev, nil

}

// GetDevFromPath get device
func (uv *USBVIEW) GetDevFromPath(path string) (int, error) {
	if path == "" {
		return 0, errors.New("path is empty")
	}
	pathes := strings.Split(path, "-")
	if len(pathes) < 2 {
		return 0, errors.New("usb device is not BUS. one is BUS")
	}
	var nDev int
	busindex, err := strconv.Atoi(pathes[0])
	if err != nil {
		return 0, err
	}

	for _, bs := range uv.Buses {
		if bs.Bus == busindex {
			return uv.getDev(pathes[1:], bs.Children)
		}
	}
	return nDev, nil
}

var usbview = &USBVIEW{
	Buses: []*BusDev{},
}

func parseBus(line string, bus *BusDev) {
	// /:  Bus 04.Port 1: Dev 1, Class=root_hub, Driver=xhci_hcd/4p, 5000M
	re := regexp.MustCompile(`^/:  Bus (\d+)\.Port (\d+): Dev (\d+), Class=(.*?), Driver=.*?, .*?$`)
	items := re.FindStringSubmatch(line)
	fmt.Println(items)
	if len(items) == 5 {
		bus.Bus, _ = strconv.Atoi(items[1])
		bus.Port, _ = strconv.Atoi(items[2])
		bus.Dev, _ = strconv.Atoi(items[3])
		bus.Class = items[4]
	}
}

func parseChild(line string, child *ChildDevice) {
	//    |__ Port 1: Dev 2, If 0, Class=Hub, Driver=hub/4p, 5000M
	re := regexp.MustCompile(`^([ ]+)\|__ Port (\d+): Dev (\d+),.*?, Class=(.*?), Driver=.*?, .*?$`)
	items := re.FindStringSubmatch(line)
	fmt.Println(items)
	if len(items) == 5 {
		child.Level = (len(items[1]) / 4) - 1
		child.Port, _ = strconv.Atoi(items[2])
		child.Dev, _ = strconv.Atoi(items[3])
		child.Class = items[4]
	}
}

// RunLsusb Run ls Usb
func RunLsusb() {
	cmd := "lsusb -t"
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		FDLogger.Printf("Failed to execute command: %s, %s\n", cmd, err)
		return
	}
	fmt.Printf("result execute command: %s, %s\n", cmd, string(out))

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	var line string
	for scanner.Scan() {
		line = scanner.Text()
		fmt.Println(line) // Println will add back the final '\n'
		if strings.HasPrefix(line, "/:  Bus") {
			bus := &BusDev{
				Bus:      0,
				Port:     0,
				Dev:      0,
				Class:    "",
				Children: []*ChildDevice{},
			}
			parseBus(line, bus)
			fmt.Println(bus)
			usbview.Add(bus)
		} else {
			child := &ChildDevice{
				Level:    0,
				Port:     0,
				Class:    "",
				Dev:      0,
				Children: []*ChildDevice{},
			}
			parseChild(line, child)
			usbview.AddChild(child)
		}
	}
}

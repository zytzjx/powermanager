package main

import (
	"fmt"
	"testing"
)

func TestUsbView(t *testing.T) {
	RunLsusb()
	fmt.Println(usbview)
}

func TestGetDevFromPath(t *testing.T) {
	RunLsusb()
	dev, err := usbview.GetDevFromPath("3/1")
	if err != nil {
		t.Error(err)
	}
	t.Log(dev)
}

func TestParseBus(t *testing.T) {
	bus := BusDev{
		Bus:      0,
		Port:     0,
		Dev:      0,
		Class:    "",
		Children: []*ChildDevice{},
	}
	line := `/:  Bus 04.Port 1: Dev 1, Class=root_hub, Driver=xhci_hcd/4p, 5000M`
	parseBus(line, &bus)
	fmt.Println(bus)
	line = `/:  Bus 03.Port 1: Dev 1, Class=root_hub, Driver=xhci_hcd/14p, 480M`
	parseBus(line, &bus)
	fmt.Println(bus)
	line = `/:  Bus 02.Port 1: Dev 1, Class=root_hub, Driver=ehci-pci/2p, 480M`
	parseBus(line, &bus)
	fmt.Println(bus)
	data := `/:  Bus 04.Port 1: Dev 1, Class=root_hub, Driver=xhci_hcd/4p, 5000M
    |__ Port 1: Dev 2, If 0, Class=Hub, Driver=hub/4p, 5000M
	    |__ Port 1: Dev 3, If 0, Class=Hub, Driver=hub/4p, 5000M
/:  Bus 03.Port 1: Dev 1, Class=root_hub, Driver=xhci_hcd/14p, 480M
    |__ Port 1: Dev 25, If 0, Class=Hub, Driver=hub/4p, 480M
	    |__ Port 4: Dev 27, If 0, Class=Vendor Specific Class, Driver=ch341, 12M
	    |__ Port 1: Dev 26, If 0, Class=Hub, Driver=hub/4p, 480M
	|__ Port 2: Dev 2, If 0, Class=Vendor Specific Class, Driver=mt76x0u, 480M
	|__ Port 4: Dev 3, If 0, Class=Video, Driver=uvcvideo, 480M
	|__ Port 4: Dev 3, If 1, Class=Video, Driver=uvcvideo, 480M
	|__ Port 6: Dev 14, If 0, Class=Vendor Specific Class, Driver=ch341, 12M
	|__ Port 7: Dev 6, If 0, Class=Wireless, Driver=btusb, 12M
	|__ Port 7: Dev 6, If 1, Class=Wireless, Driver=btusb, 12M
	|__ Port 8: Dev 5, If 0, Class=Vendor Specific Class, Driver=rtsx_usb, 480M
/:  Bus 02.Port 1: Dev 1, Class=root_hub, Driver=ehci-pci/2p, 480M
    |__ Port 1: Dev 2, If 0, Class=Hub, Driver=hub/8p, 480M
/:  Bus 01.Port 1: Dev 1, Class=root_hub, Driver=ehci-pci/2p, 480M
    |__ Port 1: Dev 2, If 0, Class=Hub, Driver=hub/6p, 480M`
	fmt.Println(data)
}

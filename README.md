### go server
   go server control arduino.  

### for example:
All of these use HTTP GET
http://localhost:8010/12?on=1  
http://localhost:8010/11?on=0  


### seting ttyUSB permision
```
 sudo usermod -a -G dialout $USER
 sudo usermod -a -G tty $USER
 reboot
```
curl -H "Content-Type: application/json" -d '{"name":"xyz","password":"xyz"}' http://localhost:8010/exitsystem

```
lsusb -t
/:  Bus 04.Port 1: Dev 1, Class=root_hub, Driver=xhci_hcd/4p, 5000M
/:  Bus 03.Port 1: Dev 1, Class=root_hub, Driver=xhci_hcd/14p, 480M
    |__ Port 1: Dev 11, If 0, Class=Vendor Specific Class, Driver=ch341, 12M
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
    |__ Port 1: Dev 2, If 0, Class=Hub, Driver=hub/6p, 480M

lsusb -t
/:  Bus 04.Port 1: Dev 1, Class=root_hub, Driver=xhci_hcd/4p, 5000M
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
    |__ Port 1: Dev 2, If 0, Class=Hub, Driver=hub/6p, 480M
```
if two usb device. now need calibration file  
calibration.json
```
{
   "power":"3/1",
   "lifting":"Bus 003 Device 008: ID 1a86:7523 QinHeng Electronics HL-340 USB-Serial adapter"
}
```

### lifting control
* Get 
   
http://localhost:8010/lift/hello  
Command: AT  
http://localhost:8010/lift/status  
Command: ATS  

http://localhost:8010/lift/info  
Command: ATI

http://localhost:8010/lift/position  
http://localhost:8010/lift/position?p=[0..9]  
Command: ATP? or ATPn?

http://localhost:8010/lift/setpos?p=[0..9]&value=+2345.1234  
Command: ATP0=+2345.1234

http://localhost:8010/lift/go?p=[0..9]  
Command: ATG0

http://localhost:8010/lift/flip?flag=[+-]    
Command: ATF+/ATF-

http://localhost:8010/lift/turn?flag=[+-]    
Command: ATT+/ATT-

http://localhost:8010/lift/home  
Command: ATC  

http://localhost:8010/lift/reset  
Command: ATZ  

http://localhost:8010/lift/stop  
Command: ATSTOP    

http://localhost:8010/exitsystem  
shut down server  





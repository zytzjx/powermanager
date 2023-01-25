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

serialcalibration.json, if PC only has two USB to serial port. this config can be removed.
```
udevadm info -q path -n /dev/ttyUSB0  
{
    "power":"3-1",
    "powerbaudrate":9600,
    "lifting":"3-6",
    "liftingbaudrate":9600
 }
```

### lifting control
* Get 
   all lift command can add interval=50 param, for example: http://localhost:8010/lift/hello?interval=30
   it means send one char sleep 30ms and send next. default is 1ms
   http request paramter delay=10000(ms), means send data=>delay=>receive data: example:
   ```
   http://localhost:8010/lift/home?flag=3&delay=15000  
   ```

http://localhost:8010/lift/hello  
Command: AT  
http://localhost:8010/lift/status  
Command: ATS  

http://localhost:8010/lift/info  
Command: ATI   

http://localhost:8010/lift/position  
http://localhost:8010/lift/position?p=[-1..9]  
Command: ATP? or ATPn?   

Query Air Compress Value    
http://localhost:8010/lift/queryair   
Command: ATQ

http://localhost:8010/lift/setpos?p=[-1..9]&value=+2345.1234  
Command: ATP0=+2345.1234   

http://localhost:8010/lift/go?p=[-1..9]  
Command: ATG0  

http://localhost:8010/lift/flip?flag=[+-]    
Command: ATF+/ATF-  

http://localhost:8010/lift/turn?flag=[+-]    
Command: ATT+/ATT-  

http://localhost:8010/lift/home?flag=[123]  
Command: ATC1/ATC2/ATC3  

http://localhost:8010/lift/reset  
Command: ATZ  

http://localhost:8010/lift/stop  
Command: ATSTOP    

http://localhost:8010/lift/getlasterror  
Command: ATE?      

http://localhost:8010/lift/start  
Command: ATSTART    


http://localhost:8010/lift/wind?flag=[012]   
Command: ATW[012]  

http://localhost:8010/lift/carrier?flag=[01]   
Command: ATX[01]  

http://localhost:8010/lift/reconnect  
reconnect serial . close and reopen  


http://localhost:8010/exitsystem  
shut down server  




### Note
if not find config json file:    
|serialport|ch340|other|result|
|------|------|------|---|
|1|yes|no|control LED|
|1|no|yes|control lifting|
|2|yes|yes|Led and Lifting|
|2|yes|no|last serial LED|
|2|no|yes|last serial lifting|
|>2|any|any|need config json|

if lifting PID VID defined, software will be enhanced.

### voltage:
http://localhost:8010/level/voltage?v=1180  
http://localhost:8010/level/poweron  
http://localhost:8010/level/poweroff  

### tricolored light
http://localhost:8010/tricl/green  
http://localhost:8010/tricl/yellow  
http://localhost:8010/tricl/red  
http://localhost:8010/tricl/clean  

if you want to light multi LED. you can use  
   http://localhost:8010/x?on=1  x is 2,3,7

### keys
check redis DB
Publish key: tricoloredlight  
value is regex: "Key:\s*(\d),\s*(\d),\s*(\d)"  
group 1, 2, 3 means key 1, key2 , key3 status.  




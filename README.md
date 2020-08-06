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


if two usb device. now need calibration file  
calibration.json
```
{
   "power":"Bus 003 Device 007: ID 1a86:7523 QinHeng Electronics HL-340 USB-Serial adapter",
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





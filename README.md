### go server
   go server control arduino.  

### for example:
http://localhost:8010/12?on=1  
http://localhost:8010/11?on=0  


### seting ttyUSB permision
```
 sudo usermod -a -G dialout $USER
 sudo usermod -a -G tty $USER
 reboot
```
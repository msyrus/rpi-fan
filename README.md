# rpi-fan
Temperature controlled raspberry pi fan service

It triggers on gpio26 for cpu temperature heigher than 45 degree and
triggers off gpio26 for cpu temperature less than 35 degree.

It takes 20 samples for 20 seconds to colculate the average temperature.

## Install
```
make install
```
N.B: Install requires root user permission

## Todo
* Allow configuration
* Optimize memory usage

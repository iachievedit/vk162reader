# GPS Coordinate Reader 

VK-162 USB GPS Dongle - Remote Mount USB

A small application
that reads the output of a VK-162 USB GPS dongle and then publishes
the GPS coordinates found on a GPGLL line to a ZeroMQ socket.

Why?  Because currently my 
[weatherfxlite](https://github.com/iachievedit/weatherfxlite) application requires you to hardcode
the GPS location in a header file, and I thought well, why not let it update
if it is being fed new coordinates?

So weatherfxlite will be upgraded to listen on ZeroMQ, and if you have the
incredibly expensive GPS Aware addon license installed, you get a location-aware
weather display.  Nifty.  🙃
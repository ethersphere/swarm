# Dynamic Application Loader (DAL) Host Interface (aka JHI)

## Description
A daemon and libraries which allow user space applications to install Java applets on DAL FW and communicate with them.

## Features
* Allows multiple client applications to communicate with Intel DAL FW simultaneously.
* Autodetects the FW version and formats messages accordingly.
* Caches previously installed applets for easy reuse.

## Dependencies
```
cmake
uuid-dev
libxml2-dev
```

## How to build
```
cmake .
make
```
The output directory is ```bin_linux```.

## Build options
Release build:
```
cmake . -DCMAKE_BUILD_TYPE=Release
```

Use SysVinit instead of systemd:
```
cmake . -DINIT_SYSTEM=SysVinit
```

## How to install
```
sudo make install
```

## How to manage the daemon
```
systemctl {enable|disable|start|stop|restart|status} jhi
```

## How to check which version of JHI is installed
```
jhid -v
```

## How to run without init system integration
```jhid``` (run as root)

Alternatives:
* ```jhid -d```   (to run in the background)
* ```jhid 2>&1``` (to redirect stderr to stdout)

## Config file
Location:
```
/etc/jhi/jhi.conf
```

Options:
1. Run over mei/kernel/sockets.
2. If running over sockets, specify the IP of the server.
3. State the desired log level.
4. Change the daemon socket location.

## An integartion test - bist (for internal use)
> Note: The test works on Intel platforms SKL and newer.

Location:
```
bin_linux/bist
```
How to run:
```
./bist
```
All tests should pass.


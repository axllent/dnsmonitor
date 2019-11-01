# DNSMonitor - A simple DNS monitor written in go

[![Go Report Card](https://goreportcard.com/badge/github.com/axllent/dnsmonitor)](https://goreportcard.com/report/github.com/axllent/dnsmonitor)

DNSMonitor is a small commandline utility which queries a DNS for a specific hostname (or hostnames) on a regular interval, optionally alerting you (Pushbullet and/or Gotify) on any DNS change.

## Features

- Defaults to the network-configured DNS, however a separate DNS server can be specified
- Polling interval (default every 5 minutes)
- Supports querying of A, MX, CNAME, TXT & NS records (see usage)
- Optionally send change alerts through [Pushbullet](https://www.pushbullet.com/) and/or [Gotify](https://gotify.net/)


## Usage examples

```
dnsmonitor www.example.com
```
Monitor the **A** records of `www.example.com` on a 5-minute interval.

```
dnsmonitor mx:example.com
```
Monitor the **MX** records of `example.com` on a 5-minute interval.


```
dnsmonitor ns:example.com txt:example.com www.example.com
```
Monitor the **NS** & **TXT** records of `example.com`, and the **A** records of `www.example.com` on a 5-minute interval.


```
dnsmonitor -d 1.1.1.1 example.com -i 10
```
Monitor the **A** records of `example.com` on a 10-minute interval using the DNS server on `1.1.1.1`.

See `dnsmonitor -h` for all options.


## Notifications

Currently Pushbullet & Gotify notifications are supported.

Create a default configuration file in `~/.config/dnsmonitor.json` or use the `-c` flag to spefify an alternative configuration file and set the values:


### Pushbullet

You can find your Access Token on https://www.pushbullet.com/#settings/account and the device name is exactly as it appears in your list of devices.

```json
{
  "pushbullet_token":"<token>",
  "pushbullet_device":"<device name>"
}
```


### Gotify

Create a new App on your gotify instance which will generate a unique token. 

```json
{
  "gotify_server":"<https://your-gotify-server/>",
  "gotify_token":"<token>"
}
```


## Installing

Multiple OS/Architecture binaries are supplied with [releases](https://github.com/axllent/dnsmonitor/releases). 
Extract the binary, make it executable, and move it to a location such as `/usr/local/bin`.


## Updating

DNSMonitor comes with a built-in self-updater:

```
dnsmonitor -u
```


## Compiling from source

Go >= 1.11 required:

```
go get github.com/axllent/dnsmonitor
```

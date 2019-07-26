# go-modemmanager
A Go wrapper around ModemManager's D-Bus API
### 🚨🚧 WIP/Experimental 🚧🚨

---

# What this is
- Wrapper for ModemManager's D-Bus API in Go
- (Mostly) usable
- Not battle-tested
- Unpolished

# Usage
```bash
$ go get -u github.com/parrotmac/go-modemmanager
```

# Run Example
```
$ go run cmd/dump.go
```
On a system with two Modems, the (formatted & slightly tweaked) output might look like this:
```json
/* Listing modems */
[
  {
    "bearer_paths": [
      "/org/freedesktop/ModemManager1/Bearer/1"
    ],
    "sim_path": "/org/freedesktop/ModemManager1/SIM/285",
    "model": "MS2372h-517",
    "manufacturer": "huawei",
    "device": "/sys/devices/pci0000:00/0000:00:14.0/usb1/1-1",
    "primary_port": "ttyUSB2",
    "own_numbers": [],
    "revision": "21.328.02.01.00",
    "hardware_revision": "",
    "equipment_identifier": "866667030000000",
    "drivers": [
      "huawei_cdc_ncm",
      "option1"
    ],
    "signal_quality": {
      "percent": 77,
      "recent": true
    }
  },
  {
    "bearer_paths": [],
    "sim_path": "/org/freedesktop/ModemManager1/SIM/284",
    "model": "DW5811e Snapdragon™ X7 LTE",
    "manufacturer": "Sierra Wireless, Incorporated",
    "device": "/sys/devices/pci0000:00/0000:00:14.0/usb1/1-6",
    "primary_port": "cdc-wdm0",
    "own_numbers": [
      "+13130000000"
    ],
    "revision": "SWI9X30C_02.24.05.06",
    "hardware_revision": "EM7455B",
    "equipment_identifier": "354479080000000",
    "drivers": [
      "cdc_mbim"
    ],
    "signal_quality": {
      "percent": 0,
      "recent": false
    }
  }
]
/* Listing active Bearer(s) */
{
  "connected": true,
  "suspended": false,
  "interface": "wwx001e101f0000"
}
/* Listing SIMs */
{
  "imsi": "295050910000000",
  "operator_identifier": "29505",
  "operator_name": "SORACOM",
  "sim_identifier": "8942310017000000000"
}
{
  "imsi": "310260850000000",
  "operator_identifier": "310260",
  "operator_name": "T-Mobile",
  "sim_identifier": "8901260852000000000"
}
```

# Contributions
PRs and Issues (even for Q's) welcome

# License
ISC License (ISC)
Copyright 2019 Isaac Parker

Permission to use, copy, modify, and/or distribute this software for any purpose with or without fee is hereby granted, provided that the above copyright notice and this permission notice appear in all copies.

THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

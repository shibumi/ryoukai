// Copyright 2019 Christian Rebischke <chris@nullday.de>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This work is based on the i3status bar https://github.com/soumya92/barista
package main

import (
	"barista.run/modules/clock"
	"barista.run/modules/cputemp"
	"barista.run/modules/diskio"
	"barista.run/modules/funcs"
	"barista.run/modules/netspeed"
	"barista.run/modules/sysinfo"
	"barista.run/modules/volume"
	//"barista.run/samples/mpd"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"barista.run"
	"barista.run/bar"
	"barista.run/colors"
	"barista.run/format"
	"barista.run/modules/battery"
	"barista.run/modules/diskspace"
	"barista.run/modules/meminfo"
	"barista.run/modules/netinfo"
	"barista.run/modules/wlan"
	"barista.run/outputs"
	"github.com/martinlindhe/unit"
)

func usbDeny() bool {
	data, err := ioutil.ReadFile("/proc/sys/kernel/deny_new_usb")
	if err != nil {
		return false
	}
	dataString := strings.Split(string(data), "\n")
	out, err := strconv.ParseBool(dataString[0])
	if err != nil {
		return false
	}
	return out
}

func main() {
	colors.LoadFromMap(map[string]string{
		"good":     "#9FCA56",
		"bad":      "#CD3F45",
		"degraded": "#E6CD69",
	})

	barista.Add(funcs.Every(time.Second, func(s bar.Sink) {
		out := outputs.Text("USB")
		if usbDeny() {
			out.Color(colors.Scheme("good"))
		} else {
			out.Color(colors.Scheme("bad"))
		}
		s.Output(out)
	}))

	//barista.Add(mpd.New("").Output(func( i mpd.Info) bar.Output {
	//	out := outputs.Textf("%s %s - %s", i.PlaybackStatusIcon, i.Artist, i.Title)
	//	return out
	//}))

	barista.Add(diskspace.New("/").Output(func(i diskspace.Info) bar.Output {
		out := outputs.Textf("D: " + format.IBytesize(i.Used()) + "/" + format.IBytesize(i.Total))
		switch {
		case i.AvailFrac() < 0.1:
			out.Color(colors.Scheme("bad"))
		case i.AvailFrac() < 0.3:
			out.Color(colors.Scheme("degraded"))
		default:
			out.Color(colors.Scheme("good"))
		}
		return out
	}))

	barista.Add(diskio.New("dm-0").Output(func(i diskio.IO) bar.Output {
		out := outputs.Textf("I: %s O: %s", format.IByterate(i.Input), format.IByterate(i.Output))
		return out
	}))

	barista.Add(cputemp.New().Output(func(t unit.Temperature) bar.Output {
		tDecimal := int64(t.Celsius())
		out := outputs.Textf("T: %dC", tDecimal)
		if tDecimal >= 86 {
			out.Color(colors.Scheme("bad"))
			return out
		} else if tDecimal > 65 {
			out.Color(colors.Scheme("degraded"))
			return out
		} else {
			out.Color(colors.Scheme("good"))
			return out
		}
	}))

	barista.Add(wlan.Any().Output(func(w wlan.Info) bar.Output {
		switch {
		case w.Connected():
			out := fmt.Sprintf("W: (%s)", w.SSID)
			if len(w.IPs) > 0 {
				out += fmt.Sprintf(" %s", w.IPs[0])
			}
			return outputs.Text(out).Color(colors.Scheme("good"))
		case w.Connecting():
			return outputs.Text("W: connecting...").Color(colors.Scheme("degraded"))
		case w.Enabled():
			return outputs.Text("W: down").Color(colors.Scheme("bad"))
		default:
			return nil
		}
	}))

	barista.Add(netspeed.New("wlp3s0").Output(func(s netspeed.Speeds) bar.Output {
		if s.Connected() {
			return outputs.Textf("Rx: %s Tx: %s", format.IByterate(s.Rx), format.IByterate(s.Tx))
		}
		return nil
	}))

	barista.Add(netinfo.Prefix("e").Output(func(s netinfo.State) bar.Output {
		switch {
		case s.Connected():
			ip := "<no ip>"
			if len(s.IPs) > 0 {
				ip = s.IPs[0].String()
			}
			return outputs.Textf("E: %s", ip).Color(colors.Scheme("good"))
		case s.Connecting():
			return outputs.Text("E: connecting...").Color(colors.Scheme("degraded"))
		case s.Enabled():
			return outputs.Text("E: down").Color(colors.Scheme("bad"))
		default:
			return nil
		}
	}))

	barista.Add(netspeed.New("enp0s25").Output(func(s netspeed.Speeds) bar.Output {
		if s.Connected() {
			return outputs.Textf("Rx: %s Tx: %s", format.IByterate(s.Rx), format.IByterate(s.Tx))
		}
		return nil
	}))

	barista.Add(battery.All().Output(func(b battery.Info) bar.Output {
		if b.Status == battery.Disconnected {
			return nil
		}
		if b.Status == battery.Full {
			out := outputs.Text("B: 100%")
			out.Color(colors.Scheme("good"))
			return out
		}
		out := outputs.Textf("B: %d%% %s",
			b.RemainingPct(),
			b.RemainingTime())
		if b.PluggedIn() {
			out.Color(colors.Scheme("good"))
			return out
		}
		if b.Discharging() {
			if b.RemainingPct() < 20 {
				out.Color(colors.Scheme("bad"))
				return out
			} else if b.RemainingPct() < 50 {
				out.Color(colors.Scheme("degraded"))
				return out
			} else {
				out.Color(colors.Scheme("good"))
				return out
			}
		}
		return nil
	}))

	barista.Add(volume.DefaultSink().Output(func(v volume.Volume) bar.Output {
		if v.Mute {
			out := outputs.Textf("V: %03d", v.Pct())
			out.Color(colors.Scheme("bad"))
			return out
		}
		out := outputs.Textf("V: %03d", v.Pct())
		out.Color(colors.Scheme("good"))
		return out
	}))

	barista.Add(meminfo.New().Output(func(i meminfo.Info) bar.Output {
		if i.Available() < unit.Gigabyte {
			return outputs.Textf(`M: %s`,
				format.IBytesize(i.Available())).
				Color(colors.Scheme("bad"))
		}
		out := outputs.Textf(`M: %s/%s`,
			format.IBytesize(i["MemTotal"]-i.Available()),
			format.IBytesize(i["MemTotal"]))
		switch {
		case i.AvailFrac() < 0.2:
			out.Color(colors.Scheme("bad"))
		case i.AvailFrac() < 0.33:
			out.Color(colors.Scheme("degraded"))
		default:
			out.Color(colors.Scheme("good"))
		}
		return out
	}))

	barista.Add(sysinfo.New().Output(func(i sysinfo.Info) bar.Output {
		out := outputs.Textf("C: %.2f", i.Loads[0])
		if i.Loads[0] > 4.0 {
			out.Color(colors.Scheme("bad"))
		} else if i.Loads[0] > 2.0 {
			out.Color(colors.Scheme("degraded"))
		} else {
			out.Color(colors.Scheme("good"))
		}
		return out
	}))

	barista.Add(clock.Local().OutputFormat("2006-01-02 15:04"))

	panic(barista.Run())
}

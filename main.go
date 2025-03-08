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
	"github.com/shibumi/barista/modules/clock"

	"github.com/shibumi/barista/modules/volume"
	"github.com/shibumi/barista/modules/volume/alsa"
	"log"
	"os/exec"

	"github.com/shibumi/barista"
	"github.com/shibumi/barista/bar"
	"github.com/shibumi/barista/colors"
	"github.com/shibumi/barista/modules/battery"
	"github.com/shibumi/barista/modules/netinfo"
	"github.com/shibumi/barista/modules/wlan"
	"github.com/shibumi/barista/outputs"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
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

	barista.Add(wlan.Any().Output(func(w wlan.Info) bar.Output {
		switch {
		case w.Connected():
			var out string
			if len(w.IPs) > 0 {
				out = fmt.Sprintf("W: %s", w.IPs[0])
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
			if b.RemainingPct() < 5 {
				out.Color(colors.Scheme("bad"))
				err := exec.Command("fyi", "-t", "2000", "battery", "very low", "-u", "critical").Run()
				if err != nil {
					log.Fatal("Couldn't use fyi command")
				}
				return out
			} else if b.RemainingPct() < 20 {
				out.Color(colors.Scheme("degraded"))
				return out
			} else {
				out.Color(colors.Scheme("good"))
				return out
			}
		}
		return nil
	}))

	barista.Add(volume.New(alsa.DefaultMixer()).Output(func(v volume.Volume) bar.Output {
		if v.Mute {
			out := outputs.Textf("V: %d", v.Vol)
			out.Color(colors.Scheme("bad"))
			return out
		}
		out := outputs.Textf("V: %d", v.Vol)
		out.Color(colors.Scheme("good"))
		return out
	}))

	barista.Add(clock.Local().OutputFormat("2006-01-02 15:04"))

	panic(barista.Run())
}

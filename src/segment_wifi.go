package main

import (
	"regexp"
	"strconv"
	"strings"
)

type wifi struct {
	props          *properties
	env            environmentInfo
	State          string
	SSID           string
	RadioType      string
	Authentication string
	Channel        int
	ReceiveRate    int
	TransmitRate   int
	Signal         int
}

const (
	netshState          = "State"
	netshSSID           = "SSID"
	netshRadioType      = "Radio type"
	netshAuthentication = "Authentication"
	netshChannel        = "Channel"
	netshReceiveRate    = "Receive rate (Mbps)"
	netshTransmitRate   = "Transmit rate (Mbps)"
	netshSignal         = "Signal"
)

func (w *wifi) enabled() bool {
	// If in Linux, who wnows. Gonna need a physical linux machine to develop this since the wlan interface isn't available in wsl
	// But possible solution is to cat /proc/net/wireless, and if that file doesn't exist then don't show the segment
	// In Windows and in wsl you can use the command `netsh.exe wlan show interfaces`. Note the .exe is important for this to worw in wsl.
	if w.env.getPlatform() == windowsPlatform || w.env.isWsl() {
		cmd := "netsh"
		if !w.env.hasCommand(cmd) {
			return false
		}
		cmdResult, err := w.env.runCommand(cmd, "wlan", "show", "interfaces")
		displayError := w.props.getBool(DisplayError, false)
		if err != nil && displayError {
			w.State = "WIFI ERR"
			return true
		}
		if err != nil {
			return false
		}

		regex := regexp.MustCompile(`(.+) : (.+)`)
		lines := strings.Split(cmdResult, "\n")
		for _, line := range lines {
			matches := regex.FindStringSubmatch(line)
			if len(matches) != 3 {
				continue
			}
			name := strings.TrimSpace(matches[1])
			value := strings.TrimSpace(matches[2])

			switch name {
			case netshState:
				w.State = value
			case netshSSID:
				w.SSID = value
			case netshRadioType:
				w.RadioType = value
			case netshAuthentication:
				w.Authentication = value
			case netshChannel:
				if v, err := strconv.Atoi(value); err == nil {
					w.Channel = v
				}
			case netshReceiveRate:
				if v, err := strconv.Atoi(strings.Split(value, ".")[0]); err == nil {
					w.ReceiveRate = v
				}
			case netshTransmitRate:
				if v, err := strconv.Atoi(strings.Split(value, ".")[0]); err == nil {
					w.TransmitRate = v
				}
			case netshSignal:
				if v, err := strconv.Atoi(strings.TrimRight(value, "%")); err == nil {
					w.Signal = v
				}
			}
		}
	}

	return true
}

func (w *wifi) string() string {
	const defaultTemplate = "{{.ssid}} {{.signal}}% {{.receiveRate}}Mbps"
	segmentTemplate := w.props.getString(SegmentTemplate, defaultTemplate)
	template := &textTemplate{
		Template: segmentTemplate,
		Context:  w,
		Env:      w.env,
	}
	text, err := template.render()
	if err != nil {
		return err.Error()
	}
	return text
}

func (w *wifi) init(props *properties, env environmentInfo) {
	w.props = props
	w.env = env
}

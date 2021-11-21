package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type wifi struct {
	props          *properties
	env            environmentInfo
	Connected      bool
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
	ConnectedIcon    Property = "connected_icon"
	DisconnectedIcon Property = "disconnected_icon"
)

const defaultTemplate string = "{{if .Connected}}{{.SSID}} {{.Signal}}% {{.ReceiveRate}}Mbps{{else}}{{.State}}{{end}}"

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
			case "State":
				w.State = value
				w.Connected = value == "connected"
			case "SSID":
				w.SSID = value
			case "Radio type":
				w.RadioType = value
			case "Authentication":
				w.Authentication = value
			case "Channel":
				if v, err := strconv.Atoi(value); err == nil {
					w.Channel = v
				}
			case "Receive rate (Mbps)":
				if v, err := strconv.Atoi(strings.Split(value, ".")[0]); err == nil {
					w.ReceiveRate = v
				}
			case "Transmit rate (Mbps)":
				if v, err := strconv.Atoi(strings.Split(value, ".")[0]); err == nil {
					w.TransmitRate = v
				}
			case "Signal":
				if v, err := strconv.Atoi(strings.TrimRight(value, "%")); err == nil {
					w.Signal = v
				}
			}
		}
	}

	return true
}

func (w *wifi) string() string {
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

	var icon string
	if w.State == "connected" {
		icon = w.props.getString(ConnectedIcon, "\uFAA8")
	} else {
		icon = w.props.getString(DisconnectedIcon, "\uFAA9")
	}

	return fmt.Sprintf("%s%s", icon, text)
}

func (w *wifi) init(props *properties, env environmentInfo) {
	w.props = props
	w.env = env
}

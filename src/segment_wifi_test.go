package main

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type wifiArgs struct {
	shell    string
	platform string
	isWsl    bool

	command       string
	commandOutput string
	commandError  error
	hasCommand    bool

	displayError    bool
	segmentTemplate string
}

type netshStringArgs struct {
	state          string
	ssid           string
	radioType      string
	authentication string
	channel        int
	receiveRate    int
	transmitRate   int
	signal         int
}

func getNetshString(args *netshStringArgs) string {
	const netshString string = `
	There is 1 interface on the system:

	Name                   : Wi-Fi
	Description            : Intel(R) Wireless-AC 9560 160MHz
	GUID                   : 6bb8def2-9af2-4bd4-8be2-6bd54e46bdc9
	Physical address       : d4:3b:04:e6:10:40
	State                  : %s
	SSID                   : %s
	BSSID                  : 5c:7d:7d:82:c5:73
	Network type           : Infrastructure
	Radio type             : %s
	Authentication         : %s
	Cipher                 : CCMP
	Connection mode        : Profile
	Channel                : %d
	Receive rate (Mbps)    : %d
	Transmit rate (Mbps)   : %d
	Signal                 : %d%%
	Profile                : ohsiggy

	Hosted network status  : Not available`

	return fmt.Sprintf(netshString, args.state, args.ssid, args.radioType, args.authentication, args.channel, args.receiveRate, args.transmitRate, args.signal)
}

func bootStrapWifiWindowsPwshTest(args *wifiArgs) *wifi {
	args.platform = windowsPlatform
	args.shell = pwsh
	args.command = "netsh"
	args.isWsl = false

	env, props := bootStrapEnvironment(args)

	k := &wifi{
		env:   env,
		props: props,
	}

	return k
}

func bootStrapEnvironment(args *wifiArgs) (*MockedEnvironment, *properties) {
	env := new(MockedEnvironment)
	env.On("getPlatform", nil).Return(args.platform)
	env.On("isWsl", nil).Return(args.isWsl)
	env.On("hasCommand", args.command).Return(args.hasCommand)
	env.On("runCommand", mock.Anything, mock.Anything).Return(args.commandOutput, args.commandError)
	env.On("getShellName", nil).Return(args.shell)

	props := &properties{
		values: map[Property]interface{}{
			DisplayError:    args.displayError,
			SegmentTemplate: args.segmentTemplate,
		},
	}

	return env, props
}

func TestWifi_Enabled_ForWindowsPwsh_WhenCommandNotFound_IsNotEnabled(t *testing.T) {
	args := &wifiArgs{
		hasCommand: false,
	}
	wifi := bootStrapWifiWindowsPwshTest(args)
	assert.False(t, wifi.enabled())
}

func TestWifi_Enabled_ForWindowsPwsh_WhenRunCommandFails_IsNotEnabled(t *testing.T) {
	args := &wifiArgs{
		commandError: errors.New("Oh noes!"),
		hasCommand:   true,
	}
	wifi := bootStrapWifiWindowsPwshTest(args)
	assert.False(t, wifi.enabled())
}

func TestWifi_Enabled_ForWindowsPwsh_WhenRunCommandFailsWithDisplayError_IsEnabledWithErrorState(t *testing.T) {
	args := &wifiArgs{
		hasCommand:   true,
		commandError: errors.New("Oh noes!"),
		displayError: true,
	}
	wifi := bootStrapWifiWindowsPwshTest(args)
	assert.True(t, wifi.enabled())
	assert.Equal(t, "WIFI ERR", wifi.State)
}

func TestWifi_Enabled_ForWindowsPwsh_HappyPath_IsEnabled(t *testing.T) {
	args := &wifiArgs{
		hasCommand:    true,
		commandOutput: "",
	}
	wifi := bootStrapWifiWindowsPwshTest(args)
	assert.True(t, wifi.enabled())
}

func TestWifi_Enabled_ForWindowsPwsh_HappyPath(t *testing.T) {
	expected := &netshStringArgs{
		state:          "connected",
		ssid:           "ohsiggy",
		radioType:      "802.11ac",
		authentication: "WPA2-Personal",
		channel:        99,
		receiveRate:    500.0,
		transmitRate:   400.0,
		signal:         80,
	}

	args := &wifiArgs{
		hasCommand:    true,
		commandOutput: getNetshString(expected),
	}
	wifi := bootStrapWifiWindowsPwshTest(args)
	enabled := wifi.enabled()
	assert.True(t, enabled)
	assert.Equal(t, expected.state, wifi.State)
	assert.Equal(t, expected.ssid, wifi.SSID)
	assert.Equal(t, expected.radioType, wifi.RadioType)
	assert.Equal(t, expected.authentication, wifi.Authentication)
	assert.Equal(t, expected.channel, wifi.Channel)
	assert.Equal(t, expected.receiveRate, wifi.ReceiveRate)
	assert.Equal(t, expected.transmitRate, wifi.TransmitRate)
	assert.Equal(t, expected.signal, wifi.Signal)
}

func TestWifi_String_ForWindowsPwsh_HappyPath(t *testing.T) {
	expected := &netshStringArgs{
		state:          "connected",
		ssid:           "ohsiggy",
		radioType:      "802.11ac",
		authentication: "WPA2-Personal",
		channel:        99,
		receiveRate:    500.0,
		transmitRate:   400.0,
		signal:         80,
	}

	args := &wifiArgs{
		hasCommand:      true,
		commandOutput:   getNetshString(expected),
		segmentTemplate: "{{.State}}{{.SSID}}{{.RadioType}}{{.Authentication}}{{.Channel}}{{.ReceiveRate}}{{.TransmitRate}}{{.Signal}}",
	}
	wifi := bootStrapWifiWindowsPwshTest(args)
	assert.True(t, wifi.enabled())

	expectedString := fmt.Sprintf("%s%s%s%s%d%d%d%d",
		wifi.State, wifi.SSID, wifi.RadioType, wifi.Authentication, wifi.Channel, wifi.ReceiveRate, wifi.TransmitRate, wifi.Signal)
	assert.Equal(t, expectedString, wifi.string())
}

func TestWifi_String_ForWindowsPwsh_TemplateRenderError_ReturnsError(t *testing.T) {
	expected := &netshStringArgs{
		state:          "connected",
		ssid:           "ohsiggy",
		radioType:      "802.11ac",
		authentication: "WPA2-Personal",
		channel:        99,
		receiveRate:    500.0,
		transmitRate:   400.0,
		signal:         80,
	}

	args := &wifiArgs{
		hasCommand:      true,
		commandOutput:   getNetshString(expected),
		segmentTemplate: "{{.DoesNotExist}}",
	}
	wifi := bootStrapWifiWindowsPwshTest(args)
	assert.True(t, wifi.enabled())

	assert.Equal(t, "unable to create text based on template", wifi.string())
}

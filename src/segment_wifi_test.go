package main

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type wifiArgs struct {
	commandOutput   string
	commandError    error
	hasCommand      bool
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

func bootStrapEnvironment(args *wifiArgs) *wifi {
	env := new(MockedEnvironment)
	env.On("getPlatform", nil).Return(windowsPlatform)
	env.On("isWsl", nil).Return(false)
	env.On("hasCommand", "netsh").Return(args.hasCommand)
	env.On("runCommand", mock.Anything, mock.Anything).Return(args.commandOutput, args.commandError)

	props := &properties{
		values: map[Property]interface{}{
			DisplayError:     args.displayError,
			SegmentTemplate:  args.segmentTemplate,
			ConnectedIcon:    "",
			DisconnectedIcon: "",
		},
	}

	k := &wifi{
		env:   env,
		props: props,
	}

	return k
}

func TestWifi_Enabled_ForWindows_WhenCommandNotFound_IsNotEnabled(t *testing.T) {
	args := &wifiArgs{
		hasCommand: false,
	}
	wifi := bootStrapEnvironment(args)
	assert.False(t, wifi.enabled())
}

func TestWifi_Enabled_ForWindows_WhenRunCommandFails_IsNotEnabled(t *testing.T) {
	args := &wifiArgs{
		commandError: errors.New("intentional testing failure"),
		hasCommand:   true,
	}
	wifi := bootStrapEnvironment(args)
	assert.False(t, wifi.enabled())
}

func TestWifi_Enabled_ForWindows_WhenRunCommandFailsWithDisplayError_IsEnabledWithErrorState(t *testing.T) {
	args := &wifiArgs{
		hasCommand:   true,
		commandError: errors.New("intentional testing failure"),
		displayError: true,
	}
	wifi := bootStrapEnvironment(args)
	assert.True(t, wifi.enabled())
	assert.Equal(t, "WIFI ERR", wifi.State)
}

func TestWifi_Enabled_ForWindows_HappyPath_IsEnabled(t *testing.T) {
	args := &wifiArgs{
		hasCommand:    true,
		commandOutput: "",
	}
	wifi := bootStrapEnvironment(args)
	assert.True(t, wifi.enabled())
}

func TestWifi_Enabled_ForWindows_HappyPath(t *testing.T) {
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
	wifi := bootStrapEnvironment(args)
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

func TestWifi_String_ForWindows_HappyPath(t *testing.T) {
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
	wifi := bootStrapEnvironment(args)
	assert.True(t, wifi.enabled())

	expectedString := fmt.Sprintf("%s%s%s%s%d%d%d%d",
		wifi.State, wifi.SSID, wifi.RadioType, wifi.Authentication, wifi.Channel, wifi.ReceiveRate, wifi.TransmitRate, wifi.Signal)
	assert.Equal(t, expectedString, wifi.string())
}

func TestWifi_String_ForWindows_TemplateRenderError_ReturnsError(t *testing.T) {
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
	wifi := bootStrapEnvironment(args)
	assert.True(t, wifi.enabled())

	assert.Equal(t, "unable to create text based on template", wifi.string())
}

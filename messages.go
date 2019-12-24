// **********************************************************************
//    Copyright (c) 2018-2019 Henry Seurer
//
//   Permission is hereby granted, free of charge, to any person
//    obtaining a copy of this software and associated documentation
//    files (the "Software"), to deal in the Software without
//    restriction, including without limitation the rights to use,
//    copy, modify, merge, publish, distribute, sublicense, and/or sell
//    copies of the Software, and to permit persons to whom the
//    Software is furnished to do so, subject to the following
//    conditions:
//
//   The above copyright notice and this permission notice shall be
//   included in all copies or substantial portions of the Software.
//
//    THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
//    EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
//    OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
//    NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
//    HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//    WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
//    FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
//    OTHER DEALINGS IN THE SOFTWARE.
//
// **********************************************************************

package messages

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

type BuildVersion struct {
	Version   string                 `json:"version"`
	BuildTime string                 `json:"build_time"`
	Image     string                 `json:"image"`
	ImageID   string                 `json:"image_id"`
	Versions  map[string]interface{} `json:"versions"`
}

type MessageHeader struct {
	Status    int          `json:"status"`
	Location  string       `json:"location"`
	TimeStamp time.Time    `json:"timestamp"`
	Host      string       `json:"host"`
	Build     BuildVersion `json:"build"`
}

func GetHostName() string {
	host := os.Getenv("DOCKER_HOST_IP")
	if len(host) == 0 {
		host, _ = os.Hostname()
	}

	return host
}

type AlertMessage struct {
	Header  MessageHeader `json:"header"`
	Message string        `json:"message"`
	Source  string        `json:"source"`
}

func (a *AlertMessage) GetSource() string {
	if len(a.Source) == 0 {
		return "unknown"
	}

	return a.Source
}

type ErrorMessage struct {
	Header  MessageHeader `json:"header"`
	Message string        `json:"message"`
}

type MotionMessage struct {
	Header MessageHeader `json:"header"`
}

type TemperatureMessage struct {
	Header     MessageHeader `json:"header"`
	Celsius    float32       `json:"celsius"`
	Humidity   float32       `json:"humidity"`
	Fahrenheit float32       `json:"fahrenheit"`
}

type SuccessMessage struct {
	Header  MessageHeader `json:"header"`
	Message string        `json:"message"`
}

type AlarmMessage struct {
	Header MessageHeader `json:"header"`
}

type AlarmSensor struct {
	Status string `json:"status"`
	Number string `json:"number"`
	Name   string `json:"name"`
	Stamp  int64  `json:"stamp"`
	Id     string `json:"id"`
}

type AlarmSensors map[string]AlarmSensor

type AlarmSensorsMessage struct {
	Header       MessageHeader `json:"header"`
	AlarmSensors AlarmSensors  `json:"sensors"`
	Armed        bool          `json:"armed"`
}

type AlarmSensorMessage struct {
	Header      MessageHeader `json:"header"`
	AlarmSensor AlarmSensor   `json:"sensor"`
	Armed       bool          `json:"armed"`
}

// ThrottleEntry is a time.Duration counter, starting at Min. After every call to
// the Duration method the current timing is multiplied by Factor, but it
// never exceeds Max.
//
type ThrottleEntry struct {
	Count  uint64        `json:"count"`
	Factor float64       `json:"factor"`
	Jitter bool          `json:"jitter"`
	Min    time.Duration `json:"min"`
	Max    time.Duration `json:"max"`
	Stamp  int64         `json:"stamp"`
}

// Duration returns the duration for the current attempt before incrementing
// the attempt counter. See ForAttempt.
func (t *ThrottleEntry) Duration(minTime time.Duration, maxTime time.Duration) time.Duration {
	d := t.ForAttempt(float64(atomic.AddUint64(&t.Count, 1)-1)-1, minTime, maxTime)
	return d
}

const maxInt64 = float64(math.MaxInt64 - 512)

// ForAttempt returns the duration for a specific attempt. This is useful if
// you have a large number of independent ThrottleEntry, but don't want use
// unnecessary memory storing the back off parameters per back off. The first
// attempt should be 0.
//
func (t *ThrottleEntry) ForAttempt(attempt float64, minTime time.Duration, maxTime time.Duration) time.Duration {
	// Zero-values are nonsensical, so we use
	// them to apply defaults
	min := t.Min
	if min <= 0 {
		min = minTime
	}
	max := t.Max
	if max <= 0 {
		max = maxTime
	}
	if min >= max {
		// short-circuit
		return max
	}
	factor := t.Factor
	if factor <= 0 {
		factor = 2
	}
	//calculate this duration
	minf := float64(min)
	durf := minf * math.Pow(factor, attempt)
	if t.Jitter {
		durf = rand.Float64()*(durf-minf) + minf
	}
	//ensure float64 wont overflow int64
	if durf > maxInt64 {
		return max
	}
	dur := time.Duration(durf)
	//keep within bounds
	if dur < min {
		return min
	}
	if dur > max {
		return max
	}
	return dur
}

func (t *ThrottleEntry) Reset(minTime time.Duration, maxTime time.Duration) *ThrottleEntry {
	t.Count = 0
	t.Factor = 2
	t.Jitter = false
	t.Min = minTime
	t.Max = maxTime
	t.Stamp = time.Now().Unix()

	return t
}

// Attempt returns the current attempt counter value.
func (t *ThrottleEntry) Attempt() uint64 {
	return t.Count
}

// Copy returns a ThrottleEntry with equals constraints as the original
func (t *ThrottleEntry) Copy() *ThrottleEntry {
	return &ThrottleEntry{
		Count:  t.Count,
		Factor: t.Factor,
		Jitter: t.Jitter,
		Min:    t.Min,
		Max:    t.Max,
		Stamp:  t.Stamp,
	}
}

type ThrottleEntries map[string]ThrottleEntry

type ThrottleEntriesMessage struct {
	Header          MessageHeader   `json:"header"`
	ThrottleEntries ThrottleEntries `json:"entries"`
	Armed           bool            `json:"armed"`
}

type SystemStatus string

const (
	ONLINE    SystemStatus = "ONLINE"
	OFFLINE   SystemStatus = "OFFLINE"
	UNDEFINED SystemStatus = "UNDEFINED"
)

//noinspection GoUnusedExportedFunction
func ParseSystemState(state string) SystemStatus {
	state = strings.ToUpper(state)
	switch state {
	case "ONLINE":
		return ONLINE
	case "OFFLINE":
		return OFFLINE
	}
	return UNDEFINED
}

// SystemStatusMap[system]
type SystemStatusMap map[string]SystemStatus
type SystemStatusMessage struct {
	Header       MessageHeader   `json:"header"`
	SystemStatus SystemStatusMap `json:"message"`
}

type DeviceState string

const (
	ON      DeviceState = "ON"
	OFF     DeviceState = "OFF"
	UNKNOWN DeviceState = "UNKNOWN"
)

//noinspection GoUnusedExportedFunction
func ParseDeviceState(state string) DeviceState {
	state = strings.ToUpper(state)
	switch state {
	case "ON":
		return ON
	case "OFF":
		return OFF
	}
	return UNKNOWN
}

type LocationName string
type DeviceInfo struct {
	Device string      `json:"device"`
	State  DeviceState `json:"state"`
}

// DeviceInfoMap[location]
type DeviceInfoMap map[string]DeviceInfo
type DevicesInfoMessage struct {
	Header  MessageHeader `json:"header"`
	Devices DeviceInfoMap `json:"devices"`
}

// UPS
type UPSBattery struct {
	Charge         int     `json:"charge"`
	ChargeLow      int     `json:"charge_low"`
	ChargeWarning  int     `json:"charge_warning"`
	Runtime        int     `json:"runtime"`
	RuntimeLow     int     `json:"runtime_low"`
	Type           string  `json:"type"`
	Voltage        float32 `json:"voltage"`
	VoltageNominal float32 `json:"voltage_nominal"`
}

type UPSDriver struct {
	Name            string `json:"name"`
	PollFreq        int    `json:"poll_freq"`
	PollInterval    int    `json:"poll_interval"`
	Port            string `json:"port"`
	Synchronous     bool   `json:"synchronous"`
	Version         string `json:"version"`
	VersionData     string `json:"version_data"`
	VersionInternal string `json:"version_internal"`
}

type UPSStatus struct {
	Battery             UPSBattery `json:"battery"`
	DeviceMfr           string     `json:"device_mfr"`
	DeviceModel         string     `json:"device_model"`
	DeviceType          string     `json:"device_type"`
	InputTransferHigh   int        `json:"input_transfer_high"`
	InputTransferLow    int        `json:"input_transfer_low"`
	InputVoltage        float32    `json:"input_voltage"`
	InputVoltageNominal float32    `json:"input_voltage_nominal"`
	OutputVoltage       float32    `json:"output_voltage"`
	BeeperStatus        bool       `json:"beeper_status"`
	DelayShutdown       int        `json:"delay_shutdown"`
	DelayStart          int        `json:"delay_start"`
	Load                int        `json:"load"`
	Mfr                 string     `json:"mfr"`
	Model               string     `json:"model"`
	ProductId           string     `json:"product_id"`
	RealPowerNominal    int        `json:"real_power_nominal"`
	Status              string     `json:"status"`
	TestResult          string     `json:"test_result"`
	TimerShutdown       int        `json:"timer_shutdown"`
	TimerStart          int        `json:"timer_start"`
	VendorId            string     `json:"vendor_id"`
}

type UPSStatusMessage struct {
	Header MessageHeader `json:"header"`
	Status UPSStatus     `json:"status"`
}

//noinspection GoUnusedExportedFunction
func CreateHeader(status int, location string) MessageHeader {

	// Do we have a build version?
	//
	var build BuildVersion
	buildBytes, err := ioutil.ReadFile("/opt/build_version.json")
	if err == nil {
		err = json.Unmarshal(buildBytes, &build)
	}

	// Build Message Header
	//
	return MessageHeader{
		Status:    status,
		Location:  location,
		TimeStamp: time.Now(),
		Host:      GetHostName(),
		Build:     build,
	}
}

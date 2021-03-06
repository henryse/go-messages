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
	"log"
	"math"
	"math/rand"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

type BuildVersion struct {
	Version   string                 `json:"version,omitempty"`
	BuildTime string                 `json:"build_time,omitempty"`
	Image     string                 `json:"image,omitempty"`
	ImageID   string                 `json:"image_id,omitempty"`
	Versions  map[string]interface{} `json:"versions,omitempty"`
}

type MessageHeader struct {
	Status    int          `json:"status,omitempty"`
	Location  string       `json:"location,omitempty"`
	TimeStamp time.Time    `json:"timestamp,omitempty"`
	Host      string       `json:"host,omitempty"`
	Build     BuildVersion `json:"build,omitempty"`
}

func GetHostName() string {
	host := os.Getenv("DOCKER_HOST_IP")
	if len(host) == 0 {
		host, _ = os.Hostname()
	}

	return host
}

type AlertSeverity int

//goland:noinspection GoUnusedConst
const (
	Info    AlertSeverity = 0 // Nothing to worry about FYI
	Warning AlertSeverity = 1 // Should be looked into soon
	Error   AlertSeverity = 2 // Houston we have a problem
	Cleared AlertSeverity = 3 // The issue has been addressed
)

type AlertMessage struct {
	Header   MessageHeader `json:"header,omitempty"`
	Severity AlertSeverity `json:"severity,omitempty"`
	Source   string        `json:"source,omitempty"`
	Message  string        `json:"message,omitempty"`
}

func (a *AlertMessage) GetSource() string {
	if len(a.Source) == 0 {
		return "unknown"
	}

	return a.Source
}

type ErrorMessage struct {
	Header  MessageHeader `json:"header,omitempty"`
	Source  string        `json:"source,omitempty"`
	Message string        `json:"message,omitempty"`
}

type Text struct {
	ID        string        `json:"id"`
	Source    string        `json:"source,omitempty"`
	Location  string        `json:"location,omitempty"`
	TimeStamp time.Time     `json:"timestamp"`
	Host      string        `json:"host,omitempty"`
	Severity  AlertSeverity `json:"severity"`
	Count     int           `json:"count,omitempty"`
	Body      string        `json:"body"`
}

type Texts []Text

type TextMessage struct {
	Header MessageHeader `json:"header,omitempty"`
	Texts  Texts         `json:"texts,omitempty"`
}

type MotionMessage struct {
	Header MessageHeader `json:"header,omitempty"`
}

type TemperatureMessage struct {
	Header     MessageHeader `json:"header,omitempty"`
	Celsius    float32       `json:"celsius,omitempty"`
	Humidity   float32       `json:"humidity,omitempty"`
	Fahrenheit float32       `json:"fahrenheit,omitempty"`
	Time       time.Time     `json:"time,omitempty"`
}

type SuccessMessage struct {
	Header  MessageHeader `json:"header,omitempty"`
	Message string        `json:"message,omitempty"`
}

type Regions []string

type WeatherAlertMessage struct {
	Header   MessageHeader `json:"header,omitempty"`
	Message  string        `json:"message,omitempty"`
	Expires  time.Time     `json:"expires,omitempty"`
	Severity string        `json:"severity,omitempty"`
	Regions  Regions       `json:"Regions,omitempty"`
	Id       int64         `json:"id,omitempty"`
}

type AlarmMessage struct {
	Header MessageHeader `json:"header,omitempty"`
}

type AlarmSensor struct {
	Status string `json:"status,omitempty"`
	Number string `json:"number,omitempty"`
	Name   string `json:"name,omitempty"`
	Stamp  int64  `json:"stamp,omitempty"`
	Id     string `json:"id,omitempty"`
}

type AlarmSensors map[string]AlarmSensor

type AlarmSensorsMessage struct {
	Header       MessageHeader `json:"header,omitempty"`
	AlarmSensors AlarmSensors  `json:"sensors,omitempty"`
	Armed        bool          `json:"armed,omitempty"`
}

type AlarmSensorMessage struct {
	Header      MessageHeader `json:"header,omitempty"`
	AlarmSensor AlarmSensor   `json:"sensor,omitempty"`
	Armed       bool          `json:"armed,omitempty"`
}

// ThrottleEntry is a time.Duration counter, starting at Min. After every call to
// the Duration method the current timing is multiplied by Factor, but it
// never exceeds Max.
//
type ThrottleEntry struct {
	Count  uint64        `json:"count,omitempty"`
	Factor float64       `json:"factor,omitempty"`
	Jitter bool          `json:"jitter,omitempty"`
	Min    time.Duration `json:"min,omitempty"`
	Max    time.Duration `json:"max,omitempty"`
	Stamp  int64         `json:"stamp,omitempty"`
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
	minFloat := float64(min)
	durationFloat := minFloat * math.Pow(factor, attempt)
	if t.Jitter {
		durationFloat = rand.Float64()*(durationFloat-minFloat) + minFloat
	}
	//ensure float64 wont overflow int64
	if durationFloat > maxInt64 {
		return max
	}
	dur := time.Duration(durationFloat)
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

type ThrottleEntries []ThrottleEntry

type ThrottleEntriesMessage struct {
	Header          MessageHeader   `json:"header,omitempty"`
	ThrottleEntries ThrottleEntries `json:"entries,omitempty"`
	Armed           bool            `json:"armed,omitempty"`
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

type SystemStatusMap map[string]SystemStatus
type SystemStatusMessage struct {
	Header       MessageHeader   `json:"header,omitempty"`
	SystemStatus SystemStatusMap `json:"message,omitempty"`
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
	Device string      `json:"device,omitempty"`
	State  DeviceState `json:"state,omitempty"`
}

type DeviceInfoMap map[string]DeviceInfo
type DevicesInfoMessage struct {
	Header  MessageHeader `json:"header,omitempty"`
	Devices DeviceInfoMap `json:"devices,omitempty"`
}

type UPSBattery struct {
	Charge         int     `json:"charge,omitempty"`
	ChargeLow      int     `json:"charge_low,omitempty"`
	ChargeWarning  int     `json:"charge_warning,omitempty"`
	Runtime        int     `json:"runtime,omitempty"`
	RuntimeLow     int     `json:"runtime_low,omitempty"`
	Type           string  `json:"type,omitempty"`
	Voltage        float32 `json:"voltage,omitempty"`
	VoltageNominal float32 `json:"voltage_nominal,omitempty"`
}

type UPSDriver struct {
	Name            string `json:"name,omitempty"`
	PollFreq        int    `json:"poll_freq,omitempty"`
	PollInterval    int    `json:"poll_interval,omitempty"`
	Port            string `json:"port,omitempty"`
	Synchronous     bool   `json:"synchronous,omitempty"`
	Version         string `json:"version,omitempty"`
	VersionData     string `json:"version_data,omitempty"`
	VersionInternal string `json:"version_internal,omitempty"`
}

type UPSStatus struct {
	Name                string     `json:"name,omitempty"`
	Battery             UPSBattery `json:"battery,omitempty"`
	DeviceMfr           string     `json:"device_mfr,omitempty"`
	DeviceModel         string     `json:"device_model,omitempty"`
	DeviceType          string     `json:"device_type,omitempty"`
	InputTransferHigh   int        `json:"input_transfer_high,omitempty"`
	InputTransferLow    int        `json:"input_transfer_low,omitempty"`
	InputVoltage        float32    `json:"input_voltage,omitempty"`
	InputVoltageNominal float32    `json:"input_voltage_nominal,omitempty"`
	OutputVoltage       float32    `json:"output_voltage,omitempty"`
	BeeperStatus        bool       `json:"beeper_status,omitempty"`
	DelayShutdown       int        `json:"delay_shutdown,omitempty"`
	DelayStart          int        `json:"delay_start,omitempty"`
	Load                int        `json:"load,omitempty"`
	Mfr                 string     `json:"mfr,omitempty"`
	Model               string     `json:"model,omitempty"`
	ProductId           string     `json:"product_id,omitempty"`
	RealPowerNominal    int        `json:"real_power_nominal,omitempty"`
	Status              string     `json:"status,omitempty"`
	TestResult          string     `json:"test_result,omitempty"`
	TimerShutdown       int        `json:"timer_shutdown,omitempty"`
	TimerStart          int        `json:"timer_start,omitempty"`
	VendorId            string     `json:"vendor_id,omitempty"`
}

type UPSStatusMessage struct {
	Header  MessageHeader `json:"header,omitempty"`
	Sources []UPSStatus   `json:"sources,omitempty"`
}

type KasaListMessage struct {
	Header  MessageHeader `json:"header,omitempty"`
	Devices interface{}   `json:"devices,omitempty"`
}

type KasaSetMessage struct {
	Header MessageHeader `json:"header,omitempty"`
	Alias  string        `json:"alias,omitempty"`
	State  bool          `json:"state,omitempty"`
}

type Service struct {
	ID       string                 `json:"id,omitempty"`
	TTL      time.Duration          `json:"ttl,omitempty"`
	TTLStamp time.Time              `json:"ttl_stamp,omitempty"`
	Name     string                 `json:"name,omitempty"`
	Attrs    map[string]string      `json:"attrs,omitempty"`
	Status   string                 `json:"status,omitempty"`
	Hostname string                 `json:"hostname,omitempty"`
	HostIP   string                 `json:"host_ip,omitempty"`
	Ports    map[string]ServicePort `json:"origin,omitempty"`
}

type ServicePort struct {
	HostPort    string `json:"host_port,omitempty"`
	HostIP      string `json:"host_ip,omitempty"`
	ExposedPort string `json:"exposed_port,omitempty"`
	ExposedIP   string `json:"exposed_ip,omitempty"`
	PortType    string `json:"port_type,omitempty"`
}

//noinspection GoUnusedConst
const (
	ServiceEventStart    = "start"
	ServiceEventStop     = "stop"
	ServiceEventActive   = "active"
	ServiceEventVanished = "vanish"
)

type ServiceMessage struct {
	Header       MessageHeader `json:"header,omitempty"`
	ServiceEvent string        `json:"event,omitempty"`
	Service      Service       `json:"service,omitempty"`
}

type Observation struct {
	StationID         string    `json:"stationID,omitempty"`
	Name              string    `json:"name,omitempty"`
	ObsTimeUtc        time.Time `json:"obsTimeUtc,omitempty"`
	ObsTimeLocal      string    `json:"obsTimeLocal,omitempty"`
	Neighborhood      string    `json:"neighborhood,omitempty"`
	SoftwareType      string    `json:"softwareType,omitempty"`
	Country           string    `json:"country,omitempty"`
	SolarRadiation    string    `json:"solarRadiation,omitempty"`
	Lon               float64   `json:"longitude,omitempty"`
	RealtimeFrequency string    `json:"realtimeFrequency,omitempty"`
	Epoch             int       `json:"epoch,omitempty"`
	Lat               float64   `json:"latitude,omitempty"`
	Uv                float64   `json:"uv,omitempty"`
	Winddir           int       `json:"winddir,omitempty"`
	Humidity          int       `json:"humidity,omitempty"`
	QcStatus          int       `json:"qcStatus,omitempty"`
	Imperial          struct {
		Temp        int     `json:"temp,omitempty"`
		HeatIndex   int     `json:"heatIndex,omitempty"`
		Dewpt       int     `json:"dewpt,omitempty"`
		WindChill   int     `json:"windChill,omitempty"`
		WindSpeed   int     `json:"windSpeed,omitempty"`
		WindGust    int     `json:"windGust,omitempty"`
		Pressure    float64 `json:"pressure,omitempty"`
		PrecipRate  float64 `json:"precipRate,omitempty"`
		PrecipTotal float64 `json:"precipTotal,omitempty"`
		Elev        int     `json:"elev,omitempty"`
	} `json:"imperial,omitempty"`
}

type Observations []Observation

type WeatherMessage struct {
	Header       MessageHeader `json:"header,omitempty"`
	Observations Observations  `json:"observations,omitempty"`
}

type DaylightDate struct {
	Hours   int       `json:"hours,omitempty"`
	Sunrise time.Time `json:"sunrise,omitempty"`
	Sunset  time.Time `json:"sunset,omitempty"`
}

type DaylightDates []DaylightDate

type DaylightMessage struct {
	Header       MessageHeader `json:"header,omitempty"`
	DayLightDays DaylightDates `json:"days,omitempty"`
}

//goland:noinspection GoUnusedConst
const (
	stateActive    = "active"
	stateHarvested = "harvested"
	stateInactive  = "destroyed"

	phaseVegetative = "vegetative"
	phaseFlowering  = "flowering"

	operationInsert = "insert"
	operationUpsert = "upsert"
)

type Plant struct {
	ID        string `json:"id,omitempty"`
	Tag       string `json:"tag,omitempty"`
	Strain    string `json:"strain,omitempty"`
	Location  string `json:"location,omitempty"`
	Phase     string `json:"phase,omitempty"`
	State     string `json:"state,omitempty"`
	Group     string `json:"group,omitempty"`
	GroupType string `json:"groupType,omitempty"`
}

type Plants []Plant

type PlantMessage struct {
	Header MessageHeader `json:"header,omitempty"`
	Plant  Plant         `json:"plant,omitempty"`
}

type PlantsMessage struct {
	Header MessageHeader `json:"header,omitempty"`
	Plants Plants        `json:"plants,omitempty"`
}

type TagsMessage struct {
	Header         MessageHeader `json:"header,omitempty"`
	FloweringTags  []string      `json:"floweringTags,omitempty"`
	VegetativeTags []string      `json:"vegetativeTags,omitempty"`
	DestroyedTags  []string      `json:"destroyedTags,omitempty"`
	HarvestedTags  []string      `json:"harvestedTags,omitempty"`
}

type PlantNote struct {
	ID         string    `json:"id,omitempty"`
	Tag        string    `json:"tag,omitempty"`
	Title      string    `json:"title,omitempty"`
	Body       string    `json:"body,omitempty"`
	CreateDate time.Time `json:"createDate,omitempty"`
	UpdateDate time.Time `json:"updateDate,omitempty"`
}

type PlantNotes []PlantNote

type PlantNoteMessage struct {
	Header MessageHeader `json:"header,omitempty"`
	Note   PlantNote     `json:"note,omitempty"`
}

type PlantNotesMessage struct {
	Header MessageHeader `json:"header,omitempty"`
	Notes  PlantNotes    `json:"notes,omitempty"`
}

type LoginMessage struct {
	Header   MessageHeader     `json:"header,omitempty"`
	Settings map[string]string `json:"settings,omitempty"`
}

//noinspection GoUnusedExportedFunction
func CreateHeader(status int, location string) MessageHeader {

	// Do we have a build version?
	//
	var build BuildVersion
	buildBytes, err := ioutil.ReadFile("/opt/build_version.json")
	if err == nil {
		err = json.Unmarshal(buildBytes, &build)
		log.Println("[INFO] Reading version: ", string(buildBytes))
	} else {
		log.Println("[ERROR] Unable to read /opt/build_version.json file:", err)
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

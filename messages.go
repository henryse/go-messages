// **********************************************************************
//    Copyright (c) 2018 Henry Seurer
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
	"os"
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

type SystemStatus int

const (
	DOWN SystemStatus = 0
	UP   SystemStatus = 1
)

type SystemStatusMap map[string]SystemStatus
type SystemStatusMessage struct {
	Header       MessageHeader   `json:"header"`
	SystemStatus SystemStatusMap `json:"message"`
}

type DeviceState int

const (
	ON  DeviceState = 0
	OFF DeviceState = 1
)

type DeviceInfo struct {
	Hardware string      `json:"hardware"`
	State    DeviceState `json:"state"`
}
type DeviceInfoMap map[string]DeviceInfo
type DevicesInfoMessage struct {
	Header  MessageHeader `json:"header"`
	Devices DeviceInfoMap `json:"devices"`
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

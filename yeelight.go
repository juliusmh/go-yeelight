package yeelight

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
)

// command is send to the light bulb.
type command struct {
	ID     int           `json:"id"`
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
}

// response is returned/received by the light bulb.
type response struct {
	ID     int      `json:"id"`
	Result []string `json:"result"`
}

// Method describes the method to send to the light bulb.
type Method string

var (
	MethodSetCTABX      Method = "set_ct_abx"
	MethodSetRGB        Method = "set_rgb"
	MethodSetHSV        Method = "set_hsv"
	MethodSetBrightness Method = "set_bright"
	MethodSetPower      Method = "set_power"
	MethodToggle        Method = "toggle"
)

// Convert a Method to string
func (m *Method) String() string {
	if m == nil {
		return ""
	}
	return string(*m)
}

// Bulb struct is used to control the lights.
type Bulb struct {
	mu    sync.Mutex
	cmdID int
	conn  net.Conn
}

// NewBulb creates a new Bulb object.
func NewBulb(address string) (*Bulb, error) {
	if !strings.Contains(address, ":") {
		address = address + ":55443"
	}
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("could not dial address: %+v", err)
	}
	return &Bulb{
		conn: conn,
	}, nil
}

// Send can be used to send commands to the light bulb. Each command is defined
// by a method and possible list of arguments. If the command can not be executed
// successfully the Send method will return an error, otherwise nil.
func (b *Bulb) Send(method Method, args ...interface{}) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	cmd := command{
		ID:     b.cmdID,
		Method: method.String(),
		Params: args,
	}

	err := json.NewEncoder(b.conn).Encode(cmd)
	if err != nil {
		return fmt.Errorf("cannot write json: %+v", err)
	}

	_, err = fmt.Fprint(b.conn, "\r\n")
	if err != nil {
		return fmt.Errorf("cannot write trailer: %+v", err)
	}

	var resp response
	err = json.NewDecoder(b.conn).Decode(&resp)
	if err != nil {
		return fmt.Errorf("receiving response: %+v", err)
	}

	b.cmdID++
	return nil
}

// TurnOn will turn the light bulb on.
func (b *Bulb) TurnOn() error {
	return b.Send(MethodSetPower, "on")
}

// TurnOff will turn the light bulb off.
func (b *Bulb) TurnOff() error {
	return b.Send(MethodSetPower, "off")
}

// ColorTemp will set the light bulbs color temperature
func (b *Bulb) ColorTemp(temp int) error {
	switch {
	case temp < 1700:
		temp = 1700
	case temp > 6500:
		temp = 6500
	}
	return b.Send(MethodSetCTABX, temp)
}

// RGB will set the light bulbs red, green and blue values.
func (b *Bulb) RGB(red, green, blue int) error {
	return b.Send(MethodSetRGB, red<<16+green<<8+blue)
}

// Brightness will set the light bulbs brightness.
func (b *Bulb) Brightness(brightness int) error {
	switch {
	case brightness > 100:
		brightness = 100
	case brightness < 1:
		brightness = 1
	}
	return b.Send(MethodSetBrightness, brightness)
}

package gpio

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
)

type Pin struct {
	n  int
	mu sync.Mutex
}

func NewPin(n int) *Pin {
	return &Pin{n: n}
}

type State bool

const (
	On  State = true
	Off State = false
)

type Direction bool

const (
	Read  Direction = false
	Write Direction = true
)

func (pin *Pin) Init() error {
	pin.mu.Lock()
	defer pin.mu.Unlock()

	return ioutil.WriteFile("/sys/class/gpio/export", []byte(strconv.Itoa(pin.n)), os.ModeAppend)
}

func (pin *Pin) Close() error {
	pin.mu.Lock()
	defer pin.mu.Unlock()

	return ioutil.WriteFile("/sys/class/gpio/unexport", []byte(strconv.Itoa(pin.n)), os.ModeAppend)
}

func (pin *Pin) SetDirection(dir Direction) error {
	v := "in"
	if dir {
		v = "out"
	}
	pin.mu.Lock()
	defer pin.mu.Unlock()

	return ioutil.WriteFile(fmt.Sprintf("/sys/class/gpio/gpio%d/direction", pin.n), []byte(v), os.ModeAppend)
}

func (pin *Pin) GetDirection() (Direction, error) {
	pin.mu.Lock()
	defer pin.mu.Unlock()

	byts, err := ioutil.ReadFile(fmt.Sprintf("/sys/class/gpio/gpio%d/direction", pin.n))
	if err != nil {
		return false, err
	}
	return string(byts) == "out", nil
}

func (pin *Pin) SetState(state State) error {
	v := "0"
	if state {
		v = "1"
	}
	pin.mu.Lock()
	defer pin.mu.Unlock()

	return ioutil.WriteFile(fmt.Sprintf("/sys/class/gpio/gpio%d/value", pin.n), []byte(v), os.ModeAppend)
}

func (pin *Pin) GetState() (State, error) {
	pin.mu.Lock()
	defer pin.mu.Unlock()

	byts, err := ioutil.ReadFile(fmt.Sprintf("/sys/class/gpio/gpio%d/value", pin.n))
	if err != nil {
		return false, err
	}
	return string(byts) == "1", nil
}

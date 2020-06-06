package main

import (
	"bytes"
	"container/list"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/msyrus/rpi-fan/gpio"
)

const windowSize = 20

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill, syscall.SIGTERM)

	ctx, stop := context.WithCancel(context.Background())

	go func() {
		<-sig
		stop()
	}()

	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
	fmt.Println("stopped gracefully")
}

func run(ctx context.Context) error {
	pin := gpio.NewPin(26)
	if err := pin.Init(); err != nil {
		return err
	}
	defer pin.Close()

	if err := pin.SetDirection(gpio.Write); err != nil {
		return err
	}

	mon := monitorTemp()
	go func() {
		<-ctx.Done()
		mon.Stop()
	}()

	if err := handleFan(pin, mon.C(), 35000, 45000); err != nil {
		return err
	}

	return mon.Err()
}

type Monitor struct {
	ch   chan int
	err  error
	done chan struct{}
	once sync.Once
}

func (m *Monitor) C() <-chan int {
	return m.ch
}

func (m *Monitor) Err() error {
	return m.err
}

func (m *Monitor) Done() bool {
	select {
	case <-m.done:
		return true
	default:
		return false
	}
}

func (m *Monitor) Stop() {
	m.once.Do(func() {
		close(m.done)
		<-m.ch
	})
}

func monitorTemp() *Monitor {
	m := &Monitor{
		ch:   make(chan int, 1),
		done: make(chan struct{}),
	}

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer func() {
			m.Stop()
			ticker.Stop()
		}()

		for {
			byts, err := ioutil.ReadFile("/sys/class/thermal/thermal_zone0/temp")
			if err != nil {
				m.err = err
				return
			}

			var t int
			fmt.Fscanf(bytes.NewReader(byts), "%d", &t)

			select {
			case <-m.done:
				close(m.ch)
				return
			case m.ch <- t:
			case <-m.ch:
				m.ch <- t
			}
			<-ticker.C
		}
	}()

	return m
}

func sum(arr *list.List, sum *int, val, sz int) {
	s := *sum
	n := arr.Len()

	s += val
	arr.PushBack(val)

	if n > sz {
		val = arr.Remove(arr.Front()).(int)
		s -= val
	}

	*sum = s

	return
}

type Window struct {
	arr        []int
	sz, sum, i int
}

func NewWindow(size int) *Window {
	return &Window{
		arr: make([]int, 0, size),
		sz:  size,
	}
}

func (w *Window) Avg() int {
	if len(w.arr) == 0 {
		return 0
	}
	return w.sum / len(w.arr)
}

func (w *Window) Add(v int) {
	w.sum += v
	if len(w.arr) < w.sz {
		w.arr = append(w.arr, v)
		return
	}

	w.sum -= w.arr[w.i]
	w.arr[w.i] = v
	w.i = (w.i + 1) % w.sz
}

func handleFan(fan *gpio.Pin, temps <-chan int, minTh, maxTh int) error {
	state, err := fan.GetState()
	if err != nil {
		return err
	}

	window := NewWindow(windowSize)

	for temp := range temps {
		window.Add(temp)
		avg := window.Avg()

		switch {
		case avg < minTh && state == gpio.On:
			state = gpio.Off
			fmt.Println("turning off")
			if err := fan.SetState(state); err != nil {
				return err
			}
		case avg > maxTh && state == gpio.Off:
			state = gpio.On
			fmt.Println("turning on")
			if err := fan.SetState(state); err != nil {
				return err
			}
		}
	}

	return nil
}

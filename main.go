package main

import (
	"fmt"
	"sync"
	"time"
)

type IntBuffer struct {
	buffer         []int
	size           int
	popped         chan bool
	pushed         chan bool
	readWriteMutex sync.Mutex
}

func (ib *IntBuffer) push(i int) {
	for {
		ib.readWriteMutex.Lock()
		if len(ib.buffer) < ib.size {
			break
		}

		ib.readWriteMutex.Unlock()
		<-ib.popped
	}

	ib.buffer = append(ib.buffer, i)
	ib.readWriteMutex.Unlock()

	select {
	case ib.pushed <- true:
	default:
	}
}

func (ib *IntBuffer) pop() int {
	var bufferLength int
	for {
		ib.readWriteMutex.Lock()
		bufferLength = len(ib.buffer)

		if bufferLength > 0 {
			break
		}

		ib.readWriteMutex.Unlock()
		<-ib.pushed
	}

	lastItem := ib.buffer[bufferLength-1]
	ib.buffer = ib.buffer[:bufferLength-1]
	ib.readWriteMutex.Unlock()

	select {
	case ib.popped <- true:
	default:
	}

	return lastItem
}

func NewIntBuffer(size int) *IntBuffer {
	return &IntBuffer{
		size:   size,
		pushed: make(chan bool),
		popped: make(chan bool),
	}
}

var wg sync.WaitGroup

func main() {
	ib := NewIntBuffer(2)

	ib.push(22)
	ib.push(8)

	wg.Add(1)
	go func() {
		defer wg.Done()
		ib.push(3)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(2 * time.Second)
		fmt.Println(ib.pop())
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(2 * time.Second)
		fmt.Println(ib.pop())
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ib.push(7)
	}()

	wg.Wait()

	fmt.Printf("%+v\n", ib)
}

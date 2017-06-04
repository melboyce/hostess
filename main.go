package main

import (
	"fmt"
)

func main() {
	c := make(chan string)

	go emitDate(c)
	go emitBattery(c)
	go emitThermal(c)
	go emitWifi(c)
	go emitLoad(c)

	for {
		select {
		case msg := <-c:
			fmt.Println(msg)
		}
	}
}

package main

import (
	"github.com/weirdtales/hostess"
)

func main() {
	pp := []hostess.Patron{
		hostess.Date,
		hostess.Batt,
		hostess.Thermal,
		hostess.Load,
		hostess.Wifi,
	}
	hostess.Collect(pp)
}

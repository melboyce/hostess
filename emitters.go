package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	battpath    = "/sys/class/power_supply/BAT0"
	thermalpath = "/sys/class/thermal/thermal_zone7"
)

func emitDate(c chan string) {
	for {
		t := time.Now()
		d := t.Format("2006-01-02")
		h := t.Format("15:04")
		dt := fmt.Sprintf("date\t%s\t%s", d, h)
		c <- dt
		time.Sleep(time.Minute)
	}
}

func emitBattery(c chan string) {
	if _, err := os.Stat(battpath + "/charge_full"); err != nil {
		return
	}
	var s int
	for {
		full, err := readInt(battpath + "/charge_full")
		if err != nil {
			panic(err)
		}
		now, err := readInt(battpath + "/charge_now")
		if err != nil {
			panic(err)
		}
		status, err := readString(battpath + "/status")
		if err != nil {
			panic(err)
		}
		s = 0
		if status == "Charging" {
			s = 1
		}
		pc := now * 100 / full
		b := fmt.Sprintf("batt\t%d\t%d\t%d", pc, pc/20, s)
		c <- b
		time.Sleep(10 * time.Second)
	}
}

func emitThermal(c chan string) {
	if _, err := os.Stat(thermalpath + "/temp"); err != nil {
		return
	}
	i := 0
	for {
		temp, err := readInt(thermalpath + "/temp")
		if err != nil {
			panic(err)
		}
		temp = temp / 1000
		switch {
		case 0 <= temp && temp <= 30:
			i = 0
		case 31 <= temp && temp <= 35:
			i = 1
		case 36 <= temp && temp <= 40:
			i = 2
		case 41 <= temp && temp <= 45:
			i = 3
		case temp >= 46:
			i = 4
		}
		t := fmt.Sprintf("temp\t%d\t%d", temp, i)
		c <- t
		time.Sleep(5 * time.Second)
	}
}

func emitWifi(c chan string) {
	iface := "wlp58s0"
	var (
		o    []byte
		err  error
		qual float64
	)
	for {
		if o, err = exec.Command("iwgetid", "-r").Output(); err != nil {
			c <- fmt.Sprintf("wifi\t-\t0")
			time.Sleep(time.Minute) // TODO
			continue
		}
		ssid := strings.TrimSpace(string(o))
		if o, err = exec.Command("iwconfig", iface).Output(); err != nil {
			c <- fmt.Sprintf("wifi\t-\t0")
			time.Sleep(time.Minute) // TODO
			continue
		}
		stats := strings.TrimSpace(string(o))
		r, _ := regexp.Compile(`Quality=([0-9/]+)`)
		res := r.FindStringSubmatch(stats)
		qual = 0.0
		if len(res) >= 2 {
			els := strings.Split(res[1], "/")
			qmin, _ := strconv.Atoi(els[0])
			qmax, _ := strconv.Atoi(els[1])
			qual = float64(qmin) / float64(qmax) * 100.0
		}
		lq := int(qual)
		c <- fmt.Sprintf("wifi\t%s\t%d\t%d", ssid, lq, lq/20)
		time.Sleep(time.Minute)
	}
}

func emitLoad(c chan string) {
	for {
		la, err := readString("/proc/loadavg")
		if err != nil {
			panic(err)
		}
		lavg := strings.Join(strings.Split(la, " ")[:3], " ")
		c <- fmt.Sprintf("load\t%s", lavg)
		time.Sleep(time.Minute)
	}
}

func readInt(path string) (i int, err error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	s := strings.TrimSpace(string(b))
	i, err = strconv.Atoi(s)
	return
}

func readString(path string) (s string, err error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	s = strings.TrimSpace(string(b))
	return
}

package hostess

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const battpath = "/sys/class/power_supply/BAT0"
const thermalpath = "/sys/class/thermal/thermal_zone7"

// Patron ...
type Patron func() (time.Duration, string, error)

// Collect ...
func Collect(pp []Patron) {
	ch := make(chan string)
	for _, p := range pp {
		go func(p Patron) {
			var t time.Duration
			var msg string
			var err error
			for {
				t, msg, err = p()
				if err != nil {
					return
				}
				ch <- fmt.Sprint(msg)
				time.Sleep(t)
			}
		}(p)
	}
	for {
		select {
		case msg := <-ch:
			fmt.Println(msg)
		}
	}
}

// Date ...
func Date() (time.Duration, string, error) {
	t := time.Now()
	y := t.Format("2006-01-02")
	h := t.Format("15:04")
	return time.Minute, fmt.Sprintf("date\t%s\t%s", y, h), nil
}

// Batt ...
func Batt() (time.Duration, string, error) {
	cf := filepath.Join(battpath, "/charge_full")
	cn := filepath.Join(battpath, "/charge_now")
	st := filepath.Join(battpath, "/status")

	if _, err := os.Stat(cf); err != nil {
		return time.Hour, "", err
	}
	full, err := readInt(cf)
	if err != nil {
		return time.Hour, "", err
	}
	now, err := readInt(cn)
	if err != nil {
		return time.Hour, "", err
	}
	status, err := readString(st)
	if err != nil {
		return time.Hour, "", err
	}

	pc := now * 100 / full
	if pc > 100 {
		pc = 100
	}
	ci := 0
	if status == "Charging" {
		ci = 1
	}

	batt := fmt.Sprintf("batt\t%d\t%d\t%d", pc, pc/20, ci)
	return 10 * time.Second, batt, nil
}

// Thermal ...
func Thermal() (time.Duration, string, error) {
	tp := filepath.Join(thermalpath, "/temp")
	if _, err := os.Stat(tp); err != nil {
		return time.Hour, "", err
	}
	t, err := readInt(tp)
	if err != nil {
		return time.Hour, "", err
	}
	t = t / 1000
	i := (t - 25) * 100 / 30 / 25

	therm := fmt.Sprintf("temp\t%d\t%d", t, i)
	return 10 * time.Second, therm, nil
}

// Load ...
func Load() (time.Duration, string, error) {
	la, err := readString("/proc/loadavg")
	if err != nil {
		return time.Hour, "", err
	}
	avgs := strings.Join(strings.Split(la, " ")[:3], " ")
	return time.Minute, fmt.Sprintf("load\t%s", avgs), nil
}

// Wifi ...
func Wifi() (time.Duration, string, error) {
	dcmsg := fmt.Sprintf("wifi\t-\t0\t0")
	b, err := exec.Command("iwgetid").Output()
	if err != nil {
		// assume d/c and continue to poll
		return time.Minute, dcmsg, nil
	}
	s := string(b)
	ssid := strings.Split(s, "\"")[1]
	iface := strings.Split(s, " ")[0]
	if iface == "" {
		return time.Minute, dcmsg, nil
	}

	b, err = exec.Command("iwconfig", iface).Output()
	if err != nil {
		return time.Minute, dcmsg, nil
	}
	re := regexp.MustCompile(`Quality=([0-9/]+)`)
	m := re.FindStringSubmatch(string(b))
	q := float64(0.0)
	if len(m) >= 2 {
		e := strings.Split(m[1], "/")
		mn, _ := strconv.Atoi(e[0])
		mx, _ := strconv.Atoi(e[1])
		q = float64(mn) / float64(mx) * 100.0
	}
	lq := int(q)
	return time.Minute, fmt.Sprintf("wifi\t%s\t%d\t%d", ssid, lq, lq/20), nil
}

// readInt is a dirty convenience.
func readInt(f string) (int, error) {
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(b)))
}

// readString is a dirty convenience also.
func readString(f string) (string, error) {
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

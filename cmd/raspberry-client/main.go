package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/brian-armstrong/gpio"
)

var (
	conf = flag.Bool("configure", false, "use to detect pins used by")
)

func configure() {
	watcher := gpio.NewWatcher()

	for i := uint(1); i < 28; i++ {
		watcher.AddPin(i)
	}
	defer watcher.Close()

	ch := make(chan [2]uint)

	// Empty initial values
	go func() {
		for {
			pin, value := watcher.Watch()
			ch <- [2]uint{pin, value}
		}
	}()

	fmt.Print("performing initial measurements: ")
loop:
	for {
		select {
		case <-ch:
			fmt.Print(".")
		case <-time.After(1000 * time.Millisecond):
			break loop
		}
	}

	fmt.Println("")

	var inputFlag []string
	for _, c := range []string{"A", "B", "C", "D"} {
		fmt.Printf("press '%s':", c)
		var p [2]uint
		for p = range ch {

			if p[1] == 1 {
				// only detect pin=1
				break
			}
		}
		fmt.Printf("found %d\n", p[0])
		inputFlag = append(inputFlag, fmt.Sprintf("%s=%d", c, p[0]))
	}

	fmt.Printf("mapping complete:\nplease use -i %s flag when running\n", strings.Join(inputFlag, " -i "))
}

func main() {

	flag.Var(&devices, "i", "Some description for this param.")

	flag.Parse()

	if *conf {
		configure()
		os.Exit(0)
	}

	fmt.Println("starting")

	watcher := gpio.NewWatcher()

	for pin := range devices {
		watcher.AddPin(pin)
	}
	defer watcher.Close()

	in := make(chan string)
	matched := make(chan string)
	go tempWaitForMatch(in, matched, "ABBA")

	go func() {
		for {
			pin, value := watcher.Watch()
			if value == 0 {
				continue
			}
			fmt.Printf("read %d from gpio %d='%s'\n", value, pin, devices[pin])
			in <- devices[pin]
		}
	}()

	select {
	case c := <-matched:
		fmt.Println(c)
	}
	fmt.Println("successfully reached the end!")
}

type arrayFlags map[uint]string

var devices = arrayFlags{}

func (i arrayFlags) String() string {
	return "my string representation"
}

func (i arrayFlags) Set(value string) error {

	s := strings.Split(value, "=")

	pin, err := strconv.Atoi(s[1])
	if err != nil {
		return err
	}

	i[uint(pin)] = s[0]
	return nil
}

func tempWaitForMatch(input <-chan string, matched chan<- string, wanted string) {

	var pattern string
	for s := range input {

		// Loop over character
		for _, ss := range s {
			pattern += string(ss)
			if len(pattern) > len(wanted) {
				pattern = pattern[1:]
			}
			if pattern == wanted {
				matched <- pattern
				return
			}
		}
	}
}

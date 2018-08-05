package main

import (
	"fmt"
	"strings"

	"github.com/kyeett/escape-the-world/pkg/device-input"
	termbox "github.com/nsf/termbox-go"
)

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	termbox.Flush()

	fmt.Println("Ctrl+C or Ctrl+Q to quit program")

	in := make(chan string)
	matched := make(chan string)
	quit := make(chan bool)
	go input.WaitForMatch(in, matched, "ABBA")

	go func() {
		for {
			switch ev := termbox.PollEvent(); ev.Type {
			case termbox.EventKey:
				if ev.Key == termbox.KeyCtrlQ || ev.Key == termbox.KeyCtrlC {
					close(in)
					quit <- true
					return
				}
				s := strings.ToUpper(string(ev.Ch))
				fmt.Println("In:" + s)
				in <- s
			case termbox.EventError:
				panic(ev.Err)
			}
		}
	}()

	select {
	case m := <-matched:
		fmt.Printf("received wanted match '%s'\n", m)
	case <-quit:
	}
	fmt.Print("exit program")
}

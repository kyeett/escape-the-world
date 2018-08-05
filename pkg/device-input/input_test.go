package input

import (
	"fmt"
	"testing"
	"time"
)

func TestFullString(t *testing.T) {

	tcs := []string{
		"ABCD",
		"ABBACDCDBAAB",
		"C",
	}

	for _, tcInput := range tcs {

		t.Run(fmt.Sprintf("input %s", tcInput), func(t *testing.T) {
			t.Parallel()

			input := make(chan string)
			matched := make(chan string)

			go WaitForMatch(input, matched, tcInput)
			input <- tcInput

			select {
			case <-matched:
				// Success!
			case <-time.After(100 * time.Millisecond):
				t.Fatal("timeout while waiting for input matched")
			}
			close(input)
			close(matched)
		})
	}
}

func TestOneByOne(t *testing.T) {

	tcs := []string{
		"ABCD",
		"ABBACDCDBAAB",
		"C",
	}

	for _, tcInput := range tcs {

		t.Run(fmt.Sprintf("input %s", tcInput), func(t *testing.T) {
			t.Parallel()

			input := make(chan string)
			matched := make(chan string)

			go WaitForMatch(input, matched, tcInput)

			for _, s := range tcInput {
				input <- string(s)
			}

			select {
			case <-matched:
				// Success!
			case <-time.After(100 * time.Millisecond):
				t.Fatal("timeout while waiting for input matched")
			}
			close(input)
			close(matched)
		})
	}
}

func TestMatchPartial(t *testing.T) {

	tcs := []struct {
		in   []string
		want string
	}{
		{
			in:   []string{"A", "B", "C", "DA"},
			want: "ABCD",
		},
		{
			in:   []string{"ABBACDCD", "BA", "A", "BAAB"},
			want: "ABBACDCDBAAB",
		},
		{
			in:   []string{"A", "B", "C", "DA"},
			want: "C",
		},
	}

	for _, tc := range tcs {

		t.Run(fmt.Sprintf("input %s", tc.want), func(t *testing.T) {
			//t.Parallel()

			input := make(chan string)
			matched := make(chan string)

			go WaitForMatch(input, matched, tc.want)

			go func() {
				for _, s := range tc.in {
					input <- string(s)
				}
				close(input)
			}()

			select {
			case <-matched:
				// Success!
			case <-time.After(100 * time.Millisecond):
				t.Fatal("timeout while waiting for input matched")
			}

			close(matched)
		})
	}
}

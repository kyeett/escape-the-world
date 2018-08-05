package input

func WaitForMatch(input <-chan string, matched chan<- string, wanted string) {

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

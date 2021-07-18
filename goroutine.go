package main

func readAllErrors(amount int, errorChannel chan error) []error {
	errors := make([]error, 0)
	for i := 0; i < amount; i++ {
		err := <-errorChannel
		if err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

func runMany(amount int, function func(index int) error) []error {
	errorChannel := make(chan error)
	for i := 0; i < amount; i++ {
		index := i
		go func() { errorChannel <- function(index) }()
	}
	return readAllErrors(amount, errorChannel)
}

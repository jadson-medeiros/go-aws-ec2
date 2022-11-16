package main

import (
	"fmt"
	"os"
)

func main() {

	var (
		instanceId string
		err        error
	)

	if instanceId, err = createEC2(); err != nil {
		fmt.Printf("createEC2 error: %s", err)
		os.Exit(1)
	}

	fmt.Printf("Instace ID: %s\n", instanceId)
}

func createEC2() (string, error) {
	return "", nil
}

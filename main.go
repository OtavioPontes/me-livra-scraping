package main

import (
	"fmt"
	"log"
)

func main() {
	err := writeDepartmentsToJSON()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	err = writeTeachersToJSON()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println("Teachers data written to JSON file successfully.")
}

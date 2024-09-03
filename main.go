package main

import (
	"fmt"
	"log"
	"me-livra-scraping/src/unb"
)

func main() {
	// err := ufg.WriteDepartmentsToJSON()
	// if err != nil {
	// 	log.Fatalf("Error: %v", err)
	// }

	// err = ufg.WriteTeachersToJSON()
	// if err != nil {
	// 	log.Fatalf("Error: %v", err)
	// }

	err := unb.WriteDepartmentsToJSON()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	err = unb.WriteTeachersToJSON()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println("Teachers data written to JSON file successfully.")
}

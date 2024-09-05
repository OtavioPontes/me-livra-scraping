package main

import (
	"fmt"
	"log"
	"me-livra-scraping/src/geral"
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

	// err := unb.WriteDepartmentsToJSON()
	// if err != nil {
	// 	log.Fatalf("Error: %v", err)
	// }

	// err = unb.WriteTeachersToJSON()
	// if err != nil {
	// 	log.Fatalf("Error: %v", err)
	// }

	// err := ufmg.WriteTeachersToJSON()
	// if err != nil {
	// 	log.Fatalf("Error: %v", err)
	// }

	// err := ufrj.WriteTeachersToJSON()
	// if err != nil {
	// 	log.Fatalf("Error: %v", err)
	// }

	err := geral.WriteTeachersToJSON()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println("Teachers data written to JSON file successfully.")
}

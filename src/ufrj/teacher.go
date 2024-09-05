package ufrj

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

// Estrutura para armazenar as informações de departamento e professores
type Department struct {
	Department string    `json:"department"`
	Teachers   []Teacher `json:"teachers"`
}

type Teacher struct {
	Name string `json:"name"`
}

func WriteTeachersToJSON() error {

	// Abrir o arquivo CSV
	file, err := os.Open("files/ufrj/ufrj.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Ler o arquivo CSV
	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.Comma = ';'
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	// Mapa para armazenar os departamentos e professores
	deptMap := make(map[string][]Teacher)

	for i, record := range records {
		if i == 0 {
			continue
		}
		// O nome do professor está na primeira coluna e o departamento na terceira
		name := strings.TrimSpace(record[0])
		department := strings.TrimSpace(record[2])

		// Adiciona o professor ao departamento correspondente
		deptMap[department] = append(deptMap[department], Teacher{Name: name})
	}

	// Converter o mapa em uma lista de departamentos
	var departments []Department
	for dept, teachers := range deptMap {
		departments = append(departments, Department{
			Department: dept,
			Teachers:   teachers,
		})
	}

	// Converter para JSON
	jsonData, err := json.MarshalIndent(departments, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	// Exibir o JSON
	fmt.Println(string(jsonData))

	// Ou, salvar em um arquivo
	err = os.WriteFile("files/ufrj/teachers.json", jsonData, 0644)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

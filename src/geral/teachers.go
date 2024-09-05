package geral

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

// Estrutura para armazenar a universidade, departamento e professores
type University struct {
	Name        string       `json:"university"`
	Departments []Department `json:"departments"`
}

type Department struct {
	Department string    `json:"department"`
	Teachers   []Teacher `json:"teachers"`
}

type Teacher struct {
	Name string `json:"name"`
}

func OptimizeCsv() error {
	// Abrir o arquivo CSV
	file, err := os.Open("files/geral/servidores_utf8.csv") // Substitua pelo caminho do seu arquivo
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Criar o leitor CSV
	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.Comma = ';'

	// Ler todas as linhas do arquivo CSV
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	// Criar o novo arquivo CSV para saída
	outFile, err := os.Create("files/geral/servidores_opt.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()

	// Criar o escritor CSV para o novo arquivo
	writer := csv.NewWriter(outFile)
	defer writer.Flush()

	// Escrever o cabeçalho no novo CSV
	header := []string{"Nome", "Órgão de Exercício", "UORG de Lotação"}
	if err := writer.Write(header); err != nil {
		log.Fatal(err)
	}

	// Iterar sobre as linhas e extrair os campos desejados
	for _, record := range records {
		// Ignorar a linha de cabeçalho
		if record[0] == "Id_SERVIDOR_PORTAL" {
			continue
		}

		// Pegar o nome (índice 1), órgão de exercício (índice 23), e UORG de lotação (índice 16)
		nome := record[1]
		orgaoExercicio := record[18]
		uorgLotacao := record[16]

		// Escrever a nova linha no arquivo CSV
		newRecord := []string{nome, orgaoExercicio, uorgLotacao}
		if err := writer.Write(newRecord); err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Novo arquivo CSV criado com sucesso!")
	return nil
}

func WriteTeachersToJSON() error {
	OptimizeCsv()
	// Abrir o arquivo CSV
	file, err := os.Open("files/geral/servidores_opt.csv") // Substitua pelo caminho do seu arquivo
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Ler o arquivo CSV
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	// Mapa para armazenar universidades, departamentos e professores
	universityMap := make(map[string]map[string][]Teacher)

	// Ignora a primeira linha do CSV (cabeçalhos)
	for i, record := range records {
		if i == 0 {
			continue
		}

		// Extrair nome do professor, universidade e departamento
		name := strings.TrimSpace(record[0])
		university := strings.TrimSpace(record[1])
		department := strings.TrimSpace(record[2])

		// Adiciona o professor à universidade e ao departamento correspondente
		if _, ok := universityMap[university]; !ok {
			universityMap[university] = make(map[string][]Teacher)
		}
		universityMap[university][department] = append(universityMap[university][department], Teacher{Name: name})
	}

	// Converter o mapa para uma lista de universidades
	var universities []University
	for uni, deptMap := range universityMap {
		var departments []Department
		for dept, teachers := range deptMap {
			departments = append(departments, Department{
				Department: dept,
				Teachers:   teachers,
			})
		}
		universities = append(universities, University{
			Name:        uni,
			Departments: departments,
		})
	}

	// Converter para JSON
	jsonData, err := json.MarshalIndent(universities, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	// Exibir o JSON
	fmt.Println(string(jsonData))

	// Ou, salvar em um arquivo
	err = os.WriteFile("files/geral/teachers.json", jsonData, 0644)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

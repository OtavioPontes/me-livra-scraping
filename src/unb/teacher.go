package unb

import (
	"context"
	"encoding/json"
	"log"
	"me-livra-scraping/src/models"
	"os"
	"path/filepath"

	"github.com/chromedp/chromedp"
)

func getTeachersList(ctx context.Context, value string) ([]models.Teacher, error) {
	var names []string

	err := chromedp.Run(ctx,
		chromedp.Navigate("https://sigaa.unb.br/sigaa/public/docente/busca_docentes.jsf"),
		// Espera a página carregar
		chromedp.WaitVisible(`form:departamento`),
		// Preenche o campo do instituto
		chromedp.SetValue(`form:departamento`, value),
		// Clica no botão buscar
		chromedp.Click(`form:buscar`),
		// Espera a tabela de resultados aparecer
		chromedp.WaitVisible(`table`),
		// Extract the names using XPath
		chromedp.Evaluate(`Array.from(document.querySelectorAll(".listagem td .nome")).map(el => el.textContent)`, &names),
	)
	if err != nil {
		return nil, err
	}

	// Combina os nomes e URLs em structs `Teacher`
	teachers := make([]models.Teacher, len(names))
	for i := range names {
		log.Println(names[i])
		teachers[i].Name = titleCase(names[i])
	}

	return teachers, nil
}

func WriteTeachersToJSON() error {
	// Load department data
	departmentFilePath := filepath.Join("files", "unb", "departments.json")
	filePath := filepath.Join("files", "unb", "teachers.json")

	departmentFile, err := os.ReadFile(departmentFilePath)
	if err != nil {
		return err
	}

	var departments []models.Department
	err = json.Unmarshal(departmentFile, &departments)
	if err != nil {
		return err
	}

	teachersObject := make([]map[string]interface{}, len(departments))

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),        // Desativa o modo headless
		chromedp.Flag("disable-gpu", false),    // Não desabilita a GPU
		chromedp.Flag("start-maximized", true), // Inicia maximizado
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)

	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)

	defer cancel()

	for i, dep := range departments {
		log.Println(dep.Name)
		teachersObject[i] = map[string]interface{}{"department": dep.Name}
		teachers, err := getTeachersList(ctx, dep.Value)
		if err != nil {
			log.Printf("Failed to get teachers for department %s: %v", dep.Name, err)
			continue
		}

		teachersObject[i]["teachers"] = teachers
	}

	jsonData, err := json.MarshalIndent(teachersObject, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

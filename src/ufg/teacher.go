package ufg

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
		chromedp.Navigate("https://sigaa.sistemas.ufg.br/sigaa/public/docente/busca_docentes.jsf?aba=p-academico"),
		chromedp.SetValue(`form:departamento`, value),
		chromedp.Click(`form:buscar`),
		chromedp.WaitVisible(`table`),

		chromedp.Evaluate(`Array.from(document.querySelectorAll('.listagem td .nome')).map(name => name.textContent);`, &names),
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
	departmentFilePath := filepath.Join("files", "ufg", "departments.json")
	filePath := filepath.Join("files", "ufg", "teachers.json")

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

	ctx, cancel := chromedp.NewContext(context.Background())
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

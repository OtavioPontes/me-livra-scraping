package ufmg

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"me-livra-scraping/src/models"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/chromedp/chromedp"
)

func getTeachersList(ctx context.Context, value int) ([]models.UfmgTeacher, error) {
	var namesAndInstitutes []struct {
		Name      string `json:"name"`
		Institute string `json:"institute"`
	}

	err := chromedp.Run(ctx,
		chromedp.Navigate(fmt.Sprintf(`https://ufmg.br/busca/?q=Professores&aba=pessoas&pagina=%d`, value)),

		// Extract the names using XPath
		chromedp.Evaluate(`
			Array.from(document.querySelectorAll(".people__info")).map(el => {
				const name = el.querySelector("h3").textContent.trim();
				const institute = el.querySelector("p").textContent.trim();
				return { name, institute };
			})
		`, &namesAndInstitutes),
	)
	if err != nil {
		return nil, err
	}

	// Combina os nomes e URLs em structs `Teacher`
	teachers := make([]models.UfmgTeacher, len(namesAndInstitutes))
	for i := range namesAndInstitutes {
		log.Println(namesAndInstitutes[i])
		teachers[i].Name = titleCase(namesAndInstitutes[i].Name)
		teachers[i].Institute = titleCase(namesAndInstitutes[i].Institute)
	}

	return teachers, nil
}

func WriteTeachersToJSON() error {

	type DepartmentWithTeacher struct {
		Department string           `json:"department"`
		Teachers   []models.Teacher `json:"teachers"`
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),        // Desativa o modo headless
		chromedp.Flag("disable-gpu", false),    // Não desabilita a GPU
		chromedp.Flag("start-maximized", true), // Inicia maximizado
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)

	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)

	defer cancel()

	var teachers []models.UfmgTeacher

	for i := 1; i < 256; i++ {
		log.Println(i)

		list, err := getTeachersList(ctx, i)
		if err != nil {
			log.Printf("Erro ao pegar a página: %d", i)
			continue
		}
		teachers = append(teachers, list...)

	}

	departmentMap := make(map[string][]models.Teacher)
	for _, teacher := range teachers {
		departmentMap[teacher.Institute] = append(departmentMap[teacher.Institute], models.Teacher{Name: teacher.Name})

	}

	var departments []DepartmentWithTeacher
	for institute, teachers := range departmentMap {
		departments = append(departments, DepartmentWithTeacher{
			Department: institute,
			Teachers:   teachers,
		})
	}

	jsonData, err := json.MarshalIndent(departments, "", "  ")
	if err != nil {
		return err
	}

	filePath := filepath.Join("files", "ufmg", "teachers.json")

	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}
func titleCase(s string) string {
	return strings.Map(func(r rune) rune {
		return unicode.ToTitle(unicode.ToLower(r))
	}, s)
}

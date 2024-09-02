package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/gocolly/colly"
)

// Department represents a department with a name and value.
type Department struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// titleCase converts a string to title case.
func titleCase(s string) string {
	return strings.Map(func(r rune) rune {
		return unicode.ToTitle(unicode.ToLower(r))
	}, s)
}

// getDepartmentsList retrieves the list of departments from the webpage.
func getDepartmentsList() ([]Department, error) {
	var departments []Department

	c := colly.NewCollector()

	// On every option element in the select dropdown
	c.OnHTML("#form\\:departamento option", func(e *colly.HTMLElement) {
		name := e.Text
		value := e.Attr("value")

		if value != "" && value != "0" {
			department := Department{
				Name:  titleCase(strings.Split(name, "-")[0]),
				Value: strings.TrimSpace(value),
			}
			departments = append(departments, department)
		}
	})

	// Visit the page containing the departments list
	err := c.Visit("https://sigaa.sistemas.ufg.br/sigaa/public/docente/busca_docentes.jsf?aba=p-academico")
	if err != nil {
		return nil, err
	}

	return departments, nil
}

// writeDepartmentsToJSON writes the departments list to a JSON file.
func writeDepartmentsToJSON() error {
	// Define the file path
	filePath := filepath.Join("files", "departments.json")

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}

	// Retrieve the departments list
	departments, err := getDepartmentsList()
	if err != nil {
		return err
	}

	// Convert departments to JSON
	jsonDepartments, err := json.MarshalIndent(departments, "", "  ")
	if err != nil {
		return err
	}

	// Write JSON to file
	err = os.WriteFile(filePath, jsonDepartments, 0644)
	if err != nil {
		return err
	}

	return nil
}

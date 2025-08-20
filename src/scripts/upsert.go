package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"

	_ "github.com/lib/pq"
	"github.com/xrash/smetrics" // Install with: go get -u github.com/xrash/smetrics
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Department represents a department from the departments.json file
type Department struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// TeacherEntry represents an entry in the teachers.json file
type TeacherEntry struct {
	Department string    `json:"department"`
	Teachers   []Teacher `json:"teachers"`
}

// Teacher represents a teacher in the teachers.json file
type Teacher struct {
	Name string `json:"name"`
}

// Institute represents an institute in the database
type Institute struct {
	ID         int
	Name       string
	University string
	Initials   string
	AvgGrade   float64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Professor represents a professor in the database
type Professor struct {
	ID          int
	Name        string
	InstituteID int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	AvgGrade    float64
}

// NormalizeName standardizes a name for comparison
func NormalizeName(name string) string {

	result := strings.ToLower(name)

	re := regexp.MustCompile(`[^\p{L}\p{N}\s]`)
	result = re.ReplaceAllString(result, "")

	// Remove extra spaces and trim
	re = regexp.MustCompile(`\s+`)
	result = re.ReplaceAllString(result, " ")
	result = strings.TrimSpace(result)

	result = removeAccents(result)

	c := cases.Title(language.Portuguese)
	result = c.String(result)

	return result
}

// FormatPortugueseText applies proper casing for Portuguese text
// It applies title case to words but keeps prepositions, articles and conjunctions lowercase
func FormatPortugueseText(text string) string {
	if text == "" {
		return text
	}

	// First, apply title case to the entire string
	c := cases.Title(language.Portuguese)
	result := c.String(text)

	// Define Portuguese prepositions, articles and conjunctions to be in lowercase
	lowerCaseWords := []string{
		"A", "O", "As", "Os", // articles
		"Da", "De", "Do", "Das", "Dos", "Dum", "Duma", "Duns", "Dumas", // contractions with de
		"Na", "No", "Nas", "Nos", "Num", "Numa", "Nuns", "Numas", // contractions with em
		"Ao", "Aos", "À", "Às", // contractions with a
		"Pelo", "Pela", "Pelos", "Pelas", // contractions with por
		"Para", "Por", "Com", "Em", "Sem", "Sob", "Sobre", // simple prepositions
		"E", "Ou", "Mas", "Porém", "Contudo", "Todavia", "Entretanto", // conjunctions
		"Que", "Se", "Como", "Quando", "Quanto", // relative pronouns/conjunctions
	}

	// Create a map for faster lookups
	lowerCaseMap := make(map[string]bool)
	for _, word := range lowerCaseWords {
		lowerCaseMap[word] = true
	}

	// Split the string into words and apply the proper casing
	words := strings.Split(result, " ")
	for i, word := range words {
		// Keep the first and last word capitalized, regardless of what it is
		if i > 0 && i < len(words)-1 {
			// Check if the word is in our list of words to be lowercase
			if lowerCaseMap[word] {
				words[i] = strings.ToLower(word)
			}
		}
	}

	// Join the words back
	return strings.Join(words, " ")
}

func removeAccents(s string) string {
	t := make([]rune, 0, len(s))
	for _, c := range s {
		switch c {
		case 'Á', 'À', 'Ã', 'Â', 'Ä', 'á', 'à', 'ã', 'â', 'ä':
			t = append(t, 'a')
		case 'É', 'È', 'Ê', 'Ë', 'é', 'è', 'ê', 'ë':
			t = append(t, 'e')
		case 'Í', 'Ì', 'Î', 'Ï', 'í', 'ì', 'î', 'ï':
			t = append(t, 'i')
		case 'Ó', 'Ò', 'Õ', 'Ô', 'Ö', 'ó', 'ò', 'õ', 'ô', 'ö':
			t = append(t, 'o')
		case 'Ú', 'Ù', 'Û', 'Ü', 'ú', 'ù', 'û', 'ü':
			t = append(t, 'u')
		case 'Ç', 'ç':
			t = append(t, 'c')
		case 'Ñ', 'ñ':
			t = append(t, 'n')
		default:
			t = append(t, c)
		}
	}
	return string(t)
}

// IsDuplicate checks if a professor already exists in the database
func IsDuplicate(scrapedProf Professor, existing []Professor, threshold float64) bool {
	normScraped := NormalizeName(scrapedProf.Name)
	for _, prof := range existing {
		if prof.InstituteID != scrapedProf.InstituteID {
			continue
		}

		similarity := smetrics.JaroWinkler(normScraped, prof.Name, 0.7, 4)
		if similarity >= threshold {
			return true
		}
	}
	return false
}

// IsDuplicateDepartment checks if a department name is likely a duplicate of an existing one
func IsDuplicateDepartment(name string, departments map[string]string, threshold float64) (string, bool) {
	normalized := NormalizeName(name)

	// Check for exact match first
	if officialName, exists := departments[normalized]; exists {
		return officialName, true
	}

	// Track best match
	bestMatch := ""
	bestScore := 0.0

	// Check for fuzzy match
	for normalizedDept, officialName := range departments {
		// Skip empty department names
		if normalizedDept == "" || officialName == "" {
			continue
		}

		// Try different similarity algorithms
		jaroSim := smetrics.JaroWinkler(normalized, normalizedDept, 0.7, 4)

		// Check trigram similarity for abbr/acronym handling
		trigramSim := calculateTrigramSimilarity(normalized, normalizedDept)

		// Combine scores with weights
		combinedScore := (jaroSim * 0.8) + (trigramSim * 0.2)

		if combinedScore > bestScore && combinedScore >= threshold {
			bestScore = combinedScore
			bestMatch = officialName
		}

		// Debug output
		if jaroSim > 0.85 {
			fmt.Printf("Department similarity: '%s' ⟷ '%s' = %.2f (JW) / %.2f (TRI) / %.2f (Combined)\n",
				normalized, normalizedDept, jaroSim, trigramSim, combinedScore)
		}
	}

	if bestMatch != "" {
		fmt.Printf("Best department match for '%s': '%s' (score: %.2f)\n", name, bestMatch, bestScore)
		return bestMatch, true
	}

	return "", false
}

// calculateTrigramSimilarity computes trigram similarity between strings
func calculateTrigramSimilarity(s1, s2 string) float64 {
	if len(s1) < 3 || len(s2) < 3 {
		if s1 == s2 {
			return 1.0
		}
		return 0.0
	}

	// Extract trigrams
	trigrams1 := extractTrigrams(s1)
	trigrams2 := extractTrigrams(s2)

	// Count common trigrams
	common := 0
	for _, t1 := range trigrams1 {
		for _, t2 := range trigrams2 {
			if t1 == t2 {
				common++
				break
			}
		}
	}

	// Calculate dice coefficient
	total := len(trigrams1) + len(trigrams2)
	if total == 0 {
		return 0.0
	}

	return float64(2*common) / float64(total)
}

// extractTrigrams gets all trigrams from a string
func extractTrigrams(s string) []string {
	if len(s) < 3 {
		return []string{s}
	}

	trigrams := make([]string, 0, len(s)-2)
	for i := 0; i <= len(s)-3; i++ {
		trigram := s[i : i+3]
		trigrams = append(trigrams, trigram)
	}
	return trigrams
}

// generateInitials creates initials from a department name
// Ignores prepositions, articles and conjunctions
func generateInitials(name string) string {
	words := strings.Fields(strings.ToUpper(name))
	initials := ""

	// Define words to ignore when creating initials
	// These are prepositions, articles, and conjunctions in Portuguese
	ignoreWords := map[string]bool{
		"A": true, "O": true, "AS": true, "OS": true, // articles
		"DA": true, "DE": true, "DO": true, "DAS": true, "DOS": true,
		"DUM": true, "DUMA": true, "DUNS": true, "DUMAS": true, // contractions with de
		"NA": true, "NO": true, "NAS": true, "NOS": true,
		"NUM": true, "NUMA": true, "NUNS": true, "NUMAS": true, // contractions with em
		"AO": true, "AOS": true, "À": true, "ÀS": true, // contractions with a
		"PELO": true, "PELA": true, "PELOS": true, "PELAS": true, // contractions with por
		"PARA": true, "POR": true, "COM": true, "EM": true,
		"SEM": true, "SOB": true, "SOBRE": true, // simple prepositions
		"E": true, "OU": true, "MAS": true, "PORÉM": true,
		"CONTUDO": true, "TODAVIA": true, "ENTRETANTO": true, // conjunctions
		"QUE": true, "SE": true, "COMO": true, "QUANDO": true, "QUANTO": true, // relative pronouns
	}

	for _, word := range words {
		if len(word) > 0 && unicode.IsLetter(rune(word[0])) {
			// Only include this word if it's not in our ignore list
			if !ignoreWords[word] {
				initials += string(word[0])
			}
		}
	}

	// If no initials were generated (maybe all words were prepositions)
	// then use the first letter of the first word
	if initials == "" && len(words) > 0 && len(words[0]) > 0 {
		initials = string(words[0][0])
	}

	// Limit to 5 characters
	if len(initials) > 5 {
		initials = initials[:5]
	}

	return initials
}

func main() {
	// Read departments file
	deptFile, err := os.Open("/home/otavio/me-livra-scraping/files/ufg/departments.json")
	if err != nil {
		log.Fatalf("Failed to open departments file: %v", err)
	}
	defer deptFile.Close()

	deptBytes, err := ioutil.ReadAll(deptFile)
	if err != nil {
		log.Fatalf("Failed to read departments file: %v", err)
	}

	var departments []Department
	if err := json.Unmarshal(deptBytes, &departments); err != nil {
		log.Fatalf("Failed to unmarshal departments: %v", err)
	}

	// Read teachers file
	teachersFile, err := os.Open("/home/otavio/me-livra-scraping/files/ufg/teachers.json")
	if err != nil {
		log.Fatalf("Failed to open teachers file: %v", err)
	}
	defer teachersFile.Close()

	teachersBytes, err := ioutil.ReadAll(teachersFile)
	if err != nil {
		log.Fatalf("Failed to read teachers file: %v", err)
	}

	var teachersData []TeacherEntry
	if err := json.Unmarshal(teachersBytes, &teachersData); err != nil {
		log.Fatalf("Failed to unmarshal teachers: %v", err)
	}

	// Create a mapping of normalized department names to their official names
	departmentMap := make(map[string]string)
	for _, dept := range departments {
		normalizedName := NormalizeName(dept.Name)
		departmentMap[normalizedName] = normalizedName
	}

	// Connect to the database
	host := ""
	port := 5432
	user := ""
	password := ""
	dbname := ""

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Check the connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Load existing institutes from the database
	instRows, err := db.Query("SELECT id, name, university, initials FROM institutes")
	if err != nil {
		log.Fatalf("Failed to query existing institutes: %v", err)
	}
	defer instRows.Close()

	existingInstitutes := make(map[string]Institute)
	for instRows.Next() {
		var inst Institute
		if err := instRows.Scan(&inst.ID, &inst.Name, &inst.University, &inst.Initials); err != nil {
			log.Fatalf("Failed to scan institute row: %v", err)
		}
		existingInstitutes[NormalizeName(inst.Name)] = inst

		// Add existing institutes to the department map for matching
		departmentMap[NormalizeName(inst.Name)] = inst.Name
	}
	fmt.Printf("Loaded %d existing institutes from database\n", len(existingInstitutes))

	originalCountInstitutes := len(existingInstitutes)

	// Load existing professors from database to check for duplicates
	rows, err := db.Query("SELECT id, name, institute_id FROM professors")
	if err != nil {
		log.Fatalf("Failed to query existing professors: %v", err)
	}
	defer rows.Close()

	var existingProfessors []Professor
	for rows.Next() {
		var prof Professor
		if err := rows.Scan(&prof.ID, &prof.Name, &prof.InstituteID); err != nil {
			log.Fatalf("Failed to scan professor row: %v", err)
		}
		existingProfessors = append(existingProfessors, prof)
	}
	fmt.Printf("Loaded %d existing professors from database\n", len(existingProfessors))

	originalCountProfessors := len(existingProfessors)

	fmt.Println("Starting import process...")
	fmt.Println("Creating or finding institutes and professors...")

	// Statistics for reporting
	stats := struct {
		skippedDepartments   int
		emptyDepartments     int
		processedDepartments int
		addedProfessors      int
	}{}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}

	// Process each department and its teachers
	for _, entry := range teachersData {
		// Skip departments with no teachers
		if len(entry.Teachers) == 0 {
			fmt.Printf("Skipping department '%s' as it has no teachers.\n", entry.Department)
			stats.emptyDepartments++
			continue
		}

		// Count valid teachers (those with non-empty names)
		validTeacherCount := 0
		for _, teacher := range entry.Teachers {
			if strings.TrimSpace(teacher.Name) != "" {
				validTeacherCount++
			}
		}

		// Skip if there are no valid teachers
		if validTeacherCount == 0 {
			fmt.Printf("Skipping department '%s' as it has no valid teachers (only empty names).\n", entry.Department)
			stats.emptyDepartments++
			continue
		}

		stats.processedDepartments++

		departmentName := NormalizeName(entry.Department)

		// Try to find the official department name using fuzzy matching
		officialName, found := IsDuplicateDepartment(departmentName, departmentMap, 0.9)

		// If no match found, use the original name
		if !found {
			officialName = departmentName
			fmt.Printf("No matching department found for '%s'. Using original name.\n", departmentName)

			// Handle special case for UNID.ACAD.ESP/CIENC SOCIAIS APLIC
			if strings.Contains(departmentName, "UNIDACADESPCIENC SOCIAIS APLIC") {
				officialName = "UNID.ACAD.ESP/CIENC SOCIAIS APLIC" // Use the official name from departments.json
				fmt.Printf("Special case: Using official name '%s' for '%s'\n", officialName, departmentName)
			}
		} else {
			fmt.Printf("Matched department '%s' to official name '%s'\n", departmentName, officialName)
		}

		// Check if institute already exists in our loaded data
		var instituteID int
		normalizedOfficialName := NormalizeName(officialName)

		// Format the department name for display with proper Portuguese formatting
		formattedName := FormatPortugueseText(officialName)

		if existingInst, exists := existingInstitutes[normalizedOfficialName]; exists {
			instituteID = existingInst.ID
			fmt.Printf("Found existing institute: %s (ID: %d)\n", existingInst.Name, existingInst.ID)
		} else {
			// Institute doesn't exist, create it
			err = tx.QueryRow(`
				INSERT INTO institutes (name, university, initials, avg_grade, created_at, updated_at)
				VALUES ($1, $2, $3, $4, NOW(), NOW())
				ON CONFLICT (name) DO UPDATE SET updated_at = NOW()
				RETURNING id
			`, formattedName, "UFG", generateInitials(formattedName), 0).Scan(&instituteID)

			if err != nil {
				tx.Rollback()
				log.Fatalf("Failed to insert/update institute: %v", err)
			}

			// Add to our in-memory map
			existingInstitutes[normalizedOfficialName] = Institute{
				ID:         instituteID,
				Name:       formattedName,
				University: "UFG",
				Initials:   generateInitials(formattedName),
			}

			fmt.Printf("Created new institute: %s (ID: %d)\n", formattedName, instituteID)
		}

		if err != nil {
			tx.Rollback()
			log.Fatalf("Failed to insert/update institute: %v", err)
		}

		fmt.Printf("Processing institute: %s (ID: %d)\n", officialName, instituteID)

		// Process each teacher in this department
		for _, teacher := range entry.Teachers {
			professorName := NormalizeName(teacher.Name)
			if professorName == "" {
				continue // Skip empty names
			}

			// Format the professor name with proper Portuguese formatting
			formattedProfName := FormatPortugueseText(professorName)

			// Create a temporary professor object for duplicate checking
			tempProf := Professor{
				Name:        professorName,
				InstituteID: instituteID,
			}

			// Check if professor already exists using fuzzy matching
			if IsDuplicate(tempProf, existingProfessors, 0.9) {
				fmt.Printf("Skipping duplicate professor: %s\n", formattedProfName)
				continue
			}

			// Find or create professor
			var professorID int
			err = tx.QueryRow(`
                INSERT INTO professors (name, institute_id, created_at, updated_at)
                VALUES ($1, $2, NOW(), NOW())
                ON CONFLICT (name, institute_id) DO UPDATE SET updated_at = NOW()
                RETURNING id
            `, formattedProfName, instituteID).Scan(&professorID)

			if err != nil {
				tx.Rollback()
				log.Fatalf("Failed to insert/update professor %s: %v", formattedProfName, err)
			}

			fmt.Printf("Processed professor: %s (ID: %d)\n", formattedProfName, professorID)

			// Increment the counter for added professors
			stats.addedProfessors++

			// Add to existing professors list to avoid duplicates in this run
			existingProfessors = append(existingProfessors, Professor{
				ID:          professorID,
				Name:        formattedProfName,
				InstituteID: instituteID,
			})
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	// Generate report of the import
	fmt.Println("\n============= IMPORT SUMMARY =============")

	// Count institutes before and after
	var instituteCount int
	err = db.QueryRow("SELECT COUNT(*) FROM institutes").Scan(&instituteCount)
	if err != nil {
		log.Printf("Failed to count institutes: %v", err)
	} else {
		fmt.Printf("Institutes in database: %d (started with %d)\n",
			instituteCount, originalCountInstitutes)
		fmt.Printf("Institutes added: %d\n", instituteCount-originalCountInstitutes)
	}

	// Count professors before and after
	var professorCount int
	err = db.QueryRow("SELECT COUNT(*) FROM professors").Scan(&professorCount)
	if err != nil {
		log.Printf("Failed to count professors: %v", err)
	} else {
		fmt.Printf("Professors in database: %d (started with %d)\n",
			professorCount, originalCountProfessors)
		fmt.Printf("Professors added: %d\n", professorCount-originalCountProfessors)
	}

	// Detailed statistics
	fmt.Println("\n----- Processing Statistics -----")
	fmt.Printf("Departments processed: %d\n", stats.processedDepartments)
	fmt.Printf("Departments skipped (empty): %d\n", stats.emptyDepartments)
	fmt.Printf("Total departments in input: %d\n", stats.processedDepartments+stats.emptyDepartments)
	fmt.Printf("Professors added in this run: %d\n", stats.addedProfessors)

	fmt.Println("Import completed successfully!")
}

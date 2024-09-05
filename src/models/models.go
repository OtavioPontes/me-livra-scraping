package models

type Department struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Teacher struct {
	Name string `json:"name"`
}

type UfmgTeacher struct {
	Name      string `json:"name"`
	Institute string `json:"institute"`
}

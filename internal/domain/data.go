package domain

import "time"

type Projects map[string][]Project

type Project struct {
	ClientName   string   `json:"client_name"`
	Technologies []string `json:"technologies"`
	Description  string   `json:"description"`
}

type Link struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	External bool   `json:"external"`
}

type Experience struct {
	Name        string   `json:"name"`
	Slug        string   `json:"slug"`
	Role        string   `json:"role"`
	Description string   `json:"description"`
	Start       string   `json:"start"`
	End         string   `json:"end"`
	Bullets     []string `json:"bullets"`
}

type OtherExperience struct {
	Name  string `json:"name"`
	Role  string `json:"role"`
	Start string `json:"start"`
	End   string `json:"end"`
}

type Education struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Degree      string `json:"degree"`
	Graduation  string `json:"graduation"`
}

type Publication struct {
	Date         time.Time `json:"date"`
	Year         string    `json:"year"`
	Title        string    `json:"title"`
	URL          string    `json:"url"`
	Authors      string    `json:"authors"`
	Contribution string    `json:"contribution"`
	Partnership  string    `json:"partnership"`
	Publisher    string    `json:"publisher"`
	Issue        string    `json:"issue"`
}

type Competency struct {
	Name  string   `json:"name"`
	Items []string `json:"items"`
}

type Resume struct {
	Name            string            `json:"name"`
	Title           string            `json:"title"`
	Email           string            `json:"email"`
	Phone           string            `json:"phone"`
	Location        string            `json:"location"`
	Links           []Link            `json:"links"`
	Intro           string            `json:"intro"`
	Competencies    []Competency      `json:"competencies"`
	Experience      []Experience      `json:"experience"`
	OtherExperience []OtherExperience `json:"other_experience"`
	Education       []Education       `json:"education"`
	Publications    []Publication     `json:"publications"`
	Interests       []string          `json:"interests"`
}

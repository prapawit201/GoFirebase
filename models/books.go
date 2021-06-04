package models

type Books struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Author string `json:"author"`
	Status bool   `json:"status"`
}

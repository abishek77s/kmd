package starter

import "backend/item"

type App struct {
	Name        string      `json:"name"`
	Logo        string      `json:"logo"`
	Description string      `json:"description"`
	StarterPack StarterPack `json:"starterPack"`
}

type Templates struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	AppList     []App    `json:"appList"`
	Tags        []string `json:"tags"`
}

type StarterPack struct {
	Name     string       `json:"name"`
	Commands item.Command `json:"commands"`
}

package dto

type Selection struct {
	Path    string
	Version int
}

type Config struct {
	Path  string
	Value string
}

type QueryInput struct {
	Query string `json:"query"`
}

type GetCurrentInput struct {
	Path string `json:"path"`
}

type GetInput struct {
	Path    string `json:"path"`
	Version int    `json:"version"`
}

type SetInput struct {
	Path        string `json:"path"`
	Value       string `json:"value"`
	KeepCurrent bool   `json:"keepCurrent"`
}

type SetManyInput struct {
	Path        string   `json:"path"`
	Values      []string `json:"values"`
	KeepCurrent bool     `json:"keepCurrent"`
}

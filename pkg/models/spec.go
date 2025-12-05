package models

type Spec struct {
	Image      string `json:"image" yaml:"image"`
	Entrypoint string `json:"entrypoint" yaml:"entrypoint"`
	FileName   string `json:"filename" yaml:"filename"`
	Cmd        string `json:"cmd" yaml:"cmd"`
	Language   string `json:"language" yaml:"language"`
	Use        string `json:"use" yaml:"use"`
}

type SpecMap map[string]*Spec

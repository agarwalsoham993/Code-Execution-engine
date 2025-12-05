package spec

import (
	"code-runner/pkg/models"
	"github.com/ghodss/yaml"
	"os"
)

type BaseProvider struct {
	m models.SpecMap
}

func NewFileProvider(path string) *BaseProvider {
	data, _ := os.ReadFile(path)
	m := make(models.SpecMap)
	yaml.Unmarshal(data, &m)
	return &BaseProvider{m: m}
}

func (p *BaseProvider) Spec() models.SpecMap { return p.m }

func (p *BaseProvider) Get(key string) (models.Spec, bool) {
	if s, ok := p.m[key]; ok {
		if s.Use != "" { return p.Get(s.Use) }
		return *s, true
	}
	return models.Spec{}, false
}

package sandbox

import (
	"code-runner/pkg/models"
	"strings"
	"path"
	"regexp"
)

type Sandbox interface {
	ID() string
	Run(stdout, stderr chan []byte, close chan bool) error
	Kill() error
	Delete() error
}

type Provider interface {
	Prepare(spec models.Spec) error
	CreateSandbox(spec RunSpec) (Sandbox, error)
}

type RunSpec struct {
	models.Spec
	Arguments   []string
	Environment map[string]string
	Subdir      string
	HostDir     string
}

func (s RunSpec) GetAssembledHostDir() string { return path.Join(s.HostDir, s.Subdir) }
func (s RunSpec) GetEntrypoint() []string { return splitArgs(s.Entrypoint) }
func (s RunSpec) GetCommandWithArgs() []string { return append(splitArgs(s.Cmd), s.Arguments...) }
func (s RunSpec) GetEnv() []string {
	env := []string{"RUNNER_HOSTDIR=" + s.HostDir}
	for k, v := range s.Environment { env = append(env, k+"="+v) }
	return env
}

var argRx = regexp.MustCompile(`(?:[^\s"]+|"[^"]*")+`)
func splitArgs(v string) []string {
	res := argRx.FindAllString(v, -1)
	for i, v := range res { res[i] = strings.Replace(v, "\"", "", -1) }
	return res
}

package docker

import (
	"bytes"
	"code-runner/internal/config"
	"code-runner/internal/sandbox"
	"code-runner/pkg/models"
	"fmt"
	dockerclient "github.com/fsouza/go-dockerclient"
	"github.com/rs/xid"
	"github.com/zekrotja/rogu/log"
	"path"
	"path/filepath"
	"strings"
)

type Provider struct {
	cfg    *config.EnvProvider
	client *dockerclient.Client
}

func NewProvider(cfg *config.EnvProvider) (*Provider, error) {
	client, err := dockerclient.NewClientFromEnv()
	if err != nil {
		return nil, err
	}
	return &Provider{cfg: cfg, client: client}, nil
}

func (p *Provider) Prepare(spec models.Spec) error {
	repo, tag := parseImage(spec.Image)
	_, err := p.client.InspectImage(repo + ":" + tag)
	if err == dockerclient.ErrNoSuchImage {
		log.Info().Fields("image", spec.Image).Msg("Pulling image...")
		return p.client.PullImage(dockerclient.PullImageOptions{
			Repository: repo, Tag: tag,
		}, dockerclient.AuthConfiguration{})
	}
	return err
}

func (p *Provider) CreateSandbox(spec sandbox.RunSpec) (sandbox.Sandbox, error) {
	if err := p.Prepare(spec.Spec); err != nil {
		return nil, err
	}

	workingDir := path.Join("/var/tmp/exec", spec.Subdir)
	hostDir, _ := filepath.Abs(spec.GetAssembledHostDir())

	container, err := p.client.CreateContainer(dockerclient.CreateContainerOptions{
		Name: fmt.Sprintf("runner-%s-%s", spec.Language, xid.New().String()),
		Config: &dockerclient.Config{
			Image:        spec.Image,
			WorkingDir:   workingDir,
			Entrypoint:   spec.GetEntrypoint(),
			Cmd:          spec.GetCommandWithArgs(),
			Env:          spec.GetEnv(),
			OpenStdin:    true, // Enable Stdin
			StdinOnce:    true,
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
		},
		HostConfig: &dockerclient.HostConfig{
			Binds: []string{hostDir + ":" + workingDir},
		},
	})

	if err != nil {
		return nil, err
	}

	return &Sandbox{client: p.client, container: container}, nil
}

type Sandbox struct {
	client    *dockerclient.Client
	container *dockerclient.Container
}

func (s *Sandbox) ID() string { return s.container.ID }

func (s *Sandbox) Run(stdin []byte, stdout, stderr chan []byte, close chan bool) error {
	go func() {
		opts := dockerclient.AttachToContainerOptions{
			Container:    s.container.ID,
			OutputStream: &ChanWriter{stdout},
			ErrorStream:  &ChanWriter{stderr},
			Stdout:       true, Stderr: true, Stream: true,
		}
		
		// If Stdin provided, attach it
		if len(stdin) > 0 {
			opts.InputStream = bytes.NewBuffer(stdin)
			opts.Stdin = true
		}

		s.client.AttachToContainer(opts)
		close <- true
	}()
	return s.client.StartContainer(s.container.ID, nil)
}

func (s *Sandbox) Kill() error   { return s.client.KillContainer(dockerclient.KillContainerOptions{ID: s.container.ID}) }
func (s *Sandbox) Delete() error { return s.client.RemoveContainer(dockerclient.RemoveContainerOptions{ID: s.container.ID}) }

type ChanWriter struct{ C chan []byte }
func (w *ChanWriter) Write(p []byte) (int, error) {
	cp := make([]byte, len(p))
	copy(cp, p)
	w.C <- cp
	return len(p), nil
}

func parseImage(img string) (string, string) {
	split := strings.SplitN(img, ":", 2)
	if len(split) == 1 {
		return split[0], "latest"
	}
	return split[0], split[1]
}
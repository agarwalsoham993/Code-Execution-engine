package sandbox

import (
	"code-runner/internal/config"
	"code-runner/internal/file"
	"code-runner/internal/spec"
	"code-runner/pkg/models"
	"errors"
	"fmt"
	"github.com/rs/xid"
	"github.com/zekrotja/rogu/log"
	"path"
	"sync"
	"time"
)

type Manager struct {
	sandbox Provider
	spec    *spec.BaseProvider
	file    *file.LocalFileProvider
	cfg     *config.EnvProvider
	running sync.Map
}

func NewManager(sb Provider, sp *spec.BaseProvider, fp *file.LocalFileProvider, cfg *config.EnvProvider) (*Manager, error) {
	return &Manager{sandbox: sb, spec: sp, file: fp, cfg: cfg}, nil
}

func (m *Manager) RunInSandbox(submissionID string, req *models.ExecutionRequest, cout, cerr chan []byte, cstop chan bool) error {
	spc, ok := m.spec.Get(req.Language)
	if !ok {
		log.Error().Field("language", req.Language).Msg("Unsupported language specification")
		return fmt.Errorf("unsupported language: %s", req.Language)
	}

	// Use provided ID or generate one
	runId := submissionID
	if runId == "" {
		runId = xid.New().String()
	}

	log.Debug().Field("RunID", runId).Field("Language", req.Language).Msg("Starting code execution job")

	runSpc := RunSpec{
		Spec:        spc,
		Subdir:      runId,
		HostDir:     m.cfg.Config().HostRootDir,
		Arguments:   req.Arguments,
		Environment: req.Environment,
	}

	if runSpc.Cmd == "" {
		runSpc.Cmd = spc.FileName
	}

	hostDir := runSpc.GetAssembledHostDir()

	m.file.CreateDirectory(hostDir)
	m.file.CreateFile(path.Join(hostDir, spc.FileName), req.Code)
	
	// Ensure cleanup happens even if container creation fails
	defer func() {
		m.file.DeleteDirectory(hostDir)
	}()

	sbx, err := m.sandbox.CreateSandbox(runSpc)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create Docker sandbox")
		return fmt.Errorf("failed to create sandbox: %w", err)
	}
	
	// Ensure container is cleaned up
	defer func() {
		sbx.Kill()
		sbx.Delete()
		m.running.Delete(sbx.ID())
	}()

	log.Info().Field("ContainerID", sbx.ID()).Msg("Docker container created and started")
	m.running.Store(sbx.ID(), sbx)

	finished := make(chan bool)
	go func() {
		err := sbx.Run(cout, cerr, finished)
		if err != nil {
			log.Error().Err(err).Field("ContainerID", sbx.ID()).Msg("Sandbox run failed during execution")
		}
	}()

	timedOut := false
	select {
	case <-finished:
		log.Debug().Field("ContainerID", sbx.ID()).Msg("Sandbox finished execution")
	case <-time.After(time.Duration(m.cfg.Config().Sandbox.TimeoutSeconds) * time.Second):
		log.Warn().Field("ContainerID", sbx.ID()).Msg("Sandbox timed out.")
		timedOut = true
	}

	cstop <- true // signal to stop collection

	if timedOut {
		return errors.New("execution timed out")
	}

	return nil
}

func (m *Manager) Cleanup() {
	m.running.Range(func(key, value interface{}) bool {
		log.Info().Field("ContainerID", value.(Sandbox).ID()).Msg("Cleaning up container during application shutdown")
		value.(Sandbox).Kill()
		value.(Sandbox).Delete()
		return true
	})
}

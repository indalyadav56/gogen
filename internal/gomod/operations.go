package gomod

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/indalyadav56/gogen/internal/logx"
)

// Manager handles Go module operations
type Manager struct {
	projectRoot string
}

// NewManager creates a new Go module manager
func NewManager(projectRoot string) *Manager {
	return &Manager{
		projectRoot: projectRoot,
	}
}

// Init initializes a new Go module
func (m *Manager) Init(moduleName string) error {
	logx.S().Debugw("go mod init", "module", moduleName, "dir", m.projectRoot)
	cmd := exec.Command("go", "mod", "init", moduleName)
	cmd.Dir = m.projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go mod init failed: %w", err)
	}

	return nil
}

// Tidy runs go mod tidy to clean up dependencies
func (m *Manager) Tidy() error {
	logx.S().Debugw("go mod tidy", "dir", m.projectRoot)
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = m.projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go mod tidy failed: %w", err)
	}

	return nil
}

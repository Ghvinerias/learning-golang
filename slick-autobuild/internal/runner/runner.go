package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"slick-autobuild/internal/logging"
	"slick-autobuild/internal/planner"
)

// Options configures task execution.
type Options struct {
	Logger        *logging.Logger
	WorkspaceRoot string
}

// RunTask executes the given task using Docker to ensure toolchain isolation.
// MVP: minimal commands, no caching yet.
func RunTask(ctx context.Context, task planner.Task, opts Options, pkgManager string, buildScripts []string) error {
	if opts.Logger == nil {
		opts.Logger = logging.New(false)
	}
	workDir := filepath.Join(opts.WorkspaceRoot, task.Path)
	if _, err := os.Stat(workDir); err != nil {
		return fmt.Errorf("task path missing: %s: %w", task.Path, err)
	}

	image, command := dockerSpec(task, pkgManager, buildScripts)
	opts.Logger.Debug("docker run spec", map[string]interface{}{"image": image, "cmd": command})

	args := []string{
		"run", "--rm",
		"-v", fmt.Sprintf("%s:/workspace", opts.WorkspaceRoot),
		"-w", filepath.ToSlash(filepath.Join("/workspace", task.Path)),
		image,
		"bash", "-lc", command,
	}
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker build failed: %w", err)
	}
	return nil
}

func dockerSpec(task planner.Task, pkgManager string, buildScripts []string) (image string, command string) {
	switch task.Kind {
	case "dotnet":
		image = "mcr.microsoft.com/dotnet/sdk:" + task.Version
		// Basic restore + build
		command = "dotnet restore && dotnet build -c Release"
	case "node":
		image = "node:" + task.Version
		if pkgManager == "" {
			pkgManager = "npm"
		}
		if len(buildScripts) == 0 {
			buildScripts = []string{"build"}
		}
		// Single build script only (first) for MVP
		buildCmd := buildScripts[0]
		switch pkgManager {
		case "pnpm":
			command = fmt.Sprintf("corepack enable && pnpm install --frozen-lockfile || pnpm install && pnpm run %s", buildCmd)
		case "yarn":
			command = fmt.Sprintf("corepack enable && yarn install --frozen-lockfile || yarn install && yarn run %s", buildCmd)
		default:
			command = fmt.Sprintf("npm install && npm run %s", buildCmd)
		}
	default:
		image = "alpine:latest"
		command = "echo unsupported task kind"
	}
	return
}

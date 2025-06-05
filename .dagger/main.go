package main

import (
	"context"
	"fmt"

	"github.com/tuannvm/haproxy-mcp-server/.dagger/internal/dagger"
)

type HaproxyMcpServer struct {
	// +private
	Source *dagger.Directory
	// +private
	Repo string
	// +private
	Image string
	// +private
	App string
	// +private
	Backend *dagger.Backend
}

func New(
	// +optional
	// +defaultPath="/"
	// +ignore=[".git", "**/node_modules"]
	source *dagger.Directory,
	// +optional
	// +default="github.com/tuannvm/haproxy-mcp-server"
	repo string,
	// +optional
	// +default="kylepenfound/greetings-api:latest"
	image string,
	// +optional
	// +default="dagger-demo"
	app string,
) *HaproxyMcpServer {
	g := &HaproxyMcpServer{
		Source:  source,
		Repo:    repo,
		Image:   image,
		App:     app,
		Backend: dag.Backend(source), // Include all directories including internal
	}
	return g
}

// Run the CI Checks for the project
func (g *HaproxyMcpServer) Check(
	ctx context.Context,
	// Github token with permissions to comment on the pull request
	// +optional
	githubToken *dagger.Secret,
	// git commit in github
	// +optional
	commit string,
	// The model to use to debug debug tests
	// +optional
	model string,
) (string, error) {
	// Lint
	lintOut, err := g.Lint(ctx)
	if err != nil {
		if githubToken != nil {
			debugErr := g.DebugBrokenTestsPr(ctx, githubToken, commit, model)
			return "", fmt.Errorf("lint failed, attempting to debug %v %v", err, debugErr)
		}
		return "", err
	}

	// Then Build
	_, err = g.Build().Sync(ctx)
	if err != nil {
		return "", err
	}

	return lintOut + "\n\n", nil
}

// Lint the Go code in the project
func (g *HaproxyMcpServer) Lint(ctx context.Context) (string, error) {
	return g.Backend.Lint(ctx)
}

// Build the backend binary
func (g *HaproxyMcpServer) Build() *dagger.Directory {
	// The backend.Build() function already returns a directory with the binary in bin/haproxy-mcp-server
	return g.Backend.Build()
}

// Serve the backend on port 8080
func (g *HaproxyMcpServer) Serve() *dagger.Service {
	return g.Backend.Serve()
}

// Create a GitHub release
func (g *HaproxyMcpServer) Release(ctx context.Context, tag string, ghToken *dagger.Secret) (string, error) {
	// Get build
	build := g.Build()

	// Get the binary from the build directory
	binary := build.File("bin/haproxy-mcp-server")

	// Create a directory with the binary at the root for the release
	releaseDir := dag.Directory().WithFile("haproxy-mcp-server", binary)

	title := fmt.Sprintf("Release %s", tag)
	return dag.GithubRelease().Create(ctx, g.Repo, tag, title, ghToken, dagger.GithubReleaseCreateOpts{
		Assets: releaseDir,
	})
}

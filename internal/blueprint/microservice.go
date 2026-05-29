package blueprint

import "github.com/indalyadav56/gogen/internal/cli"

// buildMicroservice reuses the clean layout and adds deployment-oriented
// extras: a graceful-shutdown main, a /healthz handler, a minimal gRPC server,
// and docker-compose. Building on top of buildClean keeps the two in sync.
func buildMicroservice(cfg *cli.Config) []FileSpec {
	imp := cleanImports(cfg.ModuleName)
	files := buildClean(cfg)

	for i := range files {
		if files[i].Path == "cmd/server/main.go" {
			files[i].Template = "main_microservice.tmpl"
		}
	}

	return append(files,
		FileSpec{Path: "internal/health/health.go", Template: "health.tmpl", Package: "health"},
		FileSpec{Path: "internal/transport/grpc/server.go", Template: "grpc_server.tmpl", Package: "grpcserver", Imports: imp},
		FileSpec{Path: "docker-compose.yml", Template: "docker_compose.tmpl"},
		FileSpec{Path: "proto/.gitkeep", Template: ""},
	)
}

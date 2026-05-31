package blueprint

import "github.com/indalyadav56/gogen/internal/template"

// authImports returns the import paths for the self-contained internal/auth
// module. Auth is always laid out DDD-style (interfaces + adapters) regardless
// of the host architecture, so the same set of templates works everywhere.
func authImports(mod string) template.Imports {
	base := mod + "/internal/auth"
	return template.Imports{
		Entity:     base + "/domain/entity",
		Repository: base + "/domain/repository",
		Service:    base + "/application",
		Infra:      base + "/infrastructure/postgres",
		Handler:    base + "/transport/http/v1/handlers",
		Routes:     base + "/transport/http/v1/routes",
		DTO:        base + "/transport/http/v1/dto",
		Middleware: base + "/transport/http/middleware",
	}
}

// authFiles returns the JWT + RBAC auth module: reusable pkg/auth helpers, a
// self-contained internal/auth bounded context, and the auth migration (placed
// in migrationsDir, e.g. "migrations" or "migrations/postgres").
func authFiles(mod, migrationsDir string) []FileSpec {
	imp := authImports(mod)
	a := "internal/auth"
	return []FileSpec{
		{Path: "pkg/auth/jwt.go", Template: "pkg_auth_jwt.tmpl", Package: "auth"},
		{Path: "pkg/auth/password.go", Template: "pkg_auth_password.tmpl", Package: "auth"},

		{Path: a + "/domain/entity/user.go", Template: "auth_user_entity.tmpl", Package: "entity", Imports: imp},
		{Path: a + "/domain/entity/role.go", Template: "auth_role_entity.tmpl", Package: "entity", Imports: imp},
		{Path: a + "/domain/repository/user_repository.go", Template: "auth_user_repository.tmpl", Package: "repository", Imports: imp},
		{Path: a + "/domain/repository/role_repository.go", Template: "auth_role_repository.tmpl", Package: "repository", Imports: imp},
		{Path: a + "/application/auth_service.go", Template: "auth_service.tmpl", Package: "application", Imports: imp},
		{Path: a + "/infrastructure/postgres/user_postgres.go", Template: "auth_user_postgres.tmpl", Package: "postgres", Imports: imp},
		{Path: a + "/infrastructure/postgres/role_postgres.go", Template: "auth_role_postgres.tmpl", Package: "postgres", Imports: imp},
		{Path: a + "/transport/http/v1/dto/auth_dto.go", Template: "auth_dto.tmpl", Package: "dto", Imports: imp},
		{Path: a + "/transport/http/v1/handlers/auth_handler.go", Template: "auth_handler.tmpl", Package: "handlers", Imports: imp},
		{Path: a + "/transport/http/v1/routes/auth_routes.go", Template: "auth_routes.tmpl", Package: "routes", Imports: imp},
		{Path: a + "/transport/http/middleware/auth_middleware.go", Template: "auth_middleware.tmpl", Package: "middleware", Imports: imp},

		{Path: migrationsDir + "/000002_create_auth_tables.up.sql", Template: "auth_migration.tmpl"},
	}
}

package blueprint

// frontendFiles returns the files for the chosen frontend layer. The Go app
// serves the frontend itself (single fullstack binary): html/htmx are embedded
// static assets (+ an htmx fragment endpoint); react is a Vite app whose built
// dist/ is embedded and served with SPA fallback.
func frontendFiles(frontend string) []FileSpec {
	switch frontend {
	case "html", "htmx":
		return []FileSpec{
			{Path: "web/web.go", Template: "frontend_web_html.tmpl", Package: "web"},
			{Path: "web/index.html", Template: "frontend_index_html.tmpl"},
			{Path: "web/static/styles.css", Template: "frontend_styles.tmpl"},
		}
	case "react":
		return []FileSpec{
			{Path: "web/web.go", Template: "frontend_web_react.tmpl", Package: "web"},
			{Path: "web/package.json", Template: "frontend_package_json.tmpl"},
			{Path: "web/vite.config.js", Template: "frontend_vite_config.tmpl"},
			{Path: "web/index.html", Template: "frontend_react_index.tmpl"},
			{Path: "web/.gitignore", Template: "frontend_react_gitignore.tmpl"},
			{Path: "web/src/main.jsx", Template: "frontend_main_jsx.tmpl"},
			{Path: "web/src/App.jsx", Template: "frontend_app_jsx.tmpl"},
			{Path: "web/src/api.js", Template: "frontend_api_js.tmpl"},
			{Path: "web/dist/index.html", Template: "frontend_dist_index.tmpl"},
		}
	}
	return nil
}

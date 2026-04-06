package handler

import (
	"net/http"

	apidocs "github.com/go-park-mail-ru/2026_1_SPORT.tech/docs"
)

const swaggerUIPage = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>SPORT.tech API Docs</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
    <style>
      html { box-sizing: border-box; overflow-y: scroll; }
      *, *::before, *::after { box-sizing: inherit; }
      body { margin: 0; background: #fafafa; }
    </style>
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
      window.onload = function () {
        window.ui = SwaggerUIBundle({
          url: "/docs/openapi.yml",
          dom_id: "#swagger-ui",
          deepLinking: true,
          presets: [SwaggerUIBundle.presets.apis],
          layout: "BaseLayout",
        });
      };
    </script>
  </body>
</html>
`

func (handler *Handler) handleGetDocs(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = writer.Write([]byte(swaggerUIPage))
}

func (handler *Handler) handleGetDocsRedirect(writer http.ResponseWriter, request *http.Request) {
	http.Redirect(writer, request, "/docs/", http.StatusMovedPermanently)
}

func (handler *Handler) handleGetOpenAPISpec(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/yaml")
	_, _ = writer.Write(apidocs.OpenAPIYAML)
}

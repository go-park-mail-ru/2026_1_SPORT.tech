package httpgateway

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type DocsSpec struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

const swaggerUIDocsPage = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>SPORT.tech API Docs</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
    <style>
      html { box-sizing: border-box; overflow-y: scroll; }
      *, *::before, *::after { box-sizing: inherit; }
      body { margin: 0; background: #f5f7fb; }
      .page-header {
        padding: 16px 24px;
        border-bottom: 1px solid #dbe2ea;
        background: #fff;
        font-family: sans-serif;
      }
      .page-header h1 {
        margin: 0 0 4px;
        font-size: 20px;
      }
      .page-header p {
        margin: 0;
        color: #52606d;
      }
    </style>
  </head>
  <body>
    <div class="page-header">
      <h1>SPORT.tech API Gateway Docs</h1>
      <p>Generated from gRPC protobuf contracts via grpc-gateway and OpenAPI.</p>
    </div>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
      const specs = %s;
      window.onload = function () {
        window.ui = SwaggerUIBundle({
          urls: specs,
          "urls.primaryName": specs[0] ? specs[0].name : undefined,
          dom_id: "#swagger-ui",
          deepLinking: true,
          presets: [SwaggerUIBundle.presets.apis],
          layout: "BaseLayout"
        });
      };
    </script>
  </body>
</html>
`

func DocsHandler(specs []DocsSpec) http.Handler {
	specsJSON, err := json.Marshal(specs)
	if err != nil {
		specsJSON = []byte("[]")
	}

	page := fmt.Sprintf(swaggerUIDocsPage, specsJSON)

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = writer.Write([]byte(page))
	})
}

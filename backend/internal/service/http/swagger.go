package http

import (
	"encoding/json"
	"net/http"

	"github.com/durianpay/fullstack-boilerplate/internal/openapigen"
)

const swaggerHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <title>Payment Dashboard API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: '/openapi.json',
      dom_id: '#swagger-ui',
      presets: [SwaggerUIBundle.presets.apis],
      layout: 'BaseLayout',
      docExpansion: 'list',
      tryItOutEnabled: true,
      persistAuthorization: true,
    });
  </script>
</body>
</html>`

func openapiJSONHandler() http.HandlerFunc {
	swagger, err := openapigen.GetSwagger()
	if err != nil {
		return func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "spec unavailable", http.StatusInternalServerError)
		}
	}
	swagger.Servers = nil
	body, err := json.Marshal(swagger)
	if err != nil {
		return func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "spec marshal failed", http.StatusInternalServerError)
		}
	}
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=300")
		_, _ = w.Write(body)
	}
}

func swaggerUIHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(swaggerHTML))
	}
}

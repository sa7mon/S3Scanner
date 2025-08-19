---
weight: 100
---

# Development

A docker compose file is included which creates 4 containers:

* rabbitmq
* postgres
* app
* mitm

2 profiles are configured:

- `dev` - Standard development environment
- `dev-mitm` - Environment configured with `mitmproxy` for easier observation of HTTP traffic when debugging or adding new providers.

To bring up the dev environment run `make dev` or `make dev-mitm`. Drop into the `app` container with `docker exec -it -w /app app_dev sh`, then `go run .`
If using the `dev-mitm` profile, open `http://127.0.0.1:8081` in a browser to view and manipulate HTTP calls being made from the app container. This can be useful for analyzing HTTP responses when adding support for a new provider.

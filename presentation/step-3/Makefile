.PHONY: server
server:
	@source ./source.me && WH_URL=$(WH_URL) go run cmd/server.go

.PHONY: watch
watch:
	find . -name "*.go" -or -name "*.tmpl" | entr -r make server

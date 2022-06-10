
server:
	source ./source.me && go run cmd/server.go

watch:
	find . -name "*.go" -or -name "*.tmpl" | entr -r make server

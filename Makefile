
.PHONY: server
server:
	source ./source.me && go run cmd/server.go

.PHONY: test
test:
	go test goose

.PHONY: watch
watch:
	find . -name "*.go" -or -name "*.tmpl" | entr -r make server

.PHONY: build-image
build-image:
	podman build -t goose .

.PHONY: run-image
run-image:
	podman run --rm -it -e API_KEY=$(API_KEY) \
		-e MAIL_SMTP_PWD=$(MAIL_SMTP_PWD) \
		-e COOKIE_SECRET=$(COOKIE_SECRET) \
		-e WH_URL=$(WH_URL) \
		-p 127.0.0.1:8077:8077 goose

.PHONY: seed
seed:
	echo "Populate the database with examples"
	@gunzip db/producation.db.gz

.PHONY: unpackjs
unpackjs:
	@npm pack eyeson
		@tar -xf eyeson-1.*.tgz package/dist/eyeson.js
		@mv package/dist/eyeson.js assets/
		@rm -r eyeson-1.*.tgz package/

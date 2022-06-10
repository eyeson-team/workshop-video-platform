
.PHONY: watch
watch:
	ls *.md | entr -s 'make build'

.PHONY: build
build:
	@pandoc -t beamer -o slides.pdf slides.md
		@pandoc -o notes.pdf notes.md


GOMAINS = make.go

%.bin %.go: $(GOMAINS)
	go build -o $@ $<

default: all-offline

all-offline:
	go run make.go all --offline
.PHONY: all
all: make.bin
	./make.bin all
.PHONY: bindata
bindata: make.bin
	./make.bin bindata
.PHONY: bindata-dev
bindata-dev: make.bin
	./make.bin bindata-dev
.PHONY: browser
browser: make.bin
	./make.bin browser
.PHONY: bumpversion-patch
bumpversion-patch: make.bin
	./make.bin bumpversion-patch
.PHONY: chrome
chrome: make.bin
	./make.bin chrome
.PHONY: chrome-copy-messages
chrome-copy-messages: make.bin
	./make.bin chrome-copy-messages
.PHONY: ci
ci: make.bin
	./make.bin ci
.PHONY: clean
clean: make.bin
	./make.bin clean
.PHONY: deps
deps: make.bin
	./make.bin deps
.PHONY: dev
dev: make.bin
	./make.bin dev
.PHONY: dist
dist: make.bin
	./make.bin dist
.PHONY: dist-build
dist-build: make.bin
	./make.bin dist-build
.PHONY: dist-build-go
dist-build-go: make.bin
	./make.bin dist-build-go
.PHONY: docs
docs: make.bin
	./make.bin docs
.PHONY: fmt
fmt: make.bin
	./make.bin fmt
.PHONY: genMakefile
genMakefile: make.bin
	./make.bin genMakefile
.PHONY: govet
govet: make.bin
	./make.bin govet
.PHONY: hot
hot: make.bin
	./make.bin hot
.PHONY: hot-build
hot-build: make.bin
	./make.bin hot-build
.PHONY: lint
lint: make.bin
	./make.bin lint
.PHONY: release
release: make.bin
	./make.bin release
.PHONY: releaseChromeExt
releaseChromeExt: make.bin
	./make.bin releaseChromeExt
.PHONY: tasks
tasks: make.bin
	./make.bin tasks
.PHONY: test
test: make.bin
	./make.bin test
.PHONY: test-all
test-all: make.bin
	./make.bin test-all
.PHONY: translations-fixup
translations-fixup: make.bin
	./make.bin translations-fixup

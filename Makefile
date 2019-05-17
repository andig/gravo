.PHONY: default clean checks test build publish-images test-release

TAG_NAME := $(shell git tag -l --contains HEAD)
SHA := $(shell git rev-parse --short HEAD)
VERSION := $(if $(TAG_NAME),$(TAG_NAME),$(SHA))

BUILD_DATE := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

default: clean checks test build

clean:
	rm -rf dist/ cover.out

checks:
	golangci-lint run

test: clean
	go test ./...

build: clean
	@echo Version: $(VERSION) $(BUILD_DATE)
	go build -v -ldflags '-X "main.version=${VERSION}" -X "main.commit=${SHA}" -X "main.date=${BUILD_DATE}"' -o gravo

publish-images:
	@echo Version: $(VERSION) $(BUILD_DATE)
	seihon publish --version="$(TAG_NAME)" --image-name andig/gravo --base-runtime-image alpine --dry-run=false
	seihon publish --version="latest" --image-name andig/gravo --base-runtime-image alpine --dry-run=false

test-release:
	goreleaser --snapshot --skip-publish --rm-dist

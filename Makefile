PROJECT = sem

VERSION ?= latest
OS ?= $(shell uname -s | tr A-Z a-z)
ARCH ?= amd64
GO_BIN ?= go
GOOS ?= $(OS)
GOARCH ?= $(ARCH)
CGO_ENABLED ?= 0
GOPROXY ?= direct
BASE_LDFLAGS = -w -s
ifneq ($(OS),darwin)
LDFLAGS = $(BASE_LDFLAGS) -linkmode external -extldflags -static
else
LDFLAGS = $(BASE_LDFLAGS)
endif

GO_BUILD = env GOPROXY=$(GOPROXY) GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) $(GO_BIN) build -ldflags "$(LDFLAGS)"

.PHONY: all clean docker push fmt

all: docker

binary: $(PROJECT)

init: docker-compose.override.yaml

docker-compose.override.yaml:
	@[ -f docker-compose.$(OS).yaml ] && cp -n docker-compose.$(OS).yaml docker-compose.override.yaml || true

fmt:
	gofmt -w ./{internal,main.go}

$(PROJECT):
	$(GO_BUILD) -o $(PROJECT) ./main.go

docker:
	docker build --build-arg GOPROXY=$(GOPROXY) --build-arg VERSION=$(VERSION) -t $(PROJECT):$(VERSION) .

push:
	docker push $(PROJECT):$(VERSION)

remove_images:
	docker rmi $(PROJECT):$(VERSION) || true

clean: remove_images

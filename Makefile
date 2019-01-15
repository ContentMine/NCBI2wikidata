BASIC=Makefile
GO=go
GIT=git

$(eval VERSION:=$(shell git rev-parse HEAD))
$(eval REMOTE:=$(shell git remote get-url origin))
LD_FLAGS=-ldflags "-X main.Version=${VERSION} -X main.Remote=${REMOTE}"

all: .PHONY ingest

ingest: .PHONY
	$(GO) build $(LD_FLAGS)

fmt: .PHONY
	$(GO) fmt

vet: .PHONY
	$(GO) vet

test: .PHONY vet
	$(GO) test

.PHONY:

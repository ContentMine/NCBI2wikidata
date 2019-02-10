BASIC=Makefile
GO=go
GIT=git

$(eval VERSION:=$(shell git rev-parse HEAD)$(shell git diff --quiet || echo '*'))
$(eval REMOTE:=$(shell git remote get-url origin))
LD_FLAGS=-ldflags "-X main.Version=${VERSION} -X main.Remote=${REMOTE}"

all: .PHONY NCBI2wikidata GenerateMeshTerms

NCBI2wikidata: .PHONY check-env
	$(GO) install $(LD_FLAGS) github.com/ContentMine/NCBI2wikidata

GenerateMeshTerms: .PHONY check-env
	$(GO) install $(LD_FLAGS) github.com/ContentMine/GenerateMeshTerms

fmt: .PHONY check-env
	$(GO) fmt github.com/ContentMine/NCBI2wikidata
	$(GO) fmt github.com/ContentMine/GenerateMeshTerms
	$(GO) fmt github.com/ContentMine/EUtils

vet: .PHONY check-env
	$(GO) vet github.com/ContentMine/NCBI2wikidata
	$(GO) vet github.com/ContentMine/GenerateMeshTerms
	$(GO) vet github.com/ContentMine/EUtils

test: .PHONY vet check-env
	$(GO) test -v github.com/ContentMine/NCBI2wikidata
	$(GO) test -v github.com/ContentMine/GenerateMeshTerms
	$(GO) test -v github.com/ContentMine/EUtils

get: .PHONY
	$(GIT) submodule update --init

clean: .PHONY
	rm -r bin
	rm -r pkg

.PHONY:

check-env: .PHONY
ifndef GOPATH
    $(error GOPATH is undefined)
endif

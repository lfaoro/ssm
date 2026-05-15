FLAGS=-trimpath -buildvcs=false -tags='netgo,osusergo,static_build'
LDFLAGS=-ldflags='-w -s -extldflags -static -buildid='

default: build-static

build-static:
	@go mod tidy
	CGO_ENABLED=0 go build ${FLAGS} ${LDFLAGS} -o ./bin/ .
build-linked:
	@go mod tidy
	rm bin/ssm
	go build -ldflags='-buildid= -w -s' -trimpath -buildvcs=false -o ./bin .

clean:
	rm -rf build/*
	rm -rf bin/*
	go clean -i -r
distclean: clean
	go clean -cache
	go clean -modcache
	go clean -testcache
	go clean -fuzzcache

release: pre release-prod help
release-check:
	goreleaser check
	goreleaser healthcheck
release-prod: release-check
	goreleaser release --verbose --clean --skip=validate
release-dev:
	goreleaser release --verbose --snapshot --clean

pre: 
	@go mod tidy
	# @go fmt ./... && go vet ./...

stats:
	@python3 scripts/stats.py

test:
	go test -race -count=1 ./...

vet:
	go vet ./...

build:
	go build ./...

update:
	go get -u .

stop:
	@pkill -9 dev.sh ||:
	@pkill -9 inotify ||:
	@pkill -9 ssm ||:

.PHONY: help
help:
	build/ssm_linux_amd64_v1/ssm --help >data/help


backup: 
	rm -rf build/*
	tar -czvf ../ssm-$(shell date +%Y%m%d).tgz --exclude='.git' .

include data/tag.mk

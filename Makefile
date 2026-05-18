FLAGS=-trimpath -buildvcs=false -tags='netgo,osusergo,static_build'
LDFLAGS=-ldflags='-w -s -extldflags -static -buildid='

default: build-static

build-static:
	@go mod tidy
	CGO_ENABLED=0 go build ${FLAGS} ${LDFLAGS} -o ./bin/ .
build-linked:
	@go mod tidy
	rm -f bin/ssm
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

stats:
	@go run scripts/stats.go

test:
	go test -race -count=1 ./...

bench:
	go test -bench=. -benchmem -count=10 ./pkg/...

bench-cpu:
	go test -bench=. -benchmem -count=10 -cpuprofile=cpu.prof ./pkg/...

bench-mem:
	go test -bench=. -benchmem -count=10 -memprofile=mem.prof ./pkg/...

bench-compare:
	@if [ ! -f bench-old.txt ] || [ ! -f bench-new.txt ]; then \
		echo "usage: run benchmarks, save as bench-old.txt and bench-new.txt"; \
		exit 1; \
	fi
	benchstat bench-old.txt bench-new.txt

vet:
	go vet ./...

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		go fmt ./... && go vet ./...; \
	fi

build:
	go build -o ./bin/ssm .

go-mod-tidy-check:
	@cp go.mod go.mod.bak && cp go.sum go.sum.bak
	@go mod tidy -diff
	@mv go.mod.bak go.mod && mv go.sum.bak go.sum
	@rm -f go.mod.bak go.sum.bak

check: lint go-mod-tidy-check test build

update:
	go get -u .

stop:
	@pkill -9 dev.sh ||:
	@pkill -9 inotify ||:
	@pkill -9 ssm ||:

.PHONY: help test bench bench-cpu bench-mem bench-compare vet lint build build-static build-linked go-mod-tidy-check check update stop clean distclean release release-check release-prod release-dev pre stats backup
help:
	go run . --help >data/help


backup: 
	rm -rf build/*
	tar -czvf ../ssm-$(shell date +%Y%m%d).tgz --exclude='.git' .

include data/tag.mk

include makefile.tag

FLAGS=-trimpath -buildvcs=false -tags='netgo,osusergo,static_build'
LDFLAGS=-ldflags='-s -w -extldflags "-static"'

default: build

build:
	@go mod tidy
	CGO_ENABLED=0 go build ${FLAGS} ${LDFLAGS} -o ./bin/ .

release: pre release-prod
release-check:
	goreleaser check
	goreleaser healthcheck
	goreleaser release --snapshot --clean
release-prod:
	goreleaser release --clean --skip=announce,validate
release-dev:
	GORELEASER_CURRENT_TAG="v0.0.1" goreleaser release --clean --skip=announce,validate --snapshot --skip-publish

pre: 
	@go mod tidy
	@go fmt ./... && go vet ./...

update:
	go get -u .

stop:
	@pkill -9 dev.sh ||:
	@pkill -9 inotify ||:
	@pkill -9 ssm ||:

clean:
	rm -rf build/*

backup: 
	tar -czvf ../ssm-$(shell date +%Y%m%d).tgz --exclude='.git' .

.PHONY: build

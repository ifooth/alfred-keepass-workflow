GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags || echo "unknown version")
BUILDTIME=$(shell date -u)
GOBUILD=CGO_ENABLED=0 go build -trimpath -ldflags '-w -s -buildid='

.PHONY: update
update:
	@go get -u

.PHONY: tidy
tidy:
	@go mod tidy

.PHONY: build
build:
	GOOS=darwin GOARCH=arm64 ${GOBUILD} -o ./output/alfred-keepass-workflow *.go

.PHONY: run
run:
	export alfred_workflow_bundleid="com.ifooth.alfred-ssh-workflow" && \
	export alfred_workflow_cache="./output/cache" && \
	export alfred_workflow_data="./output/data" && \
	export config="./etc/config.yml" && \
    go run main.go

.PHONY: build-workflow
build-workflow: build
	rm -rf ./output/workflow && \
	mkdir -p ./output/workflow && \
	cd ./output/workflow && \
	cp ../../info.plist . && \
	cp ../../iterm_ssh.js . && \
	cp ../../icon.png . && \
	cp ../alfred-ssh-workflow . && \
	zip sshmgr.alfredworkflow info.plist iterm_ssh.js icon.png alfred-ssh-workflow

.PHONY: test
test:
	@go test -v ./... -cover

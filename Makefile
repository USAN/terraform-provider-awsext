TEST?=$$(go list ./... | grep -v 'vendor')
NAME=awsext
BINARY=terraform-provider-${NAME}

default: build

# build: produces the Mac (darwin/amd64) binary in the source directory.
# ~/.terraformrc dev_overrides points here, so no separate install step needed.
build:
	go build -o ${BINARY}

# build-linux: cross-compiles a linux/amd64 binary for AWS deployment and
# installs it into DominionSolution/_BuildScripts/providers/ for the CI pipeline.
DOMINION_PROVIDERS=../DominionSolution/_BuildScripts/providers
build-linux:
	GOOS=linux GOARCH=amd64 go build -o ${BINARY}_linux_amd64
	cp ${BINARY}_linux_amd64 ${DOMINION_PROVIDERS}/terraform-provider-awsext
	@echo "Linux binary installed to ${DOMINION_PROVIDERS}/terraform-provider-awsext"

release:
	goreleaser release --rm-dist --snapshot --skip-publish  --skip-sign

test:
	go test -i $(TEST) || exit 1
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

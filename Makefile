SHELL           := /bin/bash
APP_NAME        ?= aws-sqs-worker-job-controller
API_PKG         := awssqsworkerjobcontroller
API_VERSION     := v1
TEMP_DIR        := _tmp
CGO_ENABLED     ?= 1
CURRENT_DIR     := $(shell pwd)
GOBIN           ?= $(shell go env GOPATH)/bin
CODE_GEN_INPUT  := k8s.io/${APP_NAME}/pkg/apis/${API_PKG}/${API_VERSION}
CODE_GEN_OUTPUT := pkg/generated
CODE_GEN_ARGS   := --output-base ${CURRENT_DIR} --go-header-file ${CURRENT_DIR}/${TEMP_DIR}/empty.txt

all: build test lint

${TEMP_DIR}:
	@mkdir -p $@

codegen: ${TEMP_DIR}
	@touch $</empty.txt
	"$(GOBIN)/deepcopy-gen" --input-dirs "${CODE_GEN_INPUT}" -O zz_generated.deepcopy --bounding-dirs "${CODE_GEN_INPUT}" ${CODE_GEN_ARGS}
	"${GOBIN}/client-gen" --clientset-name "versioned" --input-base "" --input "${CODE_GEN_INPUT}" --output-package "${CODE_GEN_OUTPUT}/clientset" ${CODE_GEN_ARGS}
	"${GOBIN}/lister-gen" --input-dirs "${CODE_GEN_INPUT}" --output-package "${CODE_GEN_OUTPUT}/listers" ${CODE_GEN_ARGS}
	"${GOBIN}/informer-gen" --input-dirs "${CODE_GEN_INPUT}" --versioned-clientset-package "${CODE_GEN_OUTPUT}/clientset/versioned" --listers-package "${CODE_GEN_OUTPUT}/listers" --output-package "${CODE_GEN_OUTPUT}/informers" ${CODE_GEN_ARGS}

build: codegen
	CGO_ENABLED=${CGO_ENABLED} go build -ldflags="-s -w" -trimpath -tags timetzdata -o ${APP_NAME}

test:
	go test

lint:
	go vet
	golint -set_exit_status

clean:
	@rm -f ${APP_NAME} main

build-image:
	@docker build -t ${APP_NAME} .
	@docker image prune -f

lint-image:
	@docker run --rm -i hadolint/hadolint < Dockerfile

run-container:
	@docker run --env-file=.env --rm ${APP_NAME}

clean-image:
	@docker rmi -f ${APP_NAME}

.PHONY: all codegen build test lint clean build-image lint-image run-container clean-image

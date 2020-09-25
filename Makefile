SHELL           := /bin/bash
APP_NAME        := aws-sqs-worker-job-controller
APP_BIN_NAME    ?= aws-sqs-worker-job-controller
API_PKG         := awssqsworkerjobcontroller
API_VERSION     := v1
TEMP_DIR        := _tmp
CGO_ENABLED     ?= 1
CURRENT_DIR     := $(shell pwd)
GOBIN           ?= $(shell go env GOPATH)/bin
CODE_GEN_SRCS   := $(shell find pkg/apis/${API_PKG}/ -type f -name '*.go')
CODE_GEN_DIR    := github.com/supercaracal/${APP_NAME}/pkg
CODE_GEN_INPUT  := ${CODE_GEN_DIR}/apis/${API_PKG}/${API_VERSION}
CODE_GEN_OUTPUT := ${CODE_GEN_DIR}/generated
CODE_GEN_ARGS   := --output-base ${CURRENT_DIR} --go-header-file ${CURRENT_DIR}/${TEMP_DIR}/empty.txt
CODE_GEN_DEEPC  := zz_generated.deepcopy

all: build test lint

${TEMP_DIR}:
	@mkdir -p $@

${TEMP_DIR}/codegen: ${TEMP_DIR} ${CODE_GEN_SRCS}
	@touch -a ${TEMP_DIR}/empty.txt
	"$(GOBIN)/deepcopy-gen" --input-dirs "${CODE_GEN_INPUT}" -O "${CODE_GEN_DEEPC}" --bounding-dirs "${CODE_GEN_INPUT}" ${CODE_GEN_ARGS}
	"${GOBIN}/client-gen" --clientset-name "versioned" --input-base "" --input "${CODE_GEN_INPUT}" --output-package "${CODE_GEN_OUTPUT}/clientset" ${CODE_GEN_ARGS}
	"${GOBIN}/lister-gen" --input-dirs "${CODE_GEN_INPUT}" --output-package "${CODE_GEN_OUTPUT}/listers" ${CODE_GEN_ARGS}
	"${GOBIN}/informer-gen" --input-dirs "${CODE_GEN_INPUT}" --versioned-clientset-package "${CODE_GEN_OUTPUT}/clientset/versioned" --listers-package "${CODE_GEN_OUTPUT}/listers" --output-package "${CODE_GEN_OUTPUT}/informers" ${CODE_GEN_ARGS}
	@rm -f pkg/apis/${API_PKG}/${API_VERSION}/${CODE_GEN_DEEPC}.go
	@mv ${CODE_GEN_DIR}/apis/${API_PKG}/${API_VERSION}/${CODE_GEN_DEEPC}.go pkg/apis/${API_PKG}/${API_VERSION}/
	@rm -rf pkg/generated
	@mv ${CODE_GEN_OUTPUT} pkg/
	@rm -rf github.com
	@touch $@

codegen: ${TEMP_DIR}/codegen

build: codegen
	CGO_ENABLED=${CGO_ENABLED} go build -ldflags="-s -w" -trimpath -tags timetzdata -o ${APP_BIN_NAME}

test:
	go test ./...

lint:
	go vet ./...
	golint -set_exit_status ./...

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

apply-manifests:
	kubectl apply -f config/localstack.yaml
	kubectl apply -f config/aws-sqs-worker-job-controller.yaml
	kubectl apply -f config/crd.yaml
	kubectl apply -f config/sleep-awssqsworkerjob.yaml

.PHONY: all codegen build test lint clean build-image lint-image run-container clean-image

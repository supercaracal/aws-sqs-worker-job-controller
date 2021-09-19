SHELL                 := /bin/bash -e -u -o pipefail
APP_NAME              := aws-sqs-worker-job-controller
API_PKG               := supercaracal
TEMP_DIR              := _tmp
AWS_ENDPOINT_URL      := http://127.0.0.1:4566
AWS_REGION            := ap-northeast-1
AWS_ACCOUNT_ID        := 000000000000
AWS_ACCESS_KEY_ID     := AAAAAAAAAAAAAAAAAAAA
AWS_SECRET_ACCESS_KEY := AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AWS_CLI_OPTS          := --region=${AWS_REGION}
TEST_IMAGE_TAG        := latest
TEST_QUEUE_NAME       := sleep-queue.fifo
TEST_QUEUE_URL        := ${AWS_ENDPOINT_URL}/${AWS_ACCOUNT_ID}/${TEST_QUEUE_NAME}

ifdef AWS_PROFILE
	AWS_CLI_OPTS += --profile=${AWS_PROFILE}
endif

ifdef AWS_ENDPOINT_URL
	AWS_CLI_OPTS += --endpoint-url=${AWS_ENDPOINT_URL}
endif

all: build test lint

${TEMP_DIR}:
	@mkdir -p $@

${TEMP_DIR}/codegen: GOBIN                  ?= $(shell go env GOPATH)/bin
${TEMP_DIR}/codegen: API_VERSION            := v1
${TEMP_DIR}/codegen: CODE_GEN_DIR           := pkg
${TEMP_DIR}/codegen: CODE_GEN_INPUT         := ${CODE_GEN_DIR}/apis/${API_PKG}/${API_VERSION}
${TEMP_DIR}/codegen: CODE_GEN_OUTPUT        := ${CODE_GEN_DIR}/generated
${TEMP_DIR}/codegen: CURRENT_DIR            := $(shell pwd)
${TEMP_DIR}/codegen: CODE_GEN_ARGS          := --output-base=${CURRENT_DIR} --go-header-file=${CURRENT_DIR}/${TEMP_DIR}/empty.txt
${TEMP_DIR}/codegen: CODE_GEN_DEEPC         := zz_generated.deepcopy
${TEMP_DIR}/codegen: CODE_GEN_CLI_SET_NAME  := versioned
${TEMP_DIR}/codegen: ${TEMP_DIR} $(shell find pkg/apis/${API_PKG}/ -type f -name '*.go')
	@touch -a ${TEMP_DIR}/empty.txt
	${GOBIN}/deepcopy-gen ${CODE_GEN_ARGS} --input-dirs=${CODE_GEN_INPUT} --bounding-dirs=${CODE_GEN_INPUT} -O ${CODE_GEN_DEEPC}
	${GOBIN}/client-gen   ${CODE_GEN_ARGS} --input=${CODE_GEN_INPUT}      --output-package=${CODE_GEN_OUTPUT}/clientset --input-base="" --clientset-name=${CODE_GEN_CLI_SET_NAME}
	${GOBIN}/lister-gen   ${CODE_GEN_ARGS} --input-dirs=${CODE_GEN_INPUT} --output-package=${CODE_GEN_OUTPUT}/listers
	${GOBIN}/informer-gen ${CODE_GEN_ARGS} --input-dirs=${CODE_GEN_INPUT} --output-package=${CODE_GEN_OUTPUT}/informers --versioned-clientset-package=${CODE_GEN_OUTPUT}/clientset/${CODE_GEN_CLI_SET_NAME} --listers-package=${CODE_GEN_OUTPUT}/listers
	@touch $@

codegen: ${TEMP_DIR}/codegen

build: GOOS        ?= $(shell go env GOOS)
build: GOARCH      ?= $(shell go env GOARCH)
build: CGO_ENABLED ?= $(shell go env CGO_ENABLED)
build: codegen
	GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=${CGO_ENABLED} go build -ldflags="-s -w" -trimpath -tags timetzdata -o ${APP_NAME}

test:
	@go clean -testcache
	@go test -race ./...

lint:
	@go vet ./...
	@golint -set_exit_status ./...

run: TZ := Asia/Tokyo
run:
	@AWS_REGION=${AWS_REGION} \
	AWS_ENDPOINT_URL=${AWS_ENDPOINT_URL} \
	AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID} \
	AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY} \
	SELF_NAMESPACE=default \
	TZ=${TZ} \
	./${APP_NAME} \
	--kubeconfig=$$HOME/.kube/config

clean:
	@rm -f ${APP_NAME} main

build-image:
	@docker build -t ${APP_NAME}:${TEST_IMAGE_TAG} .
	@docker image prune -f

lint-image:
	@docker run --rm -i hadolint/hadolint < Dockerfile

run-container:
	@docker run --env-file=.env --rm ${APP_NAME}

clean-image:
	@docker rmi -f ${APP_NAME}:${TEST_IMAGE_TAG}

apply-manifests:
	@kubectl --context=kind-kind apply -f config/localstack.yaml
	@kubectl --context=kind-kind apply -f config/controller.yaml
	@kubectl --context=kind-kind apply -f config/crd.yaml
	@kubectl --context=kind-kind apply -f config/sleep-awssqsworkerjob.yaml

mod-replace-kube: KUBE_LIB_VER := 1.22.1
mod-replace-kube:
	@./go_mod_replace.sh ${KUBE_LIB_VER}

create-test-queue:
	@aws ${AWS_CLI_OPTS} sqs create-queue --queue-name=${TEST_QUEUE_NAME}

enqueue-test-task: BODY ?= 3
enqueue-test-task:
	@aws ${AWS_CLI_OPTS} sqs send-message --queue-url=${TEST_QUEUE_URL} --message-body=${BODY}

get-test-queue-attrs:
	@aws ${AWS_CLI_OPTS} sqs get-queue-attributes --queue-url=${TEST_QUEUE_URL}

.PHONY: all codegen build test lint run clean \
	build-image lint-image run-container clean-image \
	apply-manifests mod-replace-kube \
	create-test-queue enqueue-test-task get-test-queue-attrs

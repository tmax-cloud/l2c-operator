SDK	= operator-sdk

REGISTRY      ?= tmaxcloudck
VERSION       ?= $(shell cat version/version.go | grep -Po '(?<=")v[0-9]+\.[0-9]+\.[0-9]+(?=")')

PACKAGE_NAME  = github.com/tmax-cloud/l2c-operator

OPERATOR_NAME  = l2c-operator
OPERATOR_IMG   = $(REGISTRY)/$(OPERATOR_NAME):$(VERSION)

DB_DEPLOYER_NAME  = l2c-db-deployer
DB_DEPLOYER_IMG   = $(REGISTRY)/$(DB_DEPLOYER_NAME):$(VERSION)

SCAN_WAITER_NAME  = l2c-scan-waiter
SCAN_WAITER_IMG   = $(REGISTRY)/$(SCAN_WAITER_NAME):$(VERSION)

VSCODE_NAME  = l2c-vscode
VSCODE_IMG   = $(REGISTRY)/$(VSCODE_NAME):$(VERSION)

MTA_NAME = l2c-tup-jeus
MTA_IMG  = 192.168.6.110:5000/$(MTA_NAME):$(VERSION)

BIN = ./build/_output/bin

.PHONY: all
all: test gen build push

.PHONY: clean
clean:
	rm -rf $(BIN)

.PHONY: gen
gen:
	$(SDK) generate k8s
	$(SDK) generate crds


.PHONY: build build-operator build-db-deployer build-scan-waiter
build: build-operator build-db-deployer build-scan-waiter

build-operator:
	$(SDK) build $(OPERATOR_IMG)

build-db-deployer:
	CGO_ENABLED=0 go build -o $(BIN)/db-deployer $(PACKAGE_NAME)/cmd/db-deployer
	docker build -t $(DB_DEPLOYER_IMG) -f build/Dockerfile.db-deployer .

build-scan-waiter:
	CGO_ENABLED=0 go build -o $(BIN)/scan-waiter $(PACKAGE_NAME)/cmd/scan-waiter
	docker build -t $(SCAN_WAITER_IMG) -f build/Dockerfile.scan-waiter .


.PHONY: push push-operator push-db-deployer push-scan-waiter
push: push-operator push-db-deployer push-scan-waiter

push-operator:
	docker push $(OPERATOR_IMG)

push-db-deployer:
	docker push $(DB_DEPLOYER_IMG)

push-scan-waiter:
	docker push $(SCAN_WAITER_IMG)


.PHONY: push-latest push-operator-latest push-db-deployer-latest push-scan-waiter-latest
push-latest: push-operator-latest push-db-deployer-latest push-scan-waiter-latest
push-operator-latest:
	docker tag $(OPERATOR_IMG) $(REGISTRY)/$(OPERATOR_NAME):latest
	docker push $(REGISTRY)/$(OPERATOR_NAME):latest

push-db-deployer-latest:
	docker tag $(DB_DEPLOYER_IMG) $(REGISTRY)/$(DB_DEPLOYER_NAME):latest
	docker push $(REGISTRY)/$(DB_DEPLOYER_NAME):latest

push-scan-waiter-latest:
	docker tag $(SCAN_WAITER_IMG) $(REGISTRY)/$(SCAN_WAITER_NAME):latest
	docker push $(REGISTRY)/$(SCAN_WAITER_NAME):latest


.PHONY: test test-gen save-sha-gen compare-sha-gen test-verify save-sha-mod compare-sha-mod verify test-unit test-lint
test: test-gen test-verify test-unit test-lint

test-gen: save-sha-gen gen compare-sha-gen

save-sha-gen:
	$(eval CRDSHA1=$(shell sha512sum deploy/crds/tmax.io_tupwas_crd.yaml))
	$(eval CRDSHA2=$(shell sha512sum deploy/crds/tmax.io_tupdbs_crd.yaml))
	$(eval GENSHA=$(shell sha512sum pkg/apis/tmax/v1/zz_generated.deepcopy.go))

compare-sha-gen:
	$(eval CRDSHA1_AFTER=$(shell sha512sum deploy/crds/tmax.io_tupwas_crd.yaml))
	$(eval CRDSHA2_AFTER=$(shell sha512sum deploy/crds/tmax.io_tupdbs_crd.yaml))
	$(eval GENSHA_AFTER=$(shell sha512sum pkg/apis/tmax/v1/zz_generated.deepcopy.go))
	@if [ "${CRDSHA1_AFTER}" = "${CRDSHA1}" ]; then echo "deploy/crds/tmax.io_tupwas_crd.yaml is not changed"; else echo "deploy/crds/tmax.io_tupwas_crd.yaml file is changed"; exit 1; fi
	@if [ "${CRDSHA2_AFTER}" = "${CRDSHA2}" ]; then echo "deploy/crds/tmax.io_tupdbs_crd.yaml is not changed"; else echo "deploy/crds/tmax.io_tupdbs_crd.yaml file is changed"; exit 1; fi
	@if [ "${GENSHA_AFTER}" = "${GENSHA}" ]; then echo "zz_generated.deepcopy.go is not changed"; else echo "zz_generated.deepcopy.go file is changed"; exit 1; fi

test-verify: save-sha-mod verify compare-sha-mod

save-sha-mod:
	$(eval MODSHA=$(shell sha512sum go.mod))
	$(eval SUMSHA=$(shell sha512sum go.sum))

verify:
	go mod verify

compare-sha-mod:
	$(eval MODSHA_AFTER=$(shell sha512sum go.mod))
	$(eval SUMSHA_AFTER=$(shell sha512sum go.sum))
	@if [ "${MODSHA_AFTER}" = "${MODSHA}" ]; then echo "go.mod is not changed"; else echo "go.mod file is changed"; exit 1; fi
	@if [ "${SUMSHA_AFTER}" = "${SUMSHA}" ]; then echo "go.sum is not changed"; else echo "go.sum file is changed"; exit 1; fi

test-unit:
	go test -v ./pkg/...

test-lint:
	golangci-lint run ./... -v -E gofmt --timeout 1h0m0s


.PHONY: pre vscode mta
pre: vscode mta

vscode:
	docker build -t $(VSCODE_IMG) -f build/Dockerfile.vscode .
	docker push $(VSCODE_IMG)

mta:
	docker build -t $(MTA_IMG) -f build/Dockerfile.tupjeus .
	docker push $(MTA_IMG)


.PHONY: run-local deploy
run-local:
	$(SDK) run --local --watch-namespace=""

deploy:
	kubectl apply -f deploy/
	kubectl apply -f deploy/crds/

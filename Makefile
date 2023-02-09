CURRENT_VERSION := 0.0.1

uname_s := $(shell uname -s)
uname_m := $(shell uname -m)
l_uname_s = $(shell echo $(uname_s) | tr A-Z a-z)
l_uname_m = $(shell echo $(uname_m) | tr A-Z a-z)

KERNEL=$(l_uname_s)
ARCH=$(l_uname_m)


.PHONY: build info fmt vet test clean local-dev-install testacc

all: build

info:
	@echo "Global info"
	@echo "$(KERNEL)"
	@echo "$(ARCH)"

fmt:
	@echo " -> checking code style"
	@! gofmt -d $(shell find . -path ./vendor -prune -o -name '*.go' -print) | grep '^'

vet:
	@echo " -> vetting code"
	@go vet ./...

build: clean
	@echo " -> Building"
	mkdir -p bin
	CGO_ENABLED=0 go build -trimpath -o bin/terraform-provider-humanitec
	@echo "Built terraform-provider-humanitec"

# Run acceptance tests
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

local-dev-install: build
	@echo "Building this release $(CURRENT_VERSION) on $(KERNEL)/$(ARCH)"
	rm -rf ~/.terraform.d/plugins/registry.terraform.io/humanitec/humanitec
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/humanitec/humanitec/$(CURRENT_VERSION)/$(KERNEL)_$(ARCH)/
	cp bin/terraform-provider-humanitec ~/.terraform.d/plugins/registry.terraform.io/humanitec/humanitec/$(CURRENT_VERSION)/$(KERNEL)_$(ARCH)/

install: build
	cp bin/terraform-provider-humanitec $$GOPATH/bin/terraform-provider-humanitec

clean:
	# @git clean -f -d
	@echo "Clean"

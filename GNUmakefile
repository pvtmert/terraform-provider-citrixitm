TEST?=$$(go list ./... |grep -v 'vendor')
PKG_NAME=citrixitm

default: build

build: fmtcheck
	go install

test: fmtcheck
	go test ./...

testacc: fmtcheck
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

install-tools:
	GO111MODULE=off go get -u github.com/client9/misspell/cmd/misspell
	GO111MODULE=off go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

lint:
	@echo "==> Checking source code against linters..."
	@golangci-lint run ./$(PKG_NAME)

fmt:
	@echo "==> Fixing source code with gofmt..."
	gofmt -s -w ./$(PKG_NAME)

# Currently required by tf-deploy compile
fmtcheck:
	@echo "==> Checking source code against gofmt..."
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)

install: build
	cp $(GOPATH)/bin/terraform-provider-citrixitm $(HOME)/.terraform.d/plugins/

check-itm-env:
ifndef ITM_BASE_URL
	$(error ITM_BASE_URL is undefined)
endif
ifndef ITM_CLIENT_ID
	$(error ITM_CLIENT_ID is undefined)
endif
ifndef ITM_CLIENT_SECRET
	$(error ITM_CLIENT_SECRET is undefined)
endif

.PHONY: build fmt fmtcheck install-tools lint test test-compile testacc website website-lint website-test install

APP := ibtui
CMD_DIR := ./cmd/ibtui
GO := go

.PHONY: all build check clean count debug deps help run update
all: help

## build: Build the executable
build:
	$(GO) build -o $(APP) $(CMD_DIR)

## check: Perform lints and tests
check:
	golangci-lint run --show-stats --color always | \less -iMRFX
	govulncheck ./...
	go test ./...
	@printf "Press Enter to continue..."; read dummy

## clean: Remove any build artifacts from the directory tree
clean:
	rm -f ./$(APP)

## count: Count the LOCs in the codebase
count:
	@find . \
		-type d -name '_*' -prune -o \
		-type f -name '*.go' ! -name '_*.go' -print | \
	while IFS= read -r file; do \
		awk '\
			/^[[:space:]]*$$/ { \
				bloc++; \
				next; \
			} \
			/^[[:space:]]*\/\// { \
				cloc++; \
				next; \
			} \
			{ \
				sloc++; \
			} \
			END { \
				printf "%d %d %d\n", sloc+0, cloc+0, bloc+0; \
			} \
		' "$$file"; \
	done | \
	awk '\
		{ \
			sloc += $$1; \
			cloc += $$2; \
			bloc += $$3; \
		} \
		END { \
			total = sloc + cloc + bloc; \
			printf "SLOC : %d\n", sloc; \
			printf "CLOC : %d\n", cloc; \
			printf "BLOC : %d\n", bloc; \
			printf "TOTAL: %d\n", total; \
		} \
	'

## debug: Run a skeleton binary to test temporary code
debug:
	$(GO) run ./cmd/debug/...

## deps: Update all module dependencies, including linters
deps:
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install golang.org/x/vuln/cmd/govulncheck@latest
	$(GO) get -u ./...
	$(GO) mod tidy

## help: Print this help message
help:
	@echo ''
	@echo 'USAGE'
	@echo '    make [target]'
	@echo ''
	@echo 'TARGETS'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/   /'
	@echo ''
	@echo 'EXAMPLE'
	@echo '    "make run"'
	@echo ''

## run: Temporarily run the binary
run:
	$(GO) run $(CMD_DIR)

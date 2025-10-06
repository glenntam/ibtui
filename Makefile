.PHONY: all build clean format run update

APP := ibtui

all: format run clean

build:
	go build -o $(APP) ./cmd/tui/

clean:
	rm -f ./$(APP)

format:
	gofumpt -d -e -extra . | colordiff | \less -iMRX
	go vet ./...
	golangci-lint run -E revive,gosec,iface,ireturn,intrange,errorlint,errname,err113,makezero,mirror,misspell,mnd,nilerr,nilnesserr,nilnil,nonamedreturns,nosprintfhostport,perfsprint,prealloc,predeclared,rowserrcheck,wastedassign,wrapcheck,gocritic,sloglint,sqlclosecheck,unconvert,unparam,unqueryvet,usestdlibvars,usetesting,bodyclose,forcetypeassert --color always | \less -iMRFX
	@printf "Press Enter to continue..."; read dummy

run: build
	./$(APP)

update:
	go get -u ./...
	go mod tidy

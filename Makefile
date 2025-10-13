.PHONY: all build clean check run update

APP := ibtui

all: check run clean

build:
	go build -o $(APP) ./cmd/tui/

clean:
	rm -f ./$(APP)

check:
	gofumpt -d -e -extra . | colordiff | \less -iMRX
	go vet ./...
	golangci-lint run -E asciicheck,bidichk,bodyclose,canonicalheader,containedctx,contextcheck,copyloopvar,decorder,dogsled,dupl,dupword,durationcheck,embeddedstructfieldcheck,err113,errcheck,errchkjson,errname,errorlint,exhaustive,exptostd,fatcontext,forcetypeassert,funcorder,gocheckcompilerdirectives,gochecknoglobals,gochecksumtype,gocognit,goconst,gocritic,gocyclo,godoclint,godot,godox,goheader,gomoddirectives,gomodguard,goprintffuncname,gosec,govet,grouper,iface,importas,inamedparam,ineffassign,interfacebloat,intrange,iotamixing,ireturn,lll,loggercheck,maintidx,makezero,mirror,misspell,mnd,musttag,nakedret,nestif,nilerr,nilnesserr,nilnil,noctx,nolintlint,nonamedreturns,nosprintfhostport,perfsprint,prealloc,predeclared,promlinter,protogetter,reassign,recvcheck,revive,rowserrcheck,sloglint,sqlclosecheck,tagalign,tagliatelle,testableexamples,thelper,tparallel,unconvert,unparam,unqueryvet,unused,usestdlibvars,usetesting,wastedassign,whitespace,wrapcheck,zerologlint --color always | \less -iMRFX
	go test ./...
	@printf "Press Enter to continue..."; read dummy

run: build
	./$(APP)

update:
	go get -u ./...
	go mod tidy

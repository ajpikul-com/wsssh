run: ../wsconn/*.go *.go
	go build
	./testclient

.PHONY:  local official init
init:
	cp go.mod go.mod.local
	cp go.mod go.mod.official
	cp go.sum go.sum.local || touch go.sum.local
	cp go.sum go.sum.official || touch go.sum.official

local: ../wsconn/*.go *.go
	cp go.mod.local go.mod
	cp go.sum.local go.sum
	-GOPROXY=direct go get -u && go mod tidy && go build
	cp go.mod go.mod.local
	cp go.sum go.sum.local
	touch local

official: ../wsconn/*.go *.go
	cp go.mod.official go.mod
	cp go.sum.official go.sum
	-GOPROXY=direct go get -u && go mod tidy && go build
	cp go.mod go.mod.official
	cp go.sum go.sum.official
	-rm local

local:
	cp go.mod.local go.mod
	cp go.sum.local go.sum
	-GOPROXY=direct go get -u && go mod tidy && go build
	cp go.mod go.mod.local
	cp go.sum go.sum.local
	touch local

official:
	cp go.mod.official go.mod
	cp go.sum.official go.sum
	-GOPROXY=direct go get -u && go mod tidy && go build
	cp go.mod go.mod.official
	cp go.sum go.sum.official
	rm local

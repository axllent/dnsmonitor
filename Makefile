ifndef $(GOPATH)
    GOPATH=$(shell go env GOPATH)
    export GOPATH
endif
VERSION ?= "dev"
LDFLAGS=-ldflags "-s -w -X main.version=${VERSION}"

build = GOOS=$(1) GOARCH=$(2) go build ${LDFLAGS} -o dist/dnsmonitor_${VERSION}_$(1)_$(2) \
	&& bzip2 -f dist/dnsmonitor_${VERSION}_$(1)_$(2)

main-build: *.go
	go build ${LDFLAGS} -o dnsmonitor

clean:
	rm -rf dist dnsmonitor

release:
	rm -f dist/dnsmonitor_${VERSION}_*
	$(call build,darwin,386)
	$(call build,darwin,amd64)
	$(call build,freebsd,386)
	$(call build,freebsd,amd64)
	$(call build,freebsd,arm)
	$(call build,linux,386)
	$(call build,linux,amd64)
	$(call build,linux,arm)
	$(call build,linux,arm64)
	$(call build,netbsd,386)
	$(call build,netbsd,amd64)
	$(call build,netbsd,arm)
	$(call build,openbsd,386)
	$(call build,openbsd,amd64)
	$(call build,openbsd,arm)

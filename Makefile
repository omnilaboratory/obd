NAME=obd
NAMECLI=lncli
PACKAGE_CLI=github.com/lightningnetwork/lnd/cmd/lncli
BINDIR=bin
VERSION=$(shell git describe --tags || echo "unknown version")
BUILDTIME=$(shell date -u)
GOVERSION := $(shell go version | awk '{print $$3}')
GOBUILD=CGO_ENABLED=0 go build -trimpath -ldflags '-X "obd/cmds.Version=$(VERSION)" \
		-X "obd/cmds.BuildTime=$(BUILDTIME)" \
		-w -s -buildid='

PLATFORM_LIST = \
	darwin-amd64 \
	darwin-arm64 \
	linux-386 \
	linux-amd64 \
	linux-armv5 \
	linux-armv6 \
	linux-armv7 \
	linux-armv8 \
	linux-mips-softfloat \
	linux-mips-hardfloat \
	linux-mipsle-softfloat \
	linux-mipsle-hardfloat \
	linux-mips64 \
	linux-mips64le \
	freebsd-386 \
	freebsd-amd64 \
	freebsd-arm64

WINDOWS_ARCH_LIST = \
	windows-386 \
	windows-amd64 \
	windows-arm64 \
	windows-arm32v7

LINUX_LIST = \
	linux-amd64 \
	darwin-amd64


all: linux-amd64 darwin-amd64 windows-amd64 # Most used

docker:
	$(GOBUILD) -o $(BINDIR)/$(NAME)-$@

darwin-amd64:
	GOARCH=amd64 GOOS=darwin $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

darwin-arm64:
	GOARCH=arm64 GOOS=darwin $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

linux-386:
	GOARCH=386 GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

linux-amd64:
	GOARCH=amd64 GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@
	GOARCH=amd64 GOOS=linux $(GOBUILD) -o $(BINDIR)/tracker-$@ ./tracker
	cd lnd && GOARCH=amd64 GOOS=linux $(GOBUILD) -o ../$(BINDIR)/$(NAMECLI)-$@ $(PACKAGE_CLI)

linux-armv5:
	GOARCH=arm GOOS=linux GOARM=5 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

linux-armv6:
	GOARCH=arm GOOS=linux GOARM=6 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

linux-armv7:
	GOARCH=arm GOOS=linux GOARM=7 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

linux-armv8:
	GOARCH=arm64 GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

linux-mips-softfloat:
	GOARCH=mips GOMIPS=softfloat GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

linux-mips-hardfloat:
	GOARCH=mips GOMIPS=hardfloat GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

linux-mipsle-softfloat:
	GOARCH=mipsle GOMIPS=softfloat GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

linux-mipsle-hardfloat:
	GOARCH=mipsle GOMIPS=hardfloat GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

linux-mips64:
	GOARCH=mips64 GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

linux-mips64le:
	GOARCH=mips64le GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

freebsd-386:
	GOARCH=386 GOOS=freebsd $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

freebsd-amd64:
	GOARCH=amd64 GOOS=freebsd $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

freebsd-arm64:
	GOARCH=arm64 GOOS=freebsd $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

windows-386:
	GOARCH=386 GOOS=windows $(GOBUILD) -o $(BINDIR)/$(NAME)-$@.exe

windows-amd64:
	GOARCH=amd64 GOOS=windows $(GOBUILD) -o $(BINDIR)/$(NAME)-$@.exe

windows-arm64:
	GOARCH=arm64 GOOS=windows $(GOBUILD) -o $(BINDIR)/$(NAME)-$@.exe

windows-arm32v7:
	GOARCH=arm GOOS=windows GOARM=7 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@.exe

#gz_releases=$(addsuffix .gz, $(PLATFORM_LIST))
#zip_releases=$(addsuffix .zip, $(WINDOWS_ARCH_LIST))
#
#$(gz_releases): %.gz : %
#	chmod +x $(BINDIR)/$(NAME)-$(basename $@)
#	gzip -f -S -$(VERSION).gz $(BINDIR)/$(NAME)-$(basename $@)
#
#$(zip_releases): %.zip : %
#	zip -m -j $(BINDIR)/$(NAME)-$(basename $@)-$(VERSION).zip $(BINDIR)/$(NAME)-$(basename $@).exe
#
#all-arch: $(PLATFORM_LIST) $(WINDOWS_ARCH_LIST)
#releases: $(gz_releases) $(zip_releases)

linux_releases=$(addsuffix .gz, $(LINUX_LIST))
windows_releases=$(addsuffix .zip, windows-amd64)
$(linux_releases): %.gz : %
	chmod +x $(BINDIR)/$(NAME)-$(basename $@)
	gzip -f -S -$(VERSION).gz $(BINDIR)/$(NAME)-$(basename $@)
$(windows_releases): %.zip : %
	zip -m -j $(BINDIR)/$(NAME)-$(basename $@)-$(VERSION).zip $(BINDIR)/$(NAME)-$(basename $@).exe
#only window mac  linux amd64
releases: $(linux_releases) $(windows_releases)


lint:
	golangci-lint run --disable-all -E govet -E gofumpt -E megacheck ./...

clean:
	rm $(BINDIR)/*
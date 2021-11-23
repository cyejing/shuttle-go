NAME := shuttle
PACKAGE_NAME := github.com/cyejing/shuttle
VERSION := `git describe --tags`
COMMIT := `git rev-parse HEAD`

PLATFORM := linux
BUILD_DIR := build
VAR_SETTING := -X $(PACKAGE_NAME)/constant.Version=$(VERSION) -X $(PACKAGE_NAME)/constant.Commit=$(COMMIT)
GOBUILD = env CGO_ENABLED=0 $(GO_DIR)go build -tags "full" -trimpath -ldflags="-s -w -buildid= $(VAR_SETTING)" -o $(BUILD_DIR)

.PHONY: build

build: clean shuttles shuttlec

clean:
	rm -rf $(BUILD_DIR)
	rm -rf logs
	rm -f *.zip
	rm -f *.dat

test:
	# Disable Bloomfilter when testing
	SHADOWSOCKS_SF_CAPACITY="-1" $(GO_DIR)go test -v ./...

shuttles:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) ./cmd/shuttles

shuttlec:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) ./cmd/shuttlec

%.zip: %
	@zip -du $(NAME)-$@ -j $(BUILD_DIR)/$</*
	@zip -du $(NAME)-$@ example/*
	@echo "<<< ---- $(NAME)-$@"

release: darwin-amd64.zip darwin-arm64.zip linux-386.zip linux-amd64.zip \
	linux-arm.zip linux-arm64.zip linux-mips64.zip linux-mips64le.zip\
	linux-mips-softfloat.zip linux-mipsle-softfloat.zip freebsd-386.zip freebsd-amd64.zip\
	windows-386.zip windows-amd64.zip windows-arm.zip windows-arm64.zip

darwin-amd64:
	mkdir -p $(BUILD_DIR)/$@
	GOARCH=amd64 GOOS=darwin $(GOBUILD)/$@ ./cmd/shuttles
	GOARCH=amd64 GOOS=darwin $(GOBUILD)/$@ ./cmd/shuttlec

darwin-arm64:
	mkdir -p $(BUILD_DIR)/$@
	GOARCH=arm64 GOOS=darwin $(GOBUILD)/$@ ./cmd/shuttles
	GOARCH=arm64 GOOS=darwin $(GOBUILD)/$@ ./cmd/shuttlec

linux-386:
	mkdir -p $(BUILD_DIR)/$@
	GOARCH=386 GOOS=linux $(GOBUILD)/$@ ./cmd/shuttles
	GOARCH=386 GOOS=linux $(GOBUILD)/$@ ./cmd/shuttlec

linux-amd64:
	mkdir -p $(BUILD_DIR)/$@
	GOARCH=amd64 GOOS=linux $(GOBUILD)/$@ ./cmd/shuttles
	GOARCH=amd64 GOOS=linux $(GOBUILD)/$@ ./cmd/shuttlec

linux-arm:
	mkdir -p $(BUILD_DIR)/$@
	GOARCH=arm GOOS=linux $(GOBUILD)/$@ ./cmd/shuttles
	GOARCH=arm GOOS=linux $(GOBUILD)/$@ ./cmd/shuttlec

linux-arm64:
	mkdir -p $(BUILD_DIR)/$@
	GOARCH=arm64 GOOS=linux $(GOBUILD)/$@ ./cmd/shuttles
	GOARCH=arm64 GOOS=linux $(GOBUILD)/$@ ./cmd/shuttlec

linux-mips-softfloat:
	mkdir -p $(BUILD_DIR)/$@
	GOARCH=mips GOMIPS=softfloat GOOS=linux $(GOBUILD)/$@ ./cmd/shuttles
	GOARCH=mips GOMIPS=softfloat GOOS=linux $(GOBUILD)/$@ ./cmd/shuttlec

linux-mipsle-softfloat:
	mkdir -p $(BUILD_DIR)/$@
	GOARCH=mipsle GOMIPS=softfloat GOOS=linux $(GOBUILD)/$@ ./cmd/shuttles
	GOARCH=mipsle GOMIPS=softfloat GOOS=linux $(GOBUILD)/$@ ./cmd/shuttlec

linux-mips64:
	mkdir -p $(BUILD_DIR)/$@
	GOARCH=mips64 GOOS=linux $(GOBUILD)/$@ ./cmd/shuttles
	GOARCH=mips64 GOOS=linux $(GOBUILD)/$@ ./cmd/shuttlec

linux-mips64le:
	mkdir -p $(BUILD_DIR)/$@
	GOARCH=mips64le GOOS=linux $(GOBUILD)/$@ ./cmd/shuttles
	GOARCH=mips64le GOOS=linux $(GOBUILD)/$@ ./cmd/shuttlec

freebsd-386:
	mkdir -p $(BUILD_DIR)/$@
	GOARCH=386 GOOS=freebsd $(GOBUILD)/$@ ./cmd/shuttles
	GOARCH=386 GOOS=freebsd $(GOBUILD)/$@ ./cmd/shuttlec

freebsd-amd64:
	mkdir -p $(BUILD_DIR)/$@
	GOARCH=amd64 GOOS=freebsd $(GOBUILD)/$@ ./cmd/shuttles
	GOARCH=amd64 GOOS=freebsd $(GOBUILD)/$@ ./cmd/shuttlec

windows-386:
	mkdir -p $(BUILD_DIR)/$@
	GOARCH=386 GOOS=windows $(GOBUILD)/$@ ./cmd/shuttles
	GOARCH=386 GOOS=windows $(GOBUILD)/$@ ./cmd/shuttlec

windows-amd64:
	mkdir -p $(BUILD_DIR)/$@
	GOARCH=amd64 GOOS=windows $(GOBUILD)/$@ ./cmd/shuttles
	GOARCH=amd64 GOOS=windows $(GOBUILD)/$@ ./cmd/shuttlec

windows-arm:
	mkdir -p $(BUILD_DIR)/$@
	GOARCH=arm GOOS=windows $(GOBUILD)/$@ ./cmd/shuttles
	GOARCH=arm GOOS=windows $(GOBUILD)/$@ ./cmd/shuttlec

windows-arm64:
	mkdir -p $(BUILD_DIR)/$@
	GOARCH=arm64 GOOS=windows $(GOBUILD)/$@ ./cmd/shuttles
	GOARCH=arm64 GOOS=windows $(GOBUILD)/$@ ./cmd/shuttlec

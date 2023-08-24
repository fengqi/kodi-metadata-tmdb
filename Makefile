NAME=kodi-tmdb
VERSION=$(shell git describe --tags || echo "dev-master")
RELEASE_DIR=release
GOBUILD=CGO_ENABLED=0 go build -trimpath -ldflags '-w -s -X "main.buildVersion=$(VERSION)"'

PLATFORM_LIST = \
	darwin-amd64 \
	darwin-arm64 \
	linux-amd64 \
	linux-arm64 \
	linux-arm

WINDOWS_ARCH_LIST = windows-amd64

all: linux-amd64 linux-arm64 linux-arm darwin-amd64 darwin-arm64 windows-amd64

darwin-amd64:
	GOARCH=amd64 GOOS=darwin $(GOBUILD) -o $(RELEASE_DIR)/$(NAME)-$@
	cp example.config.json $(RELEASE_DIR)/

darwin-arm64:
	GOARCH=arm64 GOOS=darwin $(GOBUILD) -o $(RELEASE_DIR)/$(NAME)-$@
	cp example.config.json $(RELEASE_DIR)/

linux-amd64:
	GOARCH=amd64 GOOS=linux $(GOBUILD) -o $(RELEASE_DIR)/$(NAME)-$@
	cp example.config.json $(RELEASE_DIR)/

linux-arm64:
	GOARCH=arm64 GOOS=linux $(GOBUILD) -o $(RELEASE_DIR)/$(NAME)-$@
	cp example.config.json $(RELEASE_DIR)/

linux-arm:
	GOARCH=arm GOOS=linux $(GOBUILD) -o $(RELEASE_DIR)/$(NAME)-$@
	cp example.config.json $(RELEASE_DIR)/

windows-amd64:
	GOARCH=amd64 GOOS=windows $(GOBUILD) -o $(RELEASE_DIR)/$(NAME)-$@.exe
	cp example.config.json $(RELEASE_DIR)/

gz_releases=$(addsuffix .gz, $(PLATFORM_LIST))
zip_releases=$(addsuffix .zip, $(WINDOWS_ARCH_LIST))

$(gz_releases): %.gz : %
	chmod +x $(RELEASE_DIR)/$(NAME)-$(basename $@)
	zip -m -j $(RELEASE_DIR)/$(NAME)-$(basename $@)-$(VERSION).zip $(RELEASE_DIR)/$(NAME)-$(basename $@) $(RELEASE_DIR)/example.config.json

$(zip_releases): %.zip : %
	zip -m -j $(RELEASE_DIR)/$(NAME)-$(basename $@)-$(VERSION).zip $(RELEASE_DIR)/$(NAME)-$(basename $@).exe $(RELEASE_DIR)/example.config.json

release: $(gz_releases) $(zip_releases)

clean:
	rm $(RELEASE_DIR)/*

VERSION = 1.0.0
PKG = ./cmd/itags
PLATFORMS = darwin/386 darwin/amd64 linux/386 linux/amd64 linux/arm windows/386 windows/amd64

LDFLAGS = -ldflags="-X main.Version=$(VERSION)"

all: clean test dist

clean:
	@$(RM) -fr dist

test:
	@go test -cover ./...

install:
	@go install $(LDFLAGS) ./cmd/...

version:
	@go run $(LDFLAGS) ./cmd/itags/itags.go --version

dist:
	@gox -verbose $(LDFLAGS) \
		-osarch="$(PLATFORMS)" \
        -output "dist/{{.Dir}}_$(VERSION)_{{.OS}}_{{.Arch}}" $(PKG)

.PHONY: all clean test install version dist

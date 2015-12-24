include build_config.mk

all: build

build:
	$(GO) install ./...

clean:
	@rm -rf bin ./build_config.mk

test:
	$(GO) test ./... -race
.PHONY: build clean install lint link-zokrates migrate mod test zokrates

default: build

build: clean mod zokrates
	go fmt ./...
	go build -v -o ./.bin/api ./cmd/api

clean:
	rm -rf ./.bin 2>/dev/null || true
	go fix ./...
	go clean -i ./...

install: clean
	go install ./...

link-zokrates:
	go tool link -o go-zkp -extld clang -linkmode external -v zokrates.a

lint:
	./ops/lint.sh

migrate: mod
	#go build -v -o ./.bin/migrate ./cmd/migrate
	#./ops/migrate.sh

mod:
	go mod init 2>/dev/null || true
	go mod tidy
	go mod vendor 

test: build
	# no-op

integration: build
	# no-op

zokrates:
	@rm -rf .tmp/zokrates
	@mkdir -p .tmp/
	git clone --single-branch --branch makefile git@github.com:kthomas/zokrates.git .tmp/zokrates
	@pushd .tmp/zokrates && make && popd
	@echo TODO... hoist built zokrates artifacts for linking...

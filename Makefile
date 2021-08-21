.PHONY: all
all:
	@echo "distcache"
	@echo "make <cmd>"
	@echo ""
	@echo "commands:"
	@echo "  lint            - run linter"
	@echo "  proto-compat    - check protobuf for breaking changes"
	@echo "  proto-gen       - generate go files from protobufs"
	@echo "  proto-lint      - lint protobuf files"
	@echo "  test            - run all tests"

.PHONY: lint
lint:
	@golangci-lint run

.PHONY: proto-compat
proto-compat:
	@buf check breaking --against '.git#branch=main'

.PHONY: proto-gen
proto-gen:
	@./scripts/proto/gen_proto.sh

.PHONY: proto-lint
proto-lint:
	@buf lint

.PHONY: test
test:
	@go test -race -cover ./...

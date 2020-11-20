.PHONY: all build

all:
	@echo "imaged"
	@echo "make <cmd>"
	@echo ""
	@echo "commands:"
	@echo "  gen-proto       - generate go files from protobufs"

gen-proto:
	@./scripts/proto/gen_proto.sh

test:
	@go test -race -cover ./...

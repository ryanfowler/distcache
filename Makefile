.PHONY: all
all:
	@echo "imaged"
	@echo "make <cmd>"
	@echo ""
	@echo "commands:"
	@echo "  gen-proto       - generate go files from protobufs"
	@echo "  test            - run all tests"

.PHONY: gen_proto
gen-proto:
	@./scripts/proto/gen_proto.sh

.PHONY: test
test:
	@go test -race -cover ./...

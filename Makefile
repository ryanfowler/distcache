.PHONY: all build

all:
	@echo "imaged"
	@echo "make <cmd>"
	@echo ""
	@echo "commands:"
	@echo "  gen-proto       - generate go files from protobufs"

gen-proto:
	@protoc -I . --go_out=plugins=grpc:grpc/peerpb grpc/peerpb/peer.proto

.PHONY: all waf-gate waf-server clean

BIN_DIR := $(shell pwd)/bin

all: waf-gate waf-server

waf-gate:
	@mkdir -p $(BIN_DIR)
	@cd cmd/waf-gate && go build -o $(BIN_DIR)/waf-gate

waf-server:
	@mkdir -p $(BIN_DIR)
	@cd cmd/waf-server && go build -o $(BIN_DIR)/waf-server

clean:
	@rm -rf $(BIN_DIR)

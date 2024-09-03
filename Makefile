ifneq (,$(wildcard ./.env))
    include .env
    export
endif

build:
	make gen
	go build -o ./bin/app ./cmd/app

run:
	make build
	./bin/app

wire-gen:
	wire ./internal/app/

gen:
	make wire-gen

.PHONY: cert
cert:
	openssl genrsa -out cert/id_rsa 4096
	openssl rsa -in cert/id_rsa -pubout -out cert/id_rsa.pub

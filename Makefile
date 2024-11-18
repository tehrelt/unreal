ifneq (,$(wildcard ./.env))
    include .env
    export
endif

build: gen
	go build -o ./bin/app ./cmd/app

run: build
	./bin/app

wire-gen:
	wire ./internal/app/

gen: wire-gen

.PHONY: cert
cert:
	openssl genrsa -out cert/id_rsa 4096
	openssl rsa -in cert/id_rsa -pubout -out cert/id_rsa.pub

DSA_PRIVATE = cert/id_dsa
DSA_PUBLIC = cert/id_dsa.pub
dsa-keygen:  
	openssl genpkey -algorithm Ed25519 -out $(DSA_PRIVATE)
	openssl pkey -in $(DSA_PRIVATE) -pubout -out $(DSA_PUBLIC)

clean:
	rm ./internal/app/*gen.go #

.DEFAULT_GOAL := run
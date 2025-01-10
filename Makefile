CONFIG_PATH := $(USERPROFILE)\certs

.PHONY: init

init:
	mkdir ${CONFIG_PATH}

.PHONY: gencert

gencert:
	cfssl gencert \
		-initca cert/ca-csr.json | cfssljson -bare ca
	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=cert/ca-config.json \
		-profile=server \
		cert/server-csr.json | cfssljson -bare server
	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=cert/ca-config.json \
		-profile=client \
 		cert/client-csr.json | cfssljson -bare client
	move *.pem ${CONFIG_PATH}
	move *.csr ${CONFIG_PATH}

compile:
	protoc spec/*.proto \
		--go_out=. \
		--go-grpc_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative \
		--proto_path=.
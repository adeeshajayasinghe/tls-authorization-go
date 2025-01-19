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
		-cn="root" \
 		cert/client-csr.json | cfssljson -bare root-client
	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=cert/ca-config.json \
		-profile=client \
		-cn="read-only" \
 		cert/client-csr.json | cfssljson -bare read-only-client
	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=cert/ca-config.json \
		-profile=client \
		-cn="nobody" \
 		cert/client-csr.json | cfssljson -bare nobody-client
	move *.pem ${CONFIG_PATH}
	move *.csr ${CONFIG_PATH}

$(CONFIG_PATH)\model.conf:
	copy cert\model.conf "$(CONFIG_PATH)\model.conf"

$(CONFIG_PATH)\policy.csv:
	copy cert\policy.csv "$(CONFIG_PATH)\policy.csv"

.PHONY: genacl
genacl: $(CONFIG_PATH)\model.conf $(CONFIG_PATH)\policy.csv
	echo "Access control lists configured."

.PHONY: compile
compile:
	protoc spec/*.proto \
		--go_out=. \
		--go-grpc_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative \
		--proto_path=.

.PHONY: test
test:
	go test -race ./server
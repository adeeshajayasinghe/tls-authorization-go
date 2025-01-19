# Secure gRPC Logging Service with Mutual TLS in Golang

This project demonstrates building a secure gRPC server in Golang using mutual TLS (mTLS). The server supports producing and consuming in-memory logs, ensuring authenticated and encrypted communication between client and server.

## Prerequisites
- **protoc compiler**: Install the Protocol Buffers compiler.
  - On macOS:  
    ```bash
    brew install protobuf
    ```
  - On Linux (Ubuntu):  
    ```bash
    sudo apt install -y protobuf-compiler
    ```
  - On Windows: Download from the [Protocol Buffers releases](https://github.com/protocolbuffers/protobuf/releases).

## Installing Dependencies
Run the following command to install all project dependencies:
```bash
go mod tidy
```

## Project Structure
- auth/: Authorization logic in Go
- spec/: Protocol Buffers file defining gRPC service and messages
- cert/: Directory for ca configs, certificates requests and authorization model and policy files
- log/: logging functionality implementation in Go
- server/: gRPC server implementation in Go
- config/: Configurations for mutual TLS

## Running the Project
1. Compile protocol buffers
```
make compile
```
2. Initialize certificates storing directory
```
make init
```
3. Generate certificates 
```
make gencert
```
4. copy authorization model and policy files to certificate directory
```
make genacl
```
5. Test the mutula TLS
```
make test
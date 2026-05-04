.PHONY: lint
lint: # запустить линтер
	golangci-lint run --config=.golangci.yml ./...

.PHONY: test
test: # запустить тесты
	go test -v ./...

.PHONY: build
build: # собрать приложение
	go build -o bin/app cmd/main.go

.PHONY: run
run: # запустить приложение
	go run cmd/main.go

.PHONY: clean
clean: # очистить временные файлы
	rm -rf bin/

.PHONY: install-tools
install-tools:
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

.PHONY: generate-swagger swagger docs
# Повторная генерация OpenAPI (каталог docs/) из swag-комментариев в коде.
generate-swagger:
	go run github.com/swaggo/swag/cmd/swag@latest init -g ./cmd/main.go -o docs -d . --parseInternal

swagger: generate-swagger ## удобный алиас
docs: generate-swagger     ## удобный алиас

.PHONY: generate-proto
generate-proto:
	protoc \
		--proto_path=api/proto \
		--proto_path=third_party \
		--go_out=pkg/api/proto \
		--go_opt=paths=source_relative \
		--go-grpc_out=pkg/api/proto \
		--go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=pkg/api/proto \
		--grpc-gateway_opt=paths=source_relative \
		api/proto/*.proto 
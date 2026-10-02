.PHONY: run test swagger web-dev web-build

run:
	go run ./cmd/api/*.go

test:
	go test ./cmd/... ./services/... ./platform/...

swagger:
	swag init -g cmd/api/main.go -o docs

web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build
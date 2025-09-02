.PHONY: test

test:
	@go mod tidy && go run cmd/main.go
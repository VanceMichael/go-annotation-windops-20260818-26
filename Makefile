.PHONY: test race vet build run web-install web-test web-typecheck web-build check

test:
	GOTOOLCHAIN=local go test ./... -count=1
race:
	GOTOOLCHAIN=local go test -race ./... -count=1
vet:
	GOTOOLCHAIN=local go vet ./...
build:
	GOTOOLCHAIN=local go build ./...
run:
	GOTOOLCHAIN=local go run ./cmd/server
web-install:
	cd web && npm ci
web-test:
	cd web && npm test -- --run
web-typecheck:
	cd web && npm run typecheck
web-build:
	cd web && npm run build
check: test race vet build web-test web-typecheck web-build


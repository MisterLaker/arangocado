SHELL := /bin/bash

define compile
	@printf "\nBuilding: $(1)\n"

	@time CGO_ENABLED=0 GOOS=linux go build -installsuffix cgo -o build/$(1) cmd/$(1)/main.go

	@printf "\nBuilt size: "
	@ls -lah build/$(1) | awk '{print $$5}'
	@printf "\nDone building: $(1)\n\n"
endef

.PHONY: build
build:
	 $(call compile,arangocado)

.PHONY: test
test:
	@go test -v ./...

container:
	docker-buildx build \
		--platform linux/amd64,linux/arm64/v8 \
		-t mrlaker/arangocado:0.1.0 \
		--push \
		.

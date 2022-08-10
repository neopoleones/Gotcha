.PHONY: build
rbuild:
	@echo "Building ..."

	/opt/homebrew/opt/go@1.18/bin/go build Gotcha/cmd/gotcha-app
	./gotcha-app $(ARGS)

clean:
	rm gotcha-app

tests:
	go test -race -v -timeout 30s ./...

DEFAULT_GOAL := rbuid
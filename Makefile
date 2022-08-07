.PHONY: build
rbuild:
	@echo "Building ..."
	go build Gotcha/cmd/gotcha-app
	./gotcha-app $(ARGS)

clean:
	rm gotcha-app

tests:
	go test -race -v -timeout 30s ./...

DEFAULT_GOAL := rbuid
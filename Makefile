.PHONY: build
rbuild:
	@echo "Building ..."
	go build Gotcha/cmd/gotcha-app
	./gotcha-app $(ARGS)

clean:
	rm gotcha-app

DEFAULT_GOAL := rbuid
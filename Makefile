tidy:
	go mod tidy
build:
	go build -o adoc-preprocess main.go
install:
	sudo mv adoc-preprocess /usr/local/bin/
tidy:
	go mod tidy
clean:
	rm adoc
build:
	go build -o adoc main.go
install:
	sudo mv adoc-preprocess /usr/local/bin/
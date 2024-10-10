tidy:
	go mod tidy
clean:
	rm adoc
build:
	go build -o adoc main.go
install:
	sudo mv adoc /usr/local/bin/
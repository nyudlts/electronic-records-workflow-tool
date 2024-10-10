tidy:
	go mod tidy
clean:
	rm adoc
build:
	go build ain.go
install:
	sudo mv adoc /usr/local/bin/
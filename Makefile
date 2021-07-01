build: test
	mkdir -p target
	go build -o target/goauth

test:
	go test

clean:
	rm -rf target/

image: build
	docker build . -t goauth:latest

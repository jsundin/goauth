build:
	mkdir -p target
	go build -o target/goauth

clean:
	rm -rf target/

image: build
	docker build . -t goauth:latest

all: clean build run

build: clean
	go build -o server .

clean:
	rm -f server

run:
	./server
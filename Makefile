VERSION=latest
SRCDIR=src/minibank



bin/minibank: $(shell find $(SRCDIR) -name '*.go')
	docker run -rm -it -v `pwd`:/usr/app \
		-w /usr/app \
		-e GOPATH=/usr/app \
		-e CGO_ENABLED=0 \
		-e GOOS=linux \
		golang:1.9 sh -c 'go get minibank && go build -ldflags "-extldflags -static" -o $@ minibank'

minibank: bin/minibank
	docker build -t minibank:$(VERSION) -f Dockerfile bin

mysql:
	docker pull mariadb:latest

run-images: minibank mysql
	./run.sh


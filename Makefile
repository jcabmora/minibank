VERSION=latest
SRCDIR=src/minibank
PROJECT=$(shell gcloud config get-value project)
REGISTRY=gcr.io/$(PROJECT)


bin/minibank: $(shell find $(SRCDIR) -name '*.go')
	docker run --rm -it -v `pwd`:/usr/app \
		-w /usr/app \
		-e GOPATH=/usr/app \
		-e CGO_ENABLED=0 \
		-e GOOS=linux \
		golang:1.9 sh -c 'go get minibank && go build -ldflags "-extldflags -static" -o $@ minibank'

minibank: bin/minibank
	docker build -t minibank:$(VERSION) -f Dockerfile bin
	docker tag minibank:$(VERSION) $(REGISTRY)/minibank:$(VERSION)

mysql:
	docker build -t mariadb:$(VERSION) -f Dockerfile-MariaDB .
	docker tag mariadb:$(VERSION) $(REGISTRY)/mariadb:$(VERSION)

run-images: minibank mysql
	./run.sh

push-images: minibank mysql
	gcloud docker -- push $(REGISTRY)/minibank:$(VERSION)
	gcloud docker -- push $(REGISTRY)/mariadb:$(VERSION)


VERSION=latest
SRCDIR=src/minibank
PROJECT=$(shell gcloud config get-value project)
REGISTRY=gcr.io/$(PROJECT)
REPLICAS ?= 1


bin/minibank: $(shell find $(SRCDIR) -name '*.go')
	docker run --rm -it -v `pwd`:/usr/app \
		-w /usr/app \
		-e GOPATH=/usr/app \
		-e CGO_ENABLED=0 \
		-e GOOS=linux \
		golang:1.9 sh -c 'go get minibank && go build -ldflags "-extldflags -static" -o $@ minibank'

.PHONY: minibank
minibank: bin/minibank
	docker build -t minibank:$(VERSION) -f Dockerfile bin
	docker tag minibank:$(VERSION) $(REGISTRY)/minibank:$(VERSION)

.PHONY: mariadb
mariadb:
	docker build -t mariadb:$(VERSION) -f mariadb/Dockerfile .
	docker tag mariadb:$(VERSION) $(REGISTRY)/mariadb:$(VERSION)

.PHONY: run-images
run-images: minibank mariadb
	./run.sh

.PHONY: push-images
push-images: minibank mariadb
	gcloud docker -- push $(REGISTRY)/minibank:$(VERSION)
	gcloud docker -- push $(REGISTRY)/mariadb:$(VERSION)


.PHONY: prepare-venv
prepare-venv:
	virtualenv venv
	venv/bin/pip install -r load_test/requirements.txt

.PHONE: load-test
load-test: prepare-venv
	venv/bin/python load_test/collect.py --endpoint api/account/login --replicas $(REPLICAS) --tag $(TAG) --payload load_test/payload.json


.PHONY: clean
clean:
	-rm -rf venv

VERSION=latest
PROJECT=$(shell gcloud config get-value project)
REGISTRY=gcr.io/$(PROJECT)
REPLICAS ?= 1

.PHONY: update-project
update-project:
	for i in kubernetes/*.yaml; do sed -i "s/<YOUR_PROJECT_ID>/$(PROJECT)/g" $$i; done

.PHONY: mariadb
mariadb:
	docker build -t mariadb:$(VERSION) -f mariadb/Dockerfile .
	docker tag mariadb:$(VERSION) $(REGISTRY)/mariadb:$(VERSION)

.PHONY: push-images
push-images: mariadb
	gcloud docker -- push $(REGISTRY)/mariadb:$(VERSION)


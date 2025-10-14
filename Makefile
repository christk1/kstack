# Minimal Makefile for example-app demo flow

DEMO_IMAGE ?= my-app:latest
KIND_CLUSTER ?= kstack

.PHONY: demo-build demo-load demo-deploy

demo-build:
	docker build -t $(DEMO_IMAGE) ./pkg/addons/exampleapp/demo

demo-load: demo-build
	kind load docker-image $(DEMO_IMAGE) --name $(KIND_CLUSTER)

demo-deploy: demo-load
	# split DEMO_IMAGE into repo and tag (defaults tag to 'latest' if missing)
	$(eval _repo := $(shell echo $(DEMO_IMAGE) | sed -E 's/:.*$$//'))
	$(eval _tag := $(shell echo $(DEMO_IMAGE) | sed -E 's/^.*://'))
	if [ "x$(_tag)" = "x$(DEMO_IMAGE)" ]; then _tag=latest; fi; \
	./kstack addons install example-app --set image.repository=$(_repo) --set image.tag=$(_tag)

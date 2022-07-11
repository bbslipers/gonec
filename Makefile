PACKAGE_NAME := github.com/shinanca/gonec
GOLANG_CROSS_VERSION ?= v1.18

DOCKER_RUN=docker run \
	--rm \
	--privileged \
	-e CGO_ENABLED=1 \
	-v /var/run/docker.sock:/var/run/docker.sock \
	-v `pwd`:/go/src/$(PACKAGE_NAME) \
	-w /go/src/$(PACKAGE_NAME) \
	goreleaser/goreleaser-cross:${GOLANG_CROSS_VERSION} \

build:
	@$(DOCKER_RUN) build --snapshot $(flags)

snapshot:
	@$(DOCKER_RUN) release --snapshot $(flags)

release:
	@$(DOCKER_RUN) release --rm-dist $(flags)
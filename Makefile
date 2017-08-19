DOCKER_IMAGE := fiunchinho/dmz-controller
DOCKER_TAG   := latest
K8S_NAMESPACE:= default

.PHONY: build deps lint package test coverage helm publish

build:
	go build -i -o dmz-controller

deps:
	glide install

lint:
	gometalinter.v1 --install --update
	gometalinter.v1 --vendor --disable-all -E vet -E goconst -E golint -E goimports -E misspell --deadline=50s -j 11 "${PWD}/..."

package:
	env GOOS=linux go build -o dmz-controller
	docker build -t "${DOCKER_IMAGE}":"${DOCKER_TAG}" "."

test:
	go test `glide novendor`

coverage:
	bin/coverage

helm: package
	helm upgrade --install --namespace="${K8S_NAMESPACE}" "dmz-controller" "./helm/dmz-controller"

publish: package
	docker login --username "${DOCKER_USER}" --password "${DOCKER_PASS}"
	docker build -t "${DOCKER_IMAGE}" "."
	docker tag "${DOCKER_IMAGE}" "${DOCKER_IMAGE}:${DOCKER_TAG}"
	docker push "${DOCKER_IMAGE}"

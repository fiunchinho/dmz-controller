APPLICATION  := dmz-controller
LINUX        := release/${APPLICATION}-linux-amd64
DARWIN       := release/${APPLICATION}-darwin-amd64
DOCKER_IMAGE := fiunchinho/dmz-controller
K8S_NAMESPACE:= default
BIN_DIR      := $(GOPATH)/bin
GOMETALINTER := $(BIN_DIR)/gometalinter
GLIDE        := $(BIN_DIR)/glide
COVER        := $(BIN_DIR)/gocov-xml
VERSION      ?= latest


$(DARWIN):
	GOOS=linux GOARCH=amd64 go build -i -o ${DARWIN}

$(LINUX):
	GOOS=linux GOARCH=amd64 go build -i -o ${LINUX}

.PHONY: deps
deps: $(GLIDE)
	glide install

$(GOMETALINTER):
	go get -u gopkg.in/alecthomas/gometalinter.v1
	gometalinter.v1 --install --update

$(GLIDE):
	go get -u github.com/Masterminds/glide

$(COVER):
	go get -u github.com/axw/gocov/gocov
	go get -u github.com/AlekSi/gocov-xml

.PHONY: lint
lint: $(GOMETALINTER)
	gometalinter.v1 --vendor --disable-all -E vet -E goconst -E golint -E goimports -E misspell --deadline=50s -j 11 "${PWD}/..."

.PHONY: test
test: lint
	go test `glide novendor`

.PHONY: coverage
coverage: $(COVER) lint
	bin/coverage

.PHONY: helm
helm:
	helm upgrade --install --namespace="${K8S_NAMESPACE}" "${APPLICATION}" "./helm/${APPLICATION}"

.PHONY: release
release: $(LINUX)
	docker login --username "${DOCKER_USER}" --password "${DOCKER_PASS}"
	docker build -t "${DOCKER_IMAGE}" "."
	docker tag "${DOCKER_IMAGE}" "${DOCKER_IMAGE}:${VERSION}"

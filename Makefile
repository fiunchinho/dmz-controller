APP          = dmz-controller
DOCKER_IMAGE = fiunchinho/dmz-controller

.PHONY: compile lint package run test

compile: deps
	docker run --rm -w "/go/src/${APP}" \
		-v "${PWD}:/go/src/${APP}" \
		-e "GOPATH=/go/src/${APP}/Godeps/_workspace:/go" \
		-e "CGO_ENABLED=0" \
		-e "GOOS=linux" \
		golang:alpine \
		go build -v -o ${APP}

deps:
	glide install

lint: deps
	gometalinter.v1 --install --update
	gometalinter.v1 --checkstyle --vendor --disable-all -E goimports -E vet -E goconst -E staticcheck -E structcheck -E unparam -E unused -E vetshadow -E varcheck --deadline=50s -j 11 "${PWD}/..." | tee /dev/tty > checkstyle-report.xml; test ${PIPESTATUS[0]} -eq 0

package: compile
	docker build -t "${DOCKER_IMAGE}" "."

run: package
	docker run --rm -d "${DOCKER_IMAGE}"

test: deps
	docker run --rm -w "/go/src/${APP}" \
		-v "${PWD}:/go/src/${APP}" \
		-e "GOPATH=/go/src/${APP}/Godeps/_workspace:/go" \
		-e "CGO_ENABLED=0" \
		-e "GOOS=linux" \
		golang:alpine \
		go test -v -cover $(glide novendor)

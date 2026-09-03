# This how we want to name the binary output
BINARY=email-service
GOOS=linux
GOARCH=amd64
DOCKER_HUB=ghcr.io/selftrack

# These are the values we want to pass for VERSION and BUILD
VERSION=`git describe --tags --always --dirty`
BUILD=`git rev-parse HEAD`

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.Build=${BUILD}"

# Builds the project
build: clean prepare
	CGO_ENABLED=0 GOPRIVATE=github.com/selftrack/* GOOS=${GOOS} GOARCH=${GOARCH} go build ${LDFLAGS} -o builds/${VERSION}/${BINARY} cmd/main.go

dist: build
	rm -rf dist/${BINARY}-${VERSION}
	mkdir -p dist/${BINARY}-${VERSION}
	mkdir -p dist/${BINARY}-${VERSION}/bin
	mkdir -p dist/${BINARY}-${VERSION}/conf
	cp builds/${VERSION}/${BINARY} dist/${BINARY}-${VERSION}/bin
	tar -czvf dist/${BINARY}-${VERSION}.tar.gz dist/${BINARY}-${VERSION}

# Installs our project: copies binaries
install:
	go install ${LDFLAGS}

# Cleans our project: deletes binaries
clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

prepare:
	mkdir -p builds/${VERSION}

docker: build
	docker build --build-arg version=${VERSION} --build-arg ms=${BINARY} -t ${DOCKER_HUB}/${BINARY}:${VERSION} .
	docker push ${DOCKER_HUB}/${BINARY}:${VERSION}


docker-prod: build
	docker build --build-arg version=${VERSION} --build-arg ms=${BINARY} -t ${DOCKER_HUB}/${BINARY}:latest .
	docker push ${DOCKER_HUB}/${BINARY}:latest

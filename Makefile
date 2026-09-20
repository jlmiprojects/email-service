# This how we want to name the binary output
BINARY=email-service
GOOS=linux
GOARCH=amd64
DOCKER_HUB=ghcr.io/jlmiprojects

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

# The Dockerfile builds the binary itself (no host build needed first) — see
# its header comment. VERSION/BUILD_SHA are passed through so the shipped
# binary reports the right version instead of always "0.0.0".
# PLATFORM is the *deploy target*, not the build host. A plain `docker build`
# produces an image for whatever architecture the builder is, so an Apple
# Silicon machine yields linux/arm64 — which the x86 servers refuse to start
# with "exec format error". Override for another target, or pass a list for a
# multi-arch manifest:
#   make docker PLATFORM=linux/amd64,linux/arm64
PLATFORM?=linux/amd64

# buildx --push rather than build-then-push: a cross-platform image cannot be
# loaded into the local docker image store, so it goes straight to the registry.
docker:
	docker buildx build --platform ${PLATFORM} --build-arg VERSION=${VERSION} --build-arg BUILD_SHA=${BUILD} -t ${DOCKER_HUB}/${BINARY}:${VERSION} --push .

docker-prod:
	docker buildx build --platform ${PLATFORM} --build-arg VERSION=${VERSION} --build-arg BUILD_SHA=${BUILD} -t ${DOCKER_HUB}/${BINARY}:latest --push .

run:
	go run cmd/main.go

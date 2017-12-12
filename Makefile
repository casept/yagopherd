# Name of the output binary
BINARY=yagopherd
# Get version and git commit hash
VERSION=0.0.1
COMMIT=`git rev-parse HEAD`

# Pass variables to executable via linker flags
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.Commit=${COMMIT}"

# Build the binary as default target
.DEFAULT_GOAL: ${BINARY}

# Build the binary
${BINARY}:
	go build ${LDFLAGS} -o ${BINARY} -v

# Install the binary into $GOPATH/bin
install:
	go install ${LDFLAGS} -o ${BINARY} -v
# Remove the binary
clean:
	if [ -f ${BINARY} ] ; then go clean; fi

# Run tests
test:
	go test -v .

# Just so the binary can be run easily from vim
run: ${BINARY}
	./${BINARY}

.PHONY: clean install test run

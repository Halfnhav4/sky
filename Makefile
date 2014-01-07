CFLAGS=`llvm-config-3.4 --cflags`
LDFLAGS="`llvm-config-3.4 --ldflags` -Wl,-L`llvm-config-3.4 --libdir` -lLLVM-`llvm-config-3.4 --version`"
COVERPROFILE=/tmp/c.out
TEST=.
PKG=./...

all: test

flags:
	@echo "CGO_CFLAGS=$(CFLAGS) CGO_LDFLAGS=$(LDFLAGS)"

grammar:
	${MAKE} -C query/parser

test: grammar
	CGO_CFLAGS=$(CFLAGS) CGO_LDFLAGS=$(LDFLAGS) go test -v -test.run=$(TEST) $(PKG)

cover: fmt
	CGO_CFLAGS=$(CFLAGS) CGO_LDFLAGS=$(LDFLAGS) go test -v -test.run=$(TEST) -coverprofile=$(COVERPROFILE) $(PKG)
	go tool cover -html=$(COVERPROFILE)
	rm $(COVERPROFILE)

bench: grammar
	CGO_CFLAGS=$(CFLAGS) CGO_LDFLAGS=$(LDFLAGS) go test -v -test.bench=. $(PKG)

build: grammar
	CGO_CFLAGS=$(CFLAGS) CGO_LDFLAGS=$(LDFLAGS) go build .

install: grammar
	CGO_CFLAGS=$(CFLAGS) CGO_LDFLAGS=$(LDFLAGS) go install .

.PHONY: test

# generated from spec: libaql.spec.md sha256:2fafcf9ca530438a6fb11f674789ae149aa9e4e48961b854a40a93e0c0cf6a6f
#
# Makefile of libaql, the Go binding. The tools of the build are
# lemongrass and re2c; the package build calls `make build` and
# `make test` and needs no network.

SPEC_SHA256 = 2fafcf9ca530438a6fb11f674789ae149aa9e4e48961b854a40a93e0c0cf6a6f
VERSION    := $(shell cat VERSION)
NAME        = libaql

# Tools of the build and their pinned versions.
LEMONGRASS          ?= lemongrass
LEMONGRASS_TEMPLATE ?= /usr/share/lemongrass/lempar.go
LEMONGRASS_VERSION   = 0.1.1
RE2C                ?= re2c
RE2C_VERSION         = 4.6
SPEC                ?= ../../specification/libaql.spec.md
GO                  ?= go

LEMONGRASS_FLAGS = -q -l -T$(LEMONGRASS_TEMPLATE)
RE2C_FLAGS       = --lang go -8 -i --no-generation-date

.PHONY: all generate check-tools check-generated check-spec build test clean

all: build

# The tools must have the pinned versions.
check-tools:
	@test "$$($(LEMONGRASS) -x | sed -n 's/^lemongrass version //p')" = "$(LEMONGRASS_VERSION)" || \
	  { echo "check-tools: lemongrass $(LEMONGRASS_VERSION) is needed" >&2; exit 1; }
	@test "$$($(RE2C) --version)" = "re2c $(RE2C_VERSION)" || \
	  { echo "check-tools: re2c $(RE2C_VERSION) is needed" >&2; exit 1; }

# Writes grammar.go and lexer.go again. Run it after a change of grammar.y,
# lexer.re or tokens.re, and commit the result.
generate: check-tools
	$(LEMONGRASS) $(LEMONGRASS_FLAGS) grammar.y
	$(RE2C) $(RE2C_FLAGS) lexer.re -o lexer.go

# The generated files in this directory must be what the tools write now.
check-generated: check-tools
	@tmp=$$(mktemp -d) && \
	$(LEMONGRASS) $(LEMONGRASS_FLAGS) -d$$tmp grammar.y && \
	$(RE2C) $(RE2C_FLAGS) lexer.re -o $$tmp/lexer.go && \
	cmp grammar.go $$tmp/grammar.go && cmp lexer.go $$tmp/lexer.go; \
	status=$$?; rm -rf $$tmp; exit $$status

# The token table and the grammar must be those of the specification.
check-spec:
	@LEMONGRASS="$(LEMONGRASS)" sh check-spec.sh $(SPEC)

build: check-generated check-spec
	$(GO) build ./...

test: build
	$(GO) test ./...

clean:
	$(GO) clean ./...

.PHONY: dist doc

# The API reference, from the documentation comments.
doc:
	$(GO) doc -all .

# The release source tarball. It keeps the layout of the repository:
# specification/libaql.spec.md and go/aql/.
dist:
	tar -C ../.. --exclude='*.tar.gz' --transform 's,^,$(NAME)-$(VERSION)/,' \
	  -czf $(NAME)-$(VERSION).tar.gz specification/libaql.spec.md go/aql

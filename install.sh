#!/bin/sh
command -v go 1>/dev/null 2>/dev/null ||
	( echo 2>&1 'the "go" command must be installed but was not found, exiting' && \
		exit 1 )

# This is bad and the original author feels bad.
find . -name '*.go' -exec grep -E -h '^\s".*\.(com|net|org)\/' {} \;  | \
	sort | uniq | sed 's/"//g' | while read -r pkg; do go get -u $pkg ; done && \
		go install github.com/aoeu/acme/...

go get -u golang.org/x/tools/cmd/goimports 2>/dev/null
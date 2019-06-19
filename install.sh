#!/bin/sh
find . -name '*.go' -exec grep -E -h '^\s".*\.(com|net|org)\/' {} \;  | \
	sort | uniq | sed 's/"//g' | while read -r pkg; do go get -u $pkg ; done && \
		go install github.com/aoeu/acme/...
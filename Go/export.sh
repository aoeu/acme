#!/bin/sh

echoerr() {
	echo $* 1>&2
}

test -z "$GOPATH" && echoerr "GOPATH is not set, exiting" && exit 1

targetDir="$GOPATH/src/github.com/aoeu/acme"
test ! -d "$targetDir" && echoerr "$targetDir does not exist, exiting" && exit 1

targetDir="$targetDir/Go"
test ! -d "$targetDir" && mkdir "$targetDir"

find .  ! -path "./.git*" ! -path "." -exec cp -a {} $targetDir/{} \;

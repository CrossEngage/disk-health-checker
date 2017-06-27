#!/bin/bash
version=`git describe --all --always --dirty --long`
printf "package main\nconst (\nversion=\`$version\`\n)\n" | gofmt | tee g_version.go

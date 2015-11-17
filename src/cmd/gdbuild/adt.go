package main

import (
	"fmt"
	"log"

	"pault.ag/go/debian/control"
)

func autopkgtest(from string, dsc control.DSC) string {
	img := fmt.Sprintf("gdbuild/adt:%s_%s", dsc.Source, scrubForDockerTag(dsc.Version.String()))

	dockerfile := fmt.Sprintf("FROM %s\n", from)

	dockerfile += `
USER root

RUN apt-get update && apt-get install -y \
		autopkgtest \
	&& rm -rf /var/lib/apt/lists/*

RUN adt-run \
		--changes .out/*.changes \
		--apt-upgrade \
		--- null \
	&& rm -rf /var/lib/apt/lists/*
`

	err := dockerBuild(img, dockerfile)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	return img
}

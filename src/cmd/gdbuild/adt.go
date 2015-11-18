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

# use adt-virt-chroot instead of adt-virt-null so that it doesn't think it has "isolation-machine" capability
RUN adt-run \
		--changes .out/*.changes \
		--apt-upgrade \
		--- chroot / \
	; code=$? \
	&& rm -rf /var/lib/apt/lists/* \
	&& case "$code" in \
		2) echo >&2 'WARNING: some tests were skipped!' ;; \
		*) exit "$code" ;; \
	esac
`

	err := dockerBuild(img, dockerfile)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	return img
}

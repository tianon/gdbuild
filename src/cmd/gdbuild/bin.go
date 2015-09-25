package main

import (
	"fmt"
	"log"
	"path/filepath"
	"sort"
	"strings"

	"aptsources"

	"pault.ag/go/debian/control"
	"pault.ag/go/debian/dependency"
)

func binSatPossi(depArch *dependency.Arch, bin control.BinaryIndex, possi dependency.Possibility) bool {
	return !possi.Substvar &&
		(possi.Architectures == nil || possi.Architectures.Matches(depArch)) &&
		possi.Name == bin.Package &&
		(possi.Arch == nil || possi.Arch.Is(&bin.Architecture)) &&
		(possi.Version == nil || possi.Version.SatisfiedBy(bin.Version))
}

func buildBin(dscFile string) (control.DSC, string) {
	dscDir := filepath.Dir(dscFile)
	dsc, err := control.ParseDscFile(dscFile)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	if err := dsc.Validate(); err != nil {
		log.Fatalf("error, validation failed: %v\n", err)
	}

	dscMd5, err := md5sum(dscFile)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}
	img := fmt.Sprintf("gdbuild/bin:%s_%s", dsc.Source, dscMd5)

	// TODO parse this information from an image?  optional commandline parameters?
	suite := "unstable"
	sources := aptsources.SuiteSources(suite, "main")
	arch := "amd64"
	sources = sources.Append(aptsources.Source{
		Types:      []string{"deb", "deb-src"},
		URIs:       []string{"http://incoming.debian.org/debian-buildd"},
		Suites:     []string{"buildd-" + suite},
		Components: []string{"main"},
	})

	index, err := sources.FetchCandidates(arch)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	depArch, err := dependency.ParseArch(arch)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	buildEssential, err := dependency.Parse("build-essential, dpkg-dev, fakeroot")
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	allCan := true
	bins := map[string]control.BinaryIndex{}
	binsSlice := []string{} // ugh Go
RelLoop:
	for _, rel := range append(buildEssential.Relations, dsc.BuildDepends.Relations...) {
		canRel := false
		for _, possi := range rel.Possibilities {
			if possi.Substvar {
				continue
			}
			if bin, ok := bins[possi.Name]; ok && binSatPossi(depArch, bin, possi) {
				continue RelLoop
			}
		}
	PossiLoop:
		for _, possi := range rel.Possibilities {
			if possi.Substvar {
				continue
			}
			entries, ok := map[string][]control.BinaryIndex(*index)[possi.Name]
			if !ok {
				continue
			}
			for _, bin := range entries {
				if binSatPossi(depArch, bin, possi) {
					if existBin, ok := bins[bin.Package]; ok {
						log.Printf("uh oh, already chose %s=%s but want %s=%s for %q\n", existBin.Package, existBin.Version, bin.Package, bin.Version, possi)
						continue PossiLoop
					}
					bins[bin.Package] = bin
					binsSlice = append(binsSlice, bin.Package)
					canRel = true
					break PossiLoop
				}
			}
		}
		if !canRel {
			log.Printf("warning: unable to satisfy %q\n", rel)
			allCan = false
		}
	}
	sort.Strings(binsSlice)

	if !allCan {
		//log.Fatalf("Unsatisfied possi; exiting.\n")
		log.Println()
		log.Println("WARNING: Unsatisfied possi!")
		log.Println()
	}

	dockerfile := fmt.Sprintf("FROM debian:%s\n", suite)
	// TODO allow this to instead be "FROM scratch\nADD some-chroot-tarball.tar.* /\n"

	// see https://sources.debian.net/src/pbuilder/jessie/pbuilder-modules/#L306
	// and https://sources.debian.net/src/pbuilder/jessie/pbuilder-modules/#L408
	dockerfile += `
# setup environment configuration
RUN { echo '#!/bin/sh'; echo 'exit 101'; } > /usr/sbin/policy-rc.d \
	&& chmod +x /usr/sbin/policy-rc.d
RUN echo 'APT::Install-Recommends "false";' > /etc/apt/apt.conf.d/15gdbuild
`

	// setup sources.list explicitly -- don't trust the tarball/base image
	dockerfile += fmt.Sprintf(`
# setup sources.list
RUN find /etc/apt/sources.list.d -type f -exec rm -v '{}' + \
	&& echo %q | tee /etc/apt/sources.list >&2
`, sources.ListString())

	// TODO configurable
	eatMyData := true

	eatMyDataPrefix := ""
	if eatMyData {
		eatMyDataPrefix = "eatmydata "
		dockerfile += `
RUN apt-get update && apt-get install -y \
		eatmydata \
	&& rm -rf /var/lib/apt/lists/*
`
	}

	dockerfile += fmt.Sprintf(`
RUN %sapt-get update && %sapt-get install -y \
`, eatMyDataPrefix, eatMyDataPrefix)
	for _, pkg := range binsSlice {
		bin := bins[pkg]
		dockerfile += fmt.Sprintf("\t\t%s=%s \\\n", bin.Package, bin.Version)
	}
	dockerfile += "\t&& rm -rf /var/lib/apt/lists/*\n"

	files := []string{dsc.Filename}
	for _, f := range dsc.Files {
		files = append(files, filepath.Join(dscDir, f.Filename))
	}

	dockerfile += "COPY"
	for _, f := range files {
		dockerfile += " " + filepath.Base(f)
	}
	dockerfile += " /usr/src/.in/\n"

	buildCommand := fmt.Sprintf("%sdpkg-buildpackage -uc -us -d", eatMyDataPrefix)

	// TODO make "SASBS" configurable
	buildCommand = strings.Join([]string{
		buildCommand + " -S",
		buildCommand + " -A",
		buildCommand + " -S",
		buildCommand + " -B",
		buildCommand + " -S",
	}, " && ")

	dockerfile += fmt.Sprintf(`
WORKDIR /usr/src
RUN chown -R nobody:nogroup .
USER nobody:nogroup
RUN dpkg-source -x %q pkg
RUN (cd pkg && set -x && %s) \
	&& mkdir .out \
	&& { \
		echo *.changes; \
		awk '$1 == "Files:" { files = 1; next } /^ / && files { print $5 } /^[^ ]/ { files = 0 }' *.changes; \
		echo .out/; \
	} | xargs ln -v
`, ".in/"+filepath.Base(dsc.Filename), buildCommand)

	err = dockerBuild(img, dockerfile, files...)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	return *dsc, img
}

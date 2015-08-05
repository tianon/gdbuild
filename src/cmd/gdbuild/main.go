package main

import (
	"fmt"
	"log"
	"os"

	"pault.ag/go/debian/control"
	"pault.ag/go/debian/dependency"
	"pault.ag/go/resolver"
)

func main() {
	log.SetFlags(log.Lshortfile)

	con, err := control.ParseControlFile("debian/control")
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	suite := "unstable"
	arch := "amd64"

	index, err := resolver.GetBinaryIndex(
		"http://httpredir.debian.org/debian",
		suite,
		"main",
		arch,
	)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	depArch, err := dependency.ParseArch(arch)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	buildEssential, err := dependency.Parse("build-essential")
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	allCan := true
	allPossi := append(
		buildEssential.GetPossibilities(*depArch),
		append(
			con.Source.BuildDepends.GetPossibilities(*depArch),
			con.Source.BuildDependsIndep.GetPossibilities(*depArch)...,
		)...,
	)
	allBins := []control.BinaryIndex{}
	for _, possi := range allPossi {
		can, why, bins := index.ExplainSatisfies(*depArch, possi)
		if !can {
			log.Printf("%s: %s\n", possi.Name, why)
			allCan = false
		} else {
			// TODO more smarts for which dep out of bins to use
			allBins = append(allBins, bins[0])
		}
	}

	if !allCan {
		log.Fatalf("Unsatisfied possi; exiting.\n")
	}

	err = os.MkdirAll("debian/tmp", 0777)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	df, err := os.Create("debian/tmp/gdbuild-dockerfile")
	defer df.Close()
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	fmt.Fprintf(df, "FROM debian:%s\n", suite)

	fmt.Fprintf(df, "RUN apt-get update && apt-get install -y --no-install-recommends \\\n")
	for _, pkg := range allBins {
		fmt.Fprintf(df, "\t\t%s=%s \\\n", pkg.Package, pkg.Version)
	}
	fmt.Fprintf(df, "\t&& rm -rf /var/lib/apt/lists/*\n")

	fmt.Fprintf(df, "WORKDIR /usr/src/pkg\n")
	fmt.Fprintf(df, "COPY . /usr/src/pkg\n")

	fmt.Fprintf(df, "RUN dpkg-buildpackage -uc -us\n")

	df.Close()

	fmt.Printf("docker build -f %q .\n", df.Name())
}

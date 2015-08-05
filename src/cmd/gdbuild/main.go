package main

import (
	"fmt"
	"log"

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

	arch := "amd64"

	index, err := resolver.GetBinaryIndex(
		"http://httpredir.debian.org/debian",
		"unstable",
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

	buildable, why := index.ExplainSatisfiesBuildDepends(*depArch, con.Source.BuildDepends)
	fmt.Printf("%t: %s\n", buildable, why)
}

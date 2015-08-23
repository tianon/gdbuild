package main

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
)

func md5sum(path string) (string, error) {
	algo := md5.New()

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := io.Copy(algo, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(algo.Sum(nil)), nil
}

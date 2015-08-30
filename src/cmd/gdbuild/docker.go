package main

import (
	"archive/tar"
	"os"
	"os/exec"
	"path/filepath"
)

//var dockerApiVersion = "1.18" // https://github.com/docker/docker/blob/v1.6.2/api/common.go#L18

func dockerBuild(tag string, dockerfile string, files ...string) error {
	dockerfileMd5, err := md5string(dockerfile)
	if err != nil {
		return err
	}
	dockerfileFile := ".dockerfile." + dockerfileMd5

	cmd := exec.Command("docker", "build", "-f", dockerfileFile, "-t", tag, "-")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	defer stdin.Close()

	if err := cmd.Start(); err != nil {
		return err
	}

	tw := tar.NewWriter(stdin)
	defer tw.Close()

	if err := AddStringToTar(tw, dockerfileFile, dockerfile); err != nil {
		return err
	}
	if err := tw.Flush(); err != nil {
		return err
	}

	for _, file := range files {
		if err := AddFileToTar(tw, filepath.Base(file), file); err != nil {
			return err
		}
		if err := tw.Flush(); err != nil {
			return err
		}
	}

	if err := tw.Close(); err != nil {
		return err
	}

	return cmd.Wait()
}

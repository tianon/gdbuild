package main

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

func AddStringToTar(tw *tar.Writer, name, file string) error {
	hdr := &tar.Header{
		Name: name,
		Size: int64(len(file)),

		Mode: 0666,

		Uid: syscall.Getuid(),
		Gid: syscall.Getgid(),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	if _, err := tw.Write([]byte(file)); err != nil {
		return err
	}
	return nil
}

func AddFileToTar(tw *tar.Writer, name, file string) error {
	fi, err := os.Stat(file)
	if err != nil {
		return err
	}

	if fi.IsDir() {
		err = filepath.Walk(file, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			subPath, err := filepath.Rel(file, path)
			if err != nil {
				return err
			}
			return AddFileToTar(tw, name+"/"+subPath, path)
		})
		if err != nil {
			return err
		}
		return nil
	}

	hdr, err := tar.FileInfoHeader(fi, "")
	if err != nil {
		return err
	}
	hdr.Name = name

	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}

	fh, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fh.Close()

	_, err = io.Copy(tw, fh)
	if err != nil {
		return err
	}

	return nil
}

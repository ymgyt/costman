// +build mage

package main

import (
	"fmt"
	"os"

	"github.com/magefile/mage/sh"
)

const (
	version    = "0.0.3"
	app        = "costman"
	dockerRepo = "docker.io/ymgyt"
)

var Default = All

var Aliases = map[string]interface{}{
	"deploy": All,
}

func All() {
	Build()
	Tag()
	Push()
}

// A build step that requires additional params, or platform specific steps for example
func Build() error {
	return sh.RunV("docker", "build", "-t", appImage(), ".")
}

func Tag() error {
	return sh.RunV("docker", "tag", appImage(), remoteTag())
}

func Push() error {
	return sh.RunV("docker", "push", remoteTag())
}

func Debug() {
	fmt.Println(os.Getenv("MAGE_ARG"))
}

func appImage() string  { return fmt.Sprintf("%s:%s", app, version) }
func remoteTag() string { return fmt.Sprintf("%s/%s", dockerRepo, appImage()) }

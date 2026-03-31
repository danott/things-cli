package main

import "github.com/danott/things-cli/internal/cmd"

var version = "dev"

func main() {
	cmd.Execute(version)
}

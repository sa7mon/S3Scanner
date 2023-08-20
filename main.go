package main

import "github.com/sa7mon/s3scanner/cmd/s3scanner"

var version = "dev"

func main() {
	s3scanner.Run(version)
}

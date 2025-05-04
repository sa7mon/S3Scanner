package main

import "github.com/mux0x/S3Scanner/cmd/s3scanner"

var version = "dev"

func main() {
	s3scanner.Run(version)
}

package main

import (
	"github.com/similar-manga/similar/cmd"
	_ "github.com/similar-manga/similar/cmd/calculate"
	_ "github.com/similar-manga/similar/cmd/init"
	_ "github.com/similar-manga/similar/cmd/mangadex"
)

func main() {
	cmd.Execute()
}

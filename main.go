package main

import (
	"github.com/nekomangaorg/similar-processor/cmd"
	_ "github.com/nekomangaorg/similar-processor/cmd/calculate"
	_ "github.com/nekomangaorg/similar-processor/cmd/init"
	_ "github.com/nekomangaorg/similar-processor/cmd/mangadex"
	_ "github.com/nekomangaorg/similar-processor/cmd/neko"
)

func main() {
	cmd.Execute()
}

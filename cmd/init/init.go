package calculate

import (
	"bufio"
	"fmt"
	"github.com/similar-manga/similar/cmd"
	"github.com/similar-manga/similar/internal"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init command",
	Long:  `Initializes the repository for processing`,
	Run:   runInit,
}

func init() {
	cmd.RootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) {
	startProcessing := time.Now()

	createMangaDB()

	populateMangaDB()
	fmt.Printf("Initialized in %s\n\n", time.Since(startProcessing))

}

func createMangaDB() {
	fmt.Println("Creating manga.db")
	src, err := os.Open("data/default_manga.db")
	internal.CheckErr(err)
	defer src.Close()
	dst, err := os.Create("data/manga.db")
	internal.CheckErr(err)
	defer dst.Close()

	buf := make([]byte, 1024)
	for {
		n, err := src.Read(buf)
		if err != nil && err != io.EOF {
			internal.CheckErr(err)
			break
		}

		if n == 0 {
			break
		}

		if _, err := dst.Write(buf[:n]); err != nil {
			internal.CheckErr(err)
			break
		}
	}
}

func populateMangaDB() {
	files, err := os.ReadDir("data/manga/")
	fmt.Printf("Populating manga.db manga table from %d files\n", len(files))

	if err != nil {
		log.Fatal(err)
	}

	for _, fileInfo := range files {
		fmt.Printf("Populating from  %s\n", fileInfo.Name())
		openFileAndProcess(fileInfo)
	}
}

func openFileAndProcess(fileInfo os.DirEntry) {
	file, err := os.Open("data/manga/" + fileInfo.Name())
	defer file.Close()
	internal.CheckErr(err)
	scanner := bufio.NewScanner(file)
	tx, err := internal.DB.Begin()
	internal.CheckErr(err)
	for scanner.Scan() {
		split := strings.Split(scanner.Text(), ":::")
		if len(split) > 0 {
			_, err := tx.Exec("INSERT INTO MANGA(UUID, JSON) VALUES (?,?) ON CONFLICT (UUID) DO UPDATE SET JSON=excluded.JSON", split[0], split[1])
			internal.CheckErr(err)
		}
	}
	err = tx.Commit()
	internal.CheckErr(err)
}

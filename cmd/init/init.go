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
	Short: "Initialize the database data",
	Long:  `Initializes the database with the stored data in the repository`,
	Run:   runInit,
}

func init() {
	cmd.RootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) {
	fmt.Println("Begin init")
	startProcessing := time.Now()

	createMangaDB()
	populateMangaDB()
	populateMangaUpdatesMappingDB()
	fmt.Printf("Initialized in %s\n\n", time.Since(startProcessing))

}

func createMangaDB() {
	fmt.Println("Creating manga.db")
	err := copyFile("data/default_empty_data.db", "data/data.db")
	internal.CheckErr(err)
}

func copyFile(srcPath, dstPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func populateMangaUpdatesMappingDB() {
	file, err := os.Open("data/mappings/mangaupdates_new2mdex.txt")
	fmt.Printf("Populating from  %s\n", "mangaupdates_new2mdex.txt")
	defer file.Close()
	internal.CheckErr(err)
	scanner := bufio.NewScanner(file)
	tx, err := internal.DB.Begin()
	internal.CheckErr(err)
	for scanner.Scan() {
		line := scanner.Text()
		split := strings.Split(line, ":::||@!@||:::")
		if len(split) > 0 && len(line) > 0 {
			_, err := tx.Exec("INSERT INTO MANGAUPDATES_NEW(UUID, ID) VALUES (?,?) ON CONFLICT (UUID) DO UPDATE SET ID=excluded.ID", split[1], split[0])
			internal.CheckErr(err)
		}
	}
	err = tx.Commit()
	internal.CheckErr(err)
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
		split := strings.Split(scanner.Text(), ":::||@!@||:::")
		if len(split) > 0 {
			_, err := tx.Exec("INSERT INTO MANGA(UUID, DATE, JSON) VALUES (?,?,?) ON CONFLICT (UUID) DO UPDATE SET JSON=excluded.JSON", split[0], split[1], split[2])
			internal.CheckErr(err)
		}
	}
	err = tx.Commit()
	internal.CheckErr(err)
}

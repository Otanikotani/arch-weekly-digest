package crawler

import (
	"fmt"
	"github.com/tj/go-dropy"
	"google.golang.org/api/drive/v3"
	"log"
	"strings"
	"sync"
	"time"
)

//CrawlP2s crawls P2 documents updated last week
func CrawlP2s(wg *sync.WaitGroup, gdrive *drive.Service, dropboxClient *dropy.Client) {
	defer wg.Done()

	oneDayEarlier := time.Now().Add(-7 * 24 * time.Hour)
	query := fmt.Sprintf("name contains 'Prod2' and modifiedTime > '%v'",
		oneDayEarlier.Format("2006-01-02T15:04:05"))

	fileList, err := gdrive.Files.List().Q(query).Fields("files(id, name, modifiedTime)").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve files: %v", err)
	}
	fmt.Println("Files:")

	for _, file := range fileList.Files {
		parsedTime, err := time.Parse("2006-01-02T15:04:05.999Z", file.ModifiedTime)
		if err != nil {
			panic(err)
		}

		modifiedTime := parsedTime.Format("02 Jan 06 15:04")

		fmt.Printf("%v: %v\n", modifiedTime, file.Name)
	}
	saveNewP2s(fileList.Files, dropboxClient)
}

func saveNewP2s(files []*drive.File, dropboxClient *dropy.Client) {
	now := time.Now().Format("Jan 02")
	mdFileName := "/Arch/Changelog/" + now + " P2s.md"

	fullMd := "# " + now + "\n"
	for _, file := range files {
		fullMd += fmt.Sprintf("- [%v](%v)\n", file.Name, "https://docs.google.com/document/d/"+file.Id)
		fullMd += "\n\n"
	}

	err := dropboxClient.Upload(mdFileName, strings.NewReader(fullMd))
	if err != nil {
		log.Fatal(err)
	}
}

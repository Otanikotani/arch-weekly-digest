package crawler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/tj/go-dropy"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type libraryTDD struct {
	Title        string
	Body         string
	ModifiedTime time.Time
	Labels       []conflContentMetadataLabel
}

type conflSearchResult struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Title string `json:"title"`
}

type conflSearchResults struct {
	Results []conflSearchResult `json:"results"`
}

type conflContentBodyStorage struct {
	Value string `json:"value"`
}

type conflContentBody struct {
	Storage conflContentBodyStorage `json:"storage"`
}

type conflContentMetadataLabel struct {
	Name string `json:"name"`
}

type conflContentMetadataLabels struct {
	Results []conflContentMetadataLabel `json:"results"`
}

type conflContentMetadata struct {
	Labels conflContentMetadataLabels `json:"labels"`
}

type conflContentHistoryLastUpdated struct {
	When string `json:"when"`
}

type conflContentHistory struct {
	LastUpdated conflContentHistoryLastUpdated `json:"lastUpdated"`
}

type conflContent struct {
	Body     conflContentBody     `json:"body"`
	Metadata conflContentMetadata `json:"metadata"`
	History  conflContentHistory  `json:"history"`
	Title    string               `json:"title"`
}

//CrawlLibrary is for crawling arch library items
func CrawlLibrary(wg *sync.WaitGroup, user string, password string, dropboxClient *dropy.Client) {
	defer wg.Done()
	libraryTdds := getLibraryTdds(user, password)
	saveLibraryTdds(libraryTdds, dropboxClient)
}

func getLibraryTdds(user string, password string) []libraryTDD {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://confluence.devfactory.com/rest/api/content/search?limit=1000&maxResults=1000&cql=(parent=379894933)", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", "Basic "+basicAuth(user, password))
	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}

	if resp.StatusCode != 200 {
		panic("Invalid response " + resp.Status)
	}

	var confluenceArticles conflSearchResults
	err = json.NewDecoder(resp.Body).Decode(&confluenceArticles)
	if err != nil {
		panic(err)
	}

	converter := md.NewConverter("", true, nil)

	var libraryTDDs []libraryTDD

	for _, result := range confluenceArticles.Results {
		selfURI := "https://confluence.devfactory.com/rest/api/content/" +
			result.ID + "?expand=body.storage,history.lastUpdated,metadata.labels"
		fmt.Printf("Self: %v\n", selfURI)
		req, err := http.NewRequest("GET", selfURI, nil)
		if err != nil {
			panic(err)
		}
		req.Header.Add("Authorization", "Basic "+basicAuth("aartyukhov", "g5NGp7s8J"))

		resp, err := client.Do(req)

		if err != nil {
			panic(err)
		}

		if resp.StatusCode != 200 {
			panic("Invalid response " + resp.Status)
		}

		var confluenceContent conflContent
		err = json.NewDecoder(resp.Body).Decode(&confluenceContent)
		if err != nil {
			panic(err)
		}

		markdown, err := converter.ConvertString(confluenceContent.Body.Storage.Value)
		if err != nil {
			log.Fatal(err)
		}

		modifiedTime, err := time.Parse("2006-01-02T15:04:05.999Z", confluenceContent.History.LastUpdated.When)
		if err != nil {
			log.Fatal(err)
		}

		libraryTDDs = append(libraryTDDs, libraryTDD{
			Title:        confluenceContent.Title,
			Body:         markdown,
			ModifiedTime: modifiedTime,
			Labels:       confluenceContent.Metadata.Labels.Results,
		})
	}

	return libraryTDDs
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func saveLibraryTdds(libraryTdds []libraryTDD, dropboxClient *dropy.Client) {
	for _, libraryTdd := range libraryTdds {
		mdFileName := "/Arch/Library/" + strings.ReplaceAll(libraryTdd.Title+".md", "/", "-")

		fmt.Println("Write " + mdFileName)

		stat, err := dropboxClient.Stat(mdFileName)
		if err != nil {
			log.Fatal(err)
		}

		if libraryTdd.ModifiedTime.After(stat.ModTime()) {

			fullMd := libraryTdd.Body + "\n"
			for _, label := range libraryTdd.Labels {
				fullMd += "#" + label.Name + "\n"
			}

			err := dropboxClient.Upload(mdFileName, strings.NewReader(fullMd))
			if err != nil {
				log.Fatal(err)
			}
		} else {
			fmt.Printf("File %v, skipping update\n", mdFileName)
			return
		}
	}

}

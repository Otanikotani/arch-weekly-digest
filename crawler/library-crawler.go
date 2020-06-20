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

type LibraryTDD struct {
	Title        string
	Body         string
	ModifiedTime time.Time
	Labels       []ConflContentMetadataLabel
}

type ConflSearchResult struct {
	Id    string `json:"id"`
	Type  string `json:"type"`
	Title string `json:"title"`
}

type ConflSearchResults struct {
	Results []ConflSearchResult `json:"results"`
}

type ConflContentBodyStorage struct {
	Value string `json:"value"`
}

type ConflContentBody struct {
	Storage ConflContentBodyStorage `json:"storage"`
}

type ConflContentMetadataLabel struct {
	Name string `json:"name"`
}

type ConflContentMetadataLabels struct {
	Results []ConflContentMetadataLabel `json:"results"`
}

type ConflContentMetadata struct {
	Labels ConflContentMetadataLabels `json:"labels"`
}

type ConflContentHistoryLastUpdated struct {
	When string `json:"when"`
}

type ConflContentHistory struct {
	LastUpdated ConflContentHistoryLastUpdated `json:"lastUpdated"`
}

type ConflContent struct {
	Body     ConflContentBody     `json:"body"`
	Metadata ConflContentMetadata `json:"metadata"`
	History  ConflContentHistory  `json:"history"`
	Title    string               `json:"title"`
}

func CrawlLibrary(wg *sync.WaitGroup, user string, password string, dropboxClient *dropy.Client) {
	defer wg.Done()
	libraryTdds := getLibraryTdds(user, password)
	saveLibraryTdds(libraryTdds, dropboxClient)
}

func getLibraryTdds(user string, password string) []LibraryTDD {
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

	var confluenceArticles ConflSearchResults
	err = json.NewDecoder(resp.Body).Decode(&confluenceArticles)
	if err != nil {
		panic(err)
	}

	converter := md.NewConverter("", true, nil)

	var libraryTDDs []LibraryTDD

	for _, result := range confluenceArticles.Results {
		selfUri := "https://confluence.devfactory.com/rest/api/content/" +
			result.Id + "?expand=body.storage,history.lastUpdated,metadata.labels"
		fmt.Printf("Self: %v\n", selfUri)
		req, err := http.NewRequest("GET", selfUri, nil)
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

		var confluenceContent ConflContent
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

		libraryTDDs = append(libraryTDDs, LibraryTDD{
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

func saveLibraryTdds(libraryTdds []LibraryTDD, dropboxClient *dropy.Client) {
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

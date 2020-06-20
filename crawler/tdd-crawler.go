package crawler

import (
	"fmt"
	"github.com/andygrunwald/go-jira"
	"github.com/tj/go-dropy"
	"google.golang.org/api/docs/v1"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type tdd struct {
	ID            string
	Body          string
	DocumentID    string
	DocumentTitle string
}

//CrawlTdds crawls tdds in AU tickets done last week
func CrawlTdds(wg *sync.WaitGroup, gdocs *docs.Service, user string, password string, dropboxClient *dropy.Client) {
	defer wg.Done()

	doneLastWeekJql := "project in (ARCH) AND status changed to (Done) after -1w"
	jiraClient, err := jiraClient(user, password)
	if err != nil {
		log.Fatalf("Failed to create jira client: %v", err)
	}
	issues, _, err := jiraClient.Issue.Search(doneLastWeekJql, nil)
	if err != nil {
		log.Fatalf("Failed to find issues using jql %v. Error: %v", doneLastWeekJql, err)
	}

	var tdds []tdd
	for _, issue := range issues {
		docLink := issue.Fields.Unknowns["customfield_26103"]
		docLinkStr := fmt.Sprintf("%v", docLink)
		docID := strings.Split(strings.ReplaceAll(docLinkStr, "https://docs.google.com/document/d/", ""), "/")[0]

		doc, err := gdocs.Documents.Get(docID).Do()

		if err != nil {
			log.Printf("Unable to load Doc %v: %v", docID, err)
			continue
		}

		auTdds := findTDDs(doc)
		tdds = append(tdds, auTdds...)
	}
	saveNewTdds(tdds, dropboxClient)
}

func jiraClient(user string, password string) (*jira.Client, error) {
	tp := jira.BasicAuthTransport{
		Username: user,
		Password: password,
	}

	return jira.NewClient(tp.Client(), "https://jira.devfactory.com")
}

func saveNewTdds(tdds []tdd, dropboxClient *dropy.Client) {
	now := time.Now().Format("Jan 02")
	mdFileName := "/Arch/Changelog/" + now + ".md"

	fmt.Println("Write " + mdFileName)

	var tddsByDocument = make(map[string][]tdd)
	for _, tddElement := range tdds {
		if strings.TrimSpace(tddElement.Body) == "" || strings.Contains(tddElement.Body, "What specific decisions and things to do in this componen") {
			continue
		}

		if arr, ok := tddsByDocument[tddElement.DocumentID]; !ok {
			var t []tdd
			t = append(t, tddElement)
			tddsByDocument[tddElement.DocumentID] = t
		} else {
			arr = append(arr, tddElement)
			tddsByDocument[tddElement.DocumentID] = arr
		}
	}

	fullMd := "# " + now + "\n"
	for _, docTdds := range tddsByDocument {
		if len(docTdds) != 0 {
			fullMd += fmt.Sprintf("## [%v](%v)\n", docTdds[0].DocumentTitle, "https://docs.google.com/document/d/"+docTdds[0].DocumentID)
			for _, tdd := range docTdds {
				fullMd += tdd.Body + "\n"
				fullMd += "\n --- \n"
			}
		}
	}

	err := dropboxClient.Upload(mdFileName, strings.NewReader(fullMd))
	if err != nil {
		log.Fatal(err)
	}
}

func findTDDs(doc *docs.Document) []tdd {
	content := doc.Body.Content
	var tables []*docs.Table

	for _, element := range content {
		if element.Table != nil {
			tables = append(tables, element.Table)
		}
	}

	var tdds []tdd
	if tables != nil {
		for _, table := range tables {
			for _, row := range table.TableRows {
				if row.TableCells != nil && len(row.TableCells) > 1 {
					rowTitle := getContent(row.TableCells[0].Content, doc.InlineObjects)
					if strings.HasPrefix(rowTitle, "__TDD") {
						tdds = append(tdds, tdd{
							ID:            strings.ReplaceAll(rowTitle, "_", ""),
							Body:          getContent(row.TableCells[1].Content, doc.InlineObjects),
							DocumentID:    doc.DocumentId,
							DocumentTitle: doc.Title,
						})
					}
				}
			}
		}
	}

	return tdds
}

func getContent(content []*docs.StructuralElement, inlineObjects map[string]docs.InlineObject) string {
	result := ""
	for _, element := range content {
		if element.Paragraph != nil {
			for _, paragraphElement := range element.Paragraph.Elements {
				if paragraphElement.InlineObjectElement != nil {
					inlineObjectID := paragraphElement.InlineObjectElement.InlineObjectId
					if inlineObject, ok := inlineObjects[inlineObjectID]; ok {
						uri := inlineObject.InlineObjectProperties.EmbeddedObject.ImageProperties.ContentUri
						result += fmt.Sprintf("![%v](%v)\n", inlineObject.InlineObjectProperties.EmbeddedObject.Description, uri)
					} else {
						result += "Missing Image"
					}

				} else if paragraphElement.TextRun != nil {
					result += docToMd(paragraphElement.TextRun, element)
				}
			}
		}
	}
	return result
}

func docToMd(textRun *docs.TextRun, element *docs.StructuralElement) string {
	if textRun.TextStyle == nil {
		return textRun.Content
	}

	md := textRun.Content
	if len(strings.TrimSpace(md)) == 0 {
		return md
	}

	style := textRun.TextStyle
	if style.Bold {
		md = wrapBold(md)
	}
	if style.Link != nil {
		md = wrapAsLink(md, style.Link.Url)
	}
	if element.Paragraph.Bullet != nil {
		md = "- " + md
	}

	return md
}

func wrapBold(md string) string {
	if strings.HasSuffix(md, "\n") {
		return fmt.Sprintf("__%v__", strings.Split(md, "\n")[0]) + "\n"
	}
	return fmt.Sprintf("**%v**", md)
}

func wrapAsLink(md string, url string) string {
	return fmt.Sprintf("[%v](%v)", md, url)
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jessevdk/go-flags"
	"github.com/otanikotani/arch-tdd/crawler"
	"github.com/otanikotani/arch-tdd/googleapi"
	"github.com/tj/go-dropbox"
	"github.com/tj/go-dropy"
	"log"
	"os"
	"sync"
)

//Cli options or env
type Opts struct {
	GoogleClientID     string `long:"google-client-id" env:"GOOGLE_CLIENT_ID"`
	GoogleSecret       string `long:"google-secret" env:"GOOGLE_SECRET"`
	ConfluenceUser     string `long:"confluence-user" env:"CONFLUENCE_USER"`
	ConfluencePassword string `long:"confluence-password" env:"CONFLUENCE_PASSWORD"`
	DropboxToken       string `long:"dropbox-token" env:"DROPBOX_TOKEN"`
}

func handleRequest(_ context.Context) error {
	return cli()
}

func main() {
	if len(os.Args) > 1 {
		err := cli()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		lambda.Start(handleRequest)
	}
}

func cli() error {
	var opts Opts
	if _, err := flags.Parse(&opts); err != nil {
		return err
	}
	gdocs, gdrive := googleapi.NewClients(opts.GoogleClientID, opts.GoogleSecret)
	dropboxClient := dropy.New(dropbox.New(dropbox.NewConfig(opts.DropboxToken)))

	var wg sync.WaitGroup

	wg.Add(3)

	go crawler.CrawlTdds(&wg, gdocs, opts.ConfluenceUser, opts.ConfluencePassword, dropboxClient)
	go crawler.CrawlLibrary(&wg, opts.ConfluenceUser, opts.ConfluencePassword, dropboxClient)
	go crawler.CrawlP2s(&wg, gdrive, dropboxClient)

	wg.Wait()

	return nil
}

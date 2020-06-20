package googleapi

import (
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"log"
	"net/http"
	"time"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tok, err := tokenFromFile()
	if err != nil {
		log.Fatalf("Failed to find token.json file - required for auth: %v", err)
	}
	return config.Client(context.Background(), tok)
}

// Retrieves a token from a local file.
func tokenFromFile() (*oauth2.Token, error) {
	expiry, err := time.Parse("2006-01-02T15:04:05.999999999-07:00", "2020-05-09T17:18:15.870601685-04:00")
	if err != nil {
		return nil, err
	}
	tok := &oauth2.Token{
		AccessToken:  "ya29.a0Ae4lvC08Txhk4QAttYxdDcnPVy3Yh5z8nFGPnSzzLhU2Zxd2eyls2wlTesf5zQNTY8flgX_lnGl4WVgPVPD-PZsgCIpj47xIMidgtwBVWQd7aKuHroZmbN91uKetsumEkHe1Xim64Si1_NMBEAr0YT9uR7L9_mVi-Yw",
		TokenType:    "Bearer",
		RefreshToken: "1//0dZ0c3KROYHOjCgYIARAAGA0SNwF-L9IrlpViy_recKKbz6eh88d3I8k-a2GJX-f_Y0evFAXSzMN1eyAQczbhtlpkc5A8zKfzbCU",
		Expiry:       expiry,
	}
	return tok, err
}

func NewClients(clientId string, secret string) (*docs.Service, *drive.Service) {
	config := oauth2.Config{
		ClientID:     clientId,
		ClientSecret: secret,
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
		Scopes:       []string{"https://www.googleapis.com/auth/spreadsheets.readonly"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}

	client := getClient(&config)

	gdrive, err := drive.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	gdocs, err := docs.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Docs client: %v", err)
	}

	return gdocs, gdrive
}

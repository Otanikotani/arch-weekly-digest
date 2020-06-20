package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"os"
)

type myEvent struct {
	Name string `json:"name"`
}

func handleRequest(ctx context.Context, name myEvent) (string, error) {
	return fmt.Sprintf("Hello 4 %s!", name.Name), nil
}

func main() {
	if len(os.Args) > 1 {
		fmt.Println("Run as cli")
	} else {
		lambda.Start(handleRequest)
	}
}

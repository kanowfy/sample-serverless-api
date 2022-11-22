package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var isbnRegexp = regexp.MustCompile(`[0-9]{3}-[0-9]{10}`)
var errLogger = log.New(os.Stderr, "[ERROR] ", log.Llongfile)

type application struct {
	db *dynamodb.Client
}

type book struct {
	ISBN   string `json:"isbn"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

func (app *application) router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		return app.show(req)
	case "POST":
		return app.create(req)
	default:
		return clientError(http.StatusMethodNotAllowed)
	}
}

func (app *application) show(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	isbn := req.QueryStringParameters["isbn"]
	if !isbnRegexp.MatchString(isbn) {
		return clientError(http.StatusBadRequest)
	}

	b, err := app.getItem(isbn)
	if err != nil {
		return serverError(err)
	}

	if b == nil {
		return clientError(http.StatusNotFound)
	}

	js, err := json.Marshal(b)
	if err != nil {
		return serverError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
	}, nil

}

func (app *application) create(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.Headers["content-type"] != "application/json" && req.Headers["Content-Type"] != "application/json" {
		return clientError(http.StatusNotAcceptable)
	}

	b := book{}
	err := json.Unmarshal([]byte(req.Body), &b)
	if err != nil {
		return clientError(http.StatusUnprocessableEntity)
	}

	if !isbnRegexp.MatchString(b.ISBN) {
		return clientError(http.StatusBadRequest)
	}

	if b.Title == "" || b.Author == "" {
		return clientError(http.StatusBadRequest)
	}

	err = app.putItem(&b)
	if err != nil {
		return serverError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusCreated,
		Headers: map[string]string{
			"Location": fmt.Sprintf("/books?isbn=%s", b.ISBN),
		},
	}, nil

}

// helper for handling server error
func serverError(err error) (events.APIGatewayProxyResponse, error) {
	errLogger.Println(err.Error())

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       http.StatusText(http.StatusInternalServerError),
	}, nil
}

// helper for handling client related errors
func clientError(status int) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       http.StatusText(status),
	}, nil
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-southeast-1"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	db := dynamodb.NewFromConfig(cfg)

	app := &application{
		db: db,
	}

	lambda.Start(app.router)
}

package main

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func (app *application) getItem(isbn string) (*book, error) {
	b := book{ISBN: isbn}

	input := &dynamodb.GetItemInput{
		TableName: aws.String("Books"),
		Key: map[string]types.AttributeValue{
			"ISBN": &types.AttributeValueMemberS{Value: isbn},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result, err := app.db.GetItem(ctx, input)
	if err != nil {
		return nil, err
	}

	if result.Item == nil {
		return nil, nil
	}

	err = attributevalue.UnmarshalMap(result.Item, &b)
	if err != nil {
		return nil, err
	}

	return &b, nil

}

func (app *application) putItem(b *book) error {
	item, err := attributevalue.MarshalMap(*b)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String("Books"),
		Item:      item,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = app.db.PutItem(ctx, input)
	return err
}

package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

var db = dynamodb.New(session.New(), aws.NewConfig().WithRegion("eu-west-1"))

func getEmployeedata(username string) (*details, error) {

	input := &dynamodb.GetItemInput{
		TableName:      aws.String("EmployeeDirectory"),
		ConsistentRead: aws.Bool(true),
		Key: map[string]*dynamodb.AttributeValue{
			"UserName": {
				S: aws.String(username),
			},
		},
	}

	result, err := db.GetItemWithContext(context.TODO(), input, func(r *request.Request) {
		r.Handlers.Complete.PushBack(func(req *request.Request) {
			fmt.Println(req.RequestID)

		})
	})

	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}

	details := new(details)
	err = dynamodbattribute.UnmarshalMap(result.Item, details)
	if err != nil {
		return nil, err
	}

	return details, nil
}

func deleteEmployee(username string) error {

	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"UserName": {
				S: aws.String(username),
			},
		},
		TableName: aws.String("EmployeeDirectory"),
	}

	_, err := db.DeleteItem(input)

	if err != nil {
		fmt.Println(err.Error())
	}
	return err
}

func createEmployee(emp *employee) error {

	Salt := RandStringRunes()
	Password, HashPwdErr := HashPassword(emp.Password + Salt)

	if HashPwdErr != nil {
		return HashPwdErr
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String("EmployeeDirectory"),
		Item: map[string]*dynamodb.AttributeValue{
			"UserName": {
				S: aws.String(emp.UserName),
			},
			"FullName": {
				S: aws.String(emp.FullName),
			},
			"PhoneNumber": {
				S: aws.String(emp.PhoneNumber),
			},
			"EmployeeType": {
				S: aws.String(emp.EmployeeType),
			},
			"Salt": {
				S: aws.String(Salt),
			},
			"Password": {
				S: aws.String(Password),
			},
		},
	}

	_, err := db.PutItem(input)

	return err
}

func updateEmployeePassword(details *updatePassword) error {
	Salt := RandStringRunes()
	Password, HashPwdErr := HashPassword(details.Password + Salt)

	if HashPwdErr != nil {
		return HashPwdErr
	}
	input := &dynamodb.UpdateItemInput{

		TableName: aws.String("EmployeeDirectory"),
		Key: map[string]*dynamodb.AttributeValue{
			"UserName": {
				S: aws.String(details.UserName),
			},
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":Password": {
				S: aws.String(Password),
			},
			":Salt": {
				S: aws.String(Salt),
			},
		},
		UpdateExpression: aws.String("set Password = :Password, Salt = :Salt"),
		ReturnValues:     aws.String("UPDATED_NEW"),
	}
	_, err := db.UpdateItem(input)

	if err != nil {
		fmt.Println(err.Error())
	}
	return err
}

func updateEmployee(details *updateDetails) error {
	input := &dynamodb.UpdateItemInput{

		TableName: aws.String("EmployeeDirectory"),
		Key: map[string]*dynamodb.AttributeValue{
			"UserName": {
				S: aws.String(details.UserName),
			},
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":FullName": {
				S: aws.String(details.FullName),
			},
			":PhoneNumber": {
				S: aws.String(details.PhoneNumber),
			},
		},
		UpdateExpression: aws.String("set FullName = :FullName, PhoneNumber = :PhoneNumber"),
		ReturnValues:     aws.String("UPDATED_NEW"),
	}

	_, err := db.UpdateItem(input)

	if err != nil {
		fmt.Println(err.Error())
	}
	return err
}
func authenticateUser(empp *employee) (*employee, error) {

	input := &dynamodb.GetItemInput{
		TableName:      aws.String("EmployeeDirectory"),
		ConsistentRead: aws.Bool(true),
		Key: map[string]*dynamodb.AttributeValue{
			"UserName": {
				S: aws.String(empp.UserName),
			},
		},
	}

	result, err := db.GetItemWithContext(context.TODO(), input, func(r *request.Request) {
		r.Handlers.Complete.PushBack(func(req *request.Request) {
			fmt.Println(req.RequestID)

		})
	})

	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}

	emp := new(employee)
	err = dynamodbattribute.UnmarshalMap(result.Item, emp)

	if err != nil {
		return nil, err
	}

	return emp, err
}

package main

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"golang.org/x/crypto/bcrypt"
)

var usernameRegexp = regexp.MustCompile("^[a-zA-Z0-9_.+-]+@(?:(?:[a-zA-Z0-9-]+\\.)?[a-zA-Z]+\\.)?(test)\\.no$")
var errorLogger = log.New(os.Stderr, "ERROR ", log.Llongfile)
var letterRunes = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890")

func main() {

	lambda.Start(router)
}

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	proceed, employeeData := authenticate(req.Headers["Author"])

	if proceed {
		HTTPMethod := ""
		if strings.Contains(req.Path, "changePassword") {

			HTTPMethod = "PUTPassword"
		} else if req.HTTPMethod == "GET" && req.Path == "/employees" {
			HTTPMethod = "GETAll"

		} else {
			HTTPMethod = req.HTTPMethod
		}
		if employeeData.EmployeeType == "Admin" {

			switch HTTPMethod {
			case "POST":
				return create(req)
			case "GET":
				return show(req, employeeData.EmployeeType, employeeData.UserName)
			case "DELETE":
				return deleteEmp(req)
			case "PUT":
				return update(req, employeeData.UserName, employeeData.EmployeeType)
			default:
				return clientError(http.StatusMethodNotAllowed)
			}

		} else {

			switch HTTPMethod {
			case "GET":
				return show(req, employeeData.EmployeeType, employeeData.UserName)
			case "PUT":
				return update(req, employeeData.UserName, employeeData.EmployeeType)
			case "PUTPassword":
				return updatePwd(req, employeeData)

			default:
				return clientError(http.StatusMethodNotAllowed)
			}

		}

	}

	return clientError(http.StatusBadRequest)

}

func authenticate(AuthorizationHeader string) (bool, *employee) {
	request := new(employee)

	if AuthorizationHeader != "" {

		authHeader := strings.Split(AuthorizationHeader, "Basic ")

		if len(authHeader) != 2 {
			return false, request
		}

		decodedString, err := base64.StdEncoding.DecodeString(authHeader[1])

		if err != nil {
			return false, request
		}

		credentials := strings.Split(string(decodedString), ":")
		if len(credentials) != 2 {
			return false, request
		}

		UserName, Password := credentials[0], credentials[1]
		request.UserName = UserName
		request.Password = Password

		employeeData, err2 := authenticateUser(request)

		if err2 != nil {
			return false, request
		}

		Match := CheckHash(request.Password+employeeData.Salt, employeeData.Password)

		return Match, employeeData

	} else {
		return false, request
	}

}

func RandStringRunes() string {
	b := make([]rune, 20)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckHash(password, hash string) bool {

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func deleteEmp(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.PathParameters["username"] != "" {
		err := deleteEmployee(req.PathParameters["username"])
		if err != nil {
			return serverError(err)
		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
	}, nil
}
func show(req events.APIGatewayProxyRequest, employeeType string, username string) (events.APIGatewayProxyResponse, error) {
	if req.PathParameters["username"] != "" && (req.PathParameters["username"] == username || employeeType == "Admin") {

		details, err := getEmployeedata(req.PathParameters["username"])
		if err != nil {
			return serverError(err)
		}
		if details == nil {
			return clientError(http.StatusNotFound)
		}

		js, err := json.Marshal(details)
		if err != nil {
			return serverError(err)
		}

		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       string(js),
		}, nil
	} else {
		if !usernameRegexp.MatchString(req.PathParameters["username"]) {
			return clientError(http.StatusBadRequest)

		}
		return clientError(http.StatusMethodNotAllowed)
	}

}

func create(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	if req.Headers["Content-Type"] != "application/json" {
		return clientError(http.StatusNotAcceptable)
	}

	emp := new(employee)
	UnmarshallingErr := json.Unmarshal([]byte(req.Body), emp)

	if UnmarshallingErr != nil {
		return clientError(http.StatusUnprocessableEntity)
	}

	if emp.EmployeeType == "" || emp.UserName == "" || emp.Password == "" || emp.PhoneNumber == "" || emp.Name == "" {
		return clientError(http.StatusBadRequest)
	}

	details, err := getEmployeedata(emp.UserName)
	if err != nil {
		return serverError(err)
	}
	if details != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       `{"message":"User Already exist!!"}`,
		}, nil
	}

	if !usernameRegexp.MatchString(emp.UserName) {
		return clientError(http.StatusBadRequest)
	}

	PutErr := createEmployee(emp)

	if PutErr != nil {
		return serverError(PutErr)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 201,
		Body:       `{"message":"User Created!!"}`,
	}, nil
}

func updatePwd(req events.APIGatewayProxyRequest, emp *employee) (events.APIGatewayProxyResponse, error) {

	if req.PathParameters["username"] != "" && (req.PathParameters["username"] == emp.UserName || emp.EmployeeType == "Admin") {
		if req.Headers["Content-Type"] != "application/json" {
			return clientError(http.StatusNotAcceptable)
		}

		details := new(updatePassword)
		UnmarshallingErr := json.Unmarshal([]byte(req.Body), details)

		if UnmarshallingErr != nil {
			return clientError(http.StatusUnprocessableEntity)
		}

		if details.OldPassword == "" && details.Password == "" && details.ConfirmPassword == "" {
			return clientError(http.StatusBadRequest)
		}
		if details.Password != details.ConfirmPassword {
			return clientError(http.StatusBadRequest)
		}

		if !CheckHash(details.OldPassword+emp.Salt, emp.Password) {
			return clientError(http.StatusBadRequest)
		}
		details.UserName = emp.UserName
		updateErr := updateEmployeePassword(details)
		if updateErr != nil {
			return serverError(updateErr)
		}
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNoContent,
			Body:       `{"message":"Password updated Successfully!!"}`,
		}, nil

	} else {
		if !usernameRegexp.MatchString(req.PathParameters["username"]) {
			return clientError(http.StatusBadRequest)

		}
		return clientError(http.StatusMethodNotAllowed)
	}

}

func update(req events.APIGatewayProxyRequest, username string, employeeType string) (events.APIGatewayProxyResponse, error) {

	if req.PathParameters["username"] != "" && (req.PathParameters["username"] == username || employeeType == "Admin") {
		if req.Headers["Content-Type"] != "application/json" {
			return clientError(http.StatusNotAcceptable)
		}

		details := new(updateDetails)
		UnmarshallingErr := json.Unmarshal([]byte(req.Body), details)

		if UnmarshallingErr != nil {
			return clientError(http.StatusUnprocessableEntity)
		}

		if details.PhoneNumber == "" || details.Name == "" {
			return clientError(http.StatusBadRequest)
		}
		details.UserName = username
		updateErr := updateEmployee(details)
		if updateErr != nil {
			return serverError(updateErr)
		}
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNoContent,
			Body:       `{"message":"Updated Successfully!!"}`,
		}, nil

	} else {
		if !usernameRegexp.MatchString(req.PathParameters["username"]) {
			return clientError(http.StatusBadRequest)

		}
		return clientError(http.StatusMethodNotAllowed)
	}

}

func serverError(err error) (events.APIGatewayProxyResponse, error) {
	errorLogger.Println(err.Error())

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       http.StatusText(http.StatusInternalServerError),
	}, nil
}

func clientError(status int) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       http.StatusText(status),
	}, nil
}

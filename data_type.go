package main

type employee struct {
	UserName     string `json:"username"`
	EmployeeType string `json:"employeetype"`
	Password     string `json:"password"`
	Salt         string `json:"salt"`
	FullName     string `json:"fullname"`
	PhoneNumber  string `json:"phonenumber"`
}

type details struct {
	UserName     string `json:"username"`
	EmployeeType string `json:"employeetype"`
	FUllName     string `json:"fullname"`
	PhoneNumber  string `json:"phonenumber"`
}

type updateDetails struct {
	UserName    string `json:"username"`
	FullName    string `json:"fullname"`
	PhoneNumber string `json:"phonenumber"`
}

type updatePassword struct {
	UserName        string `json:"username"`
	OldPassword     string `json:"oldpassword"`
	Password        string `json:"Password"`
	ConfirmPassword string `json:"confirmpassword"`
}

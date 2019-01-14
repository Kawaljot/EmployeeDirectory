package main

type employee struct {
	UserName     string `json:"username"`
	EmployeeType string `json:"employeetype"`
	Password     string `json:"password"`
	Salt         string `json:"salt"`
	Name         string `json:"name"`
	PhoneNumber  string `json:"phonenumber"`
}

type details struct {
	UserName     string `json:"username"`
	EmployeeType string `json:"employeetype"`
	Name         string `json:"name"`
	PhoneNumber  string `json:"phonenumber"`
}

type updateDetails struct {
	UserName    string `json:"username"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phonenumber"`
}

type updatePassword struct {
	UserName        string `json:"username"`
	OldPassword     string `json:"oldpassword"`
	Password        string `json:"Password"`
	ConfirmPassword string `json:"confirmpassword"`
}

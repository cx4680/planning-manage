package customer

import "time"

type CreateCustomerRequest struct {
	CustomerName string   `json:"customerName"`
	LeaderId     string   `json:"leaderId"`
	LeaderName   string   `json:"leaderName"`
	MembersId    []string `json:"membersId"`
	MembersName  []string `json:"membersName"`
}

type PageCustomerRequest struct {
	Current      int    `json:"current"`
	Size         int    `json:"size"`
	CustomerName string `json:"customerName"`
	LeaderName   string `json:"leaderName"`
	CustomerId   int64  `json:"customerId"`
}

type CustomerResponse struct {
	ID           int64  `json:"id"`
	CustomerName string `json:"customerName"`
	LeaderId     string `json:"leaderId"`
	LeaderName   string `json:"leaderName"`
	//MembersId    []string `json:"membersId"`
	MembersName []string  `json:"membersName"`
	CreateTime  time.Time `json:"createTime"`
	UpdateTime  time.Time `json:"updateTime"`
	Editable    bool      `json:"editable"`
}

type UpdateCustomerRequest struct {
	ID           int64    `json:"id"`
	CustomerName string   `json:"customerName"`
	LeaderId     string   `json:"leaderId"`
	LeaderName   string   `json:"leaderName"`
	MembersId    []string `json:"membersId"`
	MembersName  []string `json:"membersName"`
}

package user

const (
	RequestSuccessCode = 20000
	RequestFailCode    = 50000
)

type TokenCheckRequest struct {
	ProductCode string `json:"productCode"`
	CestcToken  string `json:"cestcToken"`
}

type TokenCheckResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		DeptNum         string `json:"deptNum"`
		DisplayName     string `json:"displayName"`
		CestcToken      string `json:"cestcToken"`
		WorkNum         string `json:"workNum"`
		ExpireFreshTime int64  `json:"expireFreshTime"`
		ExpireMinite    int64  `json:"expireMinite"`
		L               string `json:"l"`
		Email           string `json:"email"`
		DeptName        string `json:"deptName"`
		Mobile          string `json:"mobile"`
		UserName        string `json:"userName"`
		DeptRoot        string `json:"deptRoot"`
		RequestIp       string `json:"requestIp"`
	} `json:"data"`
}

type QueryUserByUidRequest struct {
	ProductCode string   `json:"productCode"`
	Sign        string   `json:"sign"`
	Timestamp   string   `json:"timestamp"`
	Data        []string `json:"data"`
}

type QueryUserByUidResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    []struct {
		Highlight                  string   `json:"highlight"`
		Uid                        string   `json:"uid"`
		EmployeeNumber             string   `json:"employeeNumber"`
		PhysicalDeliveryOfficeName string   `json:"physicalDeliveryOfficeName"`
		DisplayName                string   `json:"displayName"`
		TelephoneNumber            string   `json:"telephoneNumber"`
		Mail                       string   `json:"mail"`
		L                          string   `json:"l"`
		WorkPlace                  string   `json:"workPlace"`
		ImageUrl                   string   `json:"imageUrl"`
		DepartmentNumber           string   `json:"departmentNumber"`
		DataType                   string   `json:"dataType"`
		CreateTime                 string   `json:"createTime"`
		AlternatePhone             string   `json:"alternatePhone"`
		Sex                        int      `json:"sex"`
		Groups                     []string `json:"groups"`
		Roles                      []string `json:"roles"`
		PostCode                   string   `json:"postcode"`
		PostName                   string   `json:"postname"`
	} `json:"data"`
}

type LogoutResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

type QueryUserByEmployeeNumRequest struct {
	ProductCode string `json:"productCode"`
	Sign        string `json:"sign"`
	Timestamp   string `json:"timestamp"`
	Data        struct {
		EmpNumList []string `json:"empNumList"`
	} `json:"data"`
}

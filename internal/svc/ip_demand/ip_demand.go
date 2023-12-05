package ip_demand

import (
	"github.com/gin-gonic/gin"
)

type IpDemandBaselineDto struct {
	ID           int64  `form:"id"`
	VersionId    int64  `form:"versionId"`
	Vlan         string `form:"vlan"`
	Explain      string `form:"explain"`
	Description  string `form:"description"`
	IpSuggestion string `form:"ipSuggestion"`
	AssignNum    string `form:"assignNum"`
	Remark       string `form:"remark"`
	DeviceRoleId int64  `form:"deviceRoleId"`
}

func IpDemandListDownload(c *gin.Context) {

	return
}

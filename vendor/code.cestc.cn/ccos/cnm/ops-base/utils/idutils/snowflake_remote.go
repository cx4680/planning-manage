package idutils

import (
	"context"
	"fmt"

	"code.cestc.cn/ccos/cnm/ops-base/logging"
	"code.cestc.cn/ccos/cnm/ops-base/opshttp/client"
)

type snowflakeRemote struct {
	serviceName string
	host        string
}

func newSnowflakeRemote(serviceName, host string) IdExec {
	return &snowflakeRemote{
		serviceName: serviceName,
		host:        host,
	}
}

func (s *snowflakeRemote) generate() int64 {
	type remoteId struct {
		ID int64 `json:"id"`
	}

	url := getUrl()

	var reply remoteId
	err := client.NewReq(context.Background(), client.DefaultClient()).Get(fmt.Sprintf("%s%s?serviceName=%s&host=%s", url, remoteIdPath, s.serviceName, s.host)).Response().ParseJson(&reply)
	if err != nil {
		logging.Fatal(err)
	}
	return reply.ID
}

package dataid

import (
	"math/rand"
	"time"

	"github.com/yitter/idgenerator-go/idgen"
	"gopkg.in/mgo.v2/bson"
)

func DataID() string {
	return bson.NewObjectId().Hex()
}

func InitIDWorker() {
	rand.Seed(time.Now().UnixNano())
	options := idgen.NewIdGeneratorOptions(uint16(rand.Intn(64)))
	idgen.SetIdGenerator(options)
}

// NextID 雪花 ID
func NextID() int64 {
	return idgen.NextId()
}

package context

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/utils"
	"strings"
)

const (
	maxUIDLength      = 10
	maxTestName       = 15
	maxServerName     = 15
	maxClientAlias    = 15
	maxDatatypeKey    = 15
	maxCollectionName = 15
)

func MakeTagInRPCProcess(tag2 string, collectionNum uint32, cuid string) string {
	return fmt.Sprintf("%s|%d|%.*s", tag2, collectionNum, maxUIDLength, strings.ToUpper(cuid))
}

func MakeTagInPushPull(tag2 string, collectionNum uint32, cuid string, duid string) string {
	return fmt.Sprintf("%s|%d|%.*s|%.*s",
		tag2,
		collectionNum,
		maxUIDLength, strings.ToUpper(cuid),
		maxUIDLength, strings.ToLower(duid))
}

func MakeTagInClient(collectionName string, clientAlias string, cuid string) string {
	return fmt.Sprintf("%s|%s|%.*s",
		utils.TrimLong(collectionName, maxCollectionName),
		utils.TrimLong(clientAlias, maxClientAlias),
		maxUIDLength, strings.ToUpper(cuid))
}

func MakeTagInDatatype(collectionName string, key string, cuid, duid string) string {
	return fmt.Sprintf("%s|%s|%.*s|%.*s",
		utils.TrimLong(collectionName, maxCollectionName),
		utils.TrimLong(key, maxDatatypeKey),
		maxUIDLength, strings.ToUpper(cuid),
		maxUIDLength, strings.ToLower(duid))
}

func MakeTagInServer(host string, port int) string {
	return fmt.Sprintf("%s:%d", utils.TrimLong(host, maxServerName), port)
}

func MakeTagInTest(test string) string {
	return fmt.Sprintf("%s", utils.TrimLong(test, maxTestName))
}

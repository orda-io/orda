package context

import (
	"fmt"

	"github.com/orda-io/orda/pkg/utils"
)

const (
	maxTestName       = 15
	maxServerName     = 15
	maxClientAlias    = 15
	maxDatatypeKey    = 15
	maxCollectionName = 15
)

func MakeTagInRPCProcess(tag2 string, collectionNum uint32, cuid string) string {
	return fmt.Sprintf("%s|%d|%s", tag2, collectionNum, cuid)
}

func MakeTagInPushPull(tag2 string, collectionNum uint32, cuid string, duid string) string {
	return fmt.Sprintf("%s|%d|%s|%s", tag2, collectionNum, cuid, duid)
}

func MakeTagInClient(collectionName string, clientAlias string, cuid string) string {
	return fmt.Sprintf("%s|%s|%s",
		utils.MakeShort(collectionName, maxCollectionName),
		utils.MakeShort(clientAlias, maxClientAlias),
		cuid)
}

func MakeTagInDatatype(collectionName string, key string, cuid, duid string) string {
	return fmt.Sprintf("%s|%s|%s|%s",
		utils.MakeShort(collectionName, maxCollectionName),
		utils.MakeShort(key, maxDatatypeKey),
		cuid,
		duid)
}

func MakeTagInServer(host string, port int) string {
	return fmt.Sprintf("%s:%d", utils.MakeShort(host, maxServerName), port)
}

func MakeTagInTest(test string) string {
	return utils.MakeShort(test, maxTestName)
}

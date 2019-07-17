package commons

type pushPullPack struct {
	duid       Duid
	checkPoint CheckPoint
	operations []Operation
}

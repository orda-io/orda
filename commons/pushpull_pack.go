package commons

type pushPullPack struct {
	duid       datatypeUID
	checkPoint CheckPoint
	operations []Operation
}

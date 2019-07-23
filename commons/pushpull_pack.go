package commons

import "github.com/knowhunger/ortoo/commons/model"

type pushPullPack struct {
	duid       model.Duid
	checkPoint model.CheckPoint
	operations []model.Operationer
}

package client

import (
	"github.com/knowhunger/ortoo/commons/internal/datatypes"
	"github.com/knowhunger/ortoo/commons/model"
)

//DataManager manages Ortoo datatypes regarding operations
type DataManager struct {
	//cuid model.Cuid

	dataMap map[string]model.FinalDatatype
}

//DeliverOperation delivers an operation
func (d *DataManager) DeliverOperation(wired datatypes.WiredDatatype, op model.Operation) {
	panic("implement me")
}

//DeliverTransaction delivers a transaction
func (d *DataManager) DeliverTransaction(wired datatypes.WiredDatatype, transaction []model.Operation) {
	panic("implement me")
}

func newDataManager(cuid model.Cuid) *DataManager {
	return &DataManager{
		dataMap: make(map[string]model.FinalDatatype),
	}
}

//Get returns a datatype
func (d *DataManager) Get(key string) model.FinalDatatype {
	dt, ok := d.dataMap[key]
	if ok {
		return dt.GetFinalDatatype()
	}
	return nil
}

//Link links a datatype with the datatype
func (d *DataManager) Link(dt model.FinalDatatype) {
	if _, ok := d.dataMap[dt.GetKey()]; !ok {
		d.dataMap[dt.GetKey()] = dt
	}

}

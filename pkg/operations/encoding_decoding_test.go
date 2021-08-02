package operations

import (
	"encoding/json"
	"github.com/orda-io/orda/pkg/log"
	"github.com/orda-io/orda/pkg/model"
	"testing"
)

func TestConvertingOperations(t *testing.T) {
	t.Run("Can convert operations", func(t *testing.T) {
		snapBody := &struct {
			A string
			B int32
		}{A: "abc", B: 1234}
		snapb, _ := json.Marshal(snapBody)

		sOp := NewSnapshotOperation(model.TypeOfDatatype_DOCUMENT, snapb)
		sOpModel := sOp.ToModelOperation()
		log.Logger.Infof("%v", sOp)
		log.Logger.Infof("%v", sOpModel)
		// sOpModel.Body = []byte("{\"Type\":3,\"Snapshot\":\"{\\\"nm\\\":[{\\\"t\\\":\\\"O\\\",\\\"c\\\":{\\\"c\\\":\\\"0000000000000000\\\"},\\\"o\\\":{\\\"m\\\":{},\\\"s\\\":0}}]}\"}")
		// log.Logger.Infof("%v", sOpModel)
		sOp2 := ModelToOperation(sOpModel)
		//
		// log.Logger.Infof("%v", sOpModel)
		log.Logger.Infof("%v", sOp2)

	})
}

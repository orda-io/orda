package datatypes

import (
	"bytes"
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/errors"
	"github.com/knowhunger/ortoo/ortoo/iface"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/knowhunger/ortoo/ortoo/types"
)

// BaseDatatype is the base datatype which contains
type BaseDatatype struct {
	Key      string
	id       types.DUID
	opID     *model.OperationID
	TypeOf   model.TypeOfDatatype
	state    model.StateOfDatatype
	datatype iface.Datatype
	Logger   *log.OrtooLog
}

func NewBaseDatatype(key string, t model.TypeOfDatatype, cuid types.CUID) *BaseDatatype {
	duid := types.NewDUID()
	return &BaseDatatype{
		Key:    key,
		id:     duid,
		TypeOf: t,
		opID:   model.NewOperationIDWithCUID(cuid),
		state:  model.StateOfDatatype_DUE_TO_CREATE,
		Logger: log.NewOrtooLogWithTag(fmt.Sprintf("%s:%s:%s", t, key, cuid.ShortString())),
	}
}

// GetCUID returns CUID of the client which this datatype subecribes to.
func (its *BaseDatatype) GetCUID() string {
	return types.ToUID(its.opID.CUID)
}

// GetEra returns the era of operation ID.
func (its *BaseDatatype) GetEra() uint32 {
	return its.opID.GetEra()
}

func (its *BaseDatatype) String() string {
	return fmt.Sprintf("%s", its.id)
}

func (its *BaseDatatype) executeLocalBase(op iface.Operation) (interface{}, errors.OrtooError) {
	its.SetNextOpID(op)
	// TODO: should deal with NO_OP
	return op.ExecuteLocal(its.datatype)

}

// Replay replays an already executed operation.
func (its *BaseDatatype) Replay(op iface.Operation) errors.OrtooError {
	if bytes.Compare(its.opID.CUID, op.GetID().CUID) == 0 {
		_, err := its.executeLocalBase(op)
		if err != nil { // TODO: if an operation fails to be executed, opID should be rollbacked.
			return err
		}
	} else {
		its.executeRemoteBase(op)
	}
	return nil
}

// SetNextOpID proceeds the operation ID next.
func (its *BaseDatatype) SetNextOpID(op iface.Operation) {
	op.SetOperationID(its.opID.Next())
}

func (its *BaseDatatype) executeRemoteBase(op iface.Operation) {
	op.ExecuteRemote(its.datatype)
}

// GetType returns the type of this datatype.
func (its *BaseDatatype) GetType() model.TypeOfDatatype {
	return its.TypeOf
}

// GetState returns the state of this datatype.
func (its *BaseDatatype) GetState() model.StateOfDatatype {
	return its.state
}

// SetDatatype sets the Datatype which implements this BaseDatatype.
func (its *BaseDatatype) SetDatatype(datatype iface.Datatype) {
	its.datatype = datatype
}

// GetDatatype returns the Datatype which implements this BaseDatatype.
func (its *BaseDatatype) GetDatatype() iface.Datatype {
	return its.datatype
}

// SetOpID sets the operation ID.
func (its *BaseDatatype) SetOpID(opID *model.OperationID) {
	its.opID = opID
}

// GetKey returns the key.
func (its *BaseDatatype) GetKey() string {
	return its.Key
}

// GetDUID returns DUID.
func (its *BaseDatatype) GetDUID() types.DUID {
	return its.id
}

// SetDUID sets the DUID.
func (its *BaseDatatype) SetDUID(duid types.DUID) {
	its.id = duid
}

// SetState sets the state of this datatype.
func (its *BaseDatatype) SetState(state model.StateOfDatatype) {
	its.state = state
}

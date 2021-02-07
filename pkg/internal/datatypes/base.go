package datatypes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/knowhunger/ortoo/pkg/context"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/iface"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/types"
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

// NewBaseDatatype creates a new base datatype
func NewBaseDatatype(key string, t model.TypeOfDatatype, client *model.Client) *BaseDatatype {
	duid := types.NewDUID()
	return &BaseDatatype{
		Key:    key,
		id:     duid,
		TypeOf: t,
		opID:   model.NewOperationIDWithCUID(client.CUID),
		state:  model.StateOfDatatype_DUE_TO_CREATE,
		Logger: log.NewWithTags(string(context.DATATYPE),
			context.MakeTagInDatatype(client.Collection, key, client.CUID, duid)),
	}
}

// GetCUID returns CUID of the client which this datatype subecribes to.
func (its *BaseDatatype) GetCUID() string {
	return types.UIDtoString(its.opID.CUID)
}

// GetEra returns the era of operation ID.
func (its *BaseDatatype) GetEra() uint32 {
	return its.opID.GetEra()
}

func (its *BaseDatatype) SetLogger(l *log.OrtooLog) {
	its.Logger = l
}

func (its *BaseDatatype) String() string {
	return fmt.Sprintf("%s", its.id)
}

func (its *BaseDatatype) executeLocalBase(op iface.Operation) (interface{}, errors.OrtooError) {
	its.SetNextOpID(op)
	ret, err := op.ExecuteLocal(its.datatype)
	if err != nil {
		its.opID.RollBack()
	}
	return ret, err // should deliver err
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
	op.SetID(its.opID.Next())
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

func (its *BaseDatatype) GetOpID() *model.OperationID {
	return its.opID
}

// GetMeta returns the binary of metadata of the datatype.
func (its *BaseDatatype) GetMeta() ([]byte, errors.OrtooError) {
	meta := model.DatatypeMeta{
		Key:    its.Key,
		DUID:   its.id,
		OpID:   its.opID,
		TypeOf: its.TypeOf,
		State:  its.state,
	}
	metab, err := json.Marshal(&meta)
	if err != nil {
		return nil, errors.DatatypeMarshal.New(its.Logger, meta)
	}
	return metab, nil
}

// SetMeta sets the metadata with binary metadata.
func (its *BaseDatatype) SetMeta(meta []byte) errors.OrtooError {
	m := model.DatatypeMeta{}
	if err := json.Unmarshal(meta, &m); err != nil {
		return errors.DatatypeMarshal.New(its.Logger, string(meta))
	}
	its.Key = m.Key
	its.id = m.DUID
	its.opID = m.OpID
	its.TypeOf = m.TypeOf
	its.state = m.State
	return nil
}

func (its *BaseDatatype) GetLogger() *log.OrtooLog {
	return its.Logger
}

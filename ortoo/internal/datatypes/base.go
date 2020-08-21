package datatypes

import (
	"bytes"
	"encoding/hex"
	"fmt"
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

func newBaseDatatype(key string, t model.TypeOfDatatype, cuid types.CUID) *BaseDatatype {
	duid := types.NewDUID()
	return &BaseDatatype{
		Key:    key,
		id:     duid,
		TypeOf: t,
		opID:   model.NewOperationIDWithCUID(cuid),
		state:  model.StateOfDatatype_DUE_TO_CREATE,
		Logger: log.NewOrtooLogWithTag(fmt.Sprintf("%s:%s", key, hex.EncodeToString(cuid)[:8])),
	}
}

// GetCUID returns CUID of the client which this datatype subecribes to.
func (b *BaseDatatype) GetCUID() string {
	return hex.EncodeToString(b.opID.CUID)
}

// GetEra returns the era of operation ID.
func (b *BaseDatatype) GetEra() uint32 {
	return b.opID.GetEra()
}

func (b *BaseDatatype) String() string {
	return fmt.Sprintf("%s", b.id)
}

func (b *BaseDatatype) executeLocalBase(op iface.Operation) (interface{}, error) {
	b.SetNextOpID(op)
	// TODO: should deal with NO_OP
	return op.ExecuteLocal(b.datatype)

}

// Replay replays an already executed operation.
func (b *BaseDatatype) Replay(op iface.Operation) error {
	if bytes.Compare(b.opID.CUID, op.GetID().CUID) == 0 {
		_, err := b.executeLocalBase(op)
		if err != nil { // TODO: if an operation fails to be executed, opID should be rollbacked.
			return log.OrtooErrorf(err, "fail to replay local operation")
		}
	} else {
		b.executeRemoteBase(op)
	}
	return nil
}

// SetNextOpID proceeds the operation ID next.
func (b *BaseDatatype) SetNextOpID(op iface.Operation) {
	op.SetOperationID(b.opID.Next())
}

func (b *BaseDatatype) executeRemoteBase(op iface.Operation) {
	op.ExecuteRemote(b.datatype)
}

// GetType returns the type of this datatype.
func (b *BaseDatatype) GetType() model.TypeOfDatatype {
	return b.TypeOf
}

// GetState returns the state of this datatype.
func (b *BaseDatatype) GetState() model.StateOfDatatype {
	return b.state
}

// SetDatatype sets the Datatype which implements this BaseDatatype.
func (b *BaseDatatype) SetDatatype(datatype iface.Datatype) {
	b.datatype = datatype
}

// GetDatatype returns the Datatype which implements this BaseDatatype.
func (b *BaseDatatype) GetDatatype() iface.Datatype {
	return b.datatype
}

// SetOpID sets the operation ID.
func (b *BaseDatatype) SetOpID(opID *model.OperationID) {
	b.opID = opID
}

// GetKey returns the key.
func (b *BaseDatatype) GetKey() string {
	return b.Key
}

// GetDUID returns DUID.
func (b *BaseDatatype) GetDUID() types.DUID {
	return b.id
}

// SetDUID sets the DUID.
func (b *BaseDatatype) SetDUID(duid types.DUID) {
	b.id = duid
}

// SetState sets the state of this datatype.
func (b *BaseDatatype) SetState(state model.StateOfDatatype) {
	b.state = state
}

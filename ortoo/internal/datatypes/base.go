package datatypes

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
	operations "github.com/knowhunger/ortoo/ortoo/operations"
)

// BaseDatatype is the base datatype which contains
type BaseDatatype struct {
	Key      string
	id       model.DUID
	opID     *model.OperationID
	TypeOf   model.TypeOfDatatype
	state    model.StateOfDatatype
	datatype model.Datatype
	Logger   *log.OrtooLog
}

// PublicBaseDatatypeInterface is a public interface for a datatype.
type PublicBaseDatatypeInterface interface {
	GetType() model.TypeOfDatatype
	GetState() model.StateOfDatatype
	GetAsJSON() (string, error)
}

func newBaseDatatype(key string, t model.TypeOfDatatype, cuid model.CUID) *BaseDatatype {
	duid := model.NewDUID()
	return &BaseDatatype{
		Key:    key,
		id:     duid,
		TypeOf: t,
		opID:   model.NewOperationIDWithCuid(cuid),
		state:  model.StateOfDatatype_DUE_TO_CREATE,
		Logger: log.NewOrtooLogWithTag(fmt.Sprintf("%s", duid)[:8]),
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

func (b *BaseDatatype) executeLocalBase(op operations.Operation) (interface{}, error) {
	b.SetNextOpID(op)
	return op.ExecuteLocal(b.datatype)
}

// Replay replays an already executed operation.
func (b *BaseDatatype) Replay(op operations.Operation) error {
	if bytes.Compare(b.opID.CUID, op.GetID().CUID) == 0 {
		_, err := b.executeLocalBase(op)
		if err != nil {
			return log.OrtooErrorf(err, "fail to replay local operation")
		}
	} else {
		b.executeRemoteBase(op)
	}
	return nil
}

// SetNextOpID proceeds the operation ID next.
func (b *BaseDatatype) SetNextOpID(op operations.Operation) {
	op.SetOperationID(b.opID.Next())
}

func (b *BaseDatatype) executeRemoteBase(op operations.Operation) {
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
func (b *BaseDatatype) SetDatatype(datatype model.Datatype) {
	b.datatype = datatype
}

// GetDatatype returns the Datatype which implements this BaseDatatype.
func (b *BaseDatatype) GetDatatype() model.Datatype {
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
func (b *BaseDatatype) GetDUID() model.DUID {
	return b.id
}

// SetDUID sets the DUID.
func (b *BaseDatatype) SetDUID(duid model.DUID) {
	b.id = duid
}

// SetState sets the state of this datatype.
func (b *BaseDatatype) SetState(state model.StateOfDatatype) {
	b.state = state
}

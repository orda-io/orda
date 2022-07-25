package datatypes

import (
	"encoding/json"
	"fmt"
	"github.com/orda-io/orda/client/pkg/context"
	errors2 "github.com/orda-io/orda/client/pkg/errors"
	iface2 "github.com/orda-io/orda/client/pkg/iface"
	"github.com/orda-io/orda/client/pkg/log"
	model2 "github.com/orda-io/orda/client/pkg/model"
	"github.com/orda-io/orda/client/pkg/types"
	"github.com/orda-io/orda/client/pkg/utils"
)

// BaseDatatype is the base datatype which contains
type BaseDatatype struct {
	Key    string
	id     string
	opID   *model2.OperationID
	TypeOf model2.TypeOfDatatype
	state  model2.StateOfDatatype
	ctx    *context.DatatypeContext
	iface2.Datatype
}

// NewBaseDatatype creates a new base datatype
func NewBaseDatatype(
	key string,
	t model2.TypeOfDatatype,
	clientCtx *context.ClientContext,
	state model2.StateOfDatatype,
) *BaseDatatype {
	duid := types.NewUID()
	base := &BaseDatatype{
		Key:    key,
		id:     duid,
		TypeOf: t,
		opID:   model2.NewOperationIDWithCUID(clientCtx.Client.CUID),
		state:  state,
	}
	base.ctx = context.NewDatatypeContext(clientCtx, base)
	return base
}

// GetCUID returns CUID of the client which this datatype subscribes to.
func (its *BaseDatatype) GetCUID() string {
	return its.opID.CUID
}

// GetEra returns the era of operation ID.
func (its *BaseDatatype) GetEra() uint32 {
	return its.opID.GetEra()
}

func (its *BaseDatatype) SetLogger(l *log.OrdaLog) {
	its.ctx.SetLogger(l)
}

func (its *BaseDatatype) String() string {
	return fmt.Sprintf("%s", its.id)
}

func (its *BaseDatatype) executeLocalBase(op iface2.Operation) (interface{}, errors2.OrdaError) {
	its.SetNextOpID(op)
	if op.GetType() == model2.TypeOfOperation_TRANSACTION ||
		op.GetType() == model2.TypeOfOperation_ERROR ||
		op.GetType()%10 == 0 {
		return nil, nil
	}
	ret, err := its.ExecuteLocal(op)
	if err != nil {
		its.opID.RollBack()
	}
	return ret, err // should deliver err
}

func (its *BaseDatatype) executeRemoteBase(op iface2.Operation) {
	its.opID.SyncLamport(op.GetID().Lamport)
	_, _ = its.ExecuteRemote(op)
}

// Replay replays an already executed operation.
func (its *BaseDatatype) Replay(op iface2.Operation) errors2.OrdaError {
	if its.opID.CUID == op.GetID().CUID {
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
func (its *BaseDatatype) SetNextOpID(op iface2.Operation) {
	op.SetID(its.opID.Next())
}

// GetType returns the type of this datatype.
func (its *BaseDatatype) GetType() model2.TypeOfDatatype {
	return its.TypeOf
}

// GetState returns the state of this datatype.
func (its *BaseDatatype) GetState() model2.StateOfDatatype {
	return its.state
}

// SetOpID sets the operation ID.
func (its *BaseDatatype) SetOpID(opID *model2.OperationID) {
	its.opID = opID
}

// GetKey returns the key.
func (its *BaseDatatype) GetKey() string {
	return its.Key
}

// GetDUID returns DUID.
func (its *BaseDatatype) GetDUID() string {
	return its.id
}

// SetDUID sets the DUID.
func (its *BaseDatatype) SetDUID(duid string) {
	its.id = duid
}

// SetState sets the state of this datatype.
func (its *BaseDatatype) SetState(state model2.StateOfDatatype) {
	its.state = state
}

func (its *BaseDatatype) GetOpID() *model2.OperationID {
	return its.opID
}

// GetMeta returns the binary of metadata of the datatype.
func (its *BaseDatatype) GetMeta() ([]byte, errors2.OrdaError) {
	meta := model2.DatatypeMeta{
		Key:    its.Key,
		TypeOf: its.TypeOf,
		DUID:   its.id,
		OpID:   its.opID,
	}
	metab, err := json.Marshal(meta)
	if err != nil {
		return nil, errors2.DatatypeMarshal.New(its.ctx.L(), meta)
	}
	return metab, nil
}

// SetMeta sets the metadata with binary metadata.
func (its *BaseDatatype) SetMeta(meta []byte) errors2.OrdaError {
	m := model2.DatatypeMeta{}
	if err := json.Unmarshal(meta, &m); err != nil {
		return errors2.DatatypeMarshal.New(its.ctx.L(), string(meta))
	}
	its.Key = m.Key
	its.id = m.DUID
	its.opID = m.OpID
	its.TypeOf = m.TypeOf
	return nil
}

func (its *BaseDatatype) L() *log.OrdaLog {
	return its.ctx.L()
}

func (its *BaseDatatype) GetSummary() string {
	return fmt.Sprintf("%s(%s)", utils.MakeDefaultShort(its.Key), its.id)
}

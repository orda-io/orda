package model

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/types"
	"github.com/knowhunger/ortoo/pkg/utils"
)

// NewMessageHeader generates a message header.
func NewMessageHeader(seq uint32, typeOf TypeOfMessage, collection string, clientAlias string, cuid []byte) *MessageHeader {
	return &MessageHeader{
		Version:     ProtocolVersion,
		Seq:         seq,
		TypeOf:      typeOf,
		Collection:  collection,
		ClientAlias: clientAlias,
		Cuid:        cuid,
	}
}

func (its *MessageHeader) GetClient() string {
	return fmt.Sprintf("%s(%s)", its.ClientAlias, types.UIDtoString(its.Cuid))
}

func (its *MessageHeader) GetClientSummary() string {
	return utils.MakeSummary(its.ClientAlias, its.Cuid, true)
}

// ToString returns customized string
func (its *MessageHeader) ToString() string {
	return fmt.Sprintf("v%s|%d|%s|%s|%s",
		its.Version, its.Seq, its.TypeOf.String(), its.Collection, types.UID(its.Cuid).ShortString())
}

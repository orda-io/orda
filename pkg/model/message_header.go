package model

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/types"
)

// NewMessageHeader generates a message header.
func NewMessageHeader(typeOf RequestType, collection string, clientAlias string, cuid []byte) *Header {
	return &Header{
		Version:     ProtocolVersion,
		Collection:  collection,
		Type:        typeOf,
		ClientAlias: clientAlias,
		Cuid:        cuid,
	}
}

func (its *Header) GetClient() string {
	return fmt.Sprintf("%s(%s)", its.ClientAlias, types.UIDtoString(its.Cuid))
}

// func (its *MessageHeader) GetClientSummary() string {
// 	return utils.MakeSummary(its.ClientAlias, its.Cuid, true)
// }

// ToString returns customized string
func (its *Header) ToString() string {
	return fmt.Sprintf("v%s|%s|%s|%s",
		its.Version, its.Type, its.Collection, types.UID(its.Cuid).ShortString())
}

package model

import (
	"fmt"
	"github.com/orda-io/orda/client/pkg/constants"
)

// NewMessageHeader generates a message header.
func NewMessageHeader(typeOf RequestType) *Header {
	return &Header{
		Version: ProtocolVersion,
		Agent:   fmt.Sprintf("%s-%v", constants.SDKType, constants.Version),
		Type:    typeOf,
	}
}

// ToString returns customized string
func (its *Header) ToString() string {
	return fmt.Sprintf("%s|%s|%s", its.Version, its.Type, its.Agent)
}

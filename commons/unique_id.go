package commons

import (
	"bytes"
	"encoding/binary"
	"github.com/google/uuid"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/protocols"
)

type uniqueID struct {
	*uuid.UUID
}

func newUniqueID() (*uniqueID, error) {
	u, err := uuid.NewUUID()
	if err != nil {
		return nil, log.OrtooError(err, "cannot generate unique ID")
	}
	return &uniqueID{&u}, nil
}

func (d *uniqueID) getPb() (*protocols.PbUuid, error) {
	// []byte to int64
	bin, err := d.MarshalBinary()
	if err != nil {
		return nil, log.Logger.OrtooError(err, "fails to marshal datatype ID")
	}
	var head, tail int64
	if err = binary.Read(bytes.NewReader(bin[:8]), binary.BigEndian, &head); err != nil {
		return nil, log.OrtooError(err, "fail to encode protobuf of datatype ID")
	}
	if err = binary.Read(bytes.NewReader(bin[8:]), binary.BigEndian, &tail); err != nil {
		return nil, log.OrtooError(err, "fail to encode protobuf of datatype ID")
	}

	return &protocols.PbUuid{
		Head: head,
		Tail: tail,
	}, nil
}

func newUniqueIDFromPb(pb *protocols.PbUuid) (*uniqueID, error) {

	// int64 to []byte
	var bin []byte

	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, &pb.Head); err != nil {
		return nil, log.OrtooError(err, "fail to decode protobuf of datatype ID")
	}
	bin = append(bin, buf.Bytes()...)

	buf = new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, &pb.Tail); err != nil {
		return nil, log.OrtooError(err, "fail to decode protobuf of datatype ID")
	}
	bin = append(bin, buf.Bytes()...)

	u, err := uuid.FromBytes(bin)
	if err != nil {
		return nil, log.OrtooError(err, "fail to make uuid from binary")
	}
	return &uniqueID{UUID: &u}, nil
}

package model

import (
	"github.com/google/uuid"
	"github.com/knowhunger/ortoo/commons/log"
)

type uniqueID []byte

func newUniqueID() (uniqueID, error) {
	u, err := uuid.NewUUID()
	if err != nil {
		return nil, log.OrtooError(err, "fail to generate unique ID")
	}
	b, err := u.MarshalBinary()
	if err != nil {
		return nil, log.OrtooError(err, "fail to generate unique ID")
	}
	return uniqueID(b), nil
}

func (u uniqueID) String() string {
	uid, err := uuid.FromBytes([]byte(u))
	if err != nil {
		return "fail to make string to uuid"
	}
	return uid.String()
}

//
//
//
//func (d *uniqueID) toProtoBuf() (*model.PbUuid, error) {
//	// []byte to int64
//	bin, err := d.MarshalBinary()
//	if err != nil {
//		return nil, log.Logger.OrtooError(err, "fail to marshal datatype ID")
//	}
//	var head, tail int64
//	if err = binary.Read(bytes.NewReader(bin[:8]), binary.BigEndian, &head); err != nil {
//		return nil, log.OrtooError(err, "fail to encode protobuf of datatype ID")
//	}
//	if err = binary.Read(bytes.NewReader(bin[8:]), binary.BigEndian, &tail); err != nil {
//		return nil, log.OrtooError(err, "fail to encode protobuf of datatype ID")
//	}
//
//	return &model.PbUuid{
//		Head: head,
//		Tail: tail,
//	}, nil
//}
//
//func pbToUniqueID(pb *model.PbUuid) (*uniqueID, error) {
//
//	// int64 to []byte
//	var bin []byte
//
//	buf := new(bytes.Buffer)
//	if err := binary.Write(buf, binary.BigEndian, &pb.Head); err != nil {
//		return nil, log.OrtooError(err, "fail to decode protobuf of datatype ID")
//	}
//	bin = append(bin, buf.Bytes()...)
//
//	buf = new(bytes.Buffer)
//	if err := binary.Write(buf, binary.BigEndian, &pb.Tail); err != nil {
//		return nil, log.OrtooError(err, "fail to decode protobuf of datatype ID")
//	}
//	bin = append(bin, buf.Bytes()...)
//
//	u, err := uuid.FromBytes(bin)
//	if err != nil {
//		return nil, log.OrtooError(err, "fail to make uuid from binary")
//	}
//	return &uniqueID{UUID: &u}, nil
//}

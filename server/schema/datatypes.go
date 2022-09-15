package schema

import (
	"encoding/json"
	"github.com/orda-io/orda/client/pkg/model"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

// XXClient is used to return the type of client
const (
	RWClient = "rw"
	ROClient = "ro"
	NoClient = "no"
)

// DatatypeDoc defines a MongoDB document for datatype, stored in MongoDB
type DatatypeDoc struct {
	DUID               string `json:"_id" bson:"_id"`
	UpdatedDatatypeDoc `json:",inline" bson:",inline"`
}

// SseqSet is a set of Sseq(Server Sequence)
type SseqSet struct {
	Begin uint64 `json:"begin" bson:"begin"`
	End   uint64 `json:"end" bson:"end"`
	Safe  uint64 `json:"safe" bson:"safe"`
}

// UpdatedDatatypeDoc defines a MongoDB document for datatype, which are updated
type UpdatedDatatypeDoc struct {
	Key           string  `json:"key" bson:"key"`
	CollectionNum int32   `json:"colNum" bson:"colNum"`
	Type          string  `json:"type" bson:"type"`
	Sseq          SseqSet `json:"sseq" bson:"sseq"`
	// SseqBegin uint64                          `json:"sseqBegin" bson:"sseqBegin"`
	// SseqEnd   uint64                          `json:"sseqEnd" bson:"sseqEnd"`
	// SseqSafe  uint64                          `json:"sseqSafe" bson:"sseqSafe"`
	Visible   bool                            `json:"visible" bson:"visible"`
	CreatedAt time.Time                       `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time                       `json:"updatedAt" bson:"updatedAt"`
	RWClients map[string]*SubscribedClientDoc `json:"rwClients" bson:"rwClients"`
	ROClients map[string]*SubscribedClientDoc `json:"roClients" bson:"roClients"`
}

// DatatypeDocFields defines the fields of DatatypeDoc
var DatatypeDocFields = struct {
	DUID          string
	Key           string
	CollectionNum string
	Type          string
	SseqBegin     string
	SseqEnd       string
	SseqSafe      string
	Visible       string
	CreatedAt     string
	UpdatedAt     string
}{
	DUID:          "_id",
	Key:           "key",
	CollectionNum: "colNum",
	Type:          "type",
	SseqBegin:     "sseqBegin",
	SseqEnd:       "sseqEnd",
	SseqSafe:      "sseqSafe",
	Visible:       "visible",
	CreatedAt:     "createdAt",
	UpdatedAt:     "updatedAt",
}

// NewDatatypeDoc returns a new DatatypeDoc
func NewDatatypeDoc(duid, key string, colNum int32, typ string) *DatatypeDoc {
	return &DatatypeDoc{
		DUID: duid,
		UpdatedDatatypeDoc: UpdatedDatatypeDoc{
			Key:           key,
			CollectionNum: colNum,
			Type:          typ,
			Sseq:          SseqSet{0, 0, 0},
			Visible:       true,
			RWClients:     make(map[string]*SubscribedClientDoc),
			ROClients:     make(map[string]*SubscribedClientDoc),
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}
}

func (its *DatatypeDoc) String() string {
	if its == nil {
		return ""
	}
	b, _ := json.Marshal(its)
	return string(b)
}

// GetIndexModel returns the index models of the collection of ClientDoc
func (its *DatatypeDoc) GetIndexModel() []mongo.IndexModel {
	return []mongo.IndexModel{{
		Keys: bsonx.Doc{
			{DatatypeDocFields.CollectionNum, bsonx.Int32(1)},
			{DatatypeDocFields.Key, bsonx.Int32(1)},
		},
	}}
}

// ToUpdateBSON transforms DatatypeDoc to BSON type
func (its *DatatypeDoc) ToUpdateBSON() bson.D {
	its.UpdatedAt = time.Now()
	d := bson.D{
		{"$set", its.UpdatedDatatypeDoc},
	}

	return d
}

// GetType returns the type of datatype.
func (its *DatatypeDoc) GetType() model.TypeOfDatatype {
	return model.TypeOfDatatype(model.TypeOfDatatype_value[its.Type])
}

// GetClientInDatatypeDoc returns SubscribedClientDoc
func (its *DatatypeDoc) GetClientInDatatypeDoc(cuid string, ro bool) *SubscribedClientDoc {
	if ro {
		if info, ok := its.ROClients[cuid]; ok {
			return info
		}
	} else {
		if info, ok := its.RWClients[cuid]; ok {
			return info
		}
	}
	return nil
}

// HasClientInfo examines if the client info of cuid exists, and returns the client type
func (its *DatatypeDoc) HasClientInfo(cuid string) string {
	if _, ok := its.RWClients[cuid]; ok {
		return RWClient
	}
	if _, ok := its.ROClients[cuid]; ok {
		return ROClient
	}
	return NoClient
}

// AddNewClient adds a new RWClient
func (its *DatatypeDoc) AddNewClient(cuid string, typ int8, ro bool) *SubscribedClientDoc {
	clientDoc := &SubscribedClientDoc{
		CP:   model.NewCheckPoint(),
		Type: typ,
		At:   time.Now(),
	}
	if ro {
		its.ROClients[cuid] = clientDoc
	} else {
		its.RWClients[cuid] = clientDoc
	}
	return clientDoc
}

// SubscribedClientDoc contains the information of a Client
type SubscribedClientDoc struct {
	CP   *model.CheckPoint `bson:"cp"`
	Type int8              `bson:"t"`
	At   time.Time         `bson:"at"`
}

// GetCheckPoint returns *model.CheckPoint from SubscribedClientDoc
func (its *SubscribedClientDoc) GetCheckPoint() *model.CheckPoint {
	return its.CP
}

// UpdateAt updates at with current time
func (its *SubscribedClientDoc) UpdateAt() {
	if its == nil {
		return
	}
	its.At = time.Now()
}

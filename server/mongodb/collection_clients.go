package mongodb

import (
	"context"
	"errors"
	"fmt"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UpdateClient updates a clientDoc; if not exists, a new clientDoc is inserted.
func (m *MongoCollections) UpdateClient(ctx context.Context, client *schema.ClientDoc) error {
	result, err := m.clients.UpdateOne(ctx, schema.FilterByID(client.CUID), client.ToUpdateBSON(), schema.UpsertOption)
	if err != nil {
		return log.OrtooError(err)
	}

	if result.ModifiedCount == 1 || result.UpsertedCount == 1 {
		return nil
	}
	return log.OrtooError(errors.New("fail to update client"))
}

func (m *MongoCollections) UpdateCheckPointInClient(ctx context.Context, cuid, duid string, checkPoint *model.CheckPoint) error {

	x := schema.GetFilter().AddSetCheckPoint(duid, checkPoint)
	result, err := m.clients.UpdateOne(ctx, schema.FilterByID(cuid), bson.D(x), schema.UpsertOption)
	if err != nil {
		return log.OrtooError(err)
	}
	if result.ModifiedCount == 1 {
		return nil
	}
	log.Logger.Warnf("updated no checkpoint of %s in client %s", duid, cuid)
	return nil
}

func (m *MongoCollections) DeleteClient(ctx context.Context, cuid string) error {
	result, err := m.clients.DeleteOne(ctx, schema.FilterByID(cuid))
	if err != nil {
		return log.OrtooError(err)
	}
	if result.DeletedCount == 1 {
		return nil
	}
	log.Logger.Warnf("fail to find a client to delete: %s", cuid)
	return nil
}

func (m *MongoCollections) GetCheckPointFromClient(ctx context.Context, cuid string, duid string) (*model.CheckPoint, error) {
	opts := options.FindOne()
	projectField := fmt.Sprintf("%s.%s", cuid, duid)
	opts.SetProjection(bson.M{projectField: 1})
	sr := m.clients.FindOne(ctx, schema.FilterByID(cuid), opts)
	if err := sr.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, log.OrtooError(err)
	}
	var client schema.ClientDoc
	if err := sr.Decode(&client); err != nil {
		return nil, log.OrtooError(err)
	}
	checkPoint, ok := client.CheckPoints[duid]
	if !ok {
		return nil, nil
	}
	return checkPoint, nil
}

// GetClient gets a client with CUID
func (m *MongoCollections) GetClientWithoutCheckPoints(ctx context.Context, cuid string) (*schema.ClientDoc, error) {
	return m.getClient(ctx, cuid, false)
}

func (m *MongoCollections) GetClient(ctx context.Context, cuid string) (*schema.ClientDoc, error) {
	return m.getClient(ctx, cuid, true)
}

func (m *MongoCollections) getClient(ctx context.Context, cuid string, withCheckPoint bool) (*schema.ClientDoc, error) {
	opts := options.FindOne()
	if !withCheckPoint {
		opts.SetProjection(bson.M{schema.ClientDocFields.CheckPoints: 0})
	}
	sr := m.clients.FindOne(ctx, schema.FilterByID(cuid), opts)
	if err := sr.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, log.OrtooError(err)
	}

	var client schema.ClientDoc
	if err := sr.Decode(&client); err != nil {
		return nil, log.OrtooError(err)
	}
	return &client, nil
}

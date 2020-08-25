package mongodb

import (
	"context"
	"errors"
	"fmt"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/knowhunger/ortoo/pkg/model"
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

// UpdateCheckPointInClient updates a CheckPoint for the given datatype in a client.
func (m *MongoCollections) UpdateCheckPointInClient(ctx context.Context, cuid, duid string, checkPoint *model.CheckPoint) error {

	filter := schema.GetFilter().AddSetCheckPoint(duid, checkPoint)
	result, err := m.clients.UpdateOne(ctx, schema.FilterByID(cuid), bson.D(filter), schema.UpsertOption)
	if err != nil {
		return log.OrtooError(err)
	}
	if result.ModifiedCount == 1 {
		return nil
	}
	log.Logger.Warnf("updated no checkpoint of %s in client %s", duid, cuid)
	return nil
}

// UnsubscribeDatatypeFromClient makes the specified client unsubscribe the specified datatype.
func (m *MongoCollections) UnsubscribeDatatypeFromClient(ctx context.Context, cuid, duid string) error {
	filter := schema.GetFilter().AddUnsetCheckPoint(duid)
	result, err := m.clients.UpdateOne(ctx, schema.FilterByID(cuid), bson.D(filter))
	if err != nil {
		return log.OrtooError(err)
	}
	if result.ModifiedCount == 1 {
		return nil
	}
	log.Logger.Warnf("unsubscribe no client for datatype `%s`", duid)
	return nil
}

// UnsubscribeDatatypeFromAllClient makes the specified datatype unsubscribed from all the clients.
func (m *MongoCollections) UnsubscribeDatatypeFromAllClient(ctx context.Context, duid string) error {
	findFilter := schema.GetFilter().AddExists(fmt.Sprintf("%s.%s", schema.ClientDocFields.CheckPoints, duid))
	updateFilter := schema.GetFilter().AddUnsetCheckPoint(duid)
	result, err := m.clients.UpdateMany(ctx, findFilter, bson.D(updateFilter))
	if err != nil {
		return log.OrtooError(err)
	}
	if result.ModifiedCount > 0 {
		log.Logger.Infof("unsubscribed datatype `%s` form %d clients", duid, result.ModifiedCount)
		return nil
	}
	log.Logger.Warnf("unsubscribe no client for datatype `%s`", duid)
	return nil
}

// DeleteClient deletes the specified client.
func (m *MongoCollections) DeleteClient(ctx context.Context, cuid string) error {
	result, err := m.clients.DeleteOne(ctx, schema.FilterByID(cuid))
	if err != nil {
		return log.OrtooError(err)
	}
	if result.DeletedCount == 1 {
		return nil
	}
	log.Logger.Warnf("fail to find a client to delete: `%s`", cuid)
	return nil
}

// GetCheckPointFromClient returns a checkpoint for the specified datatype from the specified client.
func (m *MongoCollections) GetCheckPointFromClient(ctx context.Context, cuid string, duid string) (*model.CheckPoint, error) {
	opts := options.FindOne()
	projectField := fmt.Sprintf("checkpoints.%s", duid)
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

// GetClientWithoutCheckPoints returns a ClientDoc without CheckPoints for the specified CUID
func (m *MongoCollections) GetClientWithoutCheckPoints(ctx context.Context, cuid string) (*schema.ClientDoc, error) {
	return m.getClient(ctx, cuid, false)
}

// GetClient returns a ClientDoc for the specified CUID.
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

func (m *MongoCollections) purgeAllCollectionClients(ctx context.Context, collectionNum uint32) error {
	filter := schema.GetFilter().AddFilterEQ(schema.ClientDocFields.CollectionNum, collectionNum)
	r1, err := m.clients.DeleteMany(ctx, filter)
	if err != nil {
		return log.OrtooError(err)
	}
	if r1.DeletedCount > 0 {
		log.Logger.Infof("delete %d clients in collection %d", r1.DeletedCount, collectionNum)
		return nil
	}
	log.Logger.Warnf("delete no clients")
	return nil
}

package mongodb

import (
	"fmt"
	"github.com/orda-io/orda/server/schema"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/orda-io/orda/pkg/context"
	"github.com/orda-io/orda/pkg/errors"
	"github.com/orda-io/orda/pkg/model"
)

// UpdateClient updates a clientDoc; if not exists, a new clientDoc is inserted.
func (its *MongoCollections) UpdateClient(
	ctx context.OrdaContext,
	client *schema.ClientDoc,
) errors.OrdaError {
	result, err := its.clients.UpdateOne(ctx, schema.FilterByID(client.CUID), client.ToUpdateBSON(), schema.UpsertOption)
	if err != nil {
		return errors.ServerDBQuery.New(ctx.L(), err.Error())
	}

	if result.ModifiedCount == 1 || result.UpsertedCount == 1 {
		return nil
	}
	return errors.ServerDBQuery.New(ctx.L(), "fail to update client")
}

// UpdateCheckPointInClient updates a CheckPoint for the given datatype in a client.
func (its *MongoCollections) UpdateCheckPointInClient(
	ctx context.OrdaContext,
	cuid string,
	duid string,
	checkPoint *model.CheckPoint,
) errors.OrdaError {
	filter := schema.GetFilter().AddSetCheckPoint(duid, checkPoint)
	result, err := its.clients.UpdateOne(ctx, schema.FilterByID(cuid), bson.D(filter), schema.UpsertOption)
	if err != nil {
		return errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	if result.ModifiedCount == 1 {
		return nil
	}
	ctx.L().Warnf("updated no checkpoint of %s in client %s", duid, cuid)
	return nil
}

// UnsubscribeDatatypeFromClient makes the specified client unsubscribe the specified datatype.
func (its *MongoCollections) UnsubscribeDatatypeFromClient(
	ctx context.OrdaContext,
	cuid string,
	duid string,
) errors.OrdaError {
	filter := schema.GetFilter().AddUnsetCheckPoint(duid)
	result, err := its.clients.UpdateOne(ctx, schema.FilterByID(cuid), bson.D(filter))
	if err != nil {
		return errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	if result.ModifiedCount == 1 {
		return nil
	}
	ctx.L().Warnf("unsubscribe no client for datatype `%s`", duid)
	return nil
}

// UnsubscribeDatatypeFromAllClients makes the specified datatype unsubscribed from all the clients.
func (its *MongoCollections) UnsubscribeDatatypeFromAllClients(
	ctx context.OrdaContext,
	duid string,
) errors.OrdaError {
	findFilter := schema.GetFilter().AddExists(fmt.Sprintf("%s.%s", schema.ClientDocFields.CheckPoints, duid))
	updateFilter := schema.GetFilter().AddUnsetCheckPoint(duid)
	result, err := its.clients.UpdateMany(ctx, findFilter, bson.D(updateFilter))
	if err != nil {
		return errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	if result.ModifiedCount > 0 {
		ctx.L().Infof("unsubscribed datatype `%s` form %d clients", duid, result.ModifiedCount)
		return nil
	}
	ctx.L().Warnf("unsubscribe no client for datatype `%s`", duid)
	return nil
}

// DeleteClient deletes the specified client.
func (its *MongoCollections) DeleteClient(ctx context.OrdaContext, cuid string) errors.OrdaError {
	result, err := its.clients.DeleteOne(ctx, schema.FilterByID(cuid))
	if err != nil {
		return errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	if result.DeletedCount == 1 {
		return nil
	}
	ctx.L().Warnf("fail to find a client to delete: `%s`", cuid)
	return nil
}

// GetCheckPointFromClient returns a checkpoint for the specified datatype from the specified client.
func (its *MongoCollections) GetCheckPointFromClient(
	ctx context.OrdaContext,
	cuid string,
	duid string,
) (*model.CheckPoint, errors.OrdaError) {
	opts := options.FindOne()
	projectField := fmt.Sprintf("checkpoints.%s", duid)
	opts.SetProjection(bson.M{projectField: 1})
	sr := its.clients.FindOne(ctx, schema.FilterByID(cuid), opts)
	if err := sr.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	var client schema.ClientDoc
	if err := sr.Decode(&client); err != nil {
		return nil, errors.ServerDBDecode.New(ctx.L(), err.Error())
	}
	checkPoint, ok := client.CheckPoints[duid]
	if !ok {
		return nil, nil
	}
	return checkPoint, nil
}

// GetClientWithoutCheckPoints returns a ClientDoc without CheckPoints for the specified CUID
func (its *MongoCollections) GetClientWithoutCheckPoints(
	ctx context.OrdaContext,
	cuid string,
) (*schema.ClientDoc, errors.OrdaError) {
	return its.getClient(ctx, cuid, false)
}

// GetClient returns a ClientDoc for the specified CUID.
func (its *MongoCollections) GetClient(
	ctx context.OrdaContext,
	cuid string,
) (*schema.ClientDoc, errors.OrdaError) {
	return its.getClient(ctx, cuid, true)
}

func (its *MongoCollections) getClient(
	ctx context.OrdaContext,
	cuid string,
	withCheckPoint bool,
) (*schema.ClientDoc, errors.OrdaError) {
	opts := options.FindOne()
	if !withCheckPoint {
		opts.SetProjection(bson.M{schema.ClientDocFields.CheckPoints: 0})
	}
	sr := its.clients.FindOne(ctx, schema.FilterByID(cuid), opts)
	if err := sr.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, errors.ServerDBQuery.New(ctx.L(), err.Error())
	}

	var client schema.ClientDoc
	if err := sr.Decode(&client); err != nil {
		return nil, errors.ServerDBDecode.New(ctx.L(), err.Error())
	}
	return &client, nil
}

func (its *MongoCollections) purgeAllCollectionClients(
	ctx context.OrdaContext,
	collectionNum uint32,
) errors.OrdaError {
	filter := schema.GetFilter().AddFilterEQ(schema.ClientDocFields.CollectionNum, collectionNum)
	r1, err := its.clients.DeleteMany(ctx, filter)
	if err != nil {
		return errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	if r1.DeletedCount > 0 {
		ctx.L().Infof("delete %d clients in collection %d", r1.DeletedCount, collectionNum)
		return nil
	}
	ctx.L().Warnf("delete no clients")
	return nil
}

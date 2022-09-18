package mongodb

import (
	"github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/iface"
	"github.com/orda-io/orda/server/schema"

	"go.mongodb.org/mongo-driver/mongo"
)

// UpdateClient updates a clientDoc; if not exists, a new clientDoc is inserted.
func (its *MongoCollections) UpdateClient(
	ctx iface.OrdaContext,
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

// DeleteClient deletes the specified client.
func (its *MongoCollections) DeleteClient(ctx iface.OrdaContext, cuid string) errors.OrdaError {
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

// GetClient returns a ClientDoc for the specified CUID.
func (its *MongoCollections) GetClient(
	ctx iface.OrdaContext,
	cuid string,
) (*schema.ClientDoc, errors.OrdaError) {
	// opts := options.FindOne()
	// if !withCheckPoint {
	// 	opts.SetProjection(bson.M{schema.ClientDocFields.CheckPoints: 0})
	// }
	sr := its.clients.FindOne(ctx, schema.FilterByID(cuid))
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
	ctx iface.OrdaContext,
	collectionNum int32,
) errors.OrdaError {
	filter := schema.GetFilter().AddFilterEQ(schema.ClientDocFields.CollectionNum, collectionNum)
	r1, err := its.clients.DeleteMany(ctx, filter)
	if err != nil {
		return errors.ServerDBQuery.New(ctx.L(), err.Error())
	}
	ctx.L().Infof("delete %d clients in collection#%d", r1.DeletedCount, collectionNum)
	return nil
}

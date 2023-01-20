package pkg

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BaseModel interface {
	ToBson() interface{}
}

type InsertResult[Model BaseModel] struct {
	ID   primitive.ObjectID
	Base *Model
}

type MongoRepository[Model BaseModel] struct {
	client     *mongo.Client
	database   string
	collection string
}

func NewMongoRepository[Model BaseModel](client *mongo.Client, database, collection string) *MongoRepository[Model] {
	return &MongoRepository[Model]{
		client:     client,
		database:   database,
		collection: collection,
	}
}

func (r *MongoRepository[Model]) FindById(ctx context.Context, id string, opts ...*options.FindOneOptions) (*Model, error) {
	var result Model
	err := r.client.Database(r.database).Collection(r.collection).FindOne(ctx, bson.M{"_id": id}, opts...).Decode(&result)

	return &result, err
}

func (r *MongoRepository[Model]) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (*Model, error) {
	var result Model
	err := r.client.Database(r.database).Collection(r.collection).FindOne(ctx, filter, opts...).Decode(&result)

	return &result, err
}

func (r *MongoRepository[Model]) FindMany(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*[]Model, error) {
	var results []Model
	cursor, err := r.client.Database(r.database).Collection(r.collection).Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return &results, nil
}

func (r *MongoRepository[Model]) InsertOne(ctx context.Context, model *Model, opts ...*options.InsertOneOptions) (*InsertResult[Model], error) {
	result, err := r.client.Database(r.database).Collection(r.collection).InsertOne(ctx, model, opts...)
	if err != nil {
		return nil, err
	}

	return &InsertResult[Model]{
		Base: model,
		ID:   result.InsertedID.(primitive.ObjectID),
	}, nil
}

func (r *MongoRepository[Model]) InsertMany(ctx context.Context, models []*Model, opts ...*options.InsertManyOptions) ([]*InsertResult[Model], error) {
	var documents []interface{}

	for _, model := range models {
		documents = append(documents, (*model).ToBson())
	}

	result, err := r.client.Database(r.database).Collection(r.collection).InsertMany(ctx, documents, opts...)
	if err != nil {
		return nil, err
	}

	listIds := result.InsertedIDs

	var savedModels []*InsertResult[Model]

	for idx, id := range listIds {
		savedModels = append(savedModels, &InsertResult[Model]{
			Base: models[idx],
			ID:   id.(primitive.ObjectID),
		})
	}

	return savedModels, nil
}

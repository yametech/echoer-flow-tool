package mongo

import (
	"context"
	"github.com/yametech/echoer-flow-tool/pkg/core"
	"github.com/yametech/echoer-flow-tool/pkg/store"
	"github.com/yametech/echoer-flow-tool/pkg/store/gtm"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	metadata       = "metadata"
	version        = "version"
	metadataName   = "metadata.name"
	metadataUUID   = "metadata.uuid"
	metadataDelete = "metadata.is_delete"
)

var _ store.IStore = &Mongo{}

func getCtx(client *mongo.Client) (context.Context, context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := client.Connect(ctx); err != nil {
		return nil, nil, err
	}
	return ctx, cancel, nil
}

type Mongo struct {
	uri    string
	client *mongo.Client
}

func NewMongo(uri string) (*Mongo, error, chan error) {
	client, err := connect(uri)
	if err != nil {
		return nil, err, nil
	}

	investigationErrorChannel := make(chan error)
	go func() {
		for {
			time.Sleep(1 * time.Second)
			if err := client.Ping(context.Background(), readpref.Primary()); err != nil {
				investigationErrorChannel <- err
			}
		}
	}()

	return &Mongo{uri: uri, client: client}, nil, investigationErrorChannel
}

func connect(uri string) (*mongo.Client, error) {
	clientOptions := options.Client()
	clientOptions.SetRegistry(
		bson.NewRegistryBuilder().
			RegisterTypeMapEntry(
				bsontype.DateTime,
				reflect.TypeOf(time.Time{})).
			Build(),
	)
	clientOptions.ApplyURI(uri)
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		return nil, err
	}
	ctx, cancel, err := getCtx(client)
	defer func() { cancel() }()
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}
	return client, nil
}

func (m *Mongo) Close() error {
	ctx, cancel, err := getCtx(m.client)
	if err != nil {
		return err
	}
	defer func() { cancel() }()
	return m.client.Disconnect(ctx)
}

func (m *Mongo) List(namespace, resource, labels string, sort map[string]interface{}, skip, limit int64) ([]interface{}, int64, error) {
	ctx := context.Background()
	var filter = bson.D{{}}
	if len(labels) > 0 {
		filter = expr2labels(labels)
	}
	findOptions := options.Find()
	findOptions.SetSkip(skip).SetLimit(limit).SetSort(map2filter(sort))
	cursor, err := m.client.
		Database(namespace).
		Collection(resource).
		Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	count, err := m.client.Database(namespace).Collection(resource).CountDocuments(ctx, filter, options.Count())
	if err != nil {
		return nil, 0, err
	}
	var _results []bson.M
	if err := cursor.All(ctx, &_results); err != nil {
		return nil, 0, err
	}
	results := make([]interface{}, 0)
	for index := range _results {
		results = append(results, _results[index])
	}
	return results, count, nil
}

func (m *Mongo) ListByFilter(namespace, resource string, filter, sort map[string]interface{}, skip, limit int64) ([]interface{}, int64, error) {
	ctx := context.Background()
	findOptions := options.Find()
	findOptions.SetSkip(skip).SetLimit(limit).SetSort(sort)
	cursor, err := m.client.
		Database(namespace).
		Collection(resource).
		Find(ctx, map2filter(filter), findOptions)
	if err != nil {
		return nil, 0, err
	}
	count, err := m.client.Database(namespace).Collection(resource).CountDocuments(ctx, map2filter(filter), options.Count())
	if err != nil {
		return nil, 0, err
	}
	var _results []bson.M
	if err := cursor.All(ctx, &_results); err != nil {
		return nil, 0, err
	}
	results := make([]interface{}, 0)
	for index := range _results {
		results = append(results, _results[index])
	}
	return results, count, nil
}

func (m *Mongo) GetByFilter(namespace, resource string, result interface{}, filter map[string]interface{}) error {
	ctx := context.Background()
	findOneOptions := options.FindOne()
	singleResult := m.client.
		Database(namespace).
		Collection(resource).
		FindOne(ctx, map2filter(filter), findOneOptions)
	if err := singleResult.Decode(result); err != nil {
		if err == mongo.ErrNoDocuments {
			return store.NotFound
		}
		return err
	}
	return nil
}

func (m *Mongo) GetByUUID(namespace, resource, uuid string, result interface{}) error {
	query := bson.M{metadataUUID: uuid}
	ctx := context.Background()
	findOneOptions := options.FindOne()
	singleResult := m.client.Database(namespace).Collection(resource).FindOne(ctx, query, findOneOptions)
	if err := singleResult.Decode(result); err != nil {
		if err == mongo.ErrNoDocuments {
			return store.NotFound
		}
		return err
	}
	return nil
}

func versionMatchFilter(op *gtm.Op, resourceVersion int64) bool {
	metadata, exist := op.Data[metadata]
	if !exist {
		return false
	}
	metadataMap := metadata.(map[string]interface{})
	version, exist := metadataMap[version]
	if !exist {
		return false
	}
	if version.(int64) <= resourceVersion {
		return false
	}
	return true
}

func (m *Mongo) Create(namespace, resource string, object core.IObject) (core.IObject, error) {
	ctx := context.Background()
	object.GenerateVersion()
	_, err := m.client.Database(namespace).Collection(resource).InsertOne(ctx, object)
	if err != nil {
		return nil, err
	}
	return object, nil
}

func (m *Mongo) Apply(namespace, resource, uuid string, object core.IObject) (core.IObject, bool, error) {
	var query = bson.M{metadataUUID: uuid}

	ctx := context.Background()
	singleResult := m.client.Database(namespace).Collection(resource).FindOne(ctx, query)

	if singleResult.Err() == mongo.ErrNoDocuments {
		object.GenerateVersion()
		_, err := m.client.Database(namespace).Collection(resource).InsertOne(ctx, object)
		if err != nil {
			return nil, false, err
		}
		return object, false, nil
	}

	old := object.Clone()
	if err := singleResult.Decode(old); err != nil {
		return nil, false, err
	}

	oldMap, err := core.ToMap(old)
	if err != nil {
		return nil, false, err
	}
	objectMap, err := core.ToMap(object)
	if err != nil {
		return nil, false, err
	}

	if reflect.DeepEqual(oldMap["spec"], objectMap["spec"]) {
		return old, false, nil
	}

	oldMap["spec"] = objectMap["spec"]

	if err := core.EncodeFromMap(old, oldMap); err != nil {
		return old, false, err
	}

	upsert := true
	old.GenerateVersion() //update version
	_, err = m.client.
		Database(namespace).
		Collection(resource).
		ReplaceOne(ctx, query, old,
			options.MergeReplaceOptions(
				&options.ReplaceOptions{Upsert: &upsert},
			),
		)
	if err != nil {
		return nil, true, err
	}

	return old, true, nil
}

func (m *Mongo) Delete(namespace, resource, uuid string) error {
	query := bson.M{metadataUUID: uuid}
	ctx := context.Background()
	_, err := m.client.Database(namespace).Collection(resource).DeleteOne(ctx, query)
	if err != nil {
		return err
	}
	return nil
}

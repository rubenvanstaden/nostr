package mongodb

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"noztr/core"
)

type repository struct {
	collection *mongo.Collection

	core.Repository
}

func New(url, database, collection string) core.Repository {

	ctx := context.Background()

	// Set client options
	clientOptions := options.Client().ApplyURI(url)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	return &repository{
		collection: client.Database(database).Collection(collection),
	}
}

func (s *repository) Store(ctx context.Context, event *core.Event) error {

	const op = "mongodb.Store"

	// Insert a single document
	res, err := s.collection.InsertOne(ctx, event)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("[%s] event stored : {event_id: %s, db_id: %s}", op, event.Id, res.InsertedID)

	return nil
}

func (s *repository) FindByAuthors(ctx context.Context, authors []string) ([]core.Event, error) {

	const op = "mongodb.FindByAuthors"

	var filters []bson.M
	for _, pub := range authors {
		filter := bson.M{"pubkey": bson.M{"$regex": "^" + pub}}
		filters = append(filters, filter)
	}

	cur, err := s.collection.Find(ctx, bson.M{"$or": filters})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var results []core.Event
	for cur.Next(ctx) {
		var result core.Event
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, result)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (s *repository) FindByIdPrefix(ctx context.Context, prefixes []string) ([]core.Event, error) {

	const op = "mongodb.FindByIdPrefix"

	var filters []bson.M
	for _, prefix := range prefixes {
		filter := bson.M{"id": bson.M{"$regex": "^" + prefix}}
		filters = append(filters, filter)
	}

	cur, err := s.collection.Find(ctx, bson.M{"$or": filters})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var results []core.Event
	for cur.Next(ctx) {
		var result core.Event
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, result)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (s *repository) Find(ctx context.Context, id core.EventId) (*core.Event, error) {

	const op = "mongodb.FindById"

	filter := bson.M{"_id": bson.M{"$eq": id}}
	result := s.collection.FindOne(ctx, filter)

	var event core.Event

	err := result.Decode(&event)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found a single document: %+v\n", result)

	return &event, nil
}

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

func (s *repository) Store(ctx context.Context, e *core.Event) error {

	op := "mongodb.Store"

	// Some dummy data to add to the Database
	trainer := bson.D{
		{Key: "name", Value: "Ash"},
		{Key: "age", Value: 10},
		{Key: "city", Value: "Pallet Town"},
	}

	// Insert a single document
	res, err := s.collection.InsertOne(context.TODO(), trainer)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("[%s] event stored : {event_id: %s, db_id: %s}", op, e.Id, res.InsertedID)

	return nil
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

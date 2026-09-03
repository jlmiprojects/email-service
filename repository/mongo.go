package repository

import (
	"context"
	"log/slog"
	"time"

	"blueassetgroup.com/email-service/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoRepository struct {
	client *mongo.Client
	Config *models.Config
}

func NewMongo(config *models.Config) (*MongoRepository, error) {

	instance := new(MongoRepository)

	instance.Config = config

	var err error

	bsonOpts := &options.BSONOptions{
		UseJSONStructTags: true,
		NilSliceAsEmpty:   true,
	}

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(config.Mongo.Uri).SetServerAPIOptions(serverAPI).SetBSONOptions(bsonOpts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	instance.client, err = mongo.Connect(opts)
	if err != nil {
		return nil, err
	}

	err = instance.client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	slog.Info("Sucessfully PING mongo db")
	slog.Info("Connected to Mongo", "uri", config.Mongo.Uri)

	return instance, nil

}

func (r MongoRepository) Save(email *models.Email) error {

	var err error

	ctx, cancel := context.WithTimeout(context.Background(), r.Config.Mongo.ReadTimeout)
	defer cancel()

	db := r.client.Database(r.Config.Mongo.Database)

	coll := db.Collection(r.Config.Mongo.EmailCollection)

	email.Timestamp = time.Now()
	if email.ID.IsZero() {
		email.ID = primitive.NewObjectID()
	}

	_, err = coll.InsertOne(ctx, email)

	if err != nil {
		return err
	}

	return nil
}

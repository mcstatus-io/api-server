package main

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	CollectionUsers    string = "users"
	CollectionSessions string = "sessions"
)

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

type User struct {
	ID        string    `bson:"_id" json:"id"`
	Email     string    `bson:"email" json:"email"`
	Password  string    `bson:"password" json:"-"`
	Type      string    `bson:"type" json:"type"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}

type Session struct {
	ID        string    `bson:"_id" json:"id"`
	User      string    `bson:"user" json:"user"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}

func (c *MongoDB) Connect(uri string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	parsedURI, err := url.Parse(uri)

	if err != nil {
		return err
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))

	if err != nil {
		return err
	}

	c.Client = client
	c.Database = client.Database(strings.TrimPrefix(parsedURI.Path, "/"))

	return nil
}

func (c *MongoDB) InsertUser(document User) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	_, err := c.Database.Collection(CollectionUsers).InsertOne(ctx, document)

	return err
}

func (c *MongoDB) InsertSession(document Session) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	_, err := c.Database.Collection(CollectionSessions).InsertOne(ctx, document)

	return err
}

func (c *MongoDB) GetUserByEmail(email string) (*User, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	cur := c.Database.Collection(CollectionUsers).FindOne(ctx, bson.M{"email": email})

	if err := cur.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		return nil, err
	}

	var result User

	if err := cur.Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *MongoDB) GetUserByID(id string) (*User, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	cur := c.Database.Collection(CollectionUsers).FindOne(ctx, bson.M{"_id": id})

	if err := cur.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		return nil, err
	}

	var result User

	if err := cur.Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *MongoDB) GetSessionByID(id string) (*Session, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	cur := c.Database.Collection(CollectionSessions).FindOne(ctx, bson.M{"_id": id})

	if err := cur.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		return nil, err
	}

	var result Session

	if err := cur.Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *MongoDB) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	return c.Client.Disconnect(ctx)
}

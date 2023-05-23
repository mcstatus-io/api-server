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

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

type User struct {
	ID        string    `bson:"_id" json:"id"`
	Username  string    `bson:"username" json:"username"`
	Avatar    *string   `bson:"avatar" json:"avatar"`
	Email     string    `bson:"email" json:"-"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}

type Session struct {
	ID        string    `bson:"_id" json:"id"`
	User      string    `bson:"user" json:"user"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}

func (d *MongoDB) Connect(uri string) error {
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

	d.Client = client
	d.Database = client.Database(strings.TrimPrefix(parsedURI.Path, "/"))

	return nil
}

func (d *MongoDB) UpsertUser(filter, update bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	_, err := m.Database.Collection("users").UpdateOne(ctx, filter, update, &options.UpdateOptions{
		Upsert: PointerOf(true),
	})

	return err
}

func (d *MongoDB) InsertSession(document Session) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	_, err := m.Database.Collection("sessions").InsertOne(ctx, document)

	return err
}

func (d *MongoDB) GetSessionByID(id string) (*Session, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	cur := d.Database.Collection("sessions").FindOne(ctx, bson.M{"_id": id})

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

func (d *MongoDB) GetUserByID(id string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	cur := d.Database.Collection("users").FindOne(ctx, bson.M{"_id": id})

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

func (d *MongoDB) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	return d.Client.Disconnect(ctx)
}

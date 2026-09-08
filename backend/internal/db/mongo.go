// Package db owns the MongoDB connection and collection handles used by the
// rest of the backend. Handlers and the WebSocket hub talk to Mongo only
// through this package.
package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DB holds a connected Mongo client plus the collection handles the app
// uses. It is passed by reference into handlers so there is exactly one
// connection pool for the process's lifetime.
type DB struct {
	Client   *mongo.Client
	Database *mongo.Database

	Sessions *mongo.Collection
	Items    *mongo.Collection
	Scores   *mongo.Collection
	Comments *mongo.Collection

	// Users and AuthSessions back internal/auth (login accounts and
	// database-persisted login sessions, respectively). Named distinctly
	// from Sessions, which is the review-session domain collection.
	Users        *mongo.Collection
	AuthSessions *mongo.Collection
}

// Connect dials MongoDB at uri, selects dbName, verifies connectivity with a
// ping, and ensures the indexes the app relies on exist. uri works
// identically whether it points at a local `docker-compose` MongoDB or a
// MongoDB Atlas cluster — only the connection string changes between
// environments, which is what makes moving this to Azure later a
// configuration change rather than a code change.
func Connect(uri, dbName string) (*DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	database := client.Database(dbName)

	d := &DB{
		Client:       client,
		Database:     database,
		Sessions:     database.Collection("sessions"),
		Items:        database.Collection("items"),
		Scores:       database.Collection("scores"),
		Comments:     database.Collection("comments"),
		Users:        database.Collection("users"),
		AuthSessions: database.Collection("auth_sessions"),
	}

	if err := d.ensureIndexes(ctx); err != nil {
		return nil, err
	}

	return d, nil
}

// ensureIndexes creates the indexes the query patterns in handlers/ rely on.
// Creating an index that already exists is a no-op, so this is safe to run
// on every startup.
func (d *DB) ensureIndexes(ctx context.Context) error {
	if _, err := d.Items.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "session_id", Value: 1}, {Key: "order_index", Value: 1}},
	}); err != nil {
		return err
	}

	if _, err := d.Scores.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "item_id", Value: 1}, {Key: "reviewer_email", Value: 1}},
	}); err != nil {
		return err
	}

	if _, err := d.Comments.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "item_id", Value: 1}, {Key: "created_at", Value: 1}},
	}); err != nil {
		return err
	}

	if _, err := d.Users.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return err
	}

	// A TTL index: MongoDB deletes a session document on its own once
	// expires_at is in the past, so logins don't need a separate cleanup
	// job.
	if _, err := d.AuthSessions.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "expires_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	}); err != nil {
		return err
	}

	return nil
}

// Disconnect closes the underlying Mongo client. Call it once on shutdown.
func (d *DB) Disconnect(ctx context.Context) error {
	return d.Client.Disconnect(ctx)
}

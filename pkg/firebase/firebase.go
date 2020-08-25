package firebase

import (
	"context"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/db"
	"google.golang.org/api/option"
	"log"
)

// DBClient client for accesing to firebase realtime database
type DBClient interface {
	Fetch(ipaddr string) ([]Node, error)
	Push(v interface{}) error
	Delete() error
}

// Node interface corresponding to QueryNode
type Node interface {
	Unmarshal(v interface{}) error
}

// Reference wrapper of db.Ref
type Reference struct {
	db *db.Ref
}

// NewReference initialize func
func NewReference(db *db.Ref) *Reference {
	ref := Reference{
		db: db,
	}
	return &ref
}

// Fetch get data with a condition
func (r *Reference) Fetch(ipaddr string) ([]Node, error) {
	results, _ := r.db.OrderByChild("ipaddr").EqualTo(ipaddr).GetOrdered(context.Background())
	var node []Node
	for _, r := range results {
		node = append(node, r)
	}
	return node, nil
}

// Push save data
func (r *Reference) Push(v interface{}) error {
	_, err := r.db.Push(context.Background(), v)
	if err != nil {
		return err
	}

	return nil
}

// Delete ALL data
func (r *Reference) Delete() error {
	err := r.db.Delete(context.Background())
	if err != nil {
		return err
	}

	return nil
}

// InitFirebaseRef return Reference including db.Ref
func InitFirebaseRef(referenceName string, url string, credentialFilePath string) DBClient {
	opt := option.WithCredentialsFile(credentialFilePath)
	config := &firebase.Config{DatabaseURL: url}
	app, err := firebase.NewApp(context.Background(), config, opt)
	if err != nil {
		log.Fatalln("Error initializing database client:", err)
	}
	client, err := app.Database(context.Background())
	if err != nil {
		log.Fatalln("Error initializing database client:", err)
	}
	ref := NewReference(client.NewRef(referenceName))
	return ref
}

package firebase

import (
    "firebase.google.com/go/db"
    "context"
    "log"
//	"encoding/json"
    firebase "firebase.google.com/go"
    "google.golang.org/api/option"

)

type DBClient interface {
    Fetch(ipaddr string) ([]Node, error)
}

type Node interface {
    Unmarshal(v interface{}) (error)
}

type Reference struct {
	db      *db.Ref
}

func NewReference (db *db.Ref) *Reference {
    ref := Reference {
        db: db,
    }
    return &ref
}

func (r *Reference) Fetch(ipaddr string) ([]Node, error) {
    results, _ := r.db.OrderByChild("ipaddr").EqualTo(ipaddr).GetOrdered(context.Background())
    var node []Node
    for _, r := range results {
        node = append(node,  r)
    }
    return node, nil
}

func InitFirebaseRef(reference_name string, url string, credential_file_path string) (DBClient){
    opt := option.WithCredentialsFile(credential_file_path)
    config := &firebase.Config{DatabaseURL: url}
    app, err := firebase.NewApp(context.Background(), config, opt)
    if err != nil {
        log.Fatalln("Error initializing database client:", err)
    }
    client, err := app.Database(context.Background())
    if err != nil {
        log.Fatalln("Error initializing database client:", err)
    }
    ref := NewReference(client.NewRef(reference_name))
    return ref
}


/*func (n *Node) Unmarshal(v interface{}) (error) {
	b, err := json.Marshal(n.Value)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
    return nil
}*/

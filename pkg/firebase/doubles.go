package firebase

import (
	"encoding/json"
)

// FakeDBClient test client 
type FakeDBClient struct {
	FakeData User
}

// User test user
type User struct {
	IPAddr string
	Time   string
	Amount int
}

// NodeImp test struct corresponding to QueryNode
type NodeImp struct {
	Value interface{}
}

// Push Client method
func (f FakeDBClient) Push(v interface{}) error {
	return nil
}

// Delete Client method
func (f FakeDBClient) Delete() error {
	return nil
}

// Fetch Client method, return value stored in FakeDBClient
func (f FakeDBClient) Fetch(ipaddr string) ([]Node, error) {

	var nodes []Node
	nodes = append(nodes, NodeImp{
		Value: map[string]interface{}{
			"IPAddr": f.FakeData.IPAddr,
			"Time":   f.FakeData.Time,
			"Amount": f.FakeData.Amount,
		},
	})
	return nodes, nil
}

// Unmarshal parse interface
func (n NodeImp) Unmarshal(v interface{}) error {
	b, err := json.Marshal(n.Value)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

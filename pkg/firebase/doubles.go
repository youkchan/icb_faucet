package firebase

import(
	"encoding/json"
)

type FakeDBClient struct {
    FakeData User
}

type User struct {
    IPAddr string
    Time string
    Amount int
}

type NodeImp struct {
    Value interface{}
}

func (r FakeDBClient) Push(v interface{}) (error) {
    return nil
}

func (r FakeDBClient) Delete() (error) {
    return nil
}

func (f FakeDBClient) Fetch(ipaddr string) ([]Node, error) {

    var nodes []Node
    nodes = append(nodes, NodeImp{
        Value: map[string]interface{}{
                "IPAddr": f.FakeData.IPAddr,
                "Time": f.FakeData.Time,
                "Amount": f.FakeData.Amount,
            },
        })
     return nodes ,nil
}

func (n NodeImp) Unmarshal(v interface{}) (error) {
	b, err := json.Marshal(n.Value)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
    return nil
}

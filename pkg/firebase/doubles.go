package firebase 

import(
	"encoding/json"
)

type FakeDBClient struct {
}

type NodeImp struct {
    Value interface{}
}

func (f *FakeDBClient) Fetch(ipaddr string) ([]Node, error) {

    var nodes []Node
    nodes = append(nodes, NodeImp{
        Value: map[string]interface{}{
                "IPAddr": "2001:268:c145:92df:80dc:a420:d569:869c",
                "Time": "2020-08-23 18:46:24 +0900 JST",
                "Amount": 50,
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

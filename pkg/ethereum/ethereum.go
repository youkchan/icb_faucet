package ethereum 

import (
    /*"fmt"
    "log"
    "net/http"
    "os"
    "html/template"
    "crypto/ecdsa"
    "github.com/joho/godotenv"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    token "github.com/youkchan/icb_faucet/pkg/token"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/common/hexutil"
    "github.com/ethereum/go-ethereum/core/types"
    "golang.org/x/crypto/sha3"*/
    "regexp"
    "math/big"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    token_library "github.com/youkchan/icb_faucet/pkg/token"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "context"
    "errors"
    "log"
    "fmt"
)

type ClientFactory struct {
    Network_list []Network
}

type Client struct {
    Network Network
    Ethclient *ethclient.Client
}

type Token struct {
    ContractAddress string
}

type Network struct{
    Id int
    Endpoint string
}

func (n* Network) getName() (string, error) {
    if(n.Id == 3) {
        return "ropsten" , nil
    } else if(n.Id == 4) {
        return "rinkeby" , nil
    } else {
        log.Fatal("invalid id")
        return "", nil
    }
}

func (n* Network) validate() error {
    if(n.Id != 3 && n.Id != 4) {
        return errors.New("Network ID must be 3 or 4")
    }
    return nil
}

func NewClientFactory(network_list []Network) (*ClientFactory) {
    client_factory := ClientFactory{
        Network_list: network_list,
    }

    return &client_factory
}

func NewToken(address string) (*Token) {
    token := Token{
        ContractAddress: address,
    }

    return &token
}

func NewNetwork(network_id int, endpoint string) (*Network) {
    network := Network{
        Id: network_id,
        Endpoint: endpoint,
    }

	err := network.validate()
	if err != nil {
		log.Fatal(err)
	}
    return &network
}

func (c* ClientFactory) CreateClient(network_id int) (*Client, error) {
    var network Network
    for _, v := range c.Network_list {
        if v.Id == network_id {
            network = v
        }
    }

    if network.Id == 0 {
        return nil, errors.New("Network ID must be 3 or 4")
    }

    ethereum_client, err := ethclient.Dial(network.Endpoint)
    if err != nil {
        log.Fatal(err)
    }


    client := Client{
        Network: network,
        Ethclient: ethereum_client,
    }

    return &client, nil
}

func (c* Client) SendToken(address string, amount int, network_id int) string {
    fmt.Println(address)
    fmt.Println(amount)
    fmt.Println(network_id)

	/*if err := network.validate(); err != nil {
		log.Fatal(err)
	}

    network_name, err := network.getName();
    if err != nil {
		log.Fatal(err)
    }*/

    return address
}

/*func (c* Client) getTokenBalance(token_instance Token, address string)  (*big.Int, error) {

    return address
}*/

func (c* Client) GetEtherBalance(address string)  (*big.Int, error) {
    if !addressCheck(address) {
        return nil, errors.New("Invalid Address")
    }
    account := common.HexToAddress(address)
    return c.Ethclient.BalanceAt(context.Background(), account, nil)
}

func (c* Client) GetTokenBalance(token Token, address string)  (*big.Int, error) {
    if !addressCheck(address) {
        return nil, errors.New("Invalid Address")
    }

    instance, err := token_library.NewToken(common.HexToAddress(token.ContractAddress), c.Ethclient)
    if err != nil {
      log.Fatal(err)
    }
    account := common.HexToAddress(address)
    return instance.BalanceOf(&bind.CallOpts{}, account)
}

func addressCheck(address string) bool {
    re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
    return re.MatchString(address)
}

package ethereum

import (
    "testing"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"
    "reflect"
    "encoding/hex"
    "github.com/joho/godotenv"
    "log"
    "os"
)

func Env_load() {
    err := godotenv.Load("../../.env")
    if err != nil {
        log.Fatal("Error loading .env file")
    }
}

func TestAddressCheck_Valid(t *testing.T) {
    valid_address := "0x578D9B2d04bc99007B941787E88E4ea57D888A56"
    if !addressCheck(valid_address) {
        t.Error("invalid address check")
    }
}

func TestAddressCheck_Invalid(t *testing.T) {
    valid_address := "0x578D9B2d04bc99007B941787E88E4ea57D888Z56" //including an invalid character "Z"
    if addressCheck(valid_address) {
        t.Error("invalid address check")
    }
}


func TestNewAccount(t *testing.T) {
    valid_address := "0x578D9B2d04bc99007B941787E88E4ea57D888A56"
    account,err := NewAccount(valid_address)
    if err != nil {
        t.Error("invalid initialize")
    }

    if account.Address.Hex() != valid_address {
        t.Error("invalid initialize")
    }
}


func TestNewSendableAccount(t *testing.T) {

    test_priv_key , _ := crypto.GenerateKey()
    test_address := crypto.PubkeyToAddress(test_priv_key.PublicKey)
    test_priv_key_string := hex.EncodeToString(crypto.FromECDSA(test_priv_key))

    sendable_account,err := NewSendableAccount(test_priv_key_string)
    if err != nil {
        t.Error("invalid initialize")
    }

    if !reflect.DeepEqual(sendable_account.privateKey, test_priv_key) {
        t.Error("invalid initialize")
    }

    if !reflect.DeepEqual(sendable_account.Address, test_address) {
        t.Error("invalid initialize")
    }

}


func TestNewNetwork(t *testing.T) {
    Env_load()
    network_id := 4
    endpoint := os.Getenv("INFURA_RINKEBY")
    expected_network := Network{
        Id: network_id,
        Endpoint: endpoint,
    }
    network, err := NewNetwork(network_id, endpoint)
    if err != nil {
        t.Error("invalid initialize")
    }

    if !reflect.DeepEqual(expected_network, *network) {
        t.Error("invalid initialize")
    }


}

func TestNewNetwork_Invalid(t *testing.T) {
    Env_load()
    network_id := 1
    endpoint := os.Getenv("INFURA_RINKEBY")
    _, err := NewNetwork(network_id, endpoint)
    if err == nil {
        t.Error("invalid initialize")
    }

}


func TestNewClientFactory(t *testing.T) {
    Env_load()
    network_id := 4
    endpoint := os.Getenv("INFURA_RINKEBY")
    network_list := []Network{Network{
        Id: network_id,
        Endpoint: endpoint,
    }}
    expected_factory := ClientFactory{
        Network_list: network_list,
    }

    factory := NewClientFactory(network_list)

    if !reflect.DeepEqual(expected_factory, *factory) {
        t.Error("invalid initialize")
    }
}


func TestCreateClient(t *testing.T) {
    Env_load()
    network_id := 4
    endpoint := os.Getenv("INFURA_RINKEBY")
    network := Network{
        Id: network_id,
        Endpoint: endpoint,
    }
    network_list := []Network{network}

    ethereum_client, err := ethclient.Dial(endpoint)
    expected_client := Client{
        Network: network,
        Ethclient: ethereum_client,
    }

    factory := NewClientFactory(network_list)
    client,err := factory.CreateClient(4)
    if err != nil {
        t.Error("invalid initialize")
    }

    if !reflect.DeepEqual(expected_client.Network, client.Network) { //if each Network are equal, pass
        t.Error("invalid initialize")
    }


}


func TestSendTransaction(t *testing.T) {
    Env_load()
    network_id := 4
    endpoint := os.Getenv("INFURA_RINKEBY")
    ethereum_client, _ := ethclient.Dial(endpoint)
    network := Network{
        Id: network_id,
        Endpoint: endpoint,
    }

    client := Client{
        Network: network,
        Ethclient: ethereum_client,
    }

    token := NewToken(os.Getenv("TOKEN_RINKEBY_ADDRESS"))
    faucet_account, _ := NewSendableAccount(os.Getenv("PRIVATE_KEY"))
    toAccount, _ := NewAccount("0xfAc80Ea0c7c515e5A49D9A65564188EB8ed3F259")

    tx := client.CreateSendTokenTransaction(*token, *faucet_account, *toAccount, 100000)
    if reflect.TypeOf(tx).String() != "*types.Transaction" {
        t.Error("invalid transaction")
    }

    if tx.To().Hex() != token.ContractAddress {
        t.Error("invalid transaction")
    }
}

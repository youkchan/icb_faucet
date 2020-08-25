package ethereum

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
	"log"
	"os"
	"reflect"
	"testing"
)

func LoadEnv() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func TestAddressCheck_Valid(t *testing.T) {
	validAddress := "0x578D9B2d04bc99007B941787E88E4ea57D888A56"
	if !addressCheck(validAddress) {
		t.Error("invalid address check")
	}
}

func TestAddressCheck_Invalid(t *testing.T) {
	validAddress := "0x578D9B2d04bc99007B941787E88E4ea57D888Z56" //including an invalid character "Z"
	if addressCheck(validAddress) {
		t.Error("invalid address check")
	}
}

func TestNewAccount(t *testing.T) {
	validAddress := "0x578D9B2d04bc99007B941787E88E4ea57D888A56"
	account, err := NewAccount(validAddress)
	if err != nil {
		t.Error("invalid initialize")
	}

	if account.Address.Hex() != validAddress {
		t.Error("invalid initialize")
	}
}

func TestNewSendableAccount(t *testing.T) {

	testPrivKey, _ := crypto.GenerateKey()
	testAddress := crypto.PubkeyToAddress(testPrivKey.PublicKey)
	testPrivKeyString := hex.EncodeToString(crypto.FromECDSA(testPrivKey))

	sendableAccount, err := NewSendableAccount(testPrivKeyString)
	if err != nil {
		t.Error("invalid initialize")
	}

	if !reflect.DeepEqual(sendableAccount.privateKey, testPrivKey) {
		t.Error("invalid initialize")
	}

	if !reflect.DeepEqual(sendableAccount.Address, testAddress) {
		t.Error("invalid initialize")
	}

}

func TestNewNetwork(t *testing.T) {
	LoadEnv()
	networkID := 4
	endpoint := os.Getenv("INFURA_RINKEBY")
	expectedNetwork := Network{
		ID:       networkID,
		Endpoint: endpoint,
	}
	network, err := NewNetwork(networkID, endpoint)
	if err != nil {
		t.Error("invalid initialize")
	}

	if !reflect.DeepEqual(expectedNetwork, *network) {
		t.Error("invalid initialize")
	}

}

func TestNewNetwork_Invalid(t *testing.T) {
	LoadEnv()
	networkID := 1
	endpoint := os.Getenv("INFURA_RINKEBY")
	_, err := NewNetwork(networkID, endpoint)
	if err == nil {
		t.Error("invalid initialize")
	}

}

func TestNewClientFactory(t *testing.T) {
	LoadEnv()
	networkID := 4
	endpoint := os.Getenv("INFURA_RINKEBY")
	networkList := []Network{Network{
		ID:       networkID,
		Endpoint: endpoint,
	}}
	expectedFactory := ClientFactory{
		NetworkList: networkList,
	}

	factory := NewClientFactory(networkList)

	if !reflect.DeepEqual(expectedFactory, *factory) {
		t.Error("invalid initialize")
	}
}

func TestCreateClient(t *testing.T) {
	LoadEnv()
	networkID := 4
	endpoint := os.Getenv("INFURA_RINKEBY")
	network := Network{
		ID:       networkID,
		Endpoint: endpoint,
	}
	networkList := []Network{network}

	ethereumClient, err := ethclient.Dial(endpoint)
	expectedClient := Client{
		Network:   network,
		Ethclient: ethereumClient,
	}

	factory := NewClientFactory(networkList)
	client, err := factory.CreateClient(4)
	if err != nil {
		t.Error("invalid initialize")
	}

	if !reflect.DeepEqual(expectedClient.Network, client.Network) { //if each Network are equal, pass
		t.Error("invalid initialize")
	}

}

func TestSendTransaction(t *testing.T) {
	LoadEnv()
	networkID := 4
	endpoint := os.Getenv("INFURA_RINKEBY")
	ethereumClient, _ := ethclient.Dial(endpoint)
	network := Network{
		ID:       networkID,
		Endpoint: endpoint,
	}

	client := Client{
		Network:   network,
		Ethclient: ethereumClient,
	}

	token := NewToken(os.Getenv("TOKEN_RINKEBY_ADDRESS"))
	faucetAccount, _ := NewSendableAccount(os.Getenv("PRIVATE_KEY"))
	toAccount, _ := NewAccount("0xfAc80Ea0c7c515e5A49D9A65564188EB8ed3F259")

	tx := client.CreateSendTokenTransaction(*token, *faucetAccount, *toAccount, 100000)
	if reflect.TypeOf(tx).String() != "*types.Transaction" {
		t.Error("invalid transaction")
	}

	if tx.To().Hex() != token.ContractAddress {
		t.Error("invalid transaction")
	}
}

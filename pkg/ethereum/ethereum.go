package ethereum

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	token_library "github.com/youkchan/icb_faucet/pkg/token"
	"golang.org/x/crypto/sha3"
	"log"
	"math/big"
	"regexp"
)

// ClientFactory factory for client struct contains multiple network information
type ClientFactory struct {
	NetworkList []Network
}

// Client client for accessing ethereum
type Client struct {
	Network   Network
	Ethclient *ethclient.Client
}

// Token contains a contract address
type Token struct {
	ContractAddress string
}

// Network contains a network id according to ethereum nerwork id
type Network struct {
	ID       int
	Endpoint string
}

// SendableAccount contains privatekey
type SendableAccount struct {
	privateKey *ecdsa.PrivateKey
	Address    common.Address
}

// NewSendableAccount initialize func
func NewSendableAccount(strPrivatekey string) (*SendableAccount, error) {
	privateKey, err := crypto.HexToECDSA(strPrivatekey)
	if err != nil {
		return nil, errors.New("Invalid Private Key")
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("Invalid Private Key")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	account := SendableAccount{
		privateKey: privateKey,
		Address:    address,
	}
	return &account, nil
}

// Account just contains address not containing privatekey
type Account struct {
	Address common.Address
}

// NewAccount initialize func
func NewAccount(address string) (*Account, error) {
	if !addressCheck(address) {
		return nil, errors.New("Invalid Address")
	}

	commonAddress := common.HexToAddress(address)
	account := Account{
		Address: commonAddress,
	}
	return &account, nil
}

func (n *Network) validate() error {
	if n.ID != 3 && n.ID != 4 {
		return errors.New("Network ID must be 3 or 4")
	}
	return nil
}

// NewClientFactory initialize func
func NewClientFactory(networkList []Network) *ClientFactory {
	clientFactory := ClientFactory{
		NetworkList: networkList,
	}

	return &clientFactory
}

// NewToken initialize func
func NewToken(address string) *Token {
	token := Token{
		ContractAddress: address,
	}

	return &token
}

// NewNetwork initialize func
func NewNetwork(networkID int, endpoint string) (*Network, error) {
	network := Network{
		ID:       networkID,
		Endpoint: endpoint,
	}

	err := network.validate()
	if err != nil {
		return nil, err
	}
	return &network, nil
}

// CreateClient initialize func
func (c *ClientFactory) CreateClient(networkID int) (*Client, error) {
	var network Network
	for _, v := range c.NetworkList {
		if v.ID == networkID {
			network = v
		}
	}

	if network.ID == 0 {
		return nil, errors.New("Network ID must be 3 or 4")
	}

	ethereumClient, err := ethclient.Dial(network.Endpoint)
	if err != nil {
		return nil, errors.New("Invalid Endpoint")
	}

	client := Client{
		Network:   network,
		Ethclient: ethereumClient,
	}

	return &client, nil
}

// SendToken sendTransaction from unsignedTx
func (c *Client) SendToken(token Token, fromAccount SendableAccount, toAccount Account, amount int) string {
	tx := c.CreateSendTokenTransaction(token, fromAccount, toAccount, amount)
	signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, fromAccount.privateKey)
	if err != nil {
		log.Fatal(err)
	}

	err = c.Ethclient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}

	return signedTx.Hash().Hex()
}

// CreateSendTokenTransaction create send tx
func (c *Client) CreateSendTokenTransaction(token Token, fromAccount SendableAccount, toAccount Account, amount int) *types.Transaction {
	nonce, err := c.Ethclient.PendingNonceAt(context.Background(), fromAccount.Address)
	if err != nil {
		log.Fatal(err)
	}

	value := big.NewInt(0)
	gasLimit := uint64(2000000)

	gasPrice, err := c.Ethclient.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	toAddress := toAccount.Address
	tokenAddress := common.HexToAddress(token.ContractAddress)
	transferFnSignature := []byte("transfer(address,uint256)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4]

	paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	pIntAmount := big.NewInt(int64(amount))
	paddedAmount := common.LeftPadBytes(pIntAmount.Bytes(), 32)

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)

	tx := types.NewTransaction(nonce, tokenAddress, value, gasLimit, gasPrice, data)
	return tx
}

// GetEtherBalance return ether balance
func (c *Client) GetEtherBalance(address string) (*big.Int, error) {
	if !addressCheck(address) {
		return nil, errors.New("Invalid Address")
	}
	account := common.HexToAddress(address)
	return c.Ethclient.BalanceAt(context.Background(), account, nil)
}

// GetTokenBalance return token balance
func (c *Client) GetTokenBalance(token Token, address string) (*big.Int, error) {
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

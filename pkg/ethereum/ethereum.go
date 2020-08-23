package ethereum

import (
    "regexp"
    "math/big"
    "crypto/ecdsa"
    "context"
    "errors"
    "log"
//    "reflect"
//    "fmt"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/ethereum/go-ethereum/crypto"
    token_library "github.com/youkchan/icb_faucet/pkg/token"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/core/types"
    "golang.org/x/crypto/sha3"
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

type SendableAccount struct{
    privateKey *ecdsa.PrivateKey
    Address common.Address
}


func NewSendableAccount(str_privatekey string) (*SendableAccount, error) {
    privateKey, err := crypto.HexToECDSA(str_privatekey)
    if err != nil {
        return nil, errors.New("Invalid Private Key")
    }

    publicKey := privateKey.Public()
    publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
    if !ok {
        return nil, errors.New("Invalid Private Key")
    }

    address := crypto.PubkeyToAddress(*publicKeyECDSA)
    account := SendableAccount {
        privateKey: privateKey,
        Address: address,
    }
    return &account , nil
}

type Account struct{
    Address common.Address
}

func NewAccount(address string) (*Account, error) {
    if !addressCheck(address) {
        return nil, errors.New("Invalid Address")
    }

    commonAddress := common.HexToAddress(address)
    account := Account {
        Address: commonAddress,
    }
    return &account, nil
}

/*func (n* Network) getName() (string, error) {
    if(n.Id == 3) {
        return "ropsten" , nil
    } else if(n.Id == 4) {
        return "rinkeby" , nil
    } else {
        log.Fatal("invalid id")
        return "", nil
    }
}*/

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

func NewNetwork(network_id int, endpoint string) (*Network, error) {
    network := Network{
        Id: network_id,
        Endpoint: endpoint,
    }

	err := network.validate()
	if err != nil {
        return nil, err
	}
    return &network, nil
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
        return nil, errors.New("Invalid Endpoint")
    }

    client := Client{
        Network: network,
        Ethclient: ethereum_client,
    }

    return &client, nil
}

func (c* Client) SendToken(token Token, fromAccount SendableAccount, toAccount Account, amount int) string {
    /*nonce, err := c.Ethclient.PendingNonceAt(context.Background(), fromAccount.Address)
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

    tx := types.NewTransaction(nonce, tokenAddress, value, gasLimit, gasPrice, data)*/
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


func (c* Client) CreateSendTokenTransaction(token Token, fromAccount SendableAccount, toAccount Account, amount int)  *types.Transaction {
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

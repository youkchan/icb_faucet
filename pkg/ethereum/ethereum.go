package ethereum 

import (
    /*"fmt"
    "net/http"
    "os"
    "html/template"
    "github.com/joho/godotenv"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    token "github.com/youkchan/icb_faucet/pkg/token"
    "github.com/ethereum/go-ethereum/common/hexutil"
    "github.com/ethereum/go-ethereum/core/types"
    "golang.org/x/crypto/sha3"*/
    "regexp"
    "math/big"
    "crypto/ecdsa"
    "context"
    "errors"
    "log"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/ethereum/go-ethereum/crypto"
    token_library "github.com/youkchan/icb_faucet/pkg/token"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/core/types"
    "golang.org/x/crypto/sha3"
//    "fmt"
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


func NewSendableAccount(str_privatekey string) (*SendableAccount) {
    privateKey, err := crypto.HexToECDSA(str_privatekey)
    if err != nil {
        log.Fatal(err)
    }

    publicKey := privateKey.Public()
    publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
    if !ok {
        log.Fatal("error casting public key to ECDSA")
    }

    address := crypto.PubkeyToAddress(*publicKeyECDSA)
    account := SendableAccount {
        privateKey: privateKey,
        Address: address,
    }
    return &account
}

type Account struct{
    Address common.Address
}

func NewAccount(address string) (*Account) {
    if !addressCheck(address) {
        log.Fatal("Invalid Address")
    }

    commonAddress := common.HexToAddress(address)
    account := Account {
        Address: commonAddress,
    }
    return &account
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

func (c* Client) SendToken(token Token, fromAccount SendableAccount, toAccount Account, amount int) string {
    nonce, err := c.Ethclient.PendingNonceAt(context.Background(), fromAccount.Address)
    if err != nil {
        log.Fatal(err)
    }

    //トークン送金Transactionをテストネット送るためのgasLimit、
    value := big.NewInt(0) //（オプション）後で使用する関数NewTransactionの引数で必要になるため設定。Transactionと同時に送るETHの量を設定できます。
    gasLimit := uint64(2000000)

    //ロプステンネットワークから、現在のgasPriceを取得。トランザクションがマイニングされずに放置されることを防ぐ。
    gasPrice, err := c.Ethclient.SuggestGasPrice(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    //送金先を指定
    toAddress := toAccount.Address
    //トークンコントラクトアドレスを指定
    tokenAddress := common.HexToAddress(token.ContractAddress)
    //ERC20のどの関数を使用するか指定。https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_sendtransaction
    transferFnSignature := []byte("transfer(address,uint256)")
    //hash化し、先頭から4バイトまで取得。これで使用する関数を指定したことになる。
    hash := sha3.NewLegacyKeccak256()
    hash.Write(transferFnSignature)
    methodID := hash.Sum(nil)[:4]

    //0埋め
    paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
    //送金額を設定
    pIntAmount := big.NewInt(int64(amount))
    //0埋め
    paddedAmount := common.LeftPadBytes(pIntAmount.Bytes(), 32)

    //トランザクションで送るデータを作成
    var data []byte
    data = append(data, methodID...)
    data = append(data, paddedAddress...)
    data = append(data, paddedAmount...)

    /***** Preparing signed transaction *****/
    tx := types.NewTransaction(nonce, tokenAddress, value, gasLimit, gasPrice, data)
    //signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, privateKey)
    signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, fromAccount.privateKey)
    if err != nil {
        log.Fatal(err)
    }

    //サインしたトランザクションをRopstenNetworkに送る。
    err = c.Ethclient.SendTransaction(context.Background(), signedTx)
    if err != nil {
        log.Fatal(err)
    }

    return signedTx.Hash().Hex()
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

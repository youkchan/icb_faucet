package main

import (
    "fmt"
    "context"
    "log"
    "net/http"
    "os"
    "html/template"
    "crypto/ecdsa"
    "math/big"
    "github.com/joho/godotenv"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    token "github.com/youkchan/icb_faucet/pkg/token"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/common/hexutil"
    "github.com/ethereum/go-ethereum/core/types"
    "golang.org/x/crypto/sha3"
)

func main() {
    http.HandleFunc("/send", sendHandler)
    http.HandleFunc("/", indexHandler)

    port := os.Getenv("PORT")
    if port == "" {
            port = "8090"
            log.Printf("Defaulting to port %s", port)
    }

    log.Printf("Listening on port %s", port)
    if err := http.ListenAndServe(":"+port, nil); err != nil {
            log.Fatal(err)
    }
}

func Env_load() {
    err := godotenv.Load("../../.env")
    if err != nil {
        log.Fatal("Error loading .env file")
    }
}

func sendHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
    if err != nil {
      log.Fatal(err)
    }

	address := r.PostFormValue("address")
    fmt.Println("send")

    t, _ := template.ParseFiles("../../web/html/index.html")
    t.Execute(w, "send to " + address)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    Env_load()
    if r.URL.Path != "/" {
            http.NotFound(w, r)
            return
    }

    client, err := ethclient.Dial("https://rinkeby.infura.io/v3/deac92b380cd4b219b0c57a59cf363b1")
    if err != nil {
        log.Fatal(err)
    }

    rinkebyICBAddress := common.HexToAddress("0x5446E3481e3fe4b3082067145A47d7a0F09d5E1A")
    rinkebyICBInstance, err := token.NewToken(rinkebyICBAddress, client)
    if err != nil {
      log.Fatal(err)
    }

    account := common.HexToAddress("0x751e0e0de1881f614F40C14c175bdd12d0DCaa24")
    balance, err := client.BalanceAt(context.Background(), account, nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(balance)

    bal, err := rinkebyICBInstance.BalanceOf(&bind.CallOpts{}, account)
    if err != nil {
      log.Fatal(err)
    }

    privateKey, err := crypto.HexToECDSA(os.Getenv("PRIVATE_KEY"))
    if err != nil {
        log.Fatal(err)
    }

    publicKey := privateKey.Public()
    publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
    if !ok {
        log.Fatal("error casting public key to ECDSA")
    }

    publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
    fmt.Println(hexutil.Encode(publicKeyBytes)[4:]) // 0x049a7df67f79246283fdc93af76d4f8cdd62c4886e8cd870944e817dd0b97934fdd7719d0810951e03418205868a5c1b40b192451367f28e0088dd75e15de40c05

    //fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    fmt.Println(fromAddress)

    //Ropstenネットワークから、Nonce情報を読み取る
    nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        log.Fatal(err)
    }

    //トークン送金Transactionをテストネット送るためのgasLimit、
    value := big.NewInt(0) //（オプション）後で使用する関数NewTransactionの引数で必要になるため設定。Transactionと同時に送るETHの量を設定できます。
    gasLimit := uint64(2000000)

    //ロプステンネットワークから、現在のgasPriceを取得。トランザクションがマイニングされずに放置されることを防ぐ。
    gasPrice, err := client.SuggestGasPrice(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    //送金先を指定
    toAddress := common.HexToAddress("0xE202B444Db397F53AE05149fE2843D7841A2dCBE")
    //トークンコントラクトアドレスを指定
    tokenAddress := common.HexToAddress("0x5446E3481e3fe4b3082067145A47d7a0F09d5E1A")
    //ERC20のどの関数を使用するか指定。https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_sendtransaction
    transferFnSignature := []byte("transfer(address,uint256)")
    //hash化し、先頭から4バイトまで取得。これで使用する関数を指定したことになる。
    //hash := sha3.NewKeccak256()
    hash := sha3.NewLegacyKeccak256()
    hash.Write(transferFnSignature)
    methodID := hash.Sum(nil)[:4]

    //0埋め
    paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
    //送金額を設定
    pIntAmount := big.NewInt(int64(1000000))
    //0埋め
    paddedAmount := common.LeftPadBytes(pIntAmount.Bytes(), 32)

    //トランザクションで送るデータを作成
    var data []byte
    data = append(data, methodID...)
    data = append(data, paddedAddress...)
    data = append(data, paddedAmount...)

    /***** Preparing signed transaction *****/
    tx := types.NewTransaction(nonce, tokenAddress, value, gasLimit, gasPrice, data)
    signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, privateKey)
    if err != nil {
        log.Fatal(err)
    }

    //サインしたトランザクションをRopstenNetworkに送る。
    err = client.SendTransaction(context.Background(), signedTx)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Signed tx sent: %s", signedTx.Hash().Hex())









    //fmt.Printf("wei: %s\n", bal) // "wei: 74605500647408739782407023"
    //fmt.Fprint(w, "token balance : " + bal.String())
    //fmt.Fprint(w, "ether balance : " + balance.String())
    t, _ := template.ParseFiles("../../web/html/index.html")
    t.Execute(w, "token balance : " + bal.String() + "ether balance : " + balance.String())
}

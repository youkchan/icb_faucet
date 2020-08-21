package main

import (
    "fmt"
//    "context"
    "log"
    "net/http"
    "os"
    "strconv"
    "html/template"
//    "crypto/ecdsa"
//    "math/big"
    "github.com/joho/godotenv"
//    "github.com/ethereum/go-ethereum/accounts/abi/bind"
//    "github.com/ethereum/go-ethereum/common"
//    "github.com/ethereum/go-ethereum/ethclient"
//    token "github.com/youkchan/icb_faucet/pkg/token"
//    "github.com/ethereum/go-ethereum/crypto"
//    "github.com/ethereum/go-ethereum/common/hexutil"
//    "github.com/ethereum/go-ethereum/core/types"
//    "golang.org/x/crypto/sha3"
    ethereum "github.com/youkchan/icb_faucet/pkg/ethereum"
//    "reflect"
)

type Params struct{
    InvalidMessage string
    TxHash string
}


func main() {
    http.Handle("/web/css/", http.StripPrefix("/web/css/", http.FileServer(http.Dir("web/css/"))))
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
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }
}

func returnErrorParams(message string) Params {
    params := Params {
        InvalidMessage : message,
        TxHash : "",
    }

    return params
}


func sendHandler(w http.ResponseWriter, r *http.Request) {
    t := template.Must(template.ParseFiles("web/html/index.html"))
    var params Params
	err := r.ParseForm()
    if err != nil {
        params = returnErrorParams("Invalid Parameter")
        t.Execute(w, params)
        return
    }

	address := r.PostFormValue("address")
    account, err := ethereum.NewAccount(address)
    if err != nil {
        params = returnErrorParams("Invalid Address")
        t.Execute(w, params)
        return
    }

	amount, err := strconv.Atoi(r.PostFormValue("amount"))
    if err != nil {
        params = returnErrorParams("Invalid Amount")
        t.Execute(w, params)
        return
    }

	network, err := strconv.Atoi(r.PostFormValue("network"))
    if err != nil {
        params = returnErrorParams("Invalid Network")
        t.Execute(w, params)
        return
    }

    network_list := []ethereum.Network{
        *ethereum.NewNetwork(4, os.Getenv("INFURA_RINKEBY")),
        *ethereum.NewNetwork(3, os.Getenv("INFURA_ROPSTEN")),
    }
    client_factory := ethereum.NewClientFactory(network_list)

    ethereum_client, err := client_factory.CreateClient(network)
    if err != nil {
        params = returnErrorParams("Invalid Network")
        t.Execute(w, params)
        return
    }

    var token *ethereum.Token
    if ethereum_client.Network.Id == 4 {
        token = ethereum.NewToken(os.Getenv("TOKEN_RINKEBY_ADDRESS"))
    } else if ethereum_client.Network.Id == 3 {
        token = ethereum.NewToken(os.Getenv("TOKEN_ROPSTEN_ADDRESS"))
    }

    faucet_account := ethereum.NewSendableAccount(os.Getenv("PRIVATE_KEY"))
    txhash := ethereum_client.SendToken(*token, *faucet_account, *account, amount * 10000)

    params = Params {
        InvalidMessage : "",
        TxHash : txhash,
    }

    t.Execute(w, params)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    Env_load()
    t := template.Must(template.ParseFiles("web/html/index.html"))
    if r.URL.Path != "/" {
            http.NotFound(w, r)
            return
    }

    fmt.Println(r.RemoteAddr)

/*    network_list := []ethereum.Network{
        *ethereum.NewNetwork(4, os.Getenv("INFURA_RINKEBY")),
        *ethereum.NewNetwork(3, os.Getenv("INFURA_ROPSTEN")),
    }
    client_factory := ethereum.NewClientFactory(network_list)
    ethereum_client, err := client_factory.CreateClient(4)
    if err != nil {
        log.Fatal(err)
    }

    var token *ethereum.Token
    if ethereum_client.Network.Id == 4 {
        token = ethereum.NewToken(os.Getenv("TOKEN_RINKEBY_ADDRESS"))
    } else if ethereum_client.Network.Id == 3 {
        token = ethereum.NewToken(os.Getenv("TOKEN_ROPSTEN_ADDRESS"))
    }
    tokenBalance, err := ethereum_client.GetTokenBalance(*token, "0x751e0e0de1881f614F40C14c175bdd12d0DCaa24")
    fmt.Println(tokenBalance)
*/
    params := Params {
        InvalidMessage : "",
        //TxHash : "",
        TxHash : r.RemoteAddr,
    }
    t.Execute(w, params)
}

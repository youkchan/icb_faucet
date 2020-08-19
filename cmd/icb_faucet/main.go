package main

import (
    "fmt"
    "context"
    "log"
    "net/http"
    "os"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    token "github.com/youkchan/icb_faucet/pkg/token"
)

func main() {
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


func indexHandler(w http.ResponseWriter, r *http.Request) {
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
    //fmt.Println(balance)

    bal, err := rinkebyICBInstance.BalanceOf(&bind.CallOpts{}, account)
    if err != nil {
      log.Fatal(err)
    }
    
    //fmt.Printf("wei: %s\n", bal) // "wei: 74605500647408739782407023"
    fmt.Fprint(w, "token balance : " + bal.String())
    fmt.Fprint(w, "ether balance : " + balance.String())
}

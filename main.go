package main

import (
    //"fmt"
    "bytes"
//    "context"
    "log"
    "net/http"
    "net"
    "os"
    "strings"
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

func getIPAdress(r *http.Request) string {
    for _, h := range []string{"X-Forwarded-For", "X-Real-Ip"} {
        addresses := strings.Split(r.Header.Get(h), ",")
        // march from right to left until we get a public address
        // that will be the address right before our proxy.
        for i := len(addresses) -1 ; i >= 0; i-- {
            ip := strings.TrimSpace(addresses[i])
            // header can contain spaces too, strip those out.
            realIP := net.ParseIP(ip)
            if !realIP.IsGlobalUnicast() || isPrivateSubnet(realIP) {
                // bad address, go to next
                continue
            }
            return ip
        }
    }
    return ""
}

type ipRange struct {
    start net.IP
    end net.IP
}

// inRange - check to see if a given ip address is within a range given
func inRange(r ipRange, ipAddress net.IP) bool {
    // strcmp type byte comparison
    if bytes.Compare(ipAddress, r.start) >= 0 && bytes.Compare(ipAddress, r.end) < 0 {
        return true
    }
    return false
}

var privateRanges = []ipRange{
    ipRange{
        start: net.ParseIP("10.0.0.0"),
        end:   net.ParseIP("10.255.255.255"),
    },
    ipRange{
        start: net.ParseIP("100.64.0.0"),
        end:   net.ParseIP("100.127.255.255"),
    },
    ipRange{
        start: net.ParseIP("172.16.0.0"),
        end:   net.ParseIP("172.31.255.255"),
    },
    ipRange{
        start: net.ParseIP("192.0.0.0"),
        end:   net.ParseIP("192.0.0.255"),
    },
    ipRange{
        start: net.ParseIP("192.168.0.0"),
        end:   net.ParseIP("192.168.255.255"),
    },
    ipRange{
        start: net.ParseIP("198.18.0.0"),
        end:   net.ParseIP("198.19.255.255"),
    },
}


// isPrivateSubnet - check to see if this ip is in a private subnet
func isPrivateSubnet(ipAddress net.IP) bool {
    // my use case is only concerned with ipv4 atm
    if ipCheck := ipAddress.To4(); ipCheck != nil {
        // iterate over all our ranges
        for _, r := range privateRanges {
            // check if this ip is in a private range
            if inRange(r, ipAddress){
                return true
            }
        }
    }
    return false
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    Env_load()
    t := template.Must(template.ParseFiles("web/html/index.html"))
    if r.URL.Path != "/" {
            http.NotFound(w, r)
            return
    }

    ip := getIPAdress(r)


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
        TxHash : ip,
    }
    t.Execute(w, params)
}

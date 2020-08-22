package main

import (
    "fmt"
//    "bytes"
    //"context"
    "golang.org/x/net/context"
    "log"
    "time"
    "net/http"
//    "net"
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
    ipaddr "github.com/youkchan/icb_faucet/pkg/ipaddr"
    firebase "firebase.google.com/go"
    "firebase.google.com/go/db"
//    "firebase.google.com/go/auth"

    "google.golang.org/api/option"
    "reflect"
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
    if amount != 10 && amount != 30 && amount != 50 {
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

    if !strings.HasPrefix(r.Host, "localhost") {
        ip := ipaddr.GetIPAdress(r)
        isLimited := IsLimitedAccess(ip)
        if isLimited {
            params = returnErrorParams("Too many request, Please wait a moment")
            t.Execute(w, params)
            return
        }
    }

    faucet_account := ethereum.NewSendableAccount(os.Getenv("PRIVATE_KEY"))
    txhash := ethereum_client.SendToken(*token, *faucet_account, *account, amount * 10000)
    if !strings.HasPrefix(r.Host, "localhost") {
        ip := ipaddr.GetIPAdress(r)
        saveIPAddr(ip, amount)
    }

    params = Params {
        InvalidMessage : "",
        TxHash : txhash,
    }

    t.Execute(w, params)
}

type User struct {
    IPAddr string `json:"ipaddr"`
    Time    string `json:"time"`
    Amount    int `json:"amount"`
}

func initFirebaseRef() (*db.Ref){
    opt := option.WithCredentialsFile("serviceAccountKey.json")
    config := &firebase.Config{DatabaseURL: os.Getenv("FIREBASE_ENDPOINT")}
    app, err := firebase.NewApp(context.Background(), config, opt)
    if err != nil {
        log.Fatalln("Error initializing database client:", err)
    }
    client, err := app.Database(context.Background())
    if err != nil {
        log.Fatalln("Error initializing database client:", err)
    }
    ref := client.NewRef("users")
    return ref
}


func IsLimitedAccess(ipaddr string) bool {
    ref := initFirebaseRef()

    results, err := ref.OrderByChild("ipaddr").EqualTo(ipaddr).GetOrdered(context.Background())
    if err != nil {
        log.Fatalln(err)
    }

    jst, _ := time.LoadLocation("Asia/Tokyo")
    now := time.Now().In(jst)
    isLimited := false
    for _, r := range results {
        var u User
        if err := r.Unmarshal(&u); err != nil {
            log.Fatalln("Error unmarshaling result:", err)
        }

        var interval_time int
        if u.Amount == 10 {
            interval_time = 1
        } else if u.Amount == 30 {
            interval_time = 24
        } else if u.Amount == 50 {
            interval_time = 48
        } else {
            interval_time = 48
        }

        db_time, _ := time.Parse("2006-01-02 15:04:05 -0700 MST", u.Time)
        log.Println(now)
        log.Println(db_time)
        db_time = db_time.Add(time.Duration(interval_time) * time.Hour)
        log.Println(db_time)
        isLimited = now.Unix() < db_time.Unix()
    }

    return isLimited
}

func saveIPAddr(ipaddr string, amount int) {
    ref := initFirebaseRef()
    fmt.Println(reflect.TypeOf(ref))
    jst, _ := time.LoadLocation("Asia/Tokyo")
    now := time.Now().In(jst).Format("2006-01-02 15:04:05 -0700 MST")
    user := User {
        IPAddr: ipaddr,
        Time:    now,
        Amount:  amount,
    }
    _, err := ref.Push(context.Background() , user)
    if err != nil {
        log.Fatalln(err)
    }
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    Env_load()
    t := template.Must(template.ParseFiles("web/html/index.html"))
    if r.URL.Path != "/" {
            http.NotFound(w, r)
            return
    }

    ip := ipaddr.GetIPAdress(r)
    fmt.Println("ip")
    fmt.Println(ip)
    //usersRef := ref.Child("users")
    /*err = usersRef.Set(ctx, User{
            DateOfBirth: "June 23, 1912",
            FullName:    "Alan Turing"})*/

    //saveIPAddr("127.0.0.1", 10)
    //isLimited := IsLimitedAccess("127.0.0.1")

    //fmt.Println(isLimited)


    /*opt := option.WithCredentialsFile("serviceAccountKey.json")

    fmt.Println(cref)
    var users map[string]User

    //results, err := ref.OrderByChild("ipaddr").EqualTo(ipaddr).Get(ctx, &users)
    results := ref.Get(ctx, &users)
    fmt.Println(results)
    fmt.Println(reflect.TypeOf(results))*/


    fmt.Println(r.Host)
    fmt.Println(strings.HasPrefix(r.Host, "localhost"))
    //IsLimitedAccess("test")
    //results, err := ref.OrderByChild("full_name").EqualTo("Alan Turing").GetOrdered(ctx)
    //results, err := ref.OrderByValue().GetOrdered(ctx)
    /*fmt.Println("results")
    fmt.Println(results)

    for _, r := range results {
        var d User
        if err := r.Unmarshal(&d); err != nil {
                log.Fatalln("Error unmarshaling result:", err)
        }
        fmt.Println(d)
    }*/
    start := time.Now().Format("2006-01-02 15:04:05")
    fmt.Println(start)

    ttime, _ := time.Parse("2006-01-02 15:04:05", start)
    fmt.Println(ttime)
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
        TxHash : "",
    }
    t.Execute(w, params)
}

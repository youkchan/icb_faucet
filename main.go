package main

import (
	"github.com/joho/godotenv"
	ethereum "github.com/youkchan/icb_faucet/pkg/ethereum"
	db "github.com/youkchan/icb_faucet/pkg/firebase"
	ipaddr "github.com/youkchan/icb_faucet/pkg/ipaddr"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Params passing to template
type Params struct {
	InvalidMessage string
	Network        int
	TxHash         string
}

// User Data structure for storing IP addresses
type User struct {
	IPAddr string `json:"ipaddr"`
	Time   string `json:"time"`
	Amount int    `json:"amount"`
}

func main() {
	http.Handle("/web/css/", http.StripPrefix("/web/css/", http.FileServer(http.Dir("web/css/"))))
	http.HandleFunc("/send", handleRequest())
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

func handleRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		LoadEnv()
		var ref db.DBClient
		if strings.HasPrefix(r.Host, "localhost") || strings.HasSuffix(os.Args[0], ".test") {
			ref = db.InitFirebaseRef("users-test", os.Getenv("FIREBASE_ENDPOINT"), "serviceAccountKey.json")
		} else {
			ref = db.InitFirebaseRef("users", os.Getenv("FIREBASE_ENDPOINT"), "serviceAccountKey.json")
		}
		sendHandler(w, r, ref)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("web/html/index.html"))
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	params := Params{
		InvalidMessage: "",
		Network:        0,
		TxHash:         "",
	}
	t.Execute(w, params)
}
func sendHandler(w http.ResponseWriter, r *http.Request, ref db.DBClient) {
	LoadEnv()
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

	rinkeby, err := ethereum.NewNetwork(4, os.Getenv("INFURA_RINKEBY"))
	if err != nil {
		params = returnErrorParams("Invalid Network")
		t.Execute(w, params)
		return
	}

	ropsten, err := ethereum.NewNetwork(3, os.Getenv("INFURA_ROPSTEN"))
	if err != nil {
		params = returnErrorParams("Invalid Network")
		t.Execute(w, params)
		return
	}

	networkList := []ethereum.Network{
		*rinkeby,
		*ropsten,
	}
	clientFactory := ethereum.NewClientFactory(networkList)

	ethereumClient, err := clientFactory.CreateClient(network)
	if err != nil {
		params = returnErrorParams("Invalid Network")
		t.Execute(w, params)
		return
	}

	var token *ethereum.Token
	if ethereumClient.Network.ID == 4 {
		token = ethereum.NewToken(os.Getenv("TOKEN_RINKEBY_ADDRESS"))
	} else if ethereumClient.Network.ID == 3 {
		token = ethereum.NewToken(os.Getenv("TOKEN_ROPSTEN_ADDRESS"))
	}

	var isLimited bool
	if strings.HasPrefix(r.Host, "localhost") || strings.HasSuffix(os.Args[0], ".test") {
		isLimited = IsLimitedAccess("127.0.0.1", ref)
	} else {
		ip := ipaddr.GetIPAdress(r)
		isLimited = IsLimitedAccess(ip, ref)
	}

	if isLimited {
		params = returnErrorParams("Too many request, Please wait a moment")
		t.Execute(w, params)
		return
	}

	faucetAccount, err := ethereum.NewSendableAccount(os.Getenv("PRIVATE_KEY"))
	if err != nil {
		params = returnErrorParams("Invalid Private Key")
		t.Execute(w, params)
		return
	}

	txhash := ethereumClient.SendToken(*token, *faucetAccount, *account, amount*10000)
	if strings.HasPrefix(r.Host, "localhost") || strings.HasSuffix(os.Args[0], ".test") {
		SaveIPAddr("127.0.0.1", amount, ref)
	} else {
		ip := ipaddr.GetIPAdress(r)
		SaveIPAddr(ip, amount, ref)
	}

	params = Params{
		InvalidMessage: "",
		Network:        network,
		TxHash:         txhash,
	}

	t.Execute(w, params)
}

// LoadEnv loading .env file
func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func returnErrorParams(message string) Params {
	params := Params{
		InvalidMessage: message,
		Network:        0,
		TxHash:         "",
	}
	return params
}
func getIntervalTime(amount int) int {
	var intervalTime int
	if amount == 10 {
		intervalTime = 1
	} else if amount == 30 {
		intervalTime = 24
	} else if amount == 50 {
		intervalTime = 48
	} else {
		intervalTime = 48
	}

	return intervalTime
}

// IsLimitedAccess check whether a user requests many times
func IsLimitedAccess(ipaddr string, ref db.DBClient) bool {
	results, err := ref.Fetch(ipaddr)
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
		intervalTime := getIntervalTime(u.Amount)

		dbTime, _ := time.Parse("2006-01-02 15:04:05 -0700 MST", u.Time)
		dbTime = dbTime.Add(time.Duration(intervalTime) * time.Hour)
		isLimited = now.Unix() < dbTime.Unix()
	}

	return isLimited
}

// SaveIPAddr save user ip address in db
func SaveIPAddr(ipaddr string, amount int, ref db.DBClient) {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	now := time.Now().In(jst).Format("2006-01-02 15:04:05 -0700 MST")
	user := User{
		IPAddr: ipaddr,
		Time:   now,
		Amount: amount,
	}
	err := ref.Push(user)
	if err != nil {
		log.Fatalln(err)
	}
}

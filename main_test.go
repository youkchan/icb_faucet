package main

import (
	db "github.com/youkchan/icb_faucet/pkg/firebase"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestGetIntervalTime_10(t *testing.T) {
	intervalTime := getIntervalTime(10)
	if intervalTime != 1 {
		t.Error("invalid interval")
	}
}

func TestGetIntervalTime_30(t *testing.T) {
	intervalTime := getIntervalTime(30)
	if intervalTime != 24 {
		t.Error("invalid interval")
	}
}

func TestGetIntervalTime_50(t *testing.T) {
	intervalTime := getIntervalTime(50)
	if intervalTime != 48 {
		t.Error("invalid interval")
	}
}

func TestSaveIPAddr(t *testing.T) {
	LoadEnv()
	ref := db.InitFirebaseRef("users-test", os.Getenv("FIREBASE_ENDPOINT"), "serviceAccountKey.json")
	ref.Delete()

	jst, _ := time.LoadLocation("Asia/Tokyo")
	now := time.Now().In(jst).Format("2006-01-02 15:04:05 -0700 MST")
	testIP := "testip"
	testAmount := 50
	SaveIPAddr(testIP, testAmount, ref)
	results, _ := ref.Fetch("testip")
	r := results[0]
	var u User
	r.Unmarshal(&u)
	if u.IPAddr != testIP {
		t.Error("Invalid savedata")
	}
	if u.Amount != testAmount {
		t.Error("Invalid savedata")
	}
	if u.Time != now {
		t.Error("Invalid savedata")
	}

}

func TestIsLimitedAccess_limited(t *testing.T) {
	LoadEnv()
	jst, _ := time.LoadLocation("Asia/Tokyo")
	now := time.Now().In(jst).Format("2006-01-02 15:04:05 -0700 MST")
	user := db.User{
		IPAddr: "2001:268:c145:92df:80dc:a420:d569:869c",
		Time:   now,
		Amount: 10,
	}
	fake := db.FakeDBClient{FakeData: user}
	isLimited := IsLimitedAccess(user.IPAddr, fake)
	if !isLimited {
		t.Error("Invalid condition")
	}
}

func TestIsLimitedAccess_not_limited(t *testing.T) {
	LoadEnv()
	jst, _ := time.LoadLocation("Asia/Tokyo")
	now := time.Now().In(jst).Add(-1 * time.Hour).Add(-1 * time.Second).Format("2006-01-02 15:04:05 -0700 MST")
	user := db.User{
		IPAddr: "2001:268:c145:92df:80dc:a420:d569:869c",
		Time:   now,
		Amount: 10,
	}
	fake := db.FakeDBClient{FakeData: user}
	isLimited := IsLimitedAccess(user.IPAddr, fake)
	if isLimited {
		t.Error("Invalid condition")
	}
}

func TestSendHandler(t *testing.T) {
	LoadEnv()
	ref := db.InitFirebaseRef("users-test", os.Getenv("FIREBASE_ENDPOINT"), "serviceAccountKey.json")
	ref.Delete()

	mux := http.NewServeMux()
	mux.HandleFunc("/send", handleRequest())
	writer := httptest.NewRecorder()
	validAddress := "0x578D9B2d04bc99007B941787E88E4ea57D888A56"
	data := strings.NewReader("address=" + validAddress + "&amount=10&network=4")
	request, _ := http.NewRequest("POST", "/send", data)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mux.ServeHTTP(writer, request)

	if writer.Code != 200 {
		t.Error("invalid response code")
	}

	results, _ := ref.Fetch("127.0.0.1")
	if len(results) == 0 {
		t.Fatal("Invalid savedata")
	}
	r := results[0]
	var u User
	r.Unmarshal(&u)
	if u.IPAddr != "127.0.0.1" {
		t.Error("Invalid savedata")
	}
	if u.Amount != 10 {
		t.Error("Invalid savedata")
	}

}

func TestIndexHandler(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	writer := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/", nil)
	mux.ServeHTTP(writer, request)

	if writer.Code != 200 {
		t.Error("invalid response code")
	}
}

package datasource

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestNewSqlite(t *testing.T) {
	dbPath, _ := os.MkdirTemp("./", "test-")
	dbName := "test.db"
	defer os.RemoveAll(dbPath)

	sqlite, err := NewSqlite(fmt.Sprintf("%s/%s", dbPath, dbName))
	if err != nil {
		t.Error(err)
	}
	defer sqlite.DB.Close()

	tt := time.Now().UTC().Format("2006-01-02 15:04:05")
	rl := RequestLogs{
		CreateTime: tt,
		ClientIP:   "127.0.0.1",
		URL:        "/test/ip",
		Counter:    1,
		LastTime:   tt,
	}

	second := "3"
	if err := sqlite.InsertOrUpdateLogs(rl, second); err != nil {
		t.Error(err)
	}

	firstCount := sqlite.RateCount(rl.ClientIP, rl.URL, second)
	fmt.Println(firstCount)

	time.Sleep(4 * time.Second)
	secondCount := sqlite.RateCount(rl.ClientIP, rl.URL, second)
	fmt.Println(secondCount)

	rl.LastTime = time.Now().UTC().Format("2006-01-02 15:04:05")
	if err := sqlite.InsertOrUpdateLogs(rl, second); err != nil {
		t.Error(err)
	}

	thirdCount := sqlite.RateCount(rl.ClientIP, rl.URL, "3")
	fmt.Println(thirdCount)

	time.Sleep(1 * time.Second)
	rl.LastTime = time.Now().UTC().Format("2006-01-02 15:04:05")
	if err := sqlite.InsertOrUpdateLogs(rl, second); err != nil {
		t.Error(err)
	}

	fourthCount := sqlite.RateCount(rl.ClientIP, rl.URL, "3")
	fmt.Println(fourthCount)

	time.Sleep(4 * time.Second)
	rl.LastTime = time.Now().UTC().Format("2006-01-02 15:04:05")
	if err := sqlite.InsertOrUpdateLogs(rl, second); err != nil {
		t.Error(err)
	}

	fifthCount := sqlite.RateCount(rl.ClientIP, rl.URL, "3")
	fmt.Println(fifthCount)

	fmt.Println(sqlite.InBlacklist(rl.ClientIP))
	if err := sqlite.AddBlacklist(rl.ClientIP); err != nil {
		t.Error(err)
	}
	fmt.Println(sqlite.InBlacklist(rl.ClientIP))
}

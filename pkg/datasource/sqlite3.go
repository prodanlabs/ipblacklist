package datasource

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

const (
	dbDriverName = "sqlite3"
)

type RequestLogs struct {
	CreateTime string
	ClientIP   string
	URL        string
	Counter    int
	LastTime   string
}

type Blacklist struct {
	IP string
}

type Sqlite struct {
	*sql.DB
}

func NewSqlite(datasourceName string) (*Sqlite, error) {
	db, err := sql.Open(dbDriverName, fmt.Sprintf("file:%s?cache=shared&mode=rwc&_journal_mode=WAL", datasourceName))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	if err := crateTable(db); err != nil {
		return nil, err
	}

	return &Sqlite{
		db,
	}, nil
}

func crateTable(db *sql.DB) error {
	requestLogs := `create table if not exists "request_logs" (
                "create_time" timestamp not null,
                "client_ip" text not null,
                "url" text not null,
                "counter" text not null,
                "last_time" timestamp NOT NULL
        );`
	_, err := db.Exec(requestLogs)
	if err != nil {
		return err
	}

	blacklist := `create table if not exists "blacklist" (
         "id" INTEGER PRIMARY KEY AUTOINCREMENT,
                "ip" text not null
        );`

	_, err = db.Exec(blacklist)

	return err
}

func (c *Sqlite) insertRequestLogs(rl RequestLogs) error {
	tx, err := c.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO request_logs(create_time,client_ip,url,counter,last_time) values(?,?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(rl.CreateTime, rl.ClientIP, rl.URL, rl.Counter, rl.LastTime)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (c *Sqlite) updateRequestLogs(rl RequestLogs, second string) error {
	stmt, err := c.Prepare("update request_logs set last_time= ? , counter= ? where client_ip= ? ")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(rl.LastTime, c.RateCount(rl.ClientIP, rl.URL, second)+1, rl.ClientIP)
	if err != nil {
		return err
	}

	return nil
}

func (c *Sqlite) inRequestLogs(ip, url string) bool {
	rows, err := c.Query("SELECT client_ip FROM request_logs where client_ip= ?   and url= ?", ip, url)
	if err != nil {
		log.Printf("Failed to get %s in request_logs: %v", ip, err)
		return false
	}
	defer rows.Close()

	var clientIP []string
	for rows.Next() {
		r := ""
		err = rows.Scan(&r)
		if err != nil {
			return false
		}
		clientIP = append(clientIP, r)
	}

	if len(clientIP) == 0 {
		return false
	}

	return true
}
func (c *Sqlite) InsertOrUpdateLogs(rl RequestLogs, second string) error {
	if !c.inRequestLogs(rl.ClientIP, rl.URL) {
		return c.insertRequestLogs(rl)
	}

	return c.updateRequestLogs(rl, second)
}

func (c *Sqlite) RateCount(ip, url, second string) int {
	selectSql := fmt.Sprintf("SELECT counter FROM request_logs where client_ip= ? and url= ? and last_time >= DateTime('now','-%s second')", second)
	rows, err := c.Query(selectSql, ip, url)
	if err != nil {
		return 0
	}
	defer rows.Close()

	var counts []int
	for rows.Next() {
		i := 0
		err = rows.Scan(&i)
		if err != nil {
			return 0
		}
		counts = append(counts, i)
	}

	if len(counts) == 0 {
		return 0
	}

	return counts[0]
}

func (c *Sqlite) AddBlacklist(ip string) error {
	tx, err := c.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO blacklist(ip) values(?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(ip)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (c *Sqlite) InBlacklist(ip string) bool {
	rows, err := c.Query("SELECT ip FROM blacklist where ip= ?  ", ip)
	if err != nil {
		log.Printf("Failed to get %s in blacklist: %v", ip, err)
		return false
	}
	defer rows.Close()

	var blacklist []string
	for rows.Next() {
		r := ""
		err = rows.Scan(&r)
		if err != nil {
			return false
		}
		blacklist = append(blacklist, r)
	}

	if len(blacklist) == 0 {
		return false
	}

	return true
}

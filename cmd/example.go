package main

import (
	"os"
	"os/signal"
	"syscall"

	"fmt"
	"sync"
	"time"

	"database/sql"

	"github.com/flike/idgo"
	_ "github.com/go-sql-driver/mysql"
)

const (
	tableName              = "id_table"
	idKey                  = "mock_id"
	CreateTableNTSQLFormat = `
	CREATE TABLE IF NOT EXISTS %s (
    id bigint(20) unsigned NOT NULL auto_increment,
    id_from_idgo bigint(20) unsigned NOT NULL default '0',
    PRIMARY KEY  (id),
    KEY _idgo(id_from_idgo)
) ENGINE=Innodb DEFAULT CHARSET=utf8 `
)

var running bool = true

func GetDatabase() (*sql.DB, error) {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/test?charset=utf8")
	if err != nil {
		fmt.Printf("main:open database error:%s\n", err.Error())
		return nil, err
	}
	//create table
	CreateIDTable(db)
	return db, nil
}

func CreateIDTable(db *sql.DB) {
	sql := fmt.Sprintf(CreateTableNTSQLFormat, tableName)
	_, err := db.Exec(sql)
	if err != nil {
		s := fmt.Sprintf("CreateIDTable error:%s", err.Error())
		panic(s)
	}
}

func main() {
	var wg sync.WaitGroup
	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		sig := <-sc
		running = false
		fmt.Printf("main:Got signal:%v", sig)
	}()
	fmt.Printf("main:Mock get id process start!\n")
	db, err := GetDatabase()
	if err != nil {
		fmt.Printf("main:GetDatabase error:%s\n", err.Error())
		return
	}
	idGenerator, err := GetIdGenerator(db, idKey)
	if err != nil {
		fmt.Printf("main:GetIdGenerator error:%s\n", err.Error())
		return
	}
	wg.Add(1)
	go MockGetId(idGenerator, db, &wg)
	wg.Wait()
}

func GetIdGenerator(db *sql.DB, key string) (*idgo.MySQLIdGenerator, error) {
	idGenerator, err := idgo.NewMySQLIdGenerator(db, key)
	if err != nil {
		return nil, err
	}

	isExist, err := idGenerator.IsKeyExist()
	if err != nil {
		return nil, err
	}
	if isExist {
		err = idGenerator.Init()
		if err != nil {
			return nil, err
		}
	} else {
		err = idGenerator.Reset(1, false)
		if err != nil {
			return nil, err
		}
	}
	return idGenerator, nil
}

//get about 60 ids per second
func MockGetId(idGenerator *idgo.MySQLIdGenerator, db *sql.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	sqlFormat := "insert into %s(id_from_idgen) values(%d)"
	for running {
		id, err := idGenerator.Next()
		if err != nil {
			fmt.Printf("main:idGenerator Next error:%s\n", err.Error())
			continue
		}
		sql := fmt.Sprintf(sqlFormat, tableName, id)
		_, err = db.Exec(sql)
		if err != nil {
			fmt.Printf("main:insert into idtable error:%s,id:%d\n",
				err.Error(), id)
			continue
		}
		time.Sleep(time.Millisecond * 80)
	}
}

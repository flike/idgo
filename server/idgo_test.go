package server

import (
	"database/sql"
	"fmt"
	"sync"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

var wg sync.WaitGroup
var db *sql.DB

func init() {
	var err error

	db, err = sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/test?charset=utf8")
	if err != nil {
		fmt.Println(err.Error())
	}
}

func GetId(idGenerator *MySQLIdGenerator) {
	defer wg.Done()
	for i := 0; i < 100; i++ {
		_, err := idGenerator.Next()
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}

func TestMySQLIdgen(t *testing.T) {
	idGenerator, err := NewMySQLIdGenerator(db, "mysql_victory")
	if err != nil {
		t.Fatal(err.Error())
	}
	err = idGenerator.Reset(1, false)
	if err != nil {
		t.Fatal(err.Error())
	}
	//10 goroutine
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go GetId(idGenerator)
	}
	wg.Wait()
	id, err := idGenerator.Next()
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(id)
}

func BenchmarkMySQLIdgen(b *testing.B) {
	idGenerator, err := NewMySQLIdGenerator(db, "mysql_file")
	if err != nil {
		b.Fatal(err.Error())
	}
	err = idGenerator.Reset(1, false)
	if err != nil {
		b.Fatal(err.Error())
	}

	b.StartTimer()
	for i := 0; i < 1000; i++ {
		_, err = idGenerator.Next()
		if err != nil {
			b.Fatal(err.Error())
		}
	}

	b.StopTimer()
}

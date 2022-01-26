package server

import (
	"database/sql"
	"fmt"
	"net"
	"runtime"
	"sync"

	"github.com/flike/golog"

	"github.com/flike/idgo/config"
)

const (
	KeyRecordTableName         = "__idgo__"
	CreateRecordTableSQLFormat = `
	CREATE TABLE %s (
    k VARCHAR(255) NOT NULL,
    PRIMARY KEY (k)
) ENGINE=Innodb DEFAULT CHARSET=utf8 `

	// create key table if not exist
	CreateRecordTableNTSQLFormat = `
	CREATE TABLE IF NOT EXISTS %s (
    k VARCHAR(255) NOT NULL,
    PRIMARY KEY (k)
) ENGINE=Innodb DEFAULT CHARSET=utf8 `

	InsertKeySQLFormat  = "INSERT INTO %s (k) VALUES ('%s')"
	SelectKeySQLFormat  = "SELECT k FROM %s WHERE k = '%s'"
	SelectKeysSQLFormat = "SELECT k FROM %s"
	DeleteKeySQLFormat  = "DELETE FROM %s WHERE k = '%s'"
)

type Server struct {
	cfg *config.Config

	listener        net.Listener
	db              *sql.DB
	keyGeneratorMap map[string]*MySQLIdGenerator
	sync.RWMutex
	running bool
}

func NewServer(c *config.Config) (*Server, error) {
	s := new(Server)
	s.cfg = c

	var err error
	// init db
	proto := "mysql"
	charset := "utf8"
	// root:@tcp(127.0.0.1:3306)/test?charset=utf8
	url := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		c.DatabaseConfig.User,
		c.DatabaseConfig.Password,
		c.DatabaseConfig.Host,
		c.DatabaseConfig.Port,
		c.DatabaseConfig.DBName,
		charset,
	)

	s.db, err = sql.Open(proto, url)
	if err != nil {
		golog.Error("main", "NewServer", "open database error", 0,
			"err", err.Error(),
		)
		return nil, err
	}

	netProto := "tcp"
	s.listener, err = net.Listen(netProto, s.cfg.Addr)
	if err != nil {
		return nil, err
	}
	s.keyGeneratorMap = make(map[string]*MySQLIdGenerator)

	golog.Info("server", "NewServer", "Server running", 0,
		"netProto",
		netProto,
		"address",
		s.cfg.Addr,
	)

	return s, nil
}

func (s *Server) Init() error {
	createTableNtSQL := fmt.Sprintf(CreateRecordTableNTSQLFormat, KeyRecordTableName)
	selectKeysSQL := fmt.Sprintf(SelectKeysSQLFormat, KeyRecordTableName)
	_, err := s.db.Exec(createTableNtSQL)
	if err != nil {
		return err
	}
	rows, err := s.db.Query(selectKeysSQL)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		idGenKey := ""
		err := rows.Scan(&idGenKey)
		if err != nil {
			return err
		}
		if idGenKey != "" {
			idgen, ok := s.keyGeneratorMap[idGenKey]
			if ok == false {
				isExist, err := s.IsKeyExist(idGenKey)
				if err != nil {
					return err
				}
				if isExist {
					idgen, err = NewMySQLIdGenerator(s.db, idGenKey, BatchCount)
					if err != nil {
						return err
					}
					s.keyGeneratorMap[idGenKey] = idgen
				}
			}
		}
	}
	return nil
}

func (s *Server) Serve() error {
	s.running = true
	for s.running {
		conn, err := s.listener.Accept()
		if err != nil {
			golog.Error("server", "Run", err.Error(), 0)
			continue
		}

		go s.onConn(conn)
	}
	return nil
}

func (s *Server) onConn(conn net.Conn) error {
	defer func() {
		clientAddr := conn.RemoteAddr().String()
		r := recover()
		if err, ok := r.(error); ok {
			const size = 4096
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)] // 获得当前goroutine的stacktrace
			golog.Error("server", "onConn", "error", 0,
				"remoteAddr", clientAddr,
				"stack", string(buf),
				"err", err.Error(),
			)
			reply := &ErrorReply{
				message: err.Error(),
			}
			reply.WriteTo(conn)
		}
		conn.Close()
	}()

	for {
		request, err := NewRequest(conn)
		if err != nil {
			return err
		}

		reply := s.ServeRequest(request)
		if _, err := reply.WriteTo(conn); err != nil {
			golog.Error("server", "onConn", "reply write error", 0,
				"err", err.Error())
			return err
		}

	}
	return nil
}

func (s *Server) ServeRequest(request *Request) Reply {
	switch request.Command {
	case "GET":
		return s.handleGet(request)
	case "SET":
		return s.handleSet(request)
	case "EXISTS":
		return s.handleExists(request)
	case "DEL":
		return s.handleDel(request)
	case "SELECT":
		return s.handleSelect(request)
	default:
		return ErrMethodNotSupported
	}

	return nil
}

func (s *Server) Close() {
	s.running = false
	if s.listener != nil {
		s.listener.Close()
	}
	golog.Info("server", "close", "server closed!", 0)
}

func (s *Server) IsKeyExist(key string) (bool, error) {
	var tableName string
	var haveValue bool
	if len(key) == 0 {
		return false, nil
	}
	getKeySQL := fmt.Sprintf(GetKeySQLFormat, key)
	rows, err := s.db.Query(getKeySQL)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&tableName)
		if err != nil {
			return false, err
		}
		haveValue = true
	}
	if haveValue == false {
		return false, nil
	}
	return true, nil
}

func (s *Server) GetKey(key string) (string, error) {
	keyName := ""
	selectKeySQL := fmt.Sprintf(SelectKeySQLFormat, KeyRecordTableName, key)
	rows, err := s.db.Query(selectKeySQL)
	if err != nil {
		return keyName, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&keyName)
		if err != nil {
			return keyName, err
		}
	}
	if keyName == "" {
		return keyName, fmt.Errorf("%s:not exists key", key)
	}
	return keyName, nil
}

func (s *Server) SetKey(key string) error {
	if len(key) == 0 {
		return fmt.Errorf("%s:invalid key", key)
	}
	_, err := s.GetKey(key)
	if err == nil {
		return nil
	} else {
		insertKeySQL := fmt.Sprintf(InsertKeySQLFormat, KeyRecordTableName, key)
		_, err = s.db.Exec(insertKeySQL)
		if err != nil {
			return err
		}
		return nil
	}
}

func (s *Server) DelKey(key string) error {
	if len(key) == 0 {
		return fmt.Errorf("%s:invalid key", key)
	}
	_, err := s.GetKey(key)
	if err == nil {
		deletetKeySQL := fmt.Sprintf(DeleteKeySQLFormat, KeyRecordTableName, key)
		_, err = s.db.Exec(deletetKeySQL)
		if err != nil {
			return err
		}
		return nil
	} else {
		return nil
	}
}

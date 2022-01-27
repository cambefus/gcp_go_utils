package pgdb

/*
	Wrapper for gcp cloudsql postgresql db connections
	uses pooling to allow concurrent queries, but supports connection to only a single database at a time
*/

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
)

type DBPool struct {
	DBCon     *pgxpool.Pool
	connected bool
}

var CTxt = context.Background()

const maxConnections = 20

// NewExternalDBPool - requires TLS certificates to make the connection to the database
// expects a postgresql connection string the in the form of  "user=xxx password=xxxx host=xx.xx.xx.xx port=xxxx dbname=xxxx sslmode=?????"
func NewExternalDBPool(cparms, tls_key, tls_cert string) (*DBPool, error) {
	cert, err := tls.LoadX509KeyPair(tls_cert, tls_key)
	if err != nil {
		return nil, err
	}
	tlsc := &tls.Config{Certificates: []tls.Certificate{cert}}
	tlsc.InsecureSkipVerify = true
	return getConnection(cparms, tlsc)
}

// NewDBPool -
// expects a postgresql connection string the in the form of  "user=xxx password=xxxx host=xx.xx.xx.xx port=xxxx dbname=xxxx sslmode=?????"
func NewDBPool(cparms string) (*DBPool, error) {
	return getConnection(cparms, nil)
}

func getConnection(cparms string, tls *tls.Config) (*DBPool, error) {
	this := new(DBPool)
	cfg, e1 := pgxpool.ParseConfig(cparms)
	if e1 != nil {
		return nil, errors.New(fmt.Sprintf(`unable to parse connection parameters: %v`, e1))
	}
	cfg.ConnConfig.TLSConfig = tls
	cfg.MaxConns = maxConnections
	this.DBCon, e1 = pgxpool.ConnectConfig(CTxt, cfg)
	if e1 != nil {
		return this, errors.New(fmt.Sprintf("Unable to establish connection: %v", e1))
	}
	this.connected = true
	return this, nil // success
}

// TODO : make database logging configurable
//cfg.ConnConfig.Logger = logrusadapter.NewLogger( log.StandardLogger())
//cfg.ConnConfig.LogLevel = pgx.LogLevelDebug

// IsConnected - returns true if we have a valid connection
func (p *DBPool) IsConnected() bool {
	return p.connected
}

// Close - shut down the current database connection pool
func (p *DBPool) Close() {
	if p.connected && p.DBCon != nil {
		p.DBCon.Close()
		p.DBCon = nil
		p.connected = false
	}
}

// GetCount - execute a sql query that returns a single integer value
func (p *DBPool) GetCount(q string, args ...interface{}) (int, error) {
	count := 0
	rows, err := p.DBCon.Query(CTxt, q, args...)
	if err != nil {
		return 0, err
	} else {
		rows.Next()
		err = rows.Scan(&count)
		if err != nil {
			return 0, err
		}
		rows.Close()
	}
	return count, nil
}

func (p *DBPool) Query(q string, args ...interface{}) (pgx.Rows, error) {
	return p.DBCon.Query(CTxt, q, args...)
}

// Execute - execute a sql command that returns no rows (gives count of rows affected)
func (p *DBPool) Execute(q string, args ...interface{}) (int, error) {
	commandTag, err := p.DBCon.Exec(CTxt, q, args...)
	if err == nil {
		return int(commandTag.RowsAffected()), nil
	}
	return 0, err
}

func BToI(b bool) int {
	if b {
		return 1
	}
	return 0
}

func IToB(i int) bool {
	return i != 0
}

//IsDuplicate - returns true if err is mysql duplicate entry notification
func IsDuplicate(err error) bool {
	me, ok := err.(*pgconn.PgError)
	return ok && me.Code == pgerrcode.UniqueViolation
}

func IsForeignKeyConstraint(err error) bool {
	me, ok := err.(*pgconn.PgError)
	return ok && me.Code == pgerrcode.ForeignKeyViolation
}

func (p *DBPool) Optimize() error {
	_, err := p.Execute(`Vacuum Analyze`)
	return err
}

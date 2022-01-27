package pgdb

import (
	"github.com/cambefus/gcp_go_utils/secrets"
	"testing"
)

func Test_All(t *testing.T) {
	s, _ := secrets.InitializeFromEnvironment(`utilities_config`)
	cs := s.GetString(`CLOUDSQL`)
	tk := s.GetString(`TLS_CLIENT_KEY`)
	tc := s.GetString(`TLS_CLIENT_CERT`)
	p, e := NewExternalDBPool(cs, tk, tc)
	if e != nil {
		t.Error(e)
		return
	}

	// SET LOCAL has no effect if executed outside of a transaction
	val, err := p.Execute(`set local time zone $1`, `LOCAL`)
	if val != 0 && err != nil {
		t.Error(`Failed to execute simple query`)
	}

	cnt, err := p.Execute(`SELECT count(*) FROM information_schema.tables`)
	if cnt < 1 || err != nil {
		t.Error(`Failed to get count`, cnt, err)
	}

	q := `SELECT table_name FROM information_schema.tables where table_name like 'pg_r%'`
	rows, err := p.Query(q)
	if err != nil {
		t.Error(err)
	} else {
		for rows.Next() {
			_, e1 := p.GetCount(`SELECT count(*) FROM information_schema.tables`)
			if e1 != nil {
				t.Error(e1)
			}
		}
		rows.Close()
	}

	mc := p.DBCon.Stat().MaxConns()
	if mc != maxConnections {
		t.Error(`unexpected maxConnections: `, mc)
	}
}

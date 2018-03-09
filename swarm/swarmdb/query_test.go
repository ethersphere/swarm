// Copyright (c) 2018 Wolk Inc.  All rights reserved.

// The SWARMDB library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The SWARMDB library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package swarmdb_test

import (
	"fmt"
	//sdbc "swarmdbcommon"
	sdbc "github.com/ethereum/go-ethereum/swarm/swarmdb/swarmdbcommon"
	"reflect"
	"strings"
	"swarmdb"
	"testing"
)

func TestParseQuery(t *testing.T) {

	rawqueries := map[string]string{
		`get1`:         `select name from contacts where age >= 35`,
		`get2`:         `select name, age from contacts where email = 'rodney@wolk.com'`,
		`doublequotes`: `select name, age from contacts where email = "rodney@wolk.com"`,
		`not`:          `select name, age from contacts where email != 'rodney@wolk.com'`,
		`insert`:       `insert into contacts(email, name, age) values("bertie@gmail.com","Bertie Basset", 7)`,
		`update`:       `UPDATE contacts set age = 8, name = "Bertie B" where email = "bertie@gmail.com"`,
		`delete`:       `delete from contacts where age >= 25`,
		//`precedence`:   `select * from a where a=b and c=d or e=f`,
		//`like`:         `select name, age from contacts where email like '%wolk%'`,
		//`is`:           `select name, age from contacts where age is not null`,
		//`and`:          `select name, age from contacts where email = 'rodney@wolk.com' and age = 38`,
		//`or`:           `select name, age from contacts where email = 'rodney@wolk.com' or age = 35`,
		//`groupby`:      `select name, age from contacts where age >= 35 group by email`,
	}

	expected := make(map[string]swarmdb.QueryOption)
	expected[`get1`] = swarmdb.QueryOption{
		Type:  "Select",
		Table: "contacts",
		RequestColumns: []sdbc.Column{
			sdbc.Column{ColumnName: "name"},
		},
		Where:     swarmdb.Where{Left: "age", Right: "35", Operator: ">="},
		Ascending: 1,
	}
	expected[`get2`] = swarmdb.QueryOption{
		Type:  "Select",
		Table: "contacts",
		RequestColumns: []sdbc.Column{
			sdbc.Column{ColumnName: "name"},
			sdbc.Column{ColumnName: "age"},
		},
		Where:     swarmdb.Where{Left: "email", Right: "rodney@wolk.com", Operator: "="},
		Ascending: 1,
	}
	expected[`doublequotes`] = swarmdb.QueryOption{
		Type:  "Select",
		Table: "contacts",
		RequestColumns: []sdbc.Column{
			sdbc.Column{ColumnName: "name"},
			sdbc.Column{ColumnName: "age"},
		},
		Where:     swarmdb.Where{Left: "email", Right: "rodney@wolk.com", Operator: "="},
		Ascending: 1,
	}
	expected[`not`] = swarmdb.QueryOption{
		Type:  "Select",
		Table: "contacts",
		RequestColumns: []sdbc.Column{
			sdbc.Column{ColumnName: "name"},
			sdbc.Column{ColumnName: "age"},
		},
		Where:     swarmdb.Where{Left: "email", Right: "rodney@wolk.com", Operator: "!="},
		Ascending: 1,
	}
	expected[`insert`] = swarmdb.QueryOption{
		Type:  "Insert",
		Table: "contacts",
		Inserts: []sdbc.Row{
			sdbc.Row{
				"name":  "Bertie Basset",
				"age":   float64(7),
				"email": "bertie@gmail.com"},
		},
		Ascending: 1,
	}
	expected[`update`] = swarmdb.QueryOption{
		Type:  "Update",
		Table: "contacts",
		Update: map[string]interface{}{
			"age":  float64(8),
			"name": "Bertie B",
		},
		Where:     swarmdb.Where{Left: "email", Right: "bertie@gmail.com", Operator: "="},
		Ascending: 1,
	}
	expected[`delete`] = swarmdb.QueryOption{
		Type:      "Delete",
		Table:     "contacts",
		Where:     swarmdb.Where{Left: "age", Right: "25", Operator: ">="},
		Ascending: 1,
	}

	var fail []string
	for testid, raw := range rawqueries {

		clean, err := swarmdb.ParseQuery(raw)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(clean, expected[testid]) {
			fmt.Printf("\n[%s] raw: %s\n", testid, raw)
			fmt.Printf("clean: %+v\n", clean)
			fmt.Printf("expected: %+v\n\n", expected[testid])
			fail = append(fail, testid)
		}

	}
	if len(fail) > 0 {
		t.Fatal(fmt.Errorf("tests [%s] failed", strings.Join(fail, ",")))
	}

}

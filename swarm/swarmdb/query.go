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

package swarmdb

import (
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	//sdbc "swarmdbcommon"
	sdbc "github.com/ethereum/go-ethereum/swarm/swarmdb/swarmdbcommon"
	"github.com/xwb1989/sqlparser"
	"strconv"
)

//at the moment, only parses a query with a single un-nested where clause, i.e.
//'Select name, age from contacts where email = "rodney@wolk.com"'
//TODO: nested where clauses
func ParseQuery(rawQuery string) (query QueryOption, err error) {
	stmt, err := sqlparser.Parse(rawQuery)
	if err != nil {
		return query, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:ParseQuery] Parse [%v]", err), ErrorCode: 401, ErrorMessage: fmt.Sprintf("SQL Parsing error: [%s]", err.Error())}
	}

	switch stmt := stmt.(type) {
	case *sqlparser.Select:
		//buf := sqlparser.NewTrackedBuffer(nil)
		//stmt.Format(buf)
		//fmt.Printf("select: %v\n", buf.String())

		query.Type = "Select"
		for _, column := range stmt.SelectExprs {
			//fmt.Printf("select %d: %+v\n", i, sqlparser.String(column)) // stmt.(*sqlparser.Select).SelectExprs)
			var newcolumn sdbc.Column
			newcolumn.ColumnName = sqlparser.String(column)
			//TODO: do we need to get IndexType, ColumnType, Primary from table itself...(not here?)
			query.RequestColumns = append(query.RequestColumns, newcolumn)
		}

		//From
		//fmt.Printf("from 0: %+v \n", sqlparser.String(stmt.From[0]))
		if len(stmt.From) == 0 {
			return query, &sdbc.SWARMDBError{Message: "Invalid SQL - Missing FROM", ErrorCode: 401, ErrorMessage: "SQL Parsing Error:[Missing FROM]"}
		}
		query.Table = sqlparser.String(stmt.From[0])

		//Where & Having
		//fmt.Printf("where or having: %s \n", readable(stmt.Where.Expr))
		if stmt.Where == nil {
			log.Debug("NOT SUPPORTING SELECT WITH NO WHERE")
			return query, &sdbc.SWARMDBError{Message: fmt.Sprintf("[query:ParseQuery] WHERE missing on Update query"), ErrorCode: 444, ErrorMessage: "SELECT & UPDATE query must have WHERE"}
		}
		if stmt.Where.Type == sqlparser.WhereStr { //Where
			//fmt.Printf("type: %s\n", stmt.Where.Type)
			query.Where, err = parseWhere(stmt.Where.Expr)
			//this is where recursion for nested parentheses should take place
			if err != nil {
				return query, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:ParseQuery] parseWhere [%s]", rawQuery))
			}
		} else if stmt.Where.Type == sqlparser.HavingStr { //Having
			fmt.Printf("type: %s\n", stmt.Where.Type)
			//TODO: fill in having
			return query, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:ParseQuery] Parse Having Clause Not currently supported"), ErrorCode: 401, ErrorMessage: fmt.Sprintf("SQL Parsing error: [HAVING clause not currently supported]", err.Error())}
		}

		//TODO: GroupBy ([]Expr)
		//for _, g := range stmt.GroupBy {
		//	fmt.Printf("groupby: %s \n", readable(g))
		//}

		//TODO: OrderBy
		query.Ascending = 1 //default if nothing?

		//Limit
		return query, nil

	/* Other options inside Select:
	   type Select struct {
	   	Cache       string
	   	Comments    Comments
	   	Distinct    string
	   	Hints       string
	   	SelectExprs SelectExprs
	   	From        TableExprs
	   	Where       *Where
	   	GroupBy     GroupBy
	   	Having      *Where
	   	OrderBy     OrderBy
	   	Limit       *Limit
	   	Lock        string
	   }*/

	case *sqlparser.Insert:
		//for now, 1 row to insert only. still need to figure out multiple rows
		//i.e. INSERT INTO MyTable (id, name) VALUES (1, 'Bob'), (2, 'Peter'), (3, 'Joe')

		query.Type = "Insert"
		query.Ascending = 1 //default
		//fmt.Printf("Action: %s \n", stmt.Action)
		//fmt.Printf("Comments: %+v \n", stmt.Comments)
		//fmt.Printf("Ignore: %s \n", stmt.Ignore)
		query.Table = sqlparser.String(stmt.Table.Name)
		if len(stmt.Rows.(sqlparser.Values)) == 0 {
			return query, &sdbc.SWARMDBError{Message: fmt.Sprintf("[query:ParseQuery] Insert has no values found"), ErrorCode: 401, ErrorMessage: "SQL Parsing error: [INSERT query missing VALUES]"}
		}
		if len(stmt.Rows.(sqlparser.Values)[0]) != len(stmt.Columns) {
			return query, &sdbc.SWARMDBError{Message: fmt.Sprintf("[query:ParseQuery] Insert has mismatch # of cols & vals"), ErrorCode: 401, ErrorMessage: "SQL Parsing error: [Mismatch in number of columns and values]"}
		}
		insertCells := make(map[string]interface{})
		for i, c := range stmt.Columns {
			col := sqlparser.String(c)
			if _, ok := insertCells[col]; ok {
				return query, &sdbc.SWARMDBError{Message: fmt.Sprintf("[query:ParseQuery] Insert can't have duplicate col %s", col), ErrorCode: 401, ErrorMessage: "SQL Parsing error: [INSERT cannot have duplicate columns]"}
			}
			//only detects string and float. how to do int? does it matter
			value := sqlparser.String(stmt.Rows.(sqlparser.Values)[0][i])
			if isQuoted(value) {
				insertCells[col] = trimQuotes(value)
			} else if isNumeric(value) {
				insertCells[col], err = strconv.ParseFloat(value, 64)
				if err != nil {
					return query, &sdbc.SWARMDBError{Message: fmt.Sprintf("[query:ParseQuery] Insert can't have duplicate col %s", col), ErrorCode: 401, ErrorMessage: "SQL Parsing error: [INSERT cannot have duplicate columns]"}
				}
			} else {
				return query, &sdbc.SWARMDBError{Message: fmt.Sprintf("[query:ParseQuery] Insert value %s has unknown type", value), ErrorCode: 401, ErrorMessage: "SQL Parsing error: [Invalid value type passed in.]"}
				//TODO: more clear Message
			}
			//insertCells[col] = trimQuotes(sqlparser.String(stmt.Rows.(sqlparser.Values)[0][i]))
		}
		r := sdbc.NewRow()
		r = insertCells
		query.Inserts = append(query.Inserts, r)
		//fmt.Printf("OnDup: %+v\n", stmt.OnDup)
		//fmt.Printf("Rows: %+v\n", stmt.Rows.(sqlparser.Values))
		//fmt.Printf("Rows: %+v\n", sqlparser.String(stmt.Rows.(sqlparser.Values)))
		//for i, v := range stmt.Rows.(sqlparser.Values)[0] {
		//	fmt.Printf("row: %v %+v\n", i, sqlparser.String(v))
		//}

	case *sqlparser.Update:

		query.Type = "Update"
		//fmt.Printf("Comments: %+v \n", stmt.Comments)
		query.Table = sqlparser.String(stmt.TableExprs[0])
		query.Update = make(map[string]interface{})
		for _, expr := range stmt.Exprs {
			col := sqlparser.String(expr.Name)
			//fmt.Printf("col: %+v\n", col)
			if _, ok := query.Update[col]; ok {
				return query, &sdbc.SWARMDBError{Message: fmt.Sprintf("[query:ParseQuery] Update can't have duplicate col %s", col), ErrorCode: 401, ErrorMessage: "SQL Parsing error: [INSERT cannot have duplicate columns]"}
			}
			value := readable(expr.Expr)
			if isQuoted(value) {
				query.Update[col] = trimQuotes(value)
			} else if isNumeric(value) {
				query.Update[col], err = strconv.ParseFloat(value, 64)
				if err != nil {
					return query, &sdbc.SWARMDBError{Message: fmt.Sprintf("[query:ParseQuery] ParseFloat %s", err.Error()), ErrorCode: 401, ErrorMessage: "SQL Parsing error: [Float Value could not be parsed]"}
				}
			} else {
				return query, &sdbc.SWARMDBError{Message: fmt.Sprintf("[query:ParseQuery] Update value %s has unknown type", value), ErrorCode: 401, ErrorMessage: "SQL Parsing error: [Invalid value type passed in.]"}
			}
			//fmt.Printf("val: %v \n", query.Update[col])
		}

		// Where
		log.Debug(fmt.Sprintf("Statement: [%+v] | SqlParser: [%+v]", stmt, sqlparser.WhereStr))
		if stmt.Where == nil {
			log.Debug("NOT SUPPORTING UPDATES WITH NO WHERE")
			return query, &sdbc.SWARMDBError{Message: fmt.Sprintf("[query:ParseQuery] WHERE missing on Update query"), ErrorCode: 444, ErrorMessage: "UPDATE query must have WHERE"}
		}
		if stmt.Where.Type == sqlparser.WhereStr {
			query.Where, err = parseWhere(stmt.Where.Expr)
			//TODO: this is where recursion for nested parentheses should probably take place
			if err != nil {
				return query, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[query:ParseQuery] parseWhere %s", err.Error()))
			}
			//fmt.Printf("Where: %+v\n", query.Where)
		}

		//TODO: OrderBy
		query.Ascending = 1 //default if nothing?

		//Limit
		//fmt.Printf("Limit: %v \n", stmt.Limit)
		return query, nil
	case *sqlparser.Delete:
		query.Type = "Delete"
		if len(stmt.TableExprs) == 0 {
			return query, &sdbc.SWARMDBError{Message: fmt.Sprintf("[query:ParseQuery] DELETE TableExprs empty"), ErrorCode: 401, ErrorMessage: "SQL Parsing Error: [DELETE missing Table]"}
		}
		query.Table = sqlparser.String(stmt.TableExprs[0]) // TODO: an OK around the array in case of panic
		//fmt.Printf("Comments: %+v \n", stmt.Comments)

		//Targets
		for _, t := range stmt.Targets {
			fmt.Printf("Targets: %s\n", t.Name)
		}

		//Where
		if stmt.Where == nil {
			log.Debug("NOT SUPPORTING DELETES WITH NO WHERE")
			return query, &sdbc.SWARMDBError{Message: fmt.Sprintf("[query:ParseQuery] WHERE missing on Delete query"), ErrorCode: 444, ErrorMessage: "DELETE query must have WHERE"}
		}
		if stmt.Where.Type == sqlparser.WhereStr { //Where
			query.Where, err = parseWhere(stmt.Where.Expr)
			//TODO: this is where recursion for nested parentheses should take place
			if err != nil {
				return query, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[query:ParseQuery] parseWhere %s", err.Error()))
			}
			//fmt.Printf("Where: %+v\n", query.Where)
		}

		//TODO: OrderBy
		query.Ascending = 1 //default if nothing?

		//Limit
		//fmt.Printf("Limit: %v \n", stmt.Limit)

		return query, nil

		/* Other Options for type of Query:
		   func (*Union) iStatement()      {}
		   func (*Select) iStatement()     {}
		   func (*Insert) iStatement()     {}
		   func (*Update) iStatement()     {}
		   func (*Delete) iStatement()     {}
		   func (*Set) iStatement()        {}
		   func (*DDL) iStatement()        {}
		   func (*Show) iStatement()       {}
		   func (*Use) iStatement()        {}
		   func (*OtherRead) iStatement()  {}
		   func (*OtherAdmin) iStatement() {}
		*/

	}

	return query, err
}

func parseWhere(expr sqlparser.Expr) (where Where, err error) {

	switch expr := expr.(type) {
	case *sqlparser.OrExpr:
		where.Left = readable(expr.Left)
		where.Right = readable(expr.Right)
		where.Operator = "OR" //should be const
	case *sqlparser.AndExpr:
		where.Left = readable(expr.Left)
		where.Right = readable(expr.Right)
		where.Operator = "AND" //shoud be const
	case *sqlparser.IsExpr:
		where.Right = readable(expr.Expr)
		where.Operator = expr.Operator
	case *sqlparser.BinaryExpr:
		where.Left = readable(expr.Left)
		where.Right = readable(expr.Right)
		where.Operator = expr.Operator
	case *sqlparser.ComparisonExpr:
		where.Left = readable(expr.Left)
		where.Right = readable(expr.Right)
		where.Operator = expr.Operator
	default:
		return where, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:parseWhere] exp Type [%s] not supported", expr), ErrorCode: 401, ErrorMessage: fmt.Sprintf("SQL Parsing error: [Expression Type (%s) not currently supported]", expr)}
	}
	where.Right = trimQuotes(where.Right)

	return where, err
}

func trimQuotes(s string) string {
	if len(s) > 0 && s[0] == '\'' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == '\'' {
		s = s[:len(s)-1]
	}
	return s
}

func isQuoted(s string) bool { //string
	if (len(s) > 0) && (s[0] == '\'') && (s[len(s)-1] == '\'') {
		return true
	}
	return false
}

func isNumeric(s string) bool { //float or int
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return true
	}
	return false
}

func readable(expr sqlparser.Expr) string {
	switch expr := expr.(type) {
	case *sqlparser.OrExpr:
		return fmt.Sprintf("(%s or %s)", readable(expr.Left), readable(expr.Right))
	case *sqlparser.AndExpr:
		return fmt.Sprintf("(%s and %s)", readable(expr.Left), readable(expr.Right))
	case *sqlparser.BinaryExpr:
		return fmt.Sprintf("(%s %s %s)", readable(expr.Left), expr.Operator, readable(expr.Right))
	case *sqlparser.IsExpr:
		return fmt.Sprintf("(%s %s)", readable(expr.Expr), expr.Operator)
	case *sqlparser.ComparisonExpr:
		return fmt.Sprintf("(%s %s %s)", readable(expr.Left), expr.Operator, readable(expr.Right))
	default:
		return sqlparser.String(expr)
	}
}

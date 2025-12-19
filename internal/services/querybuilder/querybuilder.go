package querybuilder

import (
	"database/sql"
	"strings"
)

type Mapping struct {
	Key   string
	Value string
}

// -----------------------------------------------------------------------------
// BaseQueryBuilder
// -----------------------------------------------------------------------------

type BaseQueryBuilder struct {
	query    string
	mappings []Mapping
}

func (bq *BaseQueryBuilder) Query() string {
	return bq.query
}

func (bq *BaseQueryBuilder) Mappings() []Mapping {
	return bq.mappings
}

func (bq *BaseQueryBuilder) NamedArgs() []sql.NamedArg {
	values := []sql.NamedArg{}
	for _, m := range bq.mappings {
		values = append(values, sql.Named(m.Key, m.Value))
	}
	return values
}

func (bq *BaseQueryBuilder) Clear() *BaseQueryBuilder {
	bq.mappings = []Mapping{}
	bq.query = ""
	return bq
}

func (bq *BaseQueryBuilder) Select(tableName, customColumns ...string) *BaseQueryBuilder {
	if len(customColumns) > 0 {
		bq.query = "SELECT " + strings.Join(customColumns, ", ") + " FROM "+ tableName + " "
	} else {
		bq.query = "SELECT * FROM " + tableName + " "
	}
	return bq
}

func (bq *BaseQueryBuilder) Where(mappings ...Mapping) *BaseQueryBuilder {
	bq.mappings = append(bq.mappings, mappings...)
	bq.query += "WHERE "
	for i, m := range mappings {
		bq.query += m.Key + " = :" + m.Key
		if i < len(mappings)-1 {
			bq.query += " AND "
		}
	}
	return bq
}

// -----------------------------------------------------------------------------
// PressCyclesQueryBuilder
// -----------------------------------------------------------------------------

// TODO: Main package for building SQL queries for various services

type PressCyclesQueryBuilder struct {
	BaseQueryBuilder
}

func (pc *PressCyclesQueryBuilder) Select(customColumns ...string) *PressCyclesQueryBuilder {
	pc.BaseQueryBuilder.Select("press_cycles", customColumns...)
	return pc
}

func (pc *PressCyclesQueryBuilder) Where(mappings ...Mapping) *PressCyclesQueryBuilder {
	pc.BaseQueryBuilder.Where(mappings...)
	return pc
}

func (pc *PressCyclesQueryBuilder) OrderBy() *PressCyclesQueryBuilder {
	pc.query = "ORDER BY press_number ASC, stop DESC "
	return pc
}

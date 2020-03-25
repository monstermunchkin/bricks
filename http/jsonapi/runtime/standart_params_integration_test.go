// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/02/06 by Charlotte Pröller

package runtime_test

import (
	"context"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/pace/bricks/backend/postgres"
	"github.com/pace/bricks/http/jsonapi/runtime"
	"github.com/pace/bricks/maintenance/log"
	"github.com/stretchr/testify/assert"
)

type TestModel struct {
	FilterName string
}

type testValueSanitizer struct {
}

func (t *testValueSanitizer) SanitizeValue(fieldName string, value string) (interface{}, error) {
	return value, nil
}

func TestIntegrationFilterParameter(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	// Setup
	a := assert.New(t)
	db := setupDatabase(a)
	mappingNames := map[string]string{
		"test": "filter_name",
	}
	// filter
	r := httptest.NewRequest("GET", "http://abc.de/whatEver?filter[test]=b", nil)
	filterFunc, err := runtime.FilterFromRequest(r, mappingNames, &testValueSanitizer{})
	a.NoError(err)
	var modelsFilter []TestModel
	q := db.Model(&modelsFilter)
	q, err = filterFunc(q)
	a.NoError(err)
	count, _ := q.SelectAndCount()
	a.Equal(1, count)
	a.Equal("b", modelsFilter[0].FilterName)

	r = httptest.NewRequest("GET", "http://abc.de/whatEver?filter[test]=a,b", nil)
	filterFunc, err = runtime.FilterFromRequest(r, mappingNames, &testValueSanitizer{})
	a.NoError(err)
	var modelsFilter2 []TestModel
	q = db.Model(&modelsFilter2)
	q, err = filterFunc(q)
	a.NoError(err)
	count, _ = q.SelectAndCount()
	a.Equal(2, count)
	sort.Slice(modelsFilter2, func(i, j int) bool {
		return modelsFilter2[i].FilterName < modelsFilter2[j].FilterName
	})
	a.Equal("a", modelsFilter2[0].FilterName)
	a.Equal("b", modelsFilter2[1].FilterName)

	// Paging
	r = httptest.NewRequest("GET", "http://abc.de/whatEver?page[number]=0&page[size]=1", nil)
	pagingFunc, err := runtime.PagingFromRequest(r)
	var modelsPaging []TestModel
	q = db.Model(&modelsPaging)
	q, err = pagingFunc(q)
	a.NoError(err)
	err = q.Select()
	a.NoError(err)
	a.Equal(1, len(modelsPaging))

	// Sorting
	r = httptest.NewRequest("GET", "http://abc.de/whatEver?sort=-test", nil)
	sortingFunc, err := runtime.SortingFromRequest(r, mappingNames)
	var modelsSort []TestModel
	q = db.Model(&modelsSort)
	q, err = sortingFunc(q)
	a.NoError(err)
	err = q.Select()
	a.NoError(err)
	a.Equal(3, len(modelsSort))
	a.Equal("c", modelsSort[0].FilterName)
	a.Equal("b", modelsSort[1].FilterName)
	a.Equal("a", modelsSort[2].FilterName)

	// Tear Down
	db.DropTable(&TestModel{}, &orm.DropTableOptions{
		IfExists: true,
	})
}

func setupDatabase(a *assert.Assertions) *pg.DB {
	dB := postgres.DefaultConnectionPool()
	dB = dB.WithContext(log.WithContext(context.Background()))

	a.NoError(dB.CreateTable(&TestModel{}, &orm.CreateTableOptions{IfNotExists: true}))
	_, err := dB.Model(&TestModel{
		FilterName: "a",
	}).Insert()
	a.NoError(err)
	_, err = dB.Model(&TestModel{
		FilterName: "b",
	}).Insert()
	a.NoError(err)
	_, err = dB.Model(&TestModel{
		FilterName: "c",
	}).Insert()
	a.NoError(err)
	return dB
}

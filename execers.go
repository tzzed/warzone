package warzone

import (
	"fmt"

	"github.com/genjidb/genji"
)

// ExecerFunc allows to execute any SQL statement against the given DB.
type ExecerFunc func(*genji.DB) error

const (
	tableName = "warzone"

	insertAllTypes = `INSERT INTO ` + tableName + ` VALUES {
		i: 10,
		dbl: 10.10,
		b: true,
		t: "hello",
		arr: [1, "true", true],
		doc: {"foo": "bar"},
		du: 127ns,
		bb: "YmxvYlZhbHVlCg==",
		byt: "Ynl0ZXNWYWx1ZQ=="
	}`
)

// InsertAllTypes insert all supported types.
func InsertAllTypes(db *genji.DB) (ExecerFunc, func() error) {
	err := db.Exec("CREATE TABLE IF NOT EXISTS " + tableName)
	if err != nil {
		// We shouldn't expect any error while creating a new table.
		// If there is an error, we have a bigger problem so let's panic!
		panic(err)
	}

	return func(db *genji.DB) error {
		return db.Exec(insertAllTypes)
	}, nil
}

// InsertAllTypesWithTx insert all supported types within a transaction.
func InsertAllTypesWithTx(db *genji.DB) (ExecerFunc, func() error) {
	err := db.Exec("CREATE TABLE IF NOT EXISTS " + tableName)
	if err != nil {
		// We shouldn't expect any error while creating a new table.
		// If there is an error, we have a bigger problem so let's panic!
		panic(err)
	}

	tx, err := db.Begin(true)
	if err != nil {
		// If there is an error while creating a transaction, let's panic!
		panic(err)
	}

	fn := func() error {
		if r := recover(); r != nil {
			fmt.Println("recovered panic:", r)
			err := tx.Rollback()
			if err != nil {
				return err
			}
		}
		return tx.Commit()
	}

	return func(_ *genji.DB) error {
		err := tx.Exec(insertAllTypes)
		if err != nil {
			err1 := tx.Rollback()
			if err1 != nil {
				return fmt.Errorf("cannot rollback: %v, base error: %v", err1, err)
			}
			return err
		}
		return nil
	}, fn
}

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/genjidb/warzone"

	"github.com/dgraph-io/badger/v2"
	"github.com/genjidb/genji"
	"github.com/genjidb/genji/engine"
	"github.com/genjidb/genji/engine/badgerengine"
	"github.com/genjidb/genji/engine/boltengine"
	"github.com/genjidb/genji/engine/memoryengine"
)

var scenarios = map[string]func(*genji.DB) (warzone.ExecerFunc, func() error){
	"insert-all-types":         warzone.InsertAllTypes,
	"insert-all-types-with-tx": warzone.InsertAllTypesWithTx,
}

func main() {
	var (
		scenario, dbname, engine string

		n, freq int
		rm      bool
		db      *genji.DB
	)
	flag.IntVar(&n, "n", 100, "number of records to insert")
	flag.IntVar(&freq, "f", 10, "frequency of printing duration")
	flag.StringVar(&scenario, "scenario", "", "scenario to run [required]")
	flag.StringVar(&engine, "engine", "bolt", "engine to use [bolt, badger, memory]")
	flag.StringVar(&dbname, "dbname", "", "name of the db")
	flag.BoolVar(&rm, "rm", false, "remove the database")
	flag.Parse()

	if scenario == "" || (engine != "bolt" && engine != "badger" && engine != "memory") {
		flag.Usage()
		os.Exit(1)
	}

	if _, ok := scenarios[scenario]; !ok {
		fmt.Printf("unknown test: %s\n", scenario)
		fmt.Println("\nAvailable tests are:")
		for k := range scenarios {
			fmt.Println("-", k)
		}
		os.Exit(1)
	}

	// If dbname flag is empty, the DB file is named by the scenario.
	if dbname == "" {
		dbname = fmt.Sprintf("%s.db", scenario)
	}

	db, err := newEngine(engine, dbname)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	defer func() {
		// If rm flag is true, remove the DB file.
		if rm {
			if err := os.RemoveAll(dbname); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
	}()

	ef, fn := scenarios[scenario](db)
	defer func() {
		if fn != nil {
			err := fn()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}

		db.Close()
	}()

	if err := run(db, ef, n, freq); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(db *genji.DB, fn warzone.ExecerFunc, n, freq int) error {
	fmt.Println("count,duration")

	for i := 1; i <= n; i++ {
		s := time.Now()
		err := fn(db)
		if err != nil {
			return err
		}
		elapsed := time.Since(s)

		if i == 1 || i%freq == 0 || i == n {
			fmt.Printf("%d,%.2f\n", i, float32(elapsed)/1000000)
		}
	}

	return nil
}

func newEngine(e, n string) (*genji.DB, error) {
	var ng engine.Engine
	var err error

	switch e {
	case "memory":
		ng = memoryengine.NewEngine()
	case "bolt":
		ng, err = boltengine.NewEngine(n, 0660, nil)
		if err != nil {
			return nil, err
		}
	case "badger":
		ng, err = badgerengine.NewEngine(badger.DefaultOptions(n))
		if err != nil {
			return nil, err
		}
	}

	return genji.New(ng)
}

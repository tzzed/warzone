package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/genjidb/warzone"
	"github.com/hashicorp/go-multierror"

	"github.com/dgraph-io/badger/v2"
	"github.com/genjidb/genji"
	"github.com/genjidb/genji/engine"
	"github.com/genjidb/genji/engine/badgerengine"
	"github.com/genjidb/genji/engine/boltengine"
	"github.com/genjidb/genji/engine/memoryengine"
)

var scenarios = map[string]func(*genji.DB) (warzone.ScenarioFunc, func(error) error){
	"insert-all-types":         warzone.InsertAllTypes,
	"insert-all-types-with-tx": warzone.InsertAllTypesWithTx,
}

func main() {
	if err := main1(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main1() error {
	var (
		scenario, dbname, engine string

		n, freq int
		rm      bool
	)
	flag.IntVar(&n, "n", 100, "number of records to insert")
	flag.IntVar(&freq, "f", 10, "number of actions to skip between each duration printing")
	flag.StringVar(&scenario, "scenario", "", "scenario to run [required]")
	flag.StringVar(&engine, "engine", "bolt", "engine to use [bolt, badger, memory]")
	flag.StringVar(&dbname, "dbname", "", "name of the database")
	flag.BoolVar(&rm, "rm", false, "remove the database")
	flag.Parse()

	_, found := scenarios[scenario]
	switch {
	case scenario == "":
		return fmt.Errorf("flag -scenario is required")

	case !found:
		fmt.Println("Available scenarios are:")
		for k := range scenarios {
			fmt.Println("-", k)
		}
		return fmt.Errorf("unknown scenario: %v", scenario)

	case engine != "bolt" && engine != "badger" && engine != "memory":
		return fmt.Errorf("unsupported engine: %v", engine)
	}

	return run(engine, dbname, scenario, rm, n, freq)
}

func run(engine, dbname, scenario string, rm bool, n, freq int) (errs error) {
	// If dbname flag is empty, the DB file is named after the scenario.
	if dbname == "" {
		dbname = fmt.Sprintf("%s.db", scenario)
	}

	db, err := newEngine(engine, dbname)
	if err != nil {
		errs = multierror.Append(errs, err)
	}

	defer func() {
		// If rm flag is true, remove the DB file.
		if rm {
			if err := os.RemoveAll(dbname); err != nil {
				errs = multierror.Append(errs, err)
			}
		}
	}()

	ef, teardown := scenarios[scenario](db)
	defer func() {
		if teardown != nil {
			if err1 := teardown(err); err1 != nil {
				errs = multierror.Append(errs, err1)
			}
		}
		db.Close()
	}()

	defer func() {
		if r := recover(); r != nil {
			errs = multierror.Append(errs, fmt.Errorf("scenario panicked: %v", r))
		}
	}()

	err = runScenario(db, ef, n, freq)
	if err != nil {
		errs = multierror.Append(errs, err)
	}

	// errs holds all potential errors
	return errs
}

func runScenario(db *genji.DB, fn warzone.ScenarioFunc, n, freq int) error {
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

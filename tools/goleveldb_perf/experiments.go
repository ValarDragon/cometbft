package db_experiments

import (
	"context"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/docker/go-units"
)

func inserts(backendType dbm.BackendType, keySize int, valueSize int, dbPath string, ctx context.Context) []Step {
	targetStorageSize := 10 * units.GiB

	db, err := dbm.NewDB("experiment_db", backendType, dbPath)
	if err != nil {
		panic(err)
	}
	defer func(db dbm.DB) {
		err := db.Close()
		if err != nil {
			panic(err)
		}
	}(db)

	var steps []Step
	steps = append(steps, Step{
		Name:     "initial",
		Size:     dirSize(dbPath),
		Records:  0,
		Duration: 0,
		SysMem:   getSysMem(),
	})

	currentStorageSize := 0
	currentRecordings := 0
	recordingsPerStep := 1 * units.GiB / (keySize + valueSize)
	for currentStorageSize < targetStorageSize {
		select {
		case <-ctx.Done():
			return steps
		default:
			newStep := step("insert", recordingsPerStep, db, keySize, valueSize, dbPath, ctx)
			currentRecordings += recordingsPerStep
			newStep.Records = currentRecordings
			steps = append(steps, newStep)
			currentStorageSize += (keySize + valueSize) * recordingsPerStep

			// make sure process is not killed due to the overuse of memory
			if getSysMem()*units.MiB > 5*units.GiB {
				steps = append(steps, Step{Name: "memOveruse"})
				return steps
			}
		}
	}

	return steps
}

func deletions(backendType dbm.BackendType, keySize int, valueSize int, dbPath string, ctx context.Context) []Step {
	initialStorageSize := 10 * units.GiB

	db, err := dbm.NewDB("experiment_db", backendType, dbPath)
	if err != nil {
		panic(err)
	}
	defer func(db dbm.DB) {
		err := db.Close()
		if err != nil {
			panic(err)
		}
	}(db)

	var steps []Step
	initialRecordings := initialStorageSize / (keySize + valueSize)
	initialStep := step("insert", initialRecordings, db, keySize, valueSize, dbPath, ctx)
	initialStep.Records = initialRecordings
	steps = append(steps, initialStep)

	recordingsPerStep := 1 * units.GiB / (keySize + valueSize)
	currentStorageSize := (keySize + valueSize) * initialRecordings
	currentRecordings := initialRecordings
	for currentStorageSize > 0 {
		select {
		case <-ctx.Done():
			return steps
		default:
			newStep := step("delete", recordingsPerStep, db, keySize, valueSize, dbPath, ctx)
			currentRecordings -= recordingsPerStep
			newStep.Records = currentRecordings
			steps = append(steps, newStep)
			currentStorageSize -= (keySize + valueSize) * recordingsPerStep
		}
	}

	return steps
}

func batchInserts(backendType dbm.BackendType, keySize int, valueSize int, dbPath string, ctx context.Context) []Step {
	targetStorageSize := 10 * units.GiB

	db, err := dbm.NewDB("experiment_db", backendType, dbPath)
	if err != nil {
		panic(err)
	}
	defer func(db dbm.DB) {
		err := db.Close()
		if err != nil {
			panic(err)
		}
	}(db)

	var steps []Step
	steps = append(steps, Step{
		Name:     "initial",
		Size:     dirSize(dbPath),
		Records:  0,
		Duration: 0,
		SysMem:   getSysMem(),
	})

	currentStorageSize := 0
	currentRecords := 0
	recordingsPerStep := 1 * units.GiB / (keySize + valueSize)
	for currentStorageSize < targetStorageSize {
		select {
		case <-ctx.Done():
			return steps
		default:
			newStep := step("batchInsert", recordingsPerStep, db, keySize, valueSize, dbPath, ctx)
			currentRecords += recordingsPerStep
			newStep.Records = currentRecords
			steps = append(steps, newStep)
			currentStorageSize += (keySize + valueSize) * recordingsPerStep

			// make sure process is not killed due to the overuse of memory
			if getSysMem()*units.MiB > 5*units.GiB {
				steps = append(steps, Step{Name: "memOveruse"})
				return steps
			}
		}
	}

	return steps
}

func batchDeletions(backendType dbm.BackendType, keySize int, valueSize int, dbPath string, ctx context.Context) []Step {
	initialStorageSize := 10 * units.GiB

	db, err := dbm.NewDB("experiment_db", backendType, dbPath)
	if err != nil {
		panic(err)
	}
	defer func(db dbm.DB) {
		err := db.Close()
		if err != nil {
			panic(err)
		}
	}(db)

	var steps []Step
	initialRecords := initialStorageSize / (keySize + valueSize)
	initialStep := step("insert", initialRecords, db, keySize, valueSize, dbPath, ctx)
	initialStep.Records = initialRecords
	steps = append(steps, initialStep)

	recordingsPerStep := 1 * units.GiB / (keySize + valueSize)
	currentStorageSize := (keySize + valueSize) * initialRecords
	currentRecords := initialRecords
	for currentStorageSize > 0 {
		select {
		case <-ctx.Done():
			return steps
		default:
			newStep := step("batchDelete", recordingsPerStep, db, keySize, valueSize, dbPath, ctx)
			currentRecords -= recordingsPerStep
			newStep.Records = currentRecords
			steps = append(steps, newStep)
			currentStorageSize -= (keySize + valueSize) * recordingsPerStep
		}
	}

	return steps
}

func fluctuations(backendType dbm.BackendType, keySize int, valueSize int, dbPath string, ctx context.Context) []Step {
	initialStorageSize := 5 * units.GiB

	db, err := dbm.NewDB("experiment_db", backendType, dbPath)
	if err != nil {
		panic(err)
	}
	defer func(db dbm.DB) {
		err := db.Close()
		if err != nil {
			panic(err)
		}
	}(db)

	var steps []Step
	initialRecordings := initialStorageSize / (keySize + valueSize)
	initialStep := step("insert", initialRecordings, db, keySize, valueSize, dbPath, ctx)
	initialStep.Records = initialRecordings
	steps = append(steps, initialStep)

	nFluctuations := 10
	recordingsPerStep := 1 * units.GiB / (keySize + valueSize)
	currentRecords := initialRecordings
	for i := 0; i < nFluctuations; i++ {
		select {
		case <-ctx.Done():
			return steps
		default:
			if i%2 == 0 {
				newStep := step("delete", recordingsPerStep, db, keySize, valueSize, dbPath, ctx)
				currentRecords -= recordingsPerStep
				newStep.Records = currentRecords
				steps = append(steps, newStep)
			} else {
				newStep := step("insert", recordingsPerStep, db, keySize, valueSize, dbPath, ctx)
				currentRecords += recordingsPerStep
				newStep.Records = currentRecords
				steps = append(steps, newStep)
			}
		}
	}
	return steps
}

func batchedFluctuations() {

}

func fluctuationsForceCompact() {

}

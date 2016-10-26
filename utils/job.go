package utils

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

func AddJobs(count int, jobChan chan struct{}) {
	for i := 0; i < count; i++ {
		jobChan <- struct{}{}
	}
	close(jobChan)
}

func Waiting(doneChan chan struct{}, start time.Time, jobCount int, workerCount int) {
	for i := 0; i < workerCount; i++ {
		<-doneChan
	}
	close(doneChan)
	now := time.Now()
	seconds := now.Unix() - start.Unix()
	tps := int64(-1)
	if seconds > 0 {
		tps = int64(jobCount) / seconds
	}
	fmt.Printf("total %d cases, cost %d seconds, tps %d, start %s, now %s\n", jobCount, seconds, tps, start, now)
}

func HandleJob(db *sql.DB, data []string, batch int, jobChan chan struct{}, doneChan chan struct{}) {
	temp := 0
	count := 0
	for range jobChan {
		temp++
		if temp == batch {
			doExec(db, data[count-batch+1:count+1])
			temp = 0
		}
		count++
	}
	if temp > 0 {
		temp = 0
		doExec(db, data[count:len(data)])
	}
	doneChan <- struct{}{}
}

func doExec(db *sql.DB, data []string) {
	txn, err := db.Begin()
	if err != nil {
		log.Fatalf("Transaction bengin Error: %s", err)
	}
	for _, sql := range data {
		_, err := txn.Exec(sql)
		if err != nil {
			log.Fatalf("Exec sql Error: %s", err)
		}
	}
	err = txn.Commit()
	if err != nil {
		log.Fatalf("Transcation commit Error: %s", err)
	}
}
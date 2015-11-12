package main

import (
	"fmt"
	"math/rand"
	"time"
)

func PrintProgress(completed_files, statuses, finished chan string) {
	last_status := ""
	for {
		last_len := 0
		select {
		case completed_file := <-completed_files:
			fmt.Printf("\r")
			fmt.Print("completed: " + completed_file + "                                                                                  \n" + last_status) // keep track of needed spaces
			//fmt.Println("completed: "+completed_file)
			//fmt.Println(last_status)
		case status := <-statuses:
			last_status = status
			fmt.Printf("\r")
			fmt.Print(status + "                    ")
		case elapsed := <-finished:
			fmt.Println("\n" + elapsed)
			time.Sleep(1 * time.Second)
			return
		}
	}
	//completed: /local/path/filename.ext
	//completed: /local/pathfilename.ext
	//transfered 15/46 files 3/14GB 12% 24MB/s 12m16s remaining
	//completed transfer in 14m2s
}

func main() {
	completed_files := make(chan string)
	statuses := make(chan string)
	finished := make(chan string)
	go PrintProgress(completed_files, statuses, finished)

	go func(statuses chan string) {
		for {
			files_done := rand.Intn(100)
			files := files_done + rand.Intn(50)
			gb_done := rand.Intn(100)
			gb := gb_done + rand.Intn(50)
			percent := rand.Intn(100)
			speed := rand.Intn(50)
			min := rand.Intn(10)
			sec := rand.Intn(59)
			statuses <- fmt.Sprintf(
				"transfered %d/%d files %d/%dGB %d%% %dMB/s %dm%ds remaining",
				files_done,
				files,
				gb_done,
				gb,
				percent,
				speed,
				min,
				sec,
			)
			time.Sleep(1 * time.Second)
		}
	}(statuses)

	time.Sleep(1 * time.Second)
	for i := 0; i < 10; i++ {
		completed_files <- fmt.Sprintf("file_%d.mkv", rand.Intn(100))
		time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
	}
	finished <- "completed in 3m12s"
	time.Sleep(1 * time.Second)
}

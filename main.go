package main

import (
	"archive/zip"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

func main() {
	backup, err := ReadAndUnmarshal("dirs.yaml")
	if err != nil {
		log.Fatal(err)
	}
	err = backup.RunBackup()
	if err != nil {
		log.Fatal(err)
	}
}

type Backup struct {
	Destination string
	Jobs        []Job
}

type Job struct {
	Name string
	Dirs []string
}

func (bckp Backup) RunBackup() error {
	var workerCount = runtime.NumCPU()
	var wg sync.WaitGroup
	jobChan := make(chan Job)
	errChan := make(chan error, len(bckp.Jobs))
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			jobWorker(bckp.Destination, jobChan, errChan)
		}()
	}
	go func() {
		for _, job := range bckp.Jobs {
			jobChan <- job
		}
		close(jobChan)
	}()

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		if err != nil {
			return err
		}
	}
	log.Println("Finished backing up")
	return nil
}

func jobWorker(destination string, jobs <-chan Job, errChan chan<- error) {
	for job := range jobs {
		start := time.Now()
		log.Printf("Starting job: %s", job.Name)
		ZipWriter(destination, job, errChan)
		log.Printf("Finished job: %s; took %v", job.Name, time.Since(start))
	}
}

func ZipWriter(destination string, job Job, errChan chan<- error) {
	dateSuffix := time.Now().Format(time.DateOnly)
	outFile, err := os.Create(filepath.Join(destination, job.Name) + "_" + dateSuffix + ".zip")
	if err != nil {
		errChan <- err
	}
	defer outFile.Close()

	w := zip.NewWriter(outFile)
	defer w.Close()

	for _, jobDir := range job.Dirs {
		err = addFiles(w, jobDir)
		if err != nil {
			errChan <- err
			return
		}
	}
}

func addFiles(w *zip.Writer, basePath string) error {
	walker := func(path string, info os.DirEntry, err error) error {
		// log.Printf("Crawling: %#v\n", path)
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(basePath, path)
		zipPath := filepath.Join(filepath.Base(basePath), relPath)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		f, err := w.Create(zipPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(f, file)
		if err != nil {
			return err
		}

		return nil
	}
	err := filepath.WalkDir(basePath, walker)
	if err != nil {
		panic(err)
	}
	return nil
}

func ReadAndUnmarshal(file string) (*Backup, error) {
	var backup Backup
	yFile, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("error reading dirs.yaml")
	}
	err = yaml.Unmarshal(yFile, &backup)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling dir.yaml contents")
	}
	return &backup, nil
}

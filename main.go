package main

import (
	"archive/zip"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func main() {
	backup, err := ReadAndUnmarshal("dirs.yaml")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v\n", backup)
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
	var wg sync.WaitGroup
	errChan := make(chan error, len(bckp.Jobs))
	for _, job := range bckp.Jobs {
		wg.Add(1)
		go func(job Job) {
			log.Println("Starting job:", job.Name, "for Folder(s):", job.Dirs)
			bckp.ZipWriter(job, errChan)
			wg.Done()
		}(job)
	}
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

func (bckp Backup) ZipWriter(job Job, errChan chan error) {
	timeSuffix := time.Now().Format(time.DateOnly)
	outFile, err := os.Create(filepath.Join(bckp.Destination, job.Name) + "_" + timeSuffix + ".zip")
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
		}
	}
	log.Println("Finished job:", job.Name)
}

func addFiles(w *zip.Writer, basePath string) error {
	walker := func(path string, info os.FileInfo, err error) error {
		// log.Println("Crawling: %#v\n", path)
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(basePath, path)
		zipPath := filepath.Join(filepath.Base(basePath), relPath)
		if err != nil {
			return err
		}
		if info.IsDir() {
			//	path = fmt.Sprintf("%s%c", path, os.PathSeparator)
			//	_, err = w.Create(path)
			//	return err
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
	err := filepath.Walk(basePath, walker)
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

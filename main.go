package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"

	humanise "github.com/dustin/go-humanize"
)

var (
	minSizeStr = flag.String("min-size", "", "minimum size of a file to include in results, ie: 100MB, 2GB, etc (required)")
	dir        = flag.String("dir", "", "file or directory to get file sizes from (required)")
	workersNum = flag.Int("workers", runtime.GOMAXPROCS(0)*2, "number of concurrent workers, default to runtime.GOMAXPROCS * 2")

	minSize uint64
)

type filesize struct {
	filepath string
	size     uint64
}

type stats struct {
	filesizes []filesize
	fsChan    chan filesize
}

func newStats() *stats {
	s := &stats{
		fsChan: make(chan filesize),
	}
	go s.watch()
	return s
}

func (s *stats) watch() {
	// TODO(vishen): set a ctx timeout
	for {
		select {
		case fs := <-s.fsChan:
			if fs.size < minSize {
				continue
			}
			s.filesizes = append(s.filesizes, fs)
		}
	}
}

func (s *stats) printStats() {
	sort.Slice(s.filesizes, func(i, j int) bool {
		return s.filesizes[i].size > s.filesizes[j].size
	})
	for _, fs := range s.filesizes {
		size := humanise.Bytes(fs.size)
		fmt.Printf("%s -> %s\n", size, fs.filepath)
	}
}

type workers struct {
	dirCh chan string
	stats *stats
}

func newWorkers(stats *stats) *workers {
	w := &workers{
		dirCh: make(chan string),
		stats: stats,
	}
	return w
}

func (w *workers) start(dir string) {
	wg := sync.WaitGroup{}
	wrks := *workersNum
	for i := 0; i < wrks; i++ {
		go func() {
			for path := range w.dirCh {
				files, err := ioutil.ReadDir(path)
				if err != nil {
					log.Printf("error reading dir %q\n", path)
					wg.Done()
					continue
				}

				for _, fi := range files {
					fPath := filepath.Join(path, fi.Name())
					if fi.IsDir() {
						wg.Add(1)
						go func(fPath string) { w.dirCh <- fPath }(fPath)
						continue
					}

					// TODO: casting to uint64 when the number is negative will
					// likely cause incorrect results.
					size := uint64(fi.Size())

					// If path is a regular file.
					if fi.Mode().IsRegular() {
						w.stats.fsChan <- filesize{
							filepath: fPath,
							size:     size,
						}
					}
				}
				wg.Done()
			}
		}()
	}
	wg.Add(1)
	w.dirCh <- dir
	wg.Wait()
}

func usage() {
	fmt.Println("Find files recursively from a starting directory iff the file is over a certain size.")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	switch "" {
	case *minSizeStr:
		usage()
		return
	case *dir:
		usage()
		return
	}

	var err error
	minSize, err = humanise.ParseBytes(*minSizeStr)
	if err != nil {
		log.Fatal(err)
	}

	if err := run(*dir); err != nil {
		log.Fatal(err)
	}
}

func run(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	// If path is a regular file.
	if fi.Mode().IsRegular() {
		// TODO(vishen): factor this with stats.printStats()
		size := humanise.Bytes(uint64(fi.Size()))
		path := filepath.Join(path, fi.Name())
		fmt.Printf("%s -> %s\n", size, path)
		return nil
	}

	// If path is not a directory, and isn't a regular file
	// return an error.
	if !fi.IsDir() {
		return fmt.Errorf("%q was not a regular file, or directory", path)
	}

	stats := newStats()
	w := newWorkers(stats)
	fmt.Printf("Looking in %q...\n", path)
	w.start(path)
	stats.printStats()
	return nil
}

func readFiles(stats *stats, path string) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Printf("error reading dir %q\n", path)
		return
	}

	wg := sync.WaitGroup{}
	for _, fi := range files {
		fPath := filepath.Join(path, fi.Name())
		if fi.IsDir() {
			wg.Add(1)
			go func(fPath string) {
				defer wg.Done()
				readFiles(stats, fPath)
			}(fPath)
		}

		// TODO: casting to uint64 when the number is negative will
		size := uint64(fi.Size())

		// likely cause incorrect results.
		// If path is a regular file.
		if fi.Mode().IsRegular() {
			stats.fsChan <- filesize{
				filepath: fPath,
				size:     size,
			}
		}
	}
	wg.Wait()
}

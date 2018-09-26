//go:generate scribblec-param.sh Downloader.scr -d . -param Downloader github.com/nickng/httpget/Downloader -param-api Master -param-api Fetcher -param-api Server

package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"sync"

	"github.com/nickng/httpget/Downloader/Downloader/Master_1to1"
	"github.com/nickng/httpget/Downloader/Downloader/family_1/Fetcher_1toN"
	"github.com/nickng/httpget/msgsig"

	"github.com/nickng/httpget/Downloader/Downloader"
	"github.com/rhu1/scribble-go-runtime/runtime/session2"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2/shm"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2/tcp"
)

const (
	// N is the number of Fetchers.
	N        = 2
	httpHost = "127.0.0.1"
	httpPort = 6060
)

func main() {
	// Load protocol.
	protocol := Downloader.New()

	// Initialise roles.
	M := protocol.New_Master_1to1(2, 1)
	F1 := protocol.New_family_1_Fetcher_1toN(2, 1)
	F2 := protocol.New_family_1_Fetcher_1toN(2, 2)

	// Create connections.
	MtoF1, err := shm.Listen(0)
	if err != nil {
		log.Fatal(err)
	}
	MtoF2, err := shm.Listen(1)
	if err != nil {
		log.Fatal(err)
	}

	// Register messaging.
	gob.Register(msgsig.Request{})
	gob.Register(msgsig.Response{})

	waitall := new(sync.WaitGroup)
	waitall.Add(3)

	go initMaster(M, 0, 1, waitall)
	go initFetcher(F1, MtoF1, httpHost, httpPort, waitall)
	go initFetcher(F2, MtoF2, httpHost, httpPort, waitall)
	waitall.Wait()
}

func initMaster(M *Master_1to1.Master_1to1, portF1, portF2 int, wg *sync.WaitGroup) {
	if err := M.Fetcher_1toN_Dial(1, "inmem", portF1, shm.Dial, new(session2.GobFormatter)); err != nil {
		log.Fatalf("connection failed: %v", err)
	}
	if err := M.Fetcher_1toN_Dial(2, "inmem", portF2, shm.Dial, new(session2.GobFormatter)); err != nil {
		log.Fatalf("connection failed: %v", err)
	}
	fmt.Println("Master")
	M.Run(Master)
	wg.Done()
}

func initFetcher(F *Fetcher_1toN.Fetcher_1toN, M transport2.ScribListener, serverHost string, serverPort int, wg *sync.WaitGroup) {
	if err := F.Master_1to1_Accept(1, M, new(session2.GobFormatter)); err != nil {
		log.Fatalf("connection failed: %v", err)
	}
	if err := F.Server_1to1_Dial(1, serverHost, serverPort, tcp.Dial, new(msgsig.HTTPFormatter)); err != nil {
		log.Fatalf("connection failed: %v", err)
	}
	fmt.Println("Fetcher")
	F.Run(Fetcher)
	wg.Done()
}

// Fetcher is the implemenatation of Fetcher[1..N] role.
func Fetcher(s *Fetcher_1toN.Init) Fetcher_1toN.End {
	// Put implementation of Fetcher here
	url := allocURLs()
	res := allocResponse()

	s0 := s.Master_1to1_Gather_URL(url)
	req := makeRequest(url)

	s1 := s0.Server_1to1_Scatter_Request(req)
	s2 := s1.Server_1to1_Gather_Response(res)

	fetched := extractData(res)
	s3 := s2.Master_1to1_Scatter_Done(fetched)
	return *s3 // end of Fetcher
}

// Master is the implemenatation of Master role.
func Master(s *Master_1to1.Init) Master_1to1.End {
	// Put implementation of Master here
	URL := makeURL("http://127.0.0.1:6060/main.go")
	var fetched []string

	s0 := s.Foreach(func(s *Master_1to1.Init_17) Master_1to1.End {
		s0 := s.Fetcher_ItoI_Scatter_URL(URL)

		data := allocData()
		s1 := s0.Fetcher_ItoI_Gather_Done(data)
		fetched = append(fetched, data...)
		return *s1
	})
	fmt.Println("--- results ---")
	for _, fragment := range fetched {
		fmt.Print(fragment)
	}
	fmt.Println("--- end results ---")
	return *s0 // End of Master
}

// -------- helpers --------------------------------

// makeRequest creates a HTTP GET request using url.
func makeRequest(url []string) []msgsig.Request {
	return []msgsig.Request{msgsig.Request{URL: url[0]}}
}

// allocResponse allocates spaces for receiving Response from Server.
func allocResponse() []msgsig.Response {
	return []msgsig.Response{msgsig.Response{}}
}

func extractData(res []msgsig.Response) []string {
	return []string{string(res[0].Body)}
}

// allocData allocates spaces for receiving final data from a process.
func allocData() []string {
	return make([]string, 1)
}

// makeURL prepares a url to be a payload.
func makeURL(url string) []string {
	return []string{url}
}

// allocURLs allocate container for URLs.
func allocURLs() []string {
	return make([]string, 1)
}

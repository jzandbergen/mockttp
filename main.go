package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/olekukonko/tablewriter"
)

func logRequest(r *http.Request) string {

	b := new(strings.Builder)

	table := tablewriter.NewWriter(b)
	data := [][]string{
		{"Time", fmt.Sprintf("%v", time.Now())},
		{"RemoteAddr", r.RemoteAddr},
		{"Method", r.Method},
		{"Host", r.Host},
		{"RequestURI", r.RequestURI},
		{"Proto", r.Proto},
		{"Headers", "TODO"},
		{"Body", "TODO"},
		{"ContentLength", "TODO"},
		{"TransferEncoding", "TODO"},
	}
	table.AppendBulk(data)
	table.Render()
	fmt.Print(b.String())
	return b.String()
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("let's go!"))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {

	code, _ := strconv.Atoi(r.URL.Query().Get("code"))
	if code > 0 {
		fmt.Printf("Returning http status: %d\n", code)
		w.WriteHeader(code)
	}
	w.Write([]byte(logRequest(r)))
}

func debugHandler(w http.ResponseWriter, r *http.Request) {
	spew.Dump(r)
	w.Write([]byte(logRequest(r)))
}

func main() {

	var addr string
	flag.StringVar(&addr, "addr", ":8080", "Address to listen on. Default ':8080'.")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	defer cancel()
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)

	router := http.NewServeMux()
	router.HandleFunc("/", rootHandler)
	router.HandleFunc("/debug", debugHandler)
	router.HandleFunc("/ready", readyHandler)

	httpServer := &http.Server{
		Addr:        addr,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
		Handler:     router,
	}

	log.Printf("Server starting on %v\n", httpServer.Addr)
	wg.Add(1)
	go func() {
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("HTTP Server shutdown error: %v\n", err)
			os.Exit(1)
		}
		log.Printf("HTTP Server is shut down.")
		wg.Done()
	}()

	log.Printf("Waiting for SIGHUP, SIGINT, SIGTERM or SIGQUIT\n")
	signal := <-sigc
	log.Printf("Received: %v, shutting down...", signal)
	httpServer.Close()
	cancel()
	wg.Wait()
	log.Printf("All done, so long and thanks for all the fish!")
}

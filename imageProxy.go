package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"middleware"
	"net/http"
	"os"
	"os/exec"
)

func main() {

	startServer()
}

func startServer() {
	resource := []middleware.ResourceType{middleware.ResourceType{Path: "/isCredits",
		Method:  "POST",
		Handler: isCreditsHandler(middleware.IsRequestValid(isCreditsResponse(http.HandlerFunc(final))))}}
	config := middleware.ConfigType{Port: 8080, Path: "/ai", Resources: resource}
	middleware.AllowedOrigins = append(middleware.AllowedOrigins, "localhost")

	middleware.StartServer(config)
}

func final(w http.ResponseWriter, r *http.Request) {
	log.Println("Executing finalHandler")
}

func isCreditsHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func isCreditsResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
			message, status := middleware.GetErrorResponse(500, "Server unable to read body.")
			http.Error(w, message, status)
		}
		ioutil.NopCloser(r.Body)

		err = ioutil.WriteFile("~tmp.jpeg", body, 0644)
		if err != nil {
			log.Println("Create image cache... FAIL")
			log.Fatal(err)
			message, status := middleware.GetErrorResponse(500, "Server unable to create image cache.")
			http.Error(w, message, status)
		}

		cmd := exec.Command("python", "-m", "scripts.label_image", "--graph=tf_files/retrained_graph.pb", "--image=~tmp.jpeg")
		var out bytes.Buffer
		cmd.Stdout = &out
		err = cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%q", out.String())
		w.Header().Set("Content-Type", "text/html")
		log.Println("Writing response body...")
		if _, err = w.Write(out.Bytes()); err != nil {
			log.Panicf("Writing response body... FAILED \n [%s]", err.Error())
		}
		log.Println("Writing response body... DONE")

		os.Remove("~tmp.jpeg")

		next.ServeHTTP(w, r)
	})
}

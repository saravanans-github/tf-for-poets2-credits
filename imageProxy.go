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
		defer next.ServeHTTP(w, r)

		log.Println("Reading POST body...")
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Reading POST body... FAILED [%s]", err.Error());
			message, status := middleware.GetErrorResponse(500, "Server unable to read body. " + err.Error())
			http.Error(w, message, status)
			return
		}
		ioutil.NopCloser(r.Body)
		log.Println("Reading POST body... Done")

		log.Println("Creating image cache...")
		err = ioutil.WriteFile("~tmp.jpeg", body, 0644)
		if err != nil {
			log.Printf("Creating image cache... FAIL [%s]", err.Error())
			message, status := middleware.GetErrorResponse(500, "Server unable to create image cache." + err.Error())
			http.Error(w, message, status)
			return
		}
		log.Println("Creating image cache... Done")

		log.Println("Executing python tf command...")
		cmd := exec.Command("python", "-m", "scripts.label_image", "--graph=tf_files/retrained_graph.pb", "--image=~tmp.jpeg")
		var out bytes.Buffer
		cmd.Stdout = &out
		err = cmd.Run()
		if err != nil {
			log.Printf("Executing python tf command... FAILED [%s]", err.Error())
			message, status := middleware.GetErrorResponse(500, "Server unable to execute TensorFlow command. " + err.Error())
                        http.Error(w, message, status)
		}
		log.Println("Executing python tf command... Done")

		fmt.Printf("%q", out.String())
		w.Header().Set("Content-Type", "text/html")
		log.Println("Writing response body...")
		if _, err = w.Write(out.Bytes()); err != nil {
			log.Printf("Writing response body... FAILED \n [%s]", err.Error())
			message, status := middleware.GetErrorResponse(500, "Server unable to write response body. " + err.Error())
                        http.Error(w, message, status)
		}
		log.Println("Writing response body... DONE")

		os.Remove("~tmp.jpeg")
	})
}

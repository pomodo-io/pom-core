package core

import (
    "fmt"
    "log"
    "net/http"
    httpSwagger "github.com/swaggo/http-swagger"
    _ "github.com/polyzuri/pom-core/docs"
)

// @title Pomodo API
// @version 1.0
// @description This is a sample server for Pomodo API.
// @host localhost:8080
// @BasePath /
// @schemes http
// @produce application/json
// @consume application/json
func Run() {
    // Serve Swagger UI
    http.Handle("/swagger/", httpSwagger.WrapHandler)

    // Hello World endpoint
    http.HandleFunc("/", helloHandler)
    
    fmt.Println("Server starting on :8080...")
    fmt.Println("Swagger UI available at: http://localhost:8080/swagger/index.html")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal(err)
    }
}

// @Summary Hello World endpoint
// @Description Returns a hello world message
// @Tags hello
// @Accept */*
// @Produce plain
// @Success 200 {string} string "Hello, World!"
// @Router / [get]
func helloHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, World!")
}

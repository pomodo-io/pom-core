package main

import (
    "flag"
    "os"
    "os/exec"
    "path/filepath"
    "github.com/polyzuri/pom-core/internal/core"
)

// @title Pomodo.io-API
// @version 0.0.1
// @description API documentation for Pomodo.io pom-core.
// @host localhost:8080
// @BasePath /
func main() {
    cleanFlag := flag.Bool("clean", false, "Clean and regenerate Swagger documentation")
    flag.Parse()

    if *cleanFlag {
        // Get the project root directory
        projectRoot, err := os.Getwd()
        if err != nil {
            panic(err)
        }

        // Remove existing docs
        docsDir := filepath.Join(projectRoot, "docs")
        if err := os.RemoveAll(docsDir); err != nil {
            panic(err)
        }

        // Regenerate docs using swag
        cmd := exec.Command("swag", "init", "-g", "main.go", "-o", "docs")
        if err := cmd.Run(); err != nil {
            panic(err)
        }
        
        return
    }

    core.Run()
}

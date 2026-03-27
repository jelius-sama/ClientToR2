package main

import (
    "ClientToR2/internal/router"
    "ClientToR2/internal/util"
    "net/http"
    "os"
    "path/filepath"

    "github.com/jelius-sama/logger"
    "github.com/joho/godotenv"
)

var (
    // Set at compile time (use makefile)
    IS_PROD string
    PORT    string
)

func init() {
    logger.Configure(logger.Cnf{
        IsDev: logger.IsDev{
            EnvironmentVariable: nil,
            ExpectedValue:       nil,
            DirectValue:         logger.BoolPtr(IS_PROD == "FALSE"),
        },
        UseSyslog: false,
    })

    userHome, err := os.UserHomeDir()
    if err != nil {
        logger.Fatal("Couldn't get user's home directory.")
    }

    err = godotenv.Load(filepath.Join(userHome, ".config", "client-to-r2", ".env"))
    if err != nil {
        logger.Fatal("Error loading environment variables.")
    }
}

func main() {
    err := util.EnsureENV()
    if err != nil {
        logger.Fatal(err)
    }

    logger.Info("Starting server on port:", PORT)
    if err := http.ListenAndServe(PORT, router.Router()); err != nil {
        logger.Error("Failed to start server:", err)
    }
}


package main

import (
    "context"
    "fmt"
    "net/http"
    "os"
    "os/signal"
    "path/filepath"
    "syscall"
    "time"

    "github.com/jelius-sama/OpenMediaCloud/internal/mux"
    "github.com/jelius-sama/OpenMediaCloud/internal/util"

    "github.com/jelius-sama/logger"
    "github.com/joho/godotenv"
)

const VERSION = "v3.0.0"

var (
    // Set at compile time (use makefile)
    IS_PROD       string
    PORT          string
    CustomEnvPath *string
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

    if shouldExit := handleFlags(); shouldExit == true {
        os.Exit(0)
    }

    if CustomEnvPath != nil && len(*CustomEnvPath) != 0 {
        if err := godotenv.Load(*CustomEnvPath); err != nil {
            logger.Fatal(err)
        }
    } else {
        loadFromEtc := func() error {
            return godotenv.Load(filepath.Join("/etc", "OpenMediaCloud", ".env"))
        }

        userHome, err := os.UserHomeDir()

        if err != nil {
            logger.Error("Couldn't get user's home directory, loading from `/etc/OpenMediaCloud`.")
            err = loadFromEtc()
            if err != nil {
                logger.Fatal("Error loading environment variables.")
            }
        } else {
            err = godotenv.Load(filepath.Join(userHome, ".config", "OpenMediaCloud", ".env"))
            if err != nil {
                err = loadFromEtc()
                if err != nil {
                    logger.Fatal("Error loading environment variables.")
                }
            }
        }
    }
}

func main() {
    err := util.EnsureENV()
    if err != nil {
        logger.Fatal(err)
    }

    if keyPair, privKeyPath := os.Getenv("CLOUDFRONT_KEY_PAIR_ID"), os.Getenv("CLOUDFRONT_PRIVATE_KEY_PATH"); len(keyPair) != 0 && len(privKeyPath) != 0 {
        // NOTE: os.Stat doesn't necessarily mean we have read permission.
        file, err := os.OpenFile(privKeyPath, os.O_RDONLY, 0)
        if err != nil {
            logger.Fatal("Failed to read cloudfront private key:", err)
        }
        defer file.Close()
    }

    fmt.Println("\n\033[0;36mOpenMediaCloud", VERSION, "\033[0m")
    logger.Info("Starting server on port", PORT)

    var quit chan os.Signal = make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    var server *http.Server = &http.Server{
        Addr:    PORT,
        Handler: mux.Multiplexer(),
    }

    go func() {
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Fatal("Failed to start server on port "+PORT+"\n", err)
        }
    }()

    <-quit
    var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    var deadline, _ = ctx.Deadline()
    var done chan struct{} = make(chan struct{})

    var ticker *time.Ticker = time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    go func() {
        if err := server.Shutdown(ctx); err != nil {
            logger.TimedFatal("Server forced to shutdown:", err)
        }
        close(done)
    }()

    for {
        select {
        case <-done:
            logger.TimedInfo("Server stopped.")
            return

        case <-ctx.Done():
            logger.TimedInfo("Timeout reached:", ctx.Err())
            return

        case <-ticker.C:
            if term := os.Getenv("TERM"); len(term) != 0 {
                // Only show countdown in interactive terminals
                var remaining int = int(time.Until(deadline).Seconds())
                if remaining < 0 {
                    remaining = 0
                }

                fmt.Printf("\r\033[K\033[0;36m[INFO] Shutting down in %d seconds...\033[0m", remaining)
            }
        }
    }
}


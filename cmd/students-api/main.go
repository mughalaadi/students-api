package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mughalaadi/students-api/internal/config"
	"github.com/mughalaadi/students-api/internal/http/handlers/student"
	"github.com/mughalaadi/students-api/internal/storage/sqlite"
)

type sqliteStorageAdapter struct {
	*sqlite.Sqlite
}

func (a *sqliteStorageAdapter) CreateStudent(name, email string, age int) (int64, error) {
	return a.Sqlite.CreateStudent(a, name, email, age)
}

func main() {
	// load config
	cfg := config.MustLoad()
	// database connection
	storage, err := sqlite.New(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database", err.Error())
	}

	studentStore := &sqliteStorageAdapter{storage}

	slog.Info("Storage initialized", slog.String("env", cfg.Env), slog.String("version", "1.0.0"))

	// setup router
	router := http.NewServeMux()
	router.HandleFunc("POST /api/students", student.New(studentStore))
	router.HandleFunc("GET /api/students/{id}", student.GetById(studentStore))
	router.HandleFunc("GET /api/students", student.GetList(studentStore))

	// setup server
	server := http.Server{
		Addr:    cfg.Addr,
		Handler: router,
	}

	slog.Info("server started", slog.String("addr", cfg.Addr))

	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal("Failed to start server")
		}
	}()

	<-done

	slog.Info("shutting down the server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Failed to shutdown server", slog.String("error", err.Error()))
	}

	slog.Info("server stopped")

}

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/sammanbajracharya/drift/internal/app"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	var port int
	flag.IntVar(&port, "port", 6969, "Go backend server port")
	flag.Parse()

	app, err := app.NewApplication()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}
	defer app.DB.Close()

	frontendURL := os.Getenv("FRONTEND_URL")

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{frontendURL}, // or use []string{"*"} for dev
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Max age for preflight in seconds
	}))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	r.Post("/auth/register", app.UserHandler().HandleCredentialCreateUser)
	r.Post("/auth/login", app.UserHandler().HandleCredentialLogin)
	r.Post("/auth/logout", app.UserHandler().HandleLogout)

	// not done yet
	// r.Get("/auth/{provider}", app.UserHandler().HandleOAuthRedirect)
	// r.Get("/auth/{provider}/callback", app.UserHandler().HandleOAuthCallback)

	r.Route("/users", func(protected chi.Router) {
		protected.Use(app.UserHandler().SessionAuthMiddleware)

		protected.Get("/me", app.UserHandler().HandleGetMe)
		protected.Put("/me", app.UserHandler().HandleUpdateMe)
		protected.Get("/{id}", app.UserHandler().HandleGetUserById)
		protected.Delete("/{id}", app.UserHandler().HandleDeleteUser)
	})

	addr := fmt.Sprintf(":%d", port)
	app.Logger.Printf("Server listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}

package app

import (
	"context"
	"ecommerce/components/log"
	"ecommerce/repository"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/spf13/viper"
)

type App struct {
	sync.Mutex
	Name        string //name of application
	Router      *mux.Router
	DB          *gorm.DB
	Log         log.Log
	Server      *http.Server
	Respository repository.EcommerceRepository
}

type ModuleConfig interface {
	TableMigration(wg *sync.WaitGroup)
}

type Controller interface {
	RegisterRoutes(router *mux.Router)
}

func NewApp(name string, dbInstance *gorm.DB, logger log.Log, repo repository.EcommerceRepository) *App {
	return &App{
		Name:        name,
		DB:          dbInstance,
		Log:         logger,
		Respository: repo,
	}
}

func (app *App) Init() {
	app.initializeRouter()
	app.initializeServer()
}

func (app *App) initializeRouter() {
	app.Log.Print(app.Name + "  App Route Initializing")
	app.Router = mux.NewRouter().StrictSlash(true)
	app.Router = app.Router.PathPrefix("/api/v1/ecommerceapp").Subrouter()

}

func (app *App) initializeServer() {
	headers := handlers.AllowedHeaders([]string{
		"Content-Type", "X-Total-Count", "token", "Authorization",
	})
	methods := handlers.AllowedMethods([]string{
		http.MethodPost, http.MethodPut, http.MethodGet, http.MethodDelete, http.MethodOptions,
	})
	originOption := handlers.AllowedOriginValidator(app.checkOrigin)

	app.Server = &http.Server{
		Addr:         "0.0.0.0:5000",
		ReadTimeout:  time.Second * 30,
		WriteTimeout: time.Second * 30,
		IdleTimeout:  time.Second * 30,
		Handler:      handlers.CORS(headers, methods, originOption)(app.Router),
	}
	app.Log.Print("Server Created for port 5000")
}

func (app *App) StartServer() error {
	app.Log.Print("Server Time:", time.Now())
	app.Log.Print("Server Running on port:5000")

	if err := app.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		app.Log.Print("Listen and server error:", err)
		return err
	}
	return nil
}

func (app *App) ResigterControllerRoutes(controllers []Controller) {
	app.Lock()
	defer app.Unlock()
	for _, controller := range controllers {
		controller.RegisterRoutes(app.Router.NewRoute().Subrouter())
	}
}

func (app *App) MigrateTables(configs []ModuleConfig) {
	var wg sync.WaitGroup
	wg.Add(len(configs))
	for _, config := range configs {
		go config.TableMigration(&wg)
	}
	wg.Wait()
	app.Log.Print("End of Migration")
}

func (app *App) checkOrigin(origin string) bool {
	return true
}

func (app *App) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if app.DB != nil {
		sqlDB := app.DB.DB()
		sqlDB.Close()
		app.Log.Print("DB Closed")
	}

	if err := app.Server.Shutdown(ctx); err != nil {
		app.Log.Print("Failed to stop server gracefully:", err)
		return
	}
	app.Log.Print("Server Shutdown Gracefully.")
}

func NewDBConnection(logger log.Log) *gorm.DB {
	user := viper.GetString("DB_USER")
	pass := viper.GetString("DB_PASS")
	host := viper.GetString("DB_HOST")
	port := viper.GetString("DB_PORT")
	name := viper.GetString("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, pass, host, port, name)
	fmt.Println("DB DSN:", dsn)

	db, err := gorm.Open("mysql", dsn)
	if err != nil {
		logger.Print(" Failed to connect to DB:", err)
		return nil
	}

	sqlDB := db.DB()
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetMaxOpenConns(500)
	sqlDB.SetConnMaxLifetime(3 * time.Minute)

	db.LogMode(true)
	db = db.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci") //not understood
	db.BlockGlobalUpdate(true)

	logger.Print(" DB connection established successfully.")
	return db
}

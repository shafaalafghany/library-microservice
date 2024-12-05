package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/shafaalafghany/book-service/handler"
	"github.com/shafaalafghany/book-service/middleware"
	"github.com/shafaalafghany/book-service/model"
	"github.com/shafaalafghany/book-service/repository"
	"github.com/shafaalafghany/book-service/service"
	"gitlab.com/shafaalafghany/synapsis-proto/go/author"
	"gitlab.com/shafaalafghany/synapsis-proto/go/book"
	"gitlab.com/shafaalafghany/synapsis-proto/go/category"
	"gitlab.com/shafaalafghany/synapsis-proto/go/user"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	DBHost          string
	DBUser          string
	DBPassword      string
	DBPort          string
	DBName          string
	JwtSecret       string
	AppPort         string
	RedisHost       string
	RedisPort       string
	UserService     string
	AuthorService   string
	CategoryService string
}

func main() {
	_ = godotenv.Load()

	config := Config{
		JwtSecret:       os.Getenv("SECRET_KEY"),
		AppPort:         os.Getenv("APP_PORT"),
		DBHost:          os.Getenv("DB_HOST"),
		DBPort:          os.Getenv("DB_PORT"),
		DBUser:          os.Getenv("DB_USER"),
		DBPassword:      os.Getenv("DB_PASS"),
		DBName:          os.Getenv("DB_NAME"),
		UserService:     os.Getenv("USER_SERVICE"),
		AuthorService:   os.Getenv("AUTHOR_SERVICE"),
		CategoryService: os.Getenv("CATEGORY_SERVICE"),
		RedisHost:       os.Getenv("REDIS_HOST"),
		RedisPort:       os.Getenv("REDIS_PORT"),
	}

	logConfig := zap.NewDevelopmentConfig()
	logConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, err := logConfig.Build()
	if err != nil {
		panic(err)
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s", config.DBHost, config.DBUser, config.DBPassword, config.DBName, config.DBPort)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
	db.AutoMigrate(&model.Book{})
	db.AutoMigrate(&model.BorrowRecord{})

	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", config.RedisHost, config.RedisPort),
	})

	userConn, err := grpc.NewClient(config.UserService, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect user service %v", err)
	}
	defer userConn.Close()

	authorConn, err := grpc.NewClient(config.AuthorService, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect author service %v", err)
	}
	defer authorConn.Close()

	categoryConn, err := grpc.NewClient(config.CategoryService, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect category service %v", err)
	}
	defer categoryConn.Close()

	userClient := user.NewUserServiceClient(userConn)
	authorClient := author.NewAuthorServiceClient(authorConn)
	categoryClient := category.NewCategoryServiceClient(categoryConn)

	bookRepo := repository.NewBookRepository(db, logger, redisClient)
	bookService := service.NewBookService(bookRepo, logger, userClient, authorClient, categoryClient)
	bookHandler := handler.NewBookHandler(bookService, logger)

	server := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.JWTAuthInterceptor(config.JwtSecret)),
	)
	book.RegisterBookServiceServer(server, bookHandler)
	reflection.Register(server)

	listen, err := net.Listen("tcp", ":"+config.AppPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	if err := server.Serve(listen); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	fmt.Println("listened at ", config.AppPort)

}

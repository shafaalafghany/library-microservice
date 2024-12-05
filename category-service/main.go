package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/joho/godotenv"
	"github.com/shafaalafghany/category-service/handler"
	"github.com/shafaalafghany/category-service/middleware"
	"github.com/shafaalafghany/category-service/model"
	"github.com/shafaalafghany/category-service/repository"
	"github.com/shafaalafghany/category-service/service"
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
	DBHost      string
	DBUser      string
	DBPassword  string
	DBPort      string
	DBName      string
	JwtSecret   string
	AppPort     string
	UserService string
}

func main() {
	_ = godotenv.Load()

	config := Config{
		JwtSecret:   os.Getenv("SECRET_KEY"),
		AppPort:     os.Getenv("APP_PORT"),
		DBHost:      os.Getenv("DB_HOST"),
		DBPort:      os.Getenv("DB_PORT"),
		DBUser:      os.Getenv("DB_USER"),
		DBPassword:  os.Getenv("DB_PASS"),
		DBName:      os.Getenv("DB_NAME"),
		UserService: os.Getenv("USER_SERVICE"),
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
	db.AutoMigrate(&model.Category{})

	userConn, err := grpc.NewClient(config.UserService, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect user service %v", err)
	}
	defer userConn.Close()

	userClient := user.NewUserServiceClient(userConn)

	categoryRepo := repository.NewCategoryRepository(db, logger)
	categoryService := service.NewCategoryService(categoryRepo, logger, userClient)
	categoryHandler := handler.NewCategoryHandler(categoryService, logger)

	server := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.JWTAuthInterceptor(config.JwtSecret)),
	)
	category.RegisterCategoryServiceServer(server, categoryHandler)
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

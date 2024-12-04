package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/shafaalafghany/user-service/handler"
	"github.com/shafaalafghany/user-service/middleware"
	"github.com/shafaalafghany/user-service/model"
	"github.com/shafaalafghany/user-service/repository"
	"github.com/shafaalafghany/user-service/service"
	"gitlab.com/shafaalafghany/synapsis-proto/go/user"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	DBHost     string
	DBUser     string
	DBPassword string
	DBPort     string
	DBName     string
	JwtSecret  string
	AppPort    string
}

func main() {
	config := Config{
		JwtSecret:  os.Getenv("SECRET_KEY"),
		AppPort:    os.Getenv("APP_PORT"),
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PASS"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASS"),
		DBName:     os.Getenv("DB_NAME"),
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
	db.AutoMigrate(&model.User{})

	userRepo := repository.NewUserRepository(db, logger)
	userService := service.NewUserService(userRepo, logger)

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(middleware.JWTAuthInterceptor(config.JwtSecret)),
	)
	user.RegisterUserServiceServer(server, handler.NewUserHandler(userService, logger))
	reflection.Register(server)

	listen, err := net.Listen("tcp", config.AppPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	if err := server.Serve(listen); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	fmt.Println("listened at ", config.AppPort)
}

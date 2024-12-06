# Library Management App
This project consists of a set of microservices built using Go, gRPC, PostgreSQL, and Redis. Each Service is containerized with Docker and uses Docker Compose to manage dependencies and deployment.

## Services The project includes the following services: 
1. **User Service**: Handles user authentication and registration. This is the main service that other services depend on. 
2.  **Author Service**: Manages authors and interacts with the User Service for validation. 
3.  **Category Service**: Manages categories and interacts with the User Service for validation. 
4.  **Book Service**: Manages books and uses Redis for caching. This service depends on the User, Author, and Category services.
  
  ## Project Structure
  ```plain text
  . 
  ├── docker-compose.yml # Docker Compose configuration 
  ├── init-userdb.sh # Initialization script for User database 
  ├── init-authordb.sh # Initialization script for Author database 
  ├── init-categorydb.sh # Initialization script for Category database 
  ├── init-bookdb.sh # Initialization script for Book database 
  ├── user-service/ # User Service source code 
  ├── author-service/ # Author Service source code 
  ├── category-service/ # Category Service source code 
  ├── book-service/ # Book Service source code 
  └── README.md # This file
  ```

## Prerequisites

  

Prerequisite to run this application locally

  

- Docker and Docker Compose installed

- A Docker Hub account (if you're pulling images from Docker Hub)

  

## Setup

  

1.  **Clone the Repository:**

  clone this repository to your local machine:

```bash

git clone https://github.com/shafaalafghany/library-microservice.git
cd library-microservice
```
2. **Install Dependencies**
```bash
docker-compose build
```
this will build the Docker images for each services (user, author, category, and book) based on their perspective.

3. **Environment Setup**

You can start the entire application stack (including PostgreSQL and Redis) using Docker Compose:
```bash
docker-compose up -d
```
this will:
- Spin up the services defined in `docker-compose.yml`
-   Set up PostgreSQL containers with the necessary databases and users
-   Start Redis for caching
-   Run the microservices (User, Author, Category, and Book)

4. **Access Service**

Once the containers are running, the services will be available on the following ports:

- **User Service**: `grpc://localhost:3000`
- **Author Service**: `grpc://localhost:4000`
- **Category Service**: `grpc://localhost:5000`
- **Book Service**: `grpc://localhost:6000`

5. **Inter-Service Communication**

-	The services communicate with each other over gRPC.
-	Each service is configured to connect to other services using the internal Docker network (e.g., `user-service:3000`).

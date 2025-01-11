mkdir -p taxi-fare-calculator/{routes,handlers,models,database,utils,config}
cd taxi-fare-calculator
go mod init taxi-fare-calculator

# Install all required dependencies
go get github.com/gofiber/fiber/v2
go get github.com/gofiber/fiber/v2/middleware/cors
go get github.com/gofiber/fiber/v2/middleware/logger
go get github.com/joho/godotenv
go get go.mongodb.org/mongo-driver/mongo
go get go.mongodb.org/mongo-driver/mongo/options
go get go.mongodb.org/mongo-driver/bson/primitive

# Ensure dependencies are properly recorded
go mod tidy 
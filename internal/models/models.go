package models

// User represents a basic user in the system
type User struct {
    ID    int
    Name  string
    Email string
}

// Config represents application configuration
type Config struct {
    Environment string
    Debug       bool
    APIKey      string
}

// Injectable is a struct that will demonstrate dependency injection
type Injectable struct {
    UserService    interface{} `di:"userService"`
    EmailService   interface{} `di:"emailService"`
    ConfigService  interface{} `di:"configService"`
}
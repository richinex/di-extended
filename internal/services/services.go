package services

import (
	"di-extended/pkg/aop"
	"di-extended/pkg/logger"
	"fmt"

	"go.uber.org/zap"
)

// Interfaces remain the same
type UserService interface {
    GetUser(id int) string
}

type EmailService interface {
    SendEmail(to, message string) error
}

type ConfigService interface {
    GetConfig() string
}

// UserService implementation with lifecycle hooks
type userService struct {
    prefix string
    log    *zap.SugaredLogger // Changed to correct type
}

func NewUserService() UserService {
    log := logger.Get()
    log.Infow("Creating new UserService", "prefix", "USER-")
    return &userService{
        prefix: "USER-",
        log:    log,
    }
}

func (s *userService) PostConstruct() error {
    s.log.Info("PostConstruct: Initializing UserService")
    // Any initialization logic
    return nil
}

func (s *userService) PreDestroy() error {
    s.log.Info("PreDestroy: Cleaning up UserService")
    // Any cleanup logic
    return nil
}

func (s *userService) GetUser(id int) string {
    result := fmt.Sprintf("%s%d", s.prefix, id)
    s.log.Infow("Getting user",
        "id", id,
        "prefix", s.prefix,
        "result", result)
    return result
}

// EmailService implementation with lifecycle and retry
type emailService struct {
    server     string
    log        *zap.SugaredLogger // Changed to correct type
    retryCount int                `di:"retry-count"`
}

func NewEmailService() EmailService {
    log := logger.Get()
    log.Infow("Creating new EmailService", "server", "smtp.example.com")
    return &emailService{
        server: "smtp.example.com",
        log:    log,
    }
}

func (s *emailService) PostConstruct() error {
    s.log.Info("PostConstruct: Initializing EmailService")
    if s.retryCount == 0 {
        s.retryCount = 3 // default retry count
    }
    return nil
}

func (s *emailService) PreDestroy() error {
    s.log.Info("PreDestroy: Cleaning up EmailService")
    // Any cleanup logic
    return nil
}

func (s *emailService) SendEmail(to, message string) error {
    s.log.Infow("Sending email",
        "to", to,
        "server", s.server,
        "messageLength", len(message),
        "retryCount", s.retryCount)

    // Added retry logic
    var lastError error
    for attempt := 0; attempt < s.retryCount; attempt++ {
        s.log.Debugw("Sending attempt",
            "attempt", attempt+1,
            "to", to)

        fmt.Printf("Sending email to %s via %s: %s\n", to, s.server, message)

        // Simulate success
        s.log.Infow("Email sent successfully",
            "to", to,
            "server", s.server,
            "attempt", attempt+1)
        return nil
    }

    return fmt.Errorf("failed to send email after %d attempts: %v",
        s.retryCount, lastError)
}

// ConfigService implementation with profiles
type configService struct {
    env      string
    log      *zap.SugaredLogger // Changed to correct type
    profiles []string
}

func NewConfigService() ConfigService {
    log := logger.Get()
    log.Infow("Creating new ConfigService", "environment", "development")
    return &configService{
        env: "development",
        log: log,
    }
}

func (s *configService) PostConstruct() error {
    s.log.Info("PostConstruct: Initializing ConfigService")
    // Initialize based on active profiles
    if len(s.profiles) > 0 {
        s.env = s.profiles[0]
    }
    return nil
}

func (s *configService) PreDestroy() error {
    s.log.Info("PreDestroy: Cleaning up ConfigService")
    return nil
}

func (s *configService) GetConfig() string {
    result := fmt.Sprintf("Environment: %s", s.env)
    s.log.Infow("Getting config",
        "environment", s.env,
        "result", result)
    return result
}

// LoggingAspect for AOP
type LoggingAspect struct {
    Log *zap.SugaredLogger
}

func NewLoggingAspect() *LoggingAspect {
    return &LoggingAspect{
        Log: logger.Get(),
    }
}

func (a *LoggingAspect) Kind() aop.AspectKind {
    return aop.Before  // Using the correct type from aop package
}

func (a *LoggingAspect) PointCut() string {
    return ".*Service.*" // matches all service methods
}

func (a *LoggingAspect) Advice(jp *aop.JoinPoint) error {  // Using the correct JoinPoint type
    a.Log.Infow("Method call",
        "target", fmt.Sprintf("%T", jp.Target),
        "method", jp.Method.Name,
        "args", jp.Args)
    return nil
}
// package main

// import (
// 	"di-extended/internal/models"
// 	"di-extended/internal/services"
// 	"di-extended/pkg/aop"
// 	"di-extended/pkg/container"
// 	"di-extended/pkg/logger"
// 	"di-extended/pkg/reflection"
// 	"fmt"
// 	"reflect"
// )

// func main() {
//     // Initialize logger
//     logger.Initialize(true) // true for development mode with colors
//     defer logger.Sync()
//     log := logger.Get()
//     log.Info("Starting application")

//     // Create new DI container
//     log.Info("Initializing DI container")
//     di := container.NewContainer()

//     // Set up active profiles
//     log.Info("Setting up active profiles")
//     di.SetActiveProfiles("dev", "local")

//     // Create and register aspects
//     log.Info("Setting up aspects")
//     loggingAspect := &services.LoggingAspect{Log: log}
//     di.AddAspect(loggingAspect)

//     // Register lifecycle hooks
//     log.Info("Registering lifecycle hooks")
//     di.GetLifecycleManager().AddPostConstructHook(container.LifecycleHook{
//         Name:     "LoggingHook",
//         Priority: 1,
//         Handler: func(service interface{}) error {
//             log.Infow("Post-construct hook executing",
//                 "service", fmt.Sprintf("%T", service))
//             return nil
//         },
//     })

//     // Register services with scopes
//     log.Info("Registering services in container")

//     // UserService as Singleton with dev profile condition
//     if err := di.Register("userService", services.NewUserService(), container.Singleton); err != nil {
//         log.Fatalw("Failed to register userService", "error", err)
//     }

//     // EmailService as Prototype with retry configuration
//     if err := di.Register("emailService", services.NewEmailService(), container.Prototype); err != nil {
//         log.Fatalw("Failed to register emailService", "error", err)
//     }

//     // ConfigService as Singleton with profile awareness
//     if err := di.Register("configService", services.NewConfigService(), container.Singleton); err != nil {
//         log.Fatalw("Failed to register configService", "error", err)
//     }

//     // Create injectable struct
//     log.Info("Creating injectable struct")
//     injectable := &models.Injectable{}

//     // Inject dependencies
//     log.Info("Injecting dependencies")
//     if err := di.InjectStruct(injectable); err != nil {
//         log.Fatalw("Failed to inject dependencies", "error", err)
//     }

//     // Create reflection inspector
//     log.Info("Creating reflection inspector")
//     inspector := reflection.NewInspector()

//     // Inspect the injectable struct
//     log.Info("Inspecting injectable struct")
//     info, err := inspector.InspectStruct(injectable)
//     if err != nil {
//         log.Fatalw("Failed to inspect struct", "error", err)
//     }

//     // Print inspection results
//     log.Info("=== Struct Inspection Results ===")
//     fmt.Println(inspector.PrettyPrint(info))

//     // Test services with AOP
//     log.Info("=== Testing Injected Services ===")

//     // Test UserService
//     if us, ok := injectable.UserService.(services.UserService); ok {
//         result := us.GetUser(123)
//         log.Infow("Tested UserService", "result", result)
//     }

//     // Test EmailService with retry logic
//     if es, ok := injectable.EmailService.(services.EmailService); ok {
//         err := es.SendEmail("test@example.com", "Hello from DI!")
//         log.Infow("Tested EmailService", "error", err)
//     }

//     // Test ConfigService with profile awareness
//     if cs, ok := injectable.ConfigService.(services.ConfigService); ok {
//         result := cs.GetConfig()
//         log.Infow("Tested ConfigService", "result", result)
//     }

//     // Test aspect functionality
//     log.Info("=== Testing Aspect-Oriented Programming ===")
//     testAspects(di)

//     // Cleanup resources
//     log.Info("Cleaning up resources")
//     if err := di.Cleanup(); err != nil {
//         log.Errorw("Error during cleanup", "error", err)
//     }

//     log.Info("Application completed successfully")
// }

// func testAspects(di *container.Container) {
//     log := logger.Get()

//     // Create a join point for testing
//     jp := &aop.JoinPoint{
//         Target: struct{}{},
//         Method: reflect.Method{},
//         Args:   []interface{}{"test"},
//     }

//     // Execute aspects
//     if err := di.ExecuteAspects(jp); err != nil {
//         log.Errorw("Failed to execute aspects", "error", err)
//         return
//     }

//     log.Info("Aspects executed successfully")
// }

package main

import (
    "di-extended/pkg/aop"
    "di-extended/pkg/container"
    "di-extended/pkg/logger"
    "errors"
    "fmt"
    "time"
)

// Interfaces
type PaymentProcessor interface {
    ProcessPayment(amount float64, currency string) error
}

type InventoryService interface {
    CheckStock(productID string) (int, error)
    UpdateStock(productID string, quantity int) error
}

type OrderService interface {
    CreateOrder(userID string, items []OrderItem) (string, error)
}

type NotificationService interface {
    NotifyUser(userID string, message string) error
}

// Data structures
type OrderItem struct {
    ProductID string
    Quantity  int
    Price     float64
}

// Implementations with constructors
type stripePaymentProcessor struct {
    apiKey string
}

func NewPaymentProcessor(apiKey string) PaymentProcessor {
    log := logger.Get()
    log.Infow("Creating new payment processor", "apiKey", apiKey)
    return &stripePaymentProcessor{apiKey: apiKey}
}

func (s *stripePaymentProcessor) ProcessPayment(amount float64, currency string) error {
    log := logger.Get()
    log.Infow("Processing payment",
        "amount", amount,
        "currency", currency,
        "processor", "stripe")

    time.Sleep(100 * time.Millisecond)

    log.Info("Payment processed successfully")
    return nil
}

type warehouseInventoryService struct {
    database map[string]int
}

func NewInventoryService() InventoryService {
    log := logger.Get()
    service := &warehouseInventoryService{
        database: map[string]int{"PROD-1": 100, "PROD-2": 50},
    }
    log.Infow("Creating new inventory service",
        "initialStock", service.database)
    return service
}

func (w *warehouseInventoryService) CheckStock(productID string) (int, error) {
    log := logger.Get()
    log.Infow("Checking stock", "productID", productID)

    if qty, exists := w.database[productID]; exists {
        log.Infow("Stock found", "productID", productID, "quantity", qty)
        return qty, nil
    }
    log.Errorw("Product not found", "productID", productID)
    return 0, errors.New("product not found")
}

func (w *warehouseInventoryService) UpdateStock(productID string, quantity int) error {
    log := logger.Get()
    log.Infow("Updating stock",
        "productID", productID,
        "quantity", quantity,
        "currentStock", w.database[productID])

    w.database[productID] = w.database[productID] + quantity

    log.Infow("Stock updated",
        "productID", productID,
        "newStock", w.database[productID])
    return nil
}

type orderServiceImpl struct {
    PaymentProcessor PaymentProcessor  `di:"paymentService" required:"true"`
    Inventory       InventoryService  `di:"inventoryService" required:"true"`
    Notifications   NotificationService `di:"notificationService" required:"true"`
}


func NewOrderService() *orderServiceImpl {
    log := logger.Get()
    log.Info("Creating new order service")
    return &orderServiceImpl{}
}

func (o *orderServiceImpl) CreateOrder(userID string, items []OrderItem) (string, error) {
    log := logger.Get()
    log.Infow("Creating order", "userID", userID, "itemCount", len(items))

    // Check stock
    for _, item := range items {
        stock, err := o.Inventory.CheckStock(item.ProductID)
        if err != nil {
            log.Errorw("Stock check failed",
                "productID", item.ProductID,
                "error", err)
            return "", err
        }
        if stock < item.Quantity {
            log.Errorw("Insufficient stock",
                "productID", item.ProductID,
                "requested", item.Quantity,
                "available", stock)
            return "", errors.New("insufficient stock")
        }
    }

    // Calculate total
    total := 0.0
    for _, item := range items {
        total += item.Price * float64(item.Quantity)
    }
    log.Infow("Order total calculated", "total", total)

    // Process payment
    if err := o.PaymentProcessor.ProcessPayment(total, "USD"); err != nil {
        log.Errorw("Payment processing failed", "error", err)
        return "", err
    }

    // Update inventory
    for _, item := range items {
        if err := o.Inventory.UpdateStock(item.ProductID, -item.Quantity); err != nil {
            log.Errorw("Inventory update failed",
                "productID", item.ProductID,
                "error", err)
            return "", err
        }
    }

    // Notify user
    if err := o.Notifications.NotifyUser(userID, "Order placed successfully!"); err != nil {
        log.Errorw("Notification failed", "error", err)
        // Don't return error here as order is already processed
    }

    orderID := fmt.Sprintf("ORDER-%d", time.Now().Unix())
    log.Infow("Order created successfully", "orderID", orderID)
    return orderID, nil
}

type TransactionAspect struct {}

func (t *TransactionAspect) Kind() aop.AspectKind {
    return aop.Around
}

func (t *TransactionAspect) PointCut() string {
    return "OrderService.CreateOrder"
}

func (t *TransactionAspect) Advice(jp *aop.JoinPoint) error {
    log := logger.Get()
    log.Info("Starting transaction")
    defer log.Info("Ending transaction")
    return nil
}

type emailNotificationService struct {
    retryCount int
}

func NewNotificationService() NotificationService {
    log := logger.Get()
    log.Info("Creating new notification service")
    return &emailNotificationService{retryCount: 0}
}

func (e *emailNotificationService) NotifyUser(userID string, message string) error {
    log := logger.Get()
    log.Infow("Sending notification",
        "userID", userID,
        "message", message,
        "attempt", e.retryCount+1)
    return nil
}

func main() {
    // Initialize logger
    logger.Initialize(true)
    defer logger.Sync()
    log := logger.Get()

    log.Info("Starting e-commerce application")

    // Create container
    di := container.NewContainer()

    // Set up profiles
    di.SetActiveProfiles("prod")
    log.Info("Active profile: prod")

    // Register aspects
    transactionAspect := &TransactionAspect{}
    di.AddAspect(transactionAspect)
    log.Info("Transaction aspect registered")

    // Register lifecycle hooks
    di.GetLifecycleManager().AddPostConstructHook(container.LifecycleHook{
        Name: "ServiceInitializer",
        Priority: 1,
        Handler: func(service interface{}) error {
            log.Infow("Initializing service", "type", fmt.Sprintf("%T", service))
            return nil
        },
    })
    log.Info("Lifecycle hooks registered")

    // Create services using constructors
    log.Info("Creating services...")
    paymentService := NewPaymentProcessor("sk_test_123")
    inventoryService := NewInventoryService()
    orderService := NewOrderService()
    notificationService := NewNotificationService()

    // Register services with proper error handling
    log.Info("Registering services...")
    if err := di.Register("paymentService", paymentService, container.Singleton); err != nil {
        log.Fatalw("Failed to register payment service", "error", err)
    }

    if err := di.Register("inventoryService", inventoryService, container.Singleton); err != nil {
        log.Fatalw("Failed to register inventory service", "error", err)
    }

    if err := di.Register("orderService", orderService, container.Singleton); err != nil {
        log.Fatalw("Failed to register order service", "error", err)
    }

    if err := di.Register("notificationService", notificationService, container.Prototype); err != nil {
        log.Fatalw("Failed to register notification service", "error", err)
    }

    log.Info("All services registered successfully")

    // Inject dependencies with detailed error handling
    log.Info("Injecting dependencies...")
    if err := di.InjectStruct(orderService); err != nil {
        log.Fatalw("Dependency injection failed",
            "error", err,
            "service", "orderService")
    }

    // Verify injection
    if orderService.PaymentProcessor == nil {
        log.Fatal("Payment processor not injected")
    }
    if orderService.Inventory == nil {
        log.Fatal("Inventory service not injected")
    }
    if orderService.Notifications == nil {
        log.Fatal("Notification service not injected")
    }

    log.Info("Dependencies injected successfully")

    // Test the system
    log.Info("Testing order creation...")
    items := []OrderItem{
        {ProductID: "PROD-1", Quantity: 2, Price: 29.99},
        {ProductID: "PROD-2", Quantity: 1, Price: 49.99},
    }

    orderID, err := orderService.CreateOrder("USER-123", items)
    if err != nil {
        log.Errorw("Order creation failed", "error", err)
        return
    }

    log.Infow("Order created successfully",
        "orderID", orderID,
        "items", len(items))

    // Cleanup
    log.Info("Performing cleanup...")
    if err := di.Cleanup(); err != nil {
        log.Errorw("Cleanup failed", "error", err)
    }

    log.Info("Application shutdown complete")
}
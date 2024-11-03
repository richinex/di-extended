// pkg/aop/aspect.go
package aop

import (
    "reflect"
)

// AspectKind represents different types of aspect execution points
// This defines when an aspect should be executed relative to the target method
type AspectKind int

const (
    Before         AspectKind = iota  // Execute before method
    After                             // Execute after method (regardless of outcome)
    Around                            // Execute before and after method
    AfterReturning                    // Execute after method returns successfully
    AfterThrowing                     // Execute after method throws an error
)

// JoinPoint represents the context at which an aspect intercepts the program
// It contains all information about the method being executed
type JoinPoint struct {
    Target     interface{}       // The object being intercepted
    Method     reflect.Method    // Metadata about the method being called
    Args       []interface{}     // Arguments passed to the method
    ReturnVals []interface{}     // Values returned by the method
    Error      error            // Any error that occurred during method execution
}

// Aspect defines the interface for implementing cross-cutting concerns
// Examples: logging, authentication, transaction management
type Aspect interface {
    // Kind returns when this aspect should be executed
    Kind() AspectKind

    // PointCut defines which methods this aspect applies to
    // Example: "Service.*" would match all methods in services
    PointCut() string

    // Advice contains the actual cross-cutting logic to be executed
    // This is where you implement the aspect's behavior
    Advice(jp *JoinPoint) error
}

// AspectManager handles the registration and execution of aspects
// It acts as a container for all aspects in the application
type AspectManager struct {
    aspects []Aspect    // Slice of registered aspects
}

// NewAspectManager creates a new instance of AspectManager
// Initializes with an empty slice of aspects
func NewAspectManager() *AspectManager {
    return &AspectManager{
        aspects: make([]Aspect, 0),
    }
}

// AddAspect registers a new aspect with the manager
// Aspects are executed in the order they are added
func (am *AspectManager) AddAspect(aspect Aspect) {
    am.aspects = append(am.aspects, aspect)
}

// GetAspects returns all registered aspects
// Useful for inspection and debugging
func (am *AspectManager) GetAspects() []Aspect {
    return am.aspects
}

// ExecuteAspects runs all applicable aspects for a given join point
// This is called whenever an intercepted method is executed
func (am *AspectManager) ExecuteAspects(jp *JoinPoint) error {
    // Iterate through all registered aspects
    for _, aspect := range am.aspects {
        // Execute each aspect's advice
        if err := aspect.Advice(jp); err != nil {
            return err
        }
    }
    return nil
}

// This implementation allows us to:

// Define cross-cutting concerns separately
// Apply them declaratively
// Maintain clean separation of concerns
// Add/remove aspects without modifying business logic
// Handle common concerns (logging, security, etc.) uniformly

// Common use cases:

// Logging
// Authentication
// Authorization
// Transaction management
// Performance monitoring
// Error handling
// Caching
// Input validation
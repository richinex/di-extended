// pkg/container/container.go
package container

import (
    "fmt"
    "reflect"
    "sync"
    "di-extended/pkg/logger"
    "di-extended/pkg/aop"
    "go.uber.org/zap"
)

// Container represents a dependency injection container that manages services
type Container struct {
    mu              sync.RWMutex
    services        map[string]*ScopedService
    log             *zap.SugaredLogger
    lifecycleManager *LifecycleManager
    profileManager   *ProfileManager
    aspectManager    *aop.AspectManager
    parent          *Container
}

// NewContainer creates and initializes a new DI container
func NewContainer() *Container {
    return &Container{
        services:         make(map[string]*ScopedService),
        log:             logger.Get(),
        lifecycleManager: NewLifecycleManager(),
        profileManager:   &ProfileManager{
            profiles: make(map[string]*Profile),
            active:   make([]string, 0),
        },
        aspectManager:    aop.NewAspectManager(),
    }
}

// Register adds a new service to the container with the specified qualifier and scope
func (c *Container) Register(qualifier string, service interface{}, scope Scope) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.log.Infow("Registering service",
        "qualifier", qualifier,
        "type", reflect.TypeOf(service),
        "scope", scope)

    if service == nil {
        c.log.Errorw("Cannot register nil service", "qualifier", qualifier)
        return fmt.Errorf("cannot register nil service for qualifier: %s", qualifier)
    }

    if _, exists := c.services[qualifier]; exists {
        c.log.Errorw("Service already registered", "qualifier", qualifier)
        return fmt.Errorf("service already registered for qualifier: %s", qualifier)
    }

    // Create scoped service
    scopedService := &ScopedService{
        Scope:        scope,
        Factory:      func() interface{} { return service },
        Dependencies: make([]string, 0),
    }

    // Handle singleton scope initialization
    if scope == Singleton {
        scopedService.Instance = service
        if lifecycleAware, ok := service.(LifecycleAware); ok {
            // Execute post-construct hooks
            for _, hook := range c.lifecycleManager.postConstructHooks {
                if err := hook.Handler(service); err != nil {
                    return fmt.Errorf("post-construct hook failed: %w", err)
                }
            }
            if err := lifecycleAware.PostConstruct(); err != nil {
                return fmt.Errorf("post-construct failed: %w", err)
            }
        }
    }

    c.services[qualifier] = scopedService
    return nil
}

// Resolve retrieves a service from the container by its qualifier
func (c *Container) Resolve(qualifier string) (interface{}, error) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    c.log.Debugw("Resolving service", "qualifier", qualifier)

    scopedService, exists := c.services[qualifier]
    if !exists {
        if c.parent != nil {
            c.log.Debugw("Service not found in current container, checking parent",
                "qualifier", qualifier)
            return c.parent.Resolve(qualifier)
        }
        c.log.Errorw("Service not found", "qualifier", qualifier)
        return nil, fmt.Errorf("no service found for qualifier: %s", qualifier)
    }

    c.log.Debugw("Found service",
        "qualifier", qualifier,
        "scope", scopedService.Scope)

    switch scopedService.Scope {
    case Singleton:
        if scopedService.Instance == nil {
            c.log.Errorw("Singleton instance is nil", "qualifier", qualifier)
            return nil, fmt.Errorf("singleton instance is nil for qualifier: %s", qualifier)
        }
        return scopedService.Instance, nil
    case Prototype:
        instance := scopedService.Factory()
        if instance == nil {
            c.log.Errorw("Factory produced nil instance", "qualifier", qualifier)
            return nil, fmt.Errorf("factory produced nil instance for qualifier: %s", qualifier)
        }
        if lifecycleAware, ok := instance.(LifecycleAware); ok {
            for _, hook := range c.lifecycleManager.postConstructHooks {
                if err := hook.Handler(instance); err != nil {
                    return nil, fmt.Errorf("post-construct hook failed: %w", err)
                }
            }
            if err := lifecycleAware.PostConstruct(); err != nil {
                return nil, fmt.Errorf("post-construct failed: %w", err)
            }
        }
        return instance, nil
    default:
        c.log.Errorw("Unsupported scope",
            "qualifier", qualifier,
            "scope", scopedService.Scope)
        return nil, fmt.Errorf("unsupported scope: %v", scopedService.Scope)
    }
}

// InjectStruct injects dependencies into struct fields marked with "di" tags
// InjectStruct injects dependencies into struct fields marked with "di" tags
func (c *Container) InjectStruct(target interface{}) error {
    c.log.Info("Starting struct injection")

    targetValue := reflect.ValueOf(target)
    if targetValue.Kind() != reflect.Ptr {
        c.log.Errorw("Target must be a pointer", "actualKind", targetValue.Kind())
        return fmt.Errorf("target must be a pointer to struct, got: %v", targetValue.Kind())
    }

    targetValue = targetValue.Elem()
    if targetValue.Kind() != reflect.Struct {
        c.log.Errorw("Target must be a struct", "actualKind", targetValue.Kind())
        return fmt.Errorf("target must be a pointer to struct, got pointer to: %v", targetValue.Kind())
    }

    targetType := targetValue.Type()
    c.log.Infow("Processing struct for injection",
        "type", targetType.Name(),
        "numFields", targetType.NumField())

    for i := 0; i < targetType.NumField(); i++ {
        field := targetType.Field(i)
        qualifier, ok := field.Tag.Lookup("di")
        if !ok {
            c.log.Debugw("Skipping field without di tag", "field", field.Name)
            continue
        }

        c.log.Infow("Processing field for injection",
            "field", field.Name,
            "qualifier", qualifier,
            "required", field.Tag.Get("required"))

        fieldValue := targetValue.Field(i)
        if !fieldValue.CanSet() {
            c.log.Warnw("Field cannot be set", "field", field.Name)
            continue
        }

        service, err := c.Resolve(qualifier)
        if err != nil {
            if required, ok := field.Tag.Lookup("required"); ok && required == "true" {
                c.log.Errorw("Required service not found",
                    "field", field.Name,
                    "qualifier", qualifier,
                    "error", err)
                return fmt.Errorf("required service not found for field %s: %w", field.Name, err)
            }
            c.log.Warnw("Optional service not found",
                "field", field.Name,
                "qualifier", qualifier)
            continue
        }

        serviceValue := reflect.ValueOf(service)
        if !serviceValue.Type().AssignableTo(fieldValue.Type()) {
            c.log.Errorw("Type mismatch",
                "field", field.Name,
                "expectedType", fieldValue.Type(),
                "actualType", serviceValue.Type())
            return fmt.Errorf("service type %v is not assignable to field type %v",
                serviceValue.Type(), fieldValue.Type())
        }

        fieldValue.Set(serviceValue)
        c.log.Infow("Successfully injected field",
            "field", field.Name,
            "qualifier", qualifier,
            "type", serviceValue.Type())
    }

    // Handle lifecycle
    if lifecycleAware, ok := target.(LifecycleAware); ok {
        c.log.Info("Handling lifecycle for injected struct")
        for _, hook := range c.lifecycleManager.postConstructHooks {
            if err := hook.Handler(target); err != nil {
                c.log.Errorw("Post-construct hook failed", "error", err)
                return fmt.Errorf("post-construct hook failed: %w", err)
            }
        }
        if err := lifecycleAware.PostConstruct(); err != nil {
            c.log.Errorw("Post-construct failed", "error", err)
            return fmt.Errorf("post-construct failed: %w", err)
        }
    }

    c.log.Info("Completed struct injection")
    return nil
}

// SetActiveProfiles sets the active profiles
func (c *Container) SetActiveProfiles(profiles ...string) {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.profileManager.active = profiles
    c.log.Infow("Set active profiles", "profiles", profiles)
}

// AddAspect adds an aspect to the container
func (c *Container) AddAspect(aspect aop.Aspect) {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.aspectManager.AddAspect(aspect)
    c.log.Infow("Added aspect",
        "type", fmt.Sprintf("%T", aspect),
        "pointcut", aspect.PointCut())
}

// GetLifecycleManager returns the lifecycle manager
func (c *Container) GetLifecycleManager() *LifecycleManager {
    return c.lifecycleManager
}

// ExecuteAspects executes all registered aspects for a given join point
func (c *Container) ExecuteAspects(jp *aop.JoinPoint) error {
    c.mu.RLock()
    defer c.mu.RUnlock()

    for _, aspect := range c.aspectManager.GetAspects() {
        switch aspect.Kind() {
        case aop.Before:
            if err := aspect.Advice(jp); err != nil {
                return fmt.Errorf("before aspect failed: %w", err)
            }
        case aop.After:
            if err := aspect.Advice(jp); err != nil {
                return fmt.Errorf("after aspect failed: %w", err)
            }
        case aop.Around:
            if err := aspect.Advice(jp); err != nil {
                return fmt.Errorf("around aspect failed: %w", err)
            }
        case aop.AfterReturning:
            if err := aspect.Advice(jp); err != nil {
                return fmt.Errorf("after returning aspect failed: %w", err)
            }
        case aop.AfterThrowing:
            if jp.Error != nil {
                if err := aspect.Advice(jp); err != nil {
                    return fmt.Errorf("after throwing aspect failed: %w", err)
                }
            }
        }
    }

    return nil
}

// Cleanup performs cleanup of container resources
func (c *Container) Cleanup() error {
    c.mu.Lock()
    defer c.mu.Unlock()

    for qualifier, service := range c.services {
        if service.Scope == Singleton && service.Instance != nil {
            if lifecycleAware, ok := service.Instance.(LifecycleAware); ok {
                // Execute pre-destroy hooks
                for _, hook := range c.lifecycleManager.preDestroyHooks {
                    if err := hook.Handler(service.Instance); err != nil {
                        return fmt.Errorf("pre-destroy hook failed for %s: %w", qualifier, err)
                    }
                }
                if err := lifecycleAware.PreDestroy(); err != nil {
                    return fmt.Errorf("pre-destroy failed for %s: %w", qualifier, err)
                }
            }
        }
    }
    return nil
}

// Profile management
func (c *Container) IsProfileActive(profileName string) bool {
    c.mu.RLock()
    defer c.mu.RUnlock()

    for _, active := range c.profileManager.active {
        if active == profileName {
            return true
        }
    }
    return false
}

// SetParent sets the parent container for hierarchical DI
func (c *Container) SetParent(parent *Container) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.parent = parent
}
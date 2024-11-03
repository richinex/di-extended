// pkg/container/lifecycle.go
package container

// LifecycleAware defines methods for objects that need initialization and cleanup
type LifecycleAware interface {
    // PostConstruct is called after dependency injection is complete
    // Used for any initialization logic
    PostConstruct() error

    // PreDestroy is called before the object is destroyed
    // Used for cleanup logic (closing connections, freeing resources)
    PreDestroy() error
}

// LifecycleHook represents a hook that can be executed at specific lifecycle points
type LifecycleHook struct {
    Name     string                  // Identifier for the hook
    Priority int                     // Execution priority (lower numbers execute first)
    Handler  func(interface{}) error // Function to execute at lifecycle point
}

// LifecycleManager handles the execution of lifecycle hooks
type LifecycleManager struct {
    // Hooks executed after object construction/initialization
    postConstructHooks []LifecycleHook

    // Hooks executed before object destruction
    preDestroyHooks []LifecycleHook
}

// NewLifecycleManager creates a new lifecycle manager instance
func NewLifecycleManager() *LifecycleManager {
    return &LifecycleManager{
        // Initialize empty slices for both hook types
        postConstructHooks: make([]LifecycleHook, 0),
        preDestroyHooks:   make([]LifecycleHook, 0),
    }
}

// AddPostConstructHook registers a hook to run after object construction
func (lm *LifecycleManager) AddPostConstructHook(hook LifecycleHook) {
    lm.postConstructHooks = append(lm.postConstructHooks, hook)
}

// AddPreDestroyHook registers a hook to run before object destruction
func (lm *LifecycleManager) AddPreDestroyHook(hook LifecycleHook) {
    lm.preDestroyHooks = append(lm.preDestroyHooks, hook)
}
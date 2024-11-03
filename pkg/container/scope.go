// pkg/container/scope.go
package container

type Scope int

const (
    Singleton Scope = iota
    Prototype
    Request
    Session
)

type ScopedService struct {
    Instance     interface{}
    Scope        Scope
    Factory      func() interface{}
    Dependencies []string // For prototype scope dependency tracking
}
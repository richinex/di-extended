package container

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test structures
type TestService interface {
    GetName() string
}

type testServiceImpl struct {
    name string
    initialized bool
    destroyed bool
}

func (t *testServiceImpl) GetName() string {
    return t.name
}

func (t *testServiceImpl) PostConstruct() error {
    t.initialized = true
    return nil
}

func (t *testServiceImpl) PreDestroy() error {
    t.destroyed = true
    return nil
}

type TestStruct struct {
    Service  TestService `di:"testService"`
    Optional TestService `di:"optionalService" required:"false"`
    NoTag    TestService
    private  TestService `di:"privateService"`
}

func TestNewContainer(t *testing.T) {
    container := NewContainer()
    assert.NotNil(t, container)
    assert.NotNil(t, container.services)
    assert.NotNil(t, container.log)
    assert.NotNil(t, container.lifecycleManager)
    assert.NotNil(t, container.profileManager)
    assert.NotNil(t, container.aspectManager)
}

func TestContainer_Register(t *testing.T) {
    container := NewContainer()

    tests := []struct {
        name      string
        qualifier string
        service   interface{}
        scope     Scope
        wantErr   bool
    }{
        {
            name:      "valid singleton service",
            qualifier: "testService",
            service:   &testServiceImpl{name: "test"},
            scope:     Singleton,
            wantErr:   false,
        },
        {
            name:      "valid prototype service",
            qualifier: "prototypeService",
            service:   &testServiceImpl{name: "prototype"},
            scope:     Prototype,
            wantErr:   false,
        },
        {
            name:      "nil service",
            qualifier: "nilService",
            service:   nil,
            scope:     Singleton,
            wantErr:   true,
        },
        {
            name:      "duplicate service",
            qualifier: "testService",
            service:   &testServiceImpl{name: "duplicate"},
            scope:     Singleton,
            wantErr:   true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := container.Register(tt.qualifier, tt.service, tt.scope)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                // Verify service was stored
                service, exists := container.services[tt.qualifier]
                assert.True(t, exists)
                if tt.scope == Singleton {
                    assert.Equal(t, tt.service, service.Instance)
                } else {
                    assert.Nil(t, service.Instance)
                    assert.NotNil(t, service.Factory)
                }
                assert.Equal(t, tt.scope, service.Scope)
            }
        })
    }
}

func TestContainer_Resolve(t *testing.T) {
    container := NewContainer()
    testService := &testServiceImpl{name: "test"}
    prototypeService := &testServiceImpl{name: "prototype"}

    // Register services
    err := container.Register("testService", testService, Singleton)
    require.NoError(t, err)
    err = container.Register("prototypeService", prototypeService, Prototype)
    require.NoError(t, err)

    tests := []struct {
        name      string
        qualifier string
        scope     Scope
        want      interface{}
        wantErr   bool
    }{
        {
            name:      "existing singleton service",
            qualifier: "testService",
            scope:     Singleton,
            want:      testService,
            wantErr:   false,
        },
        {
            name:      "prototype service",
            qualifier: "prototypeService",
            scope:     Prototype,
            want:      nil, // Will be a new instance
            wantErr:   false,
        },
        {
            name:      "non-existent service",
            qualifier: "nonExistent",
            scope:     Singleton,
            want:      nil,
            wantErr:   true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := container.Resolve(tt.qualifier)
            if tt.wantErr {
                assert.Error(t, err)
                assert.Nil(t, got)
                return
            }

            assert.NoError(t, err)
            if tt.scope == Singleton {
                assert.Equal(t, tt.want, got)
            } else {
                assert.NotNil(t, got)
                assert.NotEqual(t, tt.want, got)
            }
        })
    }
}

func TestContainer_InjectStruct(t *testing.T) {
    container := NewContainer()
    testService := &testServiceImpl{name: "test"}

    // Register test service
    err := container.Register("testService", testService, Singleton)
    require.NoError(t, err)

    tests := []struct {
        name    string
        target  interface{}
        wantErr bool
    }{
        {
            name:    "valid struct pointer",
            target:  &TestStruct{},
            wantErr: false,
        },
        {
            name:    "non-pointer",
            target:  TestStruct{},
            wantErr: true,
        },
        {
            name:    "nil target",
            target:  nil,
            wantErr: true,
        },
        {
            name:    "pointer to non-struct",
            target:  new(string),
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := container.InjectStruct(tt.target)
            if tt.wantErr {
                assert.Error(t, err)
                return
            }

            assert.NoError(t, err)
            if ts, ok := tt.target.(*TestStruct); ok {
                assert.Equal(t, testService, ts.Service)
                assert.Nil(t, ts.Optional)  // Optional service wasn't registered
                assert.Nil(t, ts.NoTag)     // No tag, shouldn't be injected
                assert.Nil(t, ts.private)   // Private field, can't be injected
            }
        })
    }
}

func TestContainer_Lifecycle(t *testing.T) {
    container := NewContainer()
    service := &testServiceImpl{name: "lifecycle"}

    err := container.Register("lifecycleService", service, Singleton)
    require.NoError(t, err)

    // Verify PostConstruct was called
    assert.True(t, service.initialized)

    // Test cleanup
    err = container.Cleanup()
    require.NoError(t, err)

    // Verify PreDestroy was called
    assert.True(t, service.destroyed)
}

func TestContainer_Profiles(t *testing.T) {
    container := NewContainer()

    // Test profile activation
    container.SetActiveProfiles("dev", "test")
    assert.True(t, container.IsProfileActive("dev"))
    assert.True(t, container.IsProfileActive("test"))
    assert.False(t, container.IsProfileActive("prod"))
}

func TestContainer_ParentChild(t *testing.T) {
    parent := NewContainer()
    child := NewContainer()
    parentService := &testServiceImpl{name: "parent"}

    // Register service in parent
    err := parent.Register("parentService", parentService, Singleton)
    require.NoError(t, err)

    // Set parent-child relationship
    child.SetParent(parent)

    // Resolve from child should find parent's service
    resolved, err := child.Resolve("parentService")
    assert.NoError(t, err)
    assert.Equal(t, parentService, resolved)
}

func TestConcurrency(t *testing.T) {
    container := NewContainer()
    done := make(chan bool)
    const numGoroutines = 10

    // Multiple goroutines registering services
    for i := 0; i < numGoroutines; i++ {
        go func(id int) {
            service := &testServiceImpl{name: fmt.Sprintf("service-%d", id)}
            qualifier := fmt.Sprintf("service-%d", id)
            err := container.Register(qualifier, service, Singleton)
            assert.NoError(t, err)
            done <- true
        }(i)
    }

    // Wait for all goroutines
    for i := 0; i < numGoroutines; i++ {
        <-done
    }

    // Verify all services were registered
    assert.Equal(t, numGoroutines, len(container.services))
}
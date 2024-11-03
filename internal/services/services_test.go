package services

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "strings"
)

func TestNewUserService(t *testing.T) {
    // Test service creation
    service := NewUserService()
    require.NotNil(t, service)

    // Test type assertion
    _, ok := service.(*userService)
    assert.True(t, ok, "service should be of type *userService")

    // Test prefix initialization
    userSvc := service.(*userService)
    assert.Equal(t, "USER-", userSvc.prefix)
}

func TestUserService_GetUser(t *testing.T) {
    service := NewUserService()

    tests := []struct {
        name     string
        id       int
        expected string
    }{
        {
            name:     "positive id",
            id:       123,
            expected: "USER-123",
        },
        {
            name:     "zero id",
            id:       0,
            expected: "USER-0",
        },
        {
            name:     "negative id",
            id:       -1,
            expected: "USER--1",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := service.GetUser(tt.id)
            assert.Equal(t, tt.expected, result)
        })
    }
}

func TestNewEmailService(t *testing.T) {
    service := NewEmailService()
    require.NotNil(t, service)

    // Test type assertion
    emailSvc, ok := service.(*emailService)
    assert.True(t, ok, "service should be of type *emailService")

    // Test server initialization
    assert.Equal(t, "smtp.example.com", emailSvc.server)
}

func TestEmailService_SendEmail(t *testing.T) {
    service := NewEmailService()

    tests := []struct {
        name    string
        to      string
        message string
        wantErr bool
    }{
        {
            name:    "valid email",
            to:      "test@example.com",
            message: "Hello",
            wantErr: false,
        },
        {
            name:    "empty message",
            to:      "test@example.com",
            message: "",
            wantErr: false,
        },
        {
            name:    "empty recipient",
            to:      "",
            message: "Hello",
            wantErr: false, // Current implementation doesn't validate
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := service.SendEmail(tt.to, tt.message)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

func TestNewConfigService(t *testing.T) {
    service := NewConfigService()
    require.NotNil(t, service)

    // Test type assertion
    configSvc, ok := service.(*configService)
    assert.True(t, ok, "service should be of type *configService")

    // Test environment initialization
    assert.Equal(t, "development", configSvc.env)
}

func TestConfigService_GetConfig(t *testing.T) {
    service := NewConfigService()
    result := service.GetConfig()

    // Test result format and content
    assert.True(t, strings.HasPrefix(result, "Environment:"))
    assert.Contains(t, result, "development")
}
package reflection

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

type TestStruct struct {
    PublicField  string  `json:"public" di:"service"`
    privateField string
    Tagged       bool    `custom:"value"`
    NoTags      float64
}

type NestedStruct struct {
    Inner *TestStruct
}

func TestNewInspector(t *testing.T) {
    inspector := NewInspector()
    assert.NotNil(t, inspector)
    assert.NotNil(t, inspector.log)
}

func TestInspector_InspectStruct(t *testing.T) {
    inspector := NewInspector()

    tests := []struct {
        name      string
        target    interface{}
        wantName  string
        wantFields int
        wantErr   bool
    }{
        {
            name:      "valid struct",
            target:    TestStruct{PublicField: "test"},
            wantName:  "TestStruct",
            wantFields: 4,
            wantErr:   false,
        },
        {
            name:      "pointer to struct",
            target:    &TestStruct{PublicField: "test"},
            wantName:  "TestStruct",
            wantFields: 4,
            wantErr:   false,
        },
        {
            name:      "non-struct",
            target:    "not a struct",
            wantName:  "",
            wantFields: 0,
            wantErr:   true,
        },
        {
            name:      "nil target",
            target:    nil,
            wantName:  "",
            wantFields: 0,
            wantErr:   true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            info, err := inspector.InspectStruct(tt.target)
            if tt.wantErr {
                assert.Error(t, err)
                assert.Nil(t, info)
                return
            }

            require.NoError(t, err)
            require.NotNil(t, info)
            assert.Equal(t, tt.wantName, info.Name)
            assert.Equal(t, tt.wantFields, len(info.Fields))

            if tt.wantFields > 0 {
                // Check specific fields for TestStruct
                publicField := info.Fields[0]
                assert.Equal(t, "PublicField", publicField.Name)
                assert.True(t, publicField.IsExported)
                assert.Contains(t, publicField.Tags, "json")
                assert.Contains(t, publicField.Tags, "di")
            }
        })
    }
}

func TestInspector_PrettyPrint(t *testing.T) {
    inspector := NewInspector()
    testStruct := TestStruct{
        PublicField:  "test value",
        privateField: "private",
        Tagged:       true,
        NoTags:      3.14,
    }

    info, err := inspector.InspectStruct(testStruct)
    require.NoError(t, err)

    output := inspector.PrettyPrint(info)

    // Verify output contains expected information
    assert.Contains(t, output, "Struct: TestStruct")
    assert.Contains(t, output, "PublicField")
    assert.Contains(t, output, "test value")
    assert.Contains(t, output, "json: public")
    assert.Contains(t, output, "di: service")
    assert.Contains(t, output, "Tagged")
    assert.Contains(t, output, "NoTags")
    assert.Contains(t, output, "3.14")
}

func TestFieldInfoHandling(t *testing.T) {
    inspector := NewInspector()

    type ComplexStruct struct {
        Pointer *string  `json:"ptr"`
        Slice   []int    `json:"slice"`
        Map     map[string]interface{} `json:"map"`
    }

    str := "test"
    complex := ComplexStruct{
        Pointer: &str,
        Slice:   []int{1, 2, 3},
        Map:     map[string]interface{}{"key": "value"},
    }

    info, err := inspector.InspectStruct(complex)
    require.NoError(t, err)

    // Verify complex types are handled correctly
    assert.Equal(t, 3, len(info.Fields))

    // Check pointer field
    assert.Equal(t, "*string", info.Fields[0].Type)

    // Check slice field
    assert.Equal(t, "[]int", info.Fields[1].Type)

    // Check map field
    assert.Equal(t, "map[string]interface {}", info.Fields[2].Type)
}
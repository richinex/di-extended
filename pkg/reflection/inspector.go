package reflection

import (
    "fmt"
    "reflect"
    "strings"
    "di-extended/pkg/logger"
    "di-extended/pkg/container"
    "di-extended/pkg/aop"
    "go.uber.org/zap"
)

type StructInfo struct {
    Name            string
    Fields          []FieldInfo
    HasLifecycle    bool
    Scope           container.Scope
    ActiveProfiles  []string
    AspectInfo      *AspectInfo
}

type FieldInfo struct {
    Name          string
    Type          string
    Tags          map[string]string
    Value         interface{}
    IsExported    bool
    IsRequired    bool
    InjectionType string
    DefaultValue  string
}

type AspectInfo struct {
    HasAspects  bool
    PointCuts   []string
    Advices     []string
}

type Inspector struct {
    log *zap.SugaredLogger
}

func NewInspector() *Inspector {
    return &Inspector{
        log: logger.Get(),
    }
}

func (i *Inspector) InspectStruct(target interface{}) (*StructInfo, error) {
    i.log.Info("Starting struct inspection")

    if target == nil {
        i.log.Error("Target is nil")
        return nil, fmt.Errorf("target cannot be nil")
    }

    targetValue := reflect.ValueOf(target)
    targetType := targetValue.Type()

    i.log.Debugw("Analyzing target type",
        "initialType", targetType.String())

    // Handle pointer types
    if targetType.Kind() == reflect.Ptr {
        i.log.Debug("Target is a pointer, dereferencing")
        if targetValue.IsNil() {
            i.log.Error("Target pointer is nil")
            return nil, fmt.Errorf("target pointer cannot be nil")
        }
        targetValue = targetValue.Elem()
        targetType = targetValue.Type()
    }

    if targetType.Kind() != reflect.Struct {
        i.log.Errorw("Invalid target type",
            "expectedKind", "struct",
            "actualKind", targetType.Kind())
        return nil, fmt.Errorf("target must be a struct or pointer to struct, got: %v",
            targetType.Kind())
    }

    info := &StructInfo{
        Name:           targetType.Name(),
        Fields:         make([]FieldInfo, 0, targetType.NumField()),
        HasLifecycle:   i.implementsLifecycle(targetType),
        Scope:          i.determineScope(targetType),
        ActiveProfiles: i.getActiveProfiles(targetType),
        AspectInfo:     i.inspectAspects(targetType),
    }

    // Analyze each field
    for fieldIdx := 0; fieldIdx < targetType.NumField(); fieldIdx++ {
        field := targetType.Field(fieldIdx)
        fieldValue := targetValue.Field(fieldIdx)

        fieldInfo := i.inspectField(field, fieldValue)
        info.Fields = append(info.Fields, fieldInfo)
    }

    i.log.Info("Completed struct inspection")
    return info, nil
}

func (i *Inspector) inspectField(field reflect.StructField, fieldValue reflect.Value) FieldInfo {
    i.log.Debugw("Analyzing field",
        "fieldName", field.Name,
        "fieldType", field.Type.String())

    tags := i.parseTags(field)

    var value interface{}
    if fieldValue.CanInterface() {
        value = fieldValue.Interface()
        i.log.Debugw("Retrieved field value",
            "fieldName", field.Name,
            "canInterface", true)
    } else {
        i.log.Debugw("Cannot interface field value",
            "fieldName", field.Name,
            "canInterface", false)
    }

    isExported := field.PkgPath == ""
    isRequired := tags["required"] == "true"
    injectionType := tags["inject"]
    defaultValue := tags["default"]

    return FieldInfo{
        Name:          field.Name,
        Type:          field.Type.String(),
        Tags:          tags,
        Value:         value,
        IsExported:    isExported,
        IsRequired:    isRequired,
        InjectionType: injectionType,
        DefaultValue:  defaultValue,
    }
}

func (i *Inspector) parseTags(field reflect.StructField) map[string]string {
    tags := make(map[string]string)
    if field.Tag != "" {
        i.log.Debugw("Parsing field tags",
            "fieldName", field.Name,
            "rawTags", field.Tag)

        for _, tag := range strings.Split(string(field.Tag), " ") {
            parts := strings.Split(tag, ":")
            if len(parts) == 2 {
                key := strings.Trim(parts[0], "`")
                value := strings.Trim(parts[1], `"`)
                tags[key] = value
            }
        }
    }
    return tags
}

func (i *Inspector) implementsLifecycle(t reflect.Type) bool {
    lifecycleType := reflect.TypeOf((*container.LifecycleAware)(nil)).Elem()
    return t.Implements(lifecycleType) || reflect.PointerTo(t).Implements(lifecycleType)
}

func (i *Inspector) determineScope(t reflect.Type) container.Scope {
    if scope, ok := t.MethodByName("Scope"); ok {
        if scope.Type.NumOut() == 1 && scope.Type.Out(0) == reflect.TypeOf(container.Scope(0)) {
            return container.Singleton // Default to singleton if scope is defined
        }
    }
    return container.Prototype // Default to prototype if no scope is specified
}

func (i *Inspector) getActiveProfiles(t reflect.Type) []string {
    profiles := make([]string, 0)
    if profile, ok := t.MethodByName("Profiles"); ok {
        if profile.Type.NumOut() == 1 && profile.Type.Out(0).Kind() == reflect.Slice {
            // Could extract profile information if available
            // This would need to be called on an instance
        }
    }
    return profiles
}

func (i *Inspector) inspectAspects(t reflect.Type) *AspectInfo {
    aspectInfo := &AspectInfo{
        HasAspects: false,
        PointCuts:  make([]string, 0),
        Advices:    make([]string, 0),
    }

    aspectType := reflect.TypeOf((*aop.Aspect)(nil)).Elem()
    if t.Implements(aspectType) || reflect.PointerTo(t).Implements(aspectType) {
        aspectInfo.HasAspects = true
        // Extract pointcuts and advices if the type implements Aspect
        if aspect, ok := reflect.New(t).Interface().(aop.Aspect); ok {
            aspectInfo.PointCuts = append(aspectInfo.PointCuts, aspect.PointCut())
            aspectInfo.Advices = append(aspectInfo.Advices, fmt.Sprintf("%v", aspect.Kind()))
        }
    }

    return aspectInfo
}

func (i *Inspector) PrettyPrint(info *StructInfo) string {
    i.log.Info("Generating pretty print output")

    var builder strings.Builder

    builder.WriteString(fmt.Sprintf("Struct: %s\n", info.Name))
    builder.WriteString(fmt.Sprintf("Lifecycle Aware: %v\n", info.HasLifecycle))
    builder.WriteString(fmt.Sprintf("Scope: %v\n", info.Scope))

    if len(info.ActiveProfiles) > 0 {
        builder.WriteString("Active Profiles:\n")
        for _, profile := range info.ActiveProfiles {
            builder.WriteString(fmt.Sprintf("  - %s\n", profile))
        }
    }

    if info.AspectInfo.HasAspects {
        builder.WriteString("Aspects:\n")
        for i, pointcut := range info.AspectInfo.PointCuts {
            builder.WriteString(fmt.Sprintf("  Pointcut: %s\n", pointcut))
            if i < len(info.AspectInfo.Advices) {
                builder.WriteString(fmt.Sprintf("  Advice Type: %s\n", info.AspectInfo.Advices[i]))
            }
        }
    }

    builder.WriteString("Fields:\n")
    for _, field := range info.Fields {
        i.log.Debugw("Pretty printing field", "fieldName", field.Name)

        builder.WriteString(fmt.Sprintf("  - %s:\n", field.Name))
        builder.WriteString(fmt.Sprintf("    Type: %s\n", field.Type))
        builder.WriteString(fmt.Sprintf("    Exported: %v\n", field.IsExported))
        builder.WriteString(fmt.Sprintf("    Required: %v\n", field.IsRequired))

        if field.InjectionType != "" {
            builder.WriteString(fmt.Sprintf("    Injection Type: %s\n", field.InjectionType))
        }

        if field.DefaultValue != "" {
            builder.WriteString(fmt.Sprintf("    Default Value: %s\n", field.DefaultValue))
        }

        if len(field.Tags) > 0 {
            builder.WriteString("    Tags:\n")
            for key, value := range field.Tags {
                builder.WriteString(fmt.Sprintf("      %s: %s\n", key, value))
            }
        }

        if field.IsExported && field.Value != nil {
            builder.WriteString(fmt.Sprintf("    Value: %v\n", field.Value))
        }
    }

    return builder.String()
}
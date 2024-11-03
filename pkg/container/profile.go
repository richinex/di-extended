// pkg/container/profile.go
package container

// Profile represents a configuration profile for the container
type Profile struct {
    Name    string  // Profile identifier
    Active  bool    // Whether profile is currently active
    Parent  string  // Parent profile name for inheritance
    Default bool    // Whether this is the default profile
}

// ProfileManager handles profile management and activation
type ProfileManager struct {
    profiles map[string]*Profile  // Map of available profiles
    active   []string            // List of currently active profiles
}

// Condition defines an interface for conditional bean creation/activation
type Condition interface {
    // Matches checks if the condition is satisfied for the given container
    Matches(container *Container) bool
}

// ProfileCondition implements Condition for profile-based conditions
type ProfileCondition struct {
    ProfileName string  // Profile name to check for
}

// Matches checks if a specific profile is active
func (pc *ProfileCondition) Matches(container *Container) bool {
    return container.IsProfileActive(pc.ProfileName)
}


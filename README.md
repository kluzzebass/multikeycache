# multikeycache
A simple go cache that supports indexing by multiple different keys

A multi-key cache is a data structure that allows you to store and retrieve values using not just a primary key, but also through multiple secondary keys. This is particularly useful when you need to look up the same data through different unique identifiers. For example, in a user management system, you might want to find a user by their ID (primary key), email address, or username (secondary keys) - all pointing to the same user data.

This implementation provides thread-safe operations and ensures that secondary keys remain unique across different primary keys, preventing data inconsistencies.

### Example Usage

```go
// Create a new cache with secondary key names "email" and "username"
cache, err := NewMultiKeyCache[int, User, string, string]([]string{"email", "username"})
if err != nil {
    log.Fatal(err)
}

// Add a user with ID 1, email "john@example.com", and username "john123"
err = cache.Set(1, User{Name: "John"}, "john@example.com", "john123")
if err != nil {
    log.Fatal(err)
}

// Get user by primary key (ID)
user, found := cache.Get(1)
fmt.Printf("Found by ID: %v\n", found) // true

// Get user by email (secondary key)
user, found, err = cache.GetBySecondaryKey("email", "john@example.com") 
fmt.Printf("Found by email: %v\n", found) // true

// Get user by username (secondary key)
user, found, err = cache.GetBySecondaryKey("username", "john123")
fmt.Printf("Found by username: %v\n", found) // true

// Try to add another user with same email (will fail)
err = cache.Set(2, User{Name: "Jane"}, "john@example.com", "jane123")
if err != nil {
    fmt.Printf("Error: %v\n", err) // secondary key already exists
}

// Get all primary keys
keys := cache.Keys()
fmt.Printf("All primary keys: %v\n", keys)

// Get all secondary key names
secondaryKeyNames := cache.SecondaryKeyNames() 
fmt.Printf("Secondary key names: %v\n", secondaryKeyNames)

// Get all secondary keys for "email"
emailKeys := cache.SecondaryKeys("email")
fmt.Printf("All email addresses: %v\n", emailKeys)

// Get mapping of secondary keys to primary keys for "email"
emailToPK := cache.SecondaryKeyNameToKeys("email")
fmt.Printf("Email to ID mapping: %v\n", emailToPK)

// Get all values in the cache
allUsers := cache.GetAll()
fmt.Printf("All users: %v\n", allUsers)

// Get number of items in cache
size := cache.Len()
fmt.Printf("Cache size: %d\n", size)

// Delete user by ID
cache.Delete(1)

// Or delete by secondary key
err = cache.DeleteBySecondaryKey("email", "john@example.com")

// Clear the entire cache
cache.Clear()
```

# Database Integration Examples

This directory contains examples of integrating with databases in the Self-Hosted Serverless framework.

## Overview

The Self-Hosted Serverless framework provides built-in support for several databases:

- **PostgreSQL**: A powerful, open-source object-relational database system
- **SQLite**: A lightweight, file-based database
- **Redis**: An in-memory data structure store

These examples demonstrate how to use these databases in your serverless functions.

## PostgreSQL Example

The [postgres](./postgres) directory contains examples of using PostgreSQL with serverless functions.

### Direct Database Access

```go
// Function that directly accesses PostgreSQL
func (h *FunctionHandler) Execute(input map[string]interface{}) (interface{}, error) {
    // Get PostgreSQL connection
    db, err := db.GetPostgresDB()
    if err != nil {
        return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
    }

    // Execute a query
    rows, err := db.Query("SELECT id, name, email FROM users WHERE id > $1 LIMIT $2", 0, 10)
    if err != nil {
        return nil, fmt.Errorf("failed to execute query: %w", err)
    }
    defer rows.Close()

    // Process results
    var users []map[string]interface{}
    for rows.Next() {
        var id int
        var name, email string
        if err := rows.Scan(&id, &name, &email); err != nil {
            return nil, fmt.Errorf("failed to scan row: %w", err)
        }

        users = append(users, map[string]interface{}{
            "id": id,
            "name": name,
            "email": email,
        })
    }

    return map[string]interface{}{
        "users": users,
    }, nil
}
```

### Using the Database API

```sh
# Execute a query via the database API
curl -X POST http://localhost:8080/db -d '{
  "query": "SELECT id, name, email FROM users WHERE id > $1 LIMIT $2",
  "args": [0, 10]
}'
```

## SQLite Example

The [sqlite](./sqlite) directory contains examples of using SQLite with serverless functions.

### Direct Database Access

```go
// Function that directly accesses SQLite
func (h *FunctionHandler) Execute(input map[string]interface{}) (interface{}, error) {
    // Get SQLite connection
    db, err := db.GetSQLiteDB()
    if err != nil {
        return nil, fmt.Errorf("failed to connect to SQLite: %w", err)
    }

    // Execute a query
    rows, err := db.Query("SELECT id, title, content FROM notes WHERE id > ? LIMIT ?", 0, 10)
    if err != nil {
        return nil, fmt.Errorf("failed to execute query: %w", err)
    }
    defer rows.Close()

    // Process results
    var notes []map[string]interface{}
    for rows.Next() {
        var id int
        var title, content string
        if err := rows.Scan(&id, &title, &content); err != nil {
            return nil, fmt.Errorf("failed to scan row: %w", err)
        }

        notes = append(notes, map[string]interface{}{
            "id": id,
            "title": title,
            "content": content,
        })
    }

    return map[string]interface{}{
        "notes": notes,
    }, nil
}
```

## Redis Example

The [redis](./redis) directory contains examples of using Redis with serverless functions.

### Direct Redis Access

```go
// Function that directly accesses Redis
func (h *FunctionHandler) Execute(input map[string]interface{}) (interface{}, error) {
    // Get Redis client
    client, err := db.GetRedisClient()
    if err != nil {
        return nil, fmt.Errorf("failed to connect to Redis: %w", err)
    }

    // Get key from input
    key, ok := input["key"].(string)
    if !ok {
        return nil, errors.New("key is required and must be a string")
    }

    // Get value from Redis
    value, err := client.Get(context.Background(), key).Result()
    if err != nil {
        if err == redis.Nil {
            return map[string]interface{}{
                "exists": false,
            }, nil
        }
        return nil, fmt.Errorf("failed to get value from Redis: %w", err)
    }

    return map[string]interface{}{
        "exists": true,
        "value": value,
    }, nil
}
```

## Database Service Example

The [service](./service) directory contains examples of using the database service.

### Using the Database Service

```go
// Function that uses the database service
func (h *FunctionHandler) Execute(input map[string]interface{}) (interface{}, error) {
    // Get database service
    service, err := db.NewService(db.PostgreSQL)
    if err != nil {
        return nil, fmt.Errorf("failed to create database service: %w", err)
    }
    defer service.Close()

    // Execute a query
    query := "SELECT id, name, email FROM users WHERE id > $1 LIMIT $2"
    args := []interface{}{0, 10}

    rows, err := service.Query(query, args...)
    if err != nil {
        return nil, fmt.Errorf("failed to execute query: %w", err)
    }
    defer rows.Close()

    // Process results
    var users []map[string]interface{}
    for rows.Next() {
        var id int
        var name, email string
        if err := rows.Scan(&id, &name, &email); err != nil {
            return nil, fmt.Errorf("failed to scan row: %w", err)
        }

        users = append(users, map[string]interface{}{
            "id": id,
            "name": name,
            "email": email,
        })
    }

    return map[string]interface{}{
        "users": users,
    }, nil
}
```

## Setting Up Databases

### PostgreSQL

```sh
# Start PostgreSQL with Docker
docker run --name postgres -e POSTGRES_PASSWORD=postgres -p 5432:5432 -d postgres

# Create a database
docker exec -it postgres psql -U postgres -c "CREATE DATABASE serverless;"

# Create a table
docker exec -it postgres psql -U postgres -d serverless -c "
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL
);"

# Insert sample data
docker exec -it postgres psql -U postgres -d serverless -c "
INSERT INTO users (name, email) VALUES
('John Doe', 'john@example.com'),
('Jane Smith', 'jane@example.com'),
('Bob Johnson', 'bob@example.com');"
```

### SQLite

```sh
# Create a SQLite database
sqlite3 serverless.db "
CREATE TABLE notes (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT NOT NULL
);

INSERT INTO notes (title, content) VALUES
('Note 1', 'This is the first note'),
('Note 2', 'This is the second note'),
('Note 3', 'This is the third note');"
```

### Redis

```sh
# Start Redis with Docker
docker run --name redis -p 6379:6379 -d redis

# Set some keys
docker exec -it redis redis-cli set key1 "value1"
docker exec -it redis redis-cli set key2 "value2"
docker exec -it redis redis-cli set key3 "value3"
```

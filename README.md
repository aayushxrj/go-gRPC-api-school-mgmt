# gRPC School Management System

## Project Overview
A high-performance gRPC-based API for a school management system that administrative staff can use to manage students, teachers, and executive staff members. Built with Go, Protocol Buffers, and MongoDB, this system provides a robust, type-safe, and efficient alternative to traditional REST APIs.

## Table of Contents
- [Key Features](#key-features)
- [Technology Stack](#technology-stack)
- [Architecture](#architecture)
- [API Services](#api-services)
  - [Executives Service](#executives-service)
  - [Students Service](#students-service)
  - [Teachers Service](#teachers-service)
- [Message Types](#message-types)
- [Security Features](#security-features)
- [Setup and Installation](#setup-and-installation)
- [Running the Server](#running-the-server)
- [Testing](#testing)
- [Best Practices](#best-practices)
- [Common Pitfalls](#common-pitfalls)

---

## Key Features

### Core Functionality
- ✅ CRUD operations for students, teachers, and executives
- ✅ Bulk operations support (add, update, delete multiple records)
- ✅ Advanced filtering and sorting capabilities
- ✅ Pagination support for large datasets
- ✅ Class-based student grouping with class teacher management

### Authentication & Authorization
- ✅ JWT-based authentication
- ✅ Login/Logout functionality
- ✅ Password management (update, reset, forgot password)
- ✅ User deactivation capabilities
- ✅ Token-based session management

### Security & Performance
- ✅ Request interceptors for authentication
- ✅ Rate limiting (configurable per IP)
- ✅ Response time tracking
- ✅ Input validation using Protocol Buffer validation rules
- ✅ TLS/SSL support (configurable)

---

## Technology Stack

- **Language**: Go 1.25.0
- **RPC Framework**: gRPC (google.golang.org/grpc v1.75.1)
- **Protocol**: Protocol Buffers v3 (proto3)
- **Database**: MongoDB (go.mongodb.org/mongo-driver v1.17.4)
- **Authentication**: JWT (github.com/golang-jwt/jwt/v5 v5.3.0)
- **Password Hashing**: bcrypt (golang.org/x/crypto v0.39.0)
- **Validation**: protoc-gen-validate v1.2.1
- **Configuration**: godotenv v1.5.1

---

## Architecture

```
go-gRPC-api-school-mgmt/
├── cmd/grpcapi/          # Main server entry point
├── internals/
│   ├── api/
│   │   ├── handlers/     # gRPC service implementations
│   │   └── interceptors/ # Middleware (auth, rate limiting, logging)
│   ├── models/           # Data models
│   └── repositories/     # Database operations (MongoDB)
├── pkg/utils/            # Utility functions (JWT, password, error handling)
├── proto/                # Protocol Buffer definitions
│   ├── gen/              # Generated Go code from .proto files
│   └── validate/         # Validation rules
├── cert/                 # TLS certificates
└── data/                 # Sample/seed data
```

### Interceptor Chain
The server implements a chain of interceptors for cross-cutting concerns:
1. **Response Time Interceptor** - Tracks and logs request duration
2. **Authentication Interceptor** - Validates JWT tokens (except for public endpoints)
3. **Rate Limiting Interceptor** - Controls request rate per IP (optional)

---

## API Services

### Executives Service

The `ExecsService` manages administrative staff with full authentication capabilities.

#### RPC Methods

| Method | Description | Auth Required |
|--------|-------------|---------------|
| `GetExecs` | Retrieve executives with optional filtering and sorting | Yes |
| `AddExecs` | Add one or more executives | Yes |
| `UpdateExecs` | Update one or more executives | Yes |
| `DeleteExecs` | Delete executives by IDs | Yes |
| `Login` | Authenticate and receive JWT token | No |
| `Logout` | Invalidate current session token | Yes |
| `UpdatePassword` | Change password for authenticated user | Yes |
| `ResetPassword` | Reset password using reset code | No |
| `ForgotPassword` | Request password reset email | No |
| `DeactivateUser` | Deactivate user accounts | Yes |

#### Request/Response Examples

**Login**
```protobuf
message ExecLoginRequest {
    string username = 1;  // min 6 chars, alphanumeric + @.#$+-
    string password = 2;  // min 9 chars, alphanumeric + @.#$+-
}

message ExecLoginResponse {
    bool status = 1;
    string token = 2;  // JWT token
}
```

**Get Executives**
```protobuf
message GetExecsRequest {
    Exec exec = 1;                    // Filter criteria
    repeated SortField sort_by = 2;   // Sorting options
}
```

**Exec Model**
```protobuf
message Exec {
    string id = 1;
    string first_name = 2;            // Letters and spaces only
    string last_name = 3;             // Letters and spaces only
    string email = 4;                 // Valid email format
    string username = 5;              // Min 6 chars
    string password = 6;              // Min 9 chars (hashed in DB)
    string password_changed_at = 7;
    string user_created_at = 8;
    string password_reset_token = 9;
    string password_token_expires = 10;
    string role = 11;
    bool inactive_status = 12;
}
```

---

### Students Service

The `StudentsService` handles student records with pagination support.

#### RPC Methods

| Method | Description | Auth Required |
|--------|-------------|---------------|
| `GetStudents` | Retrieve students with filtering, sorting, and pagination | Yes |
| `AddStudents` | Add one or more students | Yes |
| `UpdateStudents` | Update one or more students | Yes |
| `DeleteStudents` | Delete students by IDs | Yes |

#### Request/Response Examples

**Get Students**
```protobuf
message GetStudentsRequest {
    Student student = 1;              // Filter criteria
    repeated SortField sort_by = 2;   // Sorting options
    uint32 page_number = 3;           // Pagination
    uint32 page_size = 4;             // Results per page
}
```

**Student Model**
```protobuf
message Student {
    string id = 1;
    string first_name = 2;
    string last_name = 3;
    string email = 4;
    string class = 5;  // e.g., "10th A", "12th B"
}
```

**Delete Response**
```protobuf
message DeleteStudentsConfirmation {
    string status = 1;
    repeated string deleted_ids = 2;  // IDs of deleted students
}
```

---

### Teachers Service

The `TeachersService` manages teacher records and their class assignments.

#### RPC Methods

| Method | Description | Auth Required |
|--------|-------------|---------------|
| `GetTeachers` | Retrieve teachers with filtering and sorting | Yes |
| `AddTeachers` | Add one or more teachers | Yes |
| `UpdateTeachers` | Update one or more teachers | Yes |
| `DeleteTeachers` | Delete teachers by IDs (MongoDB ObjectID format) | Yes |
| `GetStudentsByClassTeacher` | Get all students assigned to a specific teacher | Yes |
| `GetStudentCountByClassTeacher` | Get count of students for a class teacher | Yes |

#### Request/Response Examples

**Teacher Model**
```protobuf
message Teacher {
    string id = 1;
    string first_name = 2;   // Letters and spaces only
    string last_name = 3;    // Letters and spaces only
    string email = 4;        // Valid email format
    string class = 5;        // Alphanumeric and spaces
    string subject = 6;      // Alphanumeric and spaces
}
```

**Get Students by Class Teacher**
```protobuf
message TeacherId {
    string id = 1;  // Must be 24-char hex (MongoDB ObjectID)
}

// Returns: Students (list of students)
```

**Get Student Count**
```protobuf
message StudentCount {
    bool status = 1;
    int32 student_count = 2;
}
```

---

## Message Types

### Common Types

**Sort Field**
```protobuf
message SortField {
    string field = 1;  // Field name to sort by
    Order order = 2;   // ASC or DESC
}

enum Order {
    ASC = 0;
    DESC = 1;
}
```

### Validation Rules

The API uses `protoc-gen-validate` for automatic input validation:

- **Email validation**: Ensures proper email format
- **String patterns**: Regex validation for names, usernames, passwords
- **Length constraints**: Minimum/maximum string lengths
- **MongoDB ObjectID validation**: 24-character hex string pattern
- **Required fields**: Enforced at the protocol level

Examples:
```protobuf
// Email must be valid
string email = 4 [(validate.rules).string = {email: true}];

// MongoDB ObjectID (24 hex chars)
string id = 1 [(validate.rules).string = {
    min_len: 24, 
    max_len: 24, 
    pattern: "^[a-fA-F0-9]{24}$"
}];

// Name (letters and spaces only)
string first_name = 2 [(validate.rules).string = {
    pattern: "^[A-Za-z ]*$"
}];
```

---

## Security Features

### 1. JWT Authentication
- **Token Generation**: On successful login
- **Token Validation**: Via authentication interceptor
- **Token Storage**: In-memory store with automatic cleanup
- **Token Invalidation**: On logout
- **Header Format**: `Authorization: Bearer <token>`

### 2. Password Security
- **Hashing**: bcrypt algorithm
- **Minimum Length**: 9 characters
- **Character Requirements**: Alphanumeric + special chars (@.#$+-)
- **Reset Mechanism**: Token-based with expiration
- **Update Protection**: Requires current password

### 3. Rate Limiting
- **Implementation**: Per-IP rate limiting
- **Configurable**: Requests per time window
- **Reset Mechanism**: Automatic visitor count reset
- **Example**: 50 requests per minute (configurable)

### 4. Public Endpoints
The following endpoints bypass authentication:
- `/main.ExecsService/Login`
- `/main.ExecsService/ForgotPassword`
- `/main.ExecsService/ResetPassword`

### 5. TLS/SSL Support
- Certificate and key files in `cert/` directory
- Configurable via environment variables
- Can be enabled/disabled based on deployment needs

---

## Setup and Installation

### Prerequisites
- Go 1.25.0 or higher
- MongoDB (local or cloud instance)
- Protocol Buffer compiler (`protoc`)
- Go protobuf plugins

### Installation Steps

1. **Clone the repository**
   ```bash
   git clone https://github.com/aayushxrj/go-gRPC-api-school-mgmt.git
   cd go-gRPC-api-school-mgmt
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up environment variables**
   Create a `.env` file in `cmd/grpcapi/`:
   ```env
   SERVER_PORT=50051
   MONGODB_URI=mongodb://localhost:27017
   DB_NAME=school_management
   JWT_SECRET=your-secret-key-here
   CERT_FILE=../../cert/cert.pem
   KEY_FILE=../../cert/key.pem
   ```

4. **Generate Protocol Buffer code** (if modified)
   ```bash
   protoc --go_out=. --go_opt=paths=source_relative \
          --go-grpc_out=. --go-grpc_opt=paths=source_relative \
          --validate_out="lang=go:." --validate_opt=paths=source_relative \
          proto/*.proto
   ```

5. **Seed the database** (optional)
   ```bash
   # Import sample data from data/ directory
   mongoimport --db school_management --collection students --file data/students_data.json
   mongoimport --db school_management --collection teachers --file data/teachers_data.json
   mongoimport --db school_management --collection execs --file data/execs_data.json
   ```

---

## Running the Server

### Development Mode
```bash
cd cmd/grpcapi
go run server.go
```

### Production Mode (with TLS)
```bash
# Build binary
go build -o bin/server cmd/grpcapi/server.go

# Run with TLS enabled
./bin/server
```

### Using Docker (if Dockerfile exists)
```bash
docker build -t school-mgmt-grpc .
docker run -p 50051:50051 school-mgmt-grpc
```

The server will start on the port specified in `.env` (default: 50051).

---

## Testing

### Using gRPC Reflection
The server has gRPC reflection enabled, allowing tools like `grpcurl` and `grpcui`:

```bash
# List all services
grpcurl -plaintext localhost:50051 list

# List methods for a service
grpcurl -plaintext localhost:50051 list main.StudentsService

# Call a method
grpcurl -plaintext -d '{"username": "admin", "password": "password123"}' \
  localhost:50051 main.ExecsService/Login
```

### Using Postman
Postman supports gRPC natively:
1. Create new gRPC request
2. Enter server URL: `localhost:50051`
3. Import proto files or use server reflection
4. Select service and method
5. Fill in request message
6. Add metadata for auth: `authorization: Bearer <token>`

### Performance Testing with ghz
```bash
# Install ghz
go install github.com/bojand/ghz/cmd/ghz@latest

# Run benchmark (example config in ghz_config.json)
ghz --config ghz_config.json
```

### Unit Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internals/api/handlers/
```

---

## Best Practices

### 1. Modularity
- **Separation of Concerns**: Handlers, interceptors, repositories, and models are in separate packages
- **Reusable Components**: Utility functions centralized in `pkg/utils/`
- **Interface-based Design**: Allows for easy mocking and testing

### 2. Documentation
- **Proto Comments**: Document all services, methods, and messages in `.proto` files
- **Code Comments**: Explain complex business logic
- **API Documentation**: Auto-generated from proto files

### 3. Error Handling
- **gRPC Status Codes**: Use appropriate codes (e.g., `codes.InvalidArgument`, `codes.Unauthenticated`)
- **Error Messages**: Descriptive and user-friendly
- **Structured Errors**: Consistent error response format
- **Centralized Handler**: Error handling utilities in `pkg/utils/error_handler.go`

### 4. Security
- **Input Validation**: Enforced at protocol level using validation rules
- **Password Hashing**: Always hash passwords before storage
- **JWT Expiration**: Implement token expiration and refresh mechanisms
- **HTTPS/TLS**: Use in production environments
- **Environment Variables**: Never commit secrets to version control

### 5. Testing
- **Unit Tests**: Test individual handlers and utilities
- **Integration Tests**: Test end-to-end flows
- **Mocking**: Mock database and external dependencies
- **Benchmarking**: Use tools like `ghz` for performance testing

### 6. Database Operations
- **Connection Pooling**: Reuse MongoDB connections
- **Indexing**: Create indexes on frequently queried fields
- **Transactions**: Use MongoDB transactions for multi-document operations
- **Error Recovery**: Graceful handling of connection failures

### 7. Performance Optimization
- **Interceptor Chain**: Minimize overhead in interceptors
- **Database Queries**: Optimize with proper indexing and projection
- **Pagination**: Always paginate large result sets
- **Connection Reuse**: Keep database connections alive

---

## Common Pitfalls

### 1. Overcomplicating the API
- ❌ Don't create too many RPC methods for simple operations
- ✅ Use flexible filter/query patterns instead
- ✅ Leverage protobuf's optional fields for partial updates

### 2. Ignoring Security
- ❌ Don't skip authentication for "internal" endpoints
- ❌ Don't store passwords in plain text
- ❌ Don't expose sensitive data in error messages
- ✅ Always validate and sanitize inputs
- ✅ Use TLS in production
- ✅ Implement proper RBAC (Role-Based Access Control)

### 3. Poor Documentation
- ❌ Don't leave proto files undocumented
- ❌ Don't ignore the importance of examples
- ✅ Document request/response formats
- ✅ Provide usage examples
- ✅ Keep README updated

### 4. Inadequate Testing
- ❌ Don't skip unit tests for "simple" handlers
- ❌ Don't forget edge cases
- ✅ Test authentication flows thoroughly
- ✅ Test error scenarios
- ✅ Perform load testing before production

### 5. Inefficient Database Operations
- ❌ Don't fetch all records without pagination
- ❌ Don't perform N+1 queries
- ✅ Use MongoDB aggregation pipelines for complex queries
- ✅ Implement proper indexing strategy
- ✅ Monitor and optimize slow queries

### 6. Not Handling Concurrent Requests
- ❌ Don't ignore race conditions
- ❌ Don't use global state without proper synchronization
- ✅ Use mutexes or channels for shared state (see rate limiter)
- ✅ Design for stateless operations where possible

### 7. Ignoring Observability
- ❌ Don't deploy without logging
- ❌ Don't ignore performance metrics
- ✅ Log important operations and errors
- ✅ Track response times (already implemented)
- ✅ Monitor server health and resource usage

---

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code Standards
- Follow Go best practices and idioms
- Write unit tests for new features
- Update proto files before implementation
- Document all public APIs
- Run `go fmt` and `go vet` before committing

---

## License

This project is licensed under the MIT License - see the LICENSE file for details.

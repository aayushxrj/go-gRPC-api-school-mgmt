# MONGODB(Windows) 

```
netstat -ano | findstr 27017
```
```
& "C:\Program Files\MongoDB\Server\6.0\bin\mongosh.exe"
```

# MAILHOG (Send mails from :1025 to :8025)

```
go install github.com/mailhog/MailHog@latest
```

Add these to ~/.bashrc

```
export GOPATH=$HOME/go
export PATH="$PATH:$GOPATH/bin"
```
```
source ~/.bashrc
```
```
MailHog
```
```
http://localhost:8025
```

# Commands (Windows)
INSTALL protoc and add to path

```
go mod init github.com/aayushxrj/go-gRPC-api-school-mgmt
go mod tidy
```

On windows 
```
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/envoyproxy/protoc-gen-validate@latest
```
add to path
```
C:\Users\aayus\go\bin
```

```
protoc \
  -I proto \
  --go_out=proto/gen --go_opt=paths=source_relative \
  --go-grpc_out=proto/gen --go-grpc_opt=paths=source_relative \
  --validate_out="lang=go,paths=source_relative:proto/gen" \
  proto/main.proto proto/students.proto proto/execs.proto
```
```
protoc -I proto --go_out=proto/gen --go_opt=paths=source_relative --go-grpc_out=proto/gen --go-grpc_opt=paths=source_relative --validate_out="lang=go,paths=source_relative:proto/gen" proto/main.proto proto/students.proto proto/execs.proto
```
or
```
protoc `
  -I proto `
  --go_out=proto/gen --go_opt=paths=source_relative `
  --go-grpc_out=proto/gen --go-grpc_opt=paths=source_relative `
  --validate_out="lang=go,paths=source_relative:proto/gen" `
  proto/main.proto proto/students.proto proto/execs.proto
```

```
go get google.golang.org/grpc
go get github.com/envoyproxy/protoc-gen-validate
go get github.com/joho/godotenv

go get go.mongodb.org/mongo-driver/mongo
go get github.com/go-mail/mail/v2
go get github.com/golang-jwt/jwt/v5

```
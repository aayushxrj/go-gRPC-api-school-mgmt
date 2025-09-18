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

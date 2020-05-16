## Golang web service for image upload


This project was part of the assessment for Servicerocket agile developer position

to run the program:

Clone the repo

```git

git clone https://khairimusa60@bitbucket.org/khairimusa60/khairi-image-upload-service.git

```
Run main class

```go

go run main.go

```

Note: Make sure you have $GOPATH and $GOROOT setup in your enviroment


Can follow steps below to get your go setup:

```
1. Download go from https://golang.org

2. Add new system variable for GOPATH of value C:\Users\username\go

3. Add new system variable for GOROOT of value C:\Go\
```

GOPATH is the root of the workspace. Inside it need to create 3 new folders

```
1. src/: location of Go source code (for example, .go, .c, .g, .s).

2. pkg/: location of compiled package code (for example, .a).

3. bin/: location of compiled executable programs built by Go.
```

Note: make sure to clone the project to src/ folder so that dont have to configure anything. After that your are good to go

if having and issue with the 3rd party library import

```
run "go get -u <github.com>/<username>/<reponame>" command
```
  
---


## Architecure

This is the basic architecture of the application

![](https://bitbucket.org/khairimusa60/khairi-image-upload-service/raw/f1e9846436889f9d63c0e336aa3f3a37808d29ca/pictures/golang%20architecture.PNG) 

Note: actually for the current implementation it did not include any adatabases. Only the main.go class that exposes api and templates for client

API endpoints(base url: "http://localhost:8080") specificactions:

| Method | Enpoint | Description |
| ------------- | ------------- | ------------- |
| `POST` | `/` | Post to this endpoint to upload picture in a form of multipart/form-data |
| `GET` | `/about` | Get this endpoint to see the about page |

---
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

Note: make sure to clone the project to src/ folder so that dont have to configure anything. After that you are good to go

if having issues importing 3rd party libraries run command below:

```
go get -u <github.com>/<username>/<reponame>
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

## Main screen

1. This is the main screen where we can upload the images/zip file. This part shows the original image details and permanent link to it
![](https://bitbucket.org/khairimusa60/khairi-image-upload-service/raw/7783b4b58cb41882e95b41c509eac28ca4804f9a/pictures/main_screen_ss.png)

2. This is the part where the half and quater sized image being shown
![](https://bitbucket.org/khairimusa60/khairi-image-upload-service/raw/7783b4b58cb41882e95b41c509eac28ca4804f9a/pictures/main_screen_ss_1.png)

3. This is the part where all the images inside a zip file will be shown with each's parmenant link
![](https://bitbucket.org/khairimusa60/khairi-image-upload-service/raw/7783b4b58cb41882e95b41c509eac28ca4804f9a/pictures/main_screen_ss_2.png)



## Golang web service for image upload

This project was part of the assessment for Servicerocket agile developer position 

---

## Architecure

This is the basic architecture of the application

![](https://bitbucket.org/khairimusa60/khairi-image-upload-service/raw/f1e9846436889f9d63c0e336aa3f3a37808d29ca/pictures/golang%20architecture.PNG) 

Note: actually for the current implementation it did not include any adatabases. Only the main.go class that exposes api and templates for client

API endpoints(base url: "http://localhost:8080") specificactions:

| Method  | Enpoint | Description
| ------------- | ------------- | ------------- |
| POST  | "/"  | Post to this endpoint to upload picture in a form of multipart/form-data |
| GET  | "/about"  | Get this endpoint to see the about page  |

---
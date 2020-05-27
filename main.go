package main

import (
	"archive/zip"
	_ "archive/zip"
	"crypto/rand"
	"fmt"
	"html/template"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

// Photo struct
type OriginalPhoto struct {
	PhotoName   string
	PhotoPath   string
	PhotoWidth  int
	PhotoHeight int
}

// 1/2 width thumbnail photo
type HalfSizeThumbnail struct {
	HalfSizeThumbnailName   string
	HalfSizeThumbnailPath   string
	HalfSizeThumbnailWidth  int
	HalfSizeThumbnailHeight int
}

// 1/4 width thumbnail photo
type QuaterSizeThumbnail struct {
	QuaterSizeThumbnailName   string
	QuaterSizeThumbnailPath   string
	QuaterSizeThumbnailWidth  int
	QuaterSizeThumbnailHeight int
}

// photo inside the zip file
type UnzipedPhoto struct {
	UnzipedPhotoPath string
}

// PhotoModel struct
var PhotoModel struct {
	Photos               []OriginalPhoto
	HalfSizeThumbnails   []HalfSizeThumbnail
	QuaterSizeThumbnails []QuaterSizeThumbnail
	UnzipedPhotos        []UnzipedPhoto
}

// html template
var tpl *template.Template

// max image size to be uploaded
const maxUploadSize = 2 * 1024 * 1024 // 2MB

func init() {
	// Must() method must be succeed before the program shuts down
	// parseGlob() just parse the template files in templates/ folder and get it reeady to be use
	// glob patterns specifies sets of filenames wildcard character. example mv *.txt
	// glob is a bunch of file names in short
	tpl = template.Must(template.ParseGlob("templates/*.gohtml"))
}

// function to initialize all endpoints and the port
func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/about", about)
	// handle will handle request pattern from browser example img tag <img src="/public/pics/example.jpg">
	// http.Handle("/public/pics/", what to do with that route)
	// this line say that i will handle any pattern that says /public/pics/ from browser and strip the prefix /public/pics/, replaces it
	// with the file server directory of public/pics
	http.Handle("/public/pics/", http.StripPrefix("/public/pics/", http.FileServer(http.Dir("public/pics/"))))
	http.ListenAndServe(":8080", nil)

}

// handler func for about page
func about(w http.ResponseWriter, req *http.Request) {
	tpl.ExecuteTemplate(w, "about.gohtml", "About This Service")
}

// handler func for index page
func index(w http.ResponseWriter, req *http.Request) {

	// original photo
	var createdPhoto OriginalPhoto

	// Half size photo thumbnail
	var halfSizeThumbnail HalfSizeThumbnail

	// 1/4 size photo thumbnail
	var quaterSizeThumbnail QuaterSizeThumbnail

	// unziped photo details
	var unzipedPhotoDetails UnzipedPhoto

	// process form submission
	if req.Method == http.MethodPost {
		multipartFile, fileHeader, err := req.FormFile("nf")
		if err != nil {
			fmt.Println(err)
		}
		defer multipartFile.Close()

		// read the file and return in in bytes
		fileBytes, err := ioutil.ReadAll(multipartFile)
		if err != nil {
			renderError(w, "INVALID_FILE", http.StatusBadRequest)
			return
		}

		// check the file type
		detectedFileType := http.DetectContentType(fileBytes)
		switch detectedFileType {
		case "image/jpeg", "image/jpg":
		case "image/gif", "image/png":
		case "application/zip":
			break
		default:
			renderError(w, "INVALID_FILE_TYPE", http.StatusBadRequest)
			return
		}

		// check the size of the image, can get the file size on the file header
		fileSize := fileHeader.Size
		if fileSize > maxUploadSize {
			renderError(w, "CANT_PARSE_FORM SIZE EXCEEDED 2MB", http.StatusInternalServerError)
			return
		}

		// get root directory
		rootDirectory, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
		}

		// get the extension if it is an image
		ext, err := mime.ExtensionsByType(detectedFileType)
		if err != nil {
			fmt.Println(err)
		}
		// create random token for file name
		hash := randToken(12)
		// if it is a zip file
		if detectedFileType == "application/zip" {
			// get .zip extension
			extZip := strings.Split(fileHeader.Filename, ".")[1]
			// set the name of the zip file
			zipFileName := fmt.Sprintf("%x", hash) + "." + extZip

			// get path from the method from passing in root directory and the zip file name
			zipFilePath, _, _, err := generateDotDotPath(rootDirectory, zipFileName)

			newZipFile, err := os.Create(zipFilePath) // if theres already the file, it will truncate the existing one
			if err != nil {
				fmt.Println(err)
			}
			defer newZipFile.Close()

			// copy
			// seek(pointer to read data, from where to start the pointer 0 - begining, 1 - at the current position, 2 - from end )
			// this is the part where we want to mention that for the multipart file read all its content
			// before passing it inside the io.Copy() method
			// that is why this is needed
			multipartFile.Seek(0, 0) // want to seek the file pointer to the 0th byte of the data
			io.Copy(newZipFile, multipartFile)

			// read the uploaded image in public/pics path(file server)
			zipReader, err := zip.OpenReader(zipFilePath)
			if err != nil {
				log.Println(err)
			}
			defer zipReader.Close()

			// loop through the images inside the unzipped file
			for _, file := range zipReader.Reader.File {
				zippedFile, err := file.Open()
				if err != nil {
					log.Fatal(err)
				}
				defer zippedFile.Close()

				// targetDir is the file server
				targetDir := "./public/pics"
				extractedFilePath := filepath.Join(targetDir, file.Name)
				if file.FileInfo().IsDir() {
					// this one will print the unziped folder location not the image inside it
					os.MkdirAll(extractedFilePath, file.Mode())
				} else {
					// read unzipped folder and the current image inside it
					currentFileName := file.Name
					// generate html readable ../ path
					_, _, dotDotPath, err := generateDotDotPath(rootDirectory, currentFileName)
					// add the photo path to model to show the parmenant link to the front end
					unzipedPhotoDetails.UnzipedPhotoPath = dotDotPath
					PhotoModel.UnzipedPhotos = append(PhotoModel.UnzipedPhotos, unzipedPhotoDetails)

					outputFile, err := os.OpenFile(extractedFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
					if err != nil {
						log.Fatal(err)
					}
					defer outputFile.Close()

					_, err = io.Copy(outputFile, zippedFile)
					if err != nil {
						log.Fatal(err)
					}
				}
			}

		} else {
			// if file extension is not a zip
			//sprint f is just a way print the string content without it showing in the console.
			fileName := hash + ext[0]
			path, _, dotDotPath, err := generateDotDotPath(rootDirectory, fileName)
			newFile, err := os.Create(path) // if theres already the file, it will truncate the existing one
			if err != nil {
				fmt.Println(err)
			}
			defer newFile.Close()

			// copy
			// seek(pointer to read data, from where to start the pointer 0 - begining, 1 - at the current position, 2 - from end )
			// this is the part where we want to mention that for the multipart file read all its content
			// before passing it inside the io.Copy() method
			// that is why this is needed
			multipartFile.Seek(0, 0) // want to seek the file pointer to the 0th byte of the data
			io.Copy(newFile, multipartFile)

			// read the uploaded image in public/pics path(file server)
			reader, err := os.Open(path)
			if err != nil {
				fmt.Println(err)
			}

			// get the image configuration by decoding the reader that contains the file
			imgConfig, _, err := image.DecodeConfig(reader)
			if err != nil {
				fmt.Println(err)
			}

			// close the reader after all operation done
			defer reader.Close()

			// construct the photo struct and append to photoModel struct
			createdPhoto.PhotoPath = dotDotPath
			createdPhoto.PhotoName = fileName
			createdPhoto.PhotoWidth = imgConfig.Width
			createdPhoto.PhotoHeight = imgConfig.Height
			PhotoModel.Photos = append(PhotoModel.Photos, createdPhoto)

			// check if the width or height less than 128px
			if createdPhoto.PhotoWidth < 128 && createdPhoto.PhotoHeight < 128 {
				// just return the original image cause its smaller than 128
				http.Redirect(w, req, "/", http.StatusAccepted)
			} else {
				// half size
				generateThumbnail(2, imgConfig, path, detectedFileType, w, rootDirectory, halfSizeThumbnail, quaterSizeThumbnail)

				// quater size
				generateThumbnail(4, imgConfig, path, detectedFileType, w, rootDirectory, halfSizeThumbnail, quaterSizeThumbnail)
			}
		}

	}

	// this method writes into the template takes in writer, template name, and data
	tpl.ExecuteTemplate(w, "index.gohtml", PhotoModel)
}

// render error to front end
func renderError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(message))
}

// generate random token
func randToken(len int) string {
	b := make([]byte, len)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// method to generate thumbnail
func generateThumbnail(n int, imgConfig image.Config, path string, detectedFileType string, w http.ResponseWriter,
	rootDirectory string, halfSizeThumbnail HalfSizeThumbnail, quaterSizeThumbnail QuaterSizeThumbnail) {

	// load original image
	img, err := imaging.Open(path)
	if err != nil {
		renderError(w, "CANT_OPEN_DIRECTORY", http.StatusInternalServerError)
		return
	}

	// generate thumbnail based on the original img
	thumbnail := imaging.Resize(img, imgConfig.Width/n, 0, imaging.Box)

	// generate random token to name the new thumbnail file
	thumbnailName := randToken(12)
	fileEndings, err := mime.ExtensionsByType(detectedFileType)
	thumbnailFullName := thumbnailName + fileEndings[0]
	if err != nil {
		renderError(w, "CANT_READ_FILE_TYPE", http.StatusInternalServerError)
		return
	}

	thumbnailPath, _, dotDotthumbnailPath, err := generateDotDotPath(rootDirectory, thumbnailFullName)

	//save resized image
	err = imaging.Save(thumbnail, thumbnailPath)
	if err != nil {
		renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
		return
	}

	// read the newly created thumbnail image in public/pics path(file server)
	thumbnailReader, err := os.Open(thumbnailPath)
	if err != nil {
		fmt.Println(err)
	}

	// get the image configuration by decoding the reader that contains the file
	thumbnailImgConfig, _, err := image.DecodeConfig(thumbnailReader)
	if err != nil {
		fmt.Println(err)
	}

	if n == 2 {
		// set thumbnail 1/2 of file width and height
		halfSizeThumbnail.HalfSizeThumbnailPath = dotDotthumbnailPath
		halfSizeThumbnail.HalfSizeThumbnailName = thumbnailFullName
		halfSizeThumbnail.HalfSizeThumbnailHeight = thumbnailImgConfig.Height
		halfSizeThumbnail.HalfSizeThumbnailWidth = thumbnailImgConfig.Width
		PhotoModel.HalfSizeThumbnails = append(PhotoModel.HalfSizeThumbnails, halfSizeThumbnail)
	}

	if n == 4 {
		// set thumbnail 1/4 of file width and height
		quaterSizeThumbnail.QuaterSizeThumbnailPath = dotDotthumbnailPath
		quaterSizeThumbnail.QuaterSizeThumbnailName = thumbnailFullName
		quaterSizeThumbnail.QuaterSizeThumbnailHeight = thumbnailImgConfig.Height
		quaterSizeThumbnail.QuaterSizeThumbnailWidth = thumbnailImgConfig.Width
		PhotoModel.QuaterSizeThumbnails = append(PhotoModel.QuaterSizeThumbnails, quaterSizeThumbnail)
	}
}

// generate a readable path for <img> tag
func generateDotDotPath(rootDirectory string, fileName string) (path string, fowardSlashPath string, dotDotPath string, err error) {
	// this will produce the real path, path from the C: drive
	path = filepath.Join(rootDirectory, "public", "pics", fileName)
	// this replace all the \ slash of windows directory to / slash so that html can read
	fowardSlashPath = strings.Replace(path, "\\", "/", -1)
	// convert the real path to absolute path(in the root)
	dotDotPath = strings.Replace(fowardSlashPath, "C:/Users/khair/go/src/khairi-go-image-upload/", "../", -1)
	return
}

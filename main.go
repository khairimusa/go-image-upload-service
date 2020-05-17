package main

import (
	"archive/zip"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"github.com/disintegration/imaging"
	"html/template"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	_ "archive/zip"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Photo struct
type OriginalPhoto struct {
	PhotoName string
	PhotoPath string
	PhotoWidth int
	PhotoHeight int
}

// 1/2 width thumbnail photo
type HalfSizeThumbnail struct {
	HalfSizeThumbnailName string
	HalfSizeThumbnailPath string
	HalfSizeThumbnailWidth int
	HalfSizeThumbnailHeight int
}

// 1/4 width thumbnail photo
type QuaterSizeThumbnail struct {
	QuaterSizeThumbnailName string
	QuaterSizeThumbnailPath string
	QuaterSizeThumbnailWidth int
	QuaterSizeThumbnailHeight int
}

// photo inside the zip file
type UnzipedPhoto struct {
	UnzipedPhotoPath string
}

// PhotoModel struct
var PhotoModel struct{
	Photos []OriginalPhoto
	HalfSizeThumbnails [] HalfSizeThumbnail
	QuaterSizeThumbnails []QuaterSizeThumbnail
	UnzipedPhotos []UnzipedPhoto
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
	//http.Handle("/favicon.ico", http.NotFoundHandler())
	http.Handle(
		"/public/pics/",
		http.StripPrefix("/public/pics/", http.FileServer(http.Dir("public/pics/"))))
	http.ListenAndServe(":8080", nil)

}

// handler func for about page
func about(w http.ResponseWriter, req *http.Request){
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

		// check the size of the image, can get the file size on the file header
		fileSize := fileHeader.Size
		if fileSize > maxUploadSize {
			renderError(w, "CANT_PARSE_FORM SIZE EXCEEDED 2MB", http.StatusInternalServerError)
			return
		}

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

		// get root directory
		rootDirectory, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
		}

		// create sha for file name
		ext := strings.Split(fileHeader.Filename, ".")[1]
		hash := sha1.New()
		// if it is a zip file
		if (ext == "zip"){
			zipFileName := fmt.Sprintf("%x", hash.Sum(nil)) + "." + ext
			zipFilePath, _, dotDotPath, err  := generateDotDotPath(rootDirectory, zipFileName)
			log.Println(dotDotPath)

			newZipFile, err := os.Create(zipFilePath) // if theres already the file, it will truncate the existing one
			if err != nil {
				fmt.Println(err)
			}
			defer newZipFile.Close()

			// copy
			multipartFile.Seek(0, 0)
			io.Copy(newZipFile, multipartFile)

			// read the uploaded image in public/pics path(file server)
			zipReader, err := zip.OpenReader(zipFilePath)
			if err != nil {
				log.Println(err)
			}
			defer zipReader.Close()
			for _, file := range zipReader.Reader.File {
				zippedFile, err := file.Open()
				if err != nil {
					log.Fatal(err)
				}
				defer zippedFile.Close()

				targetDir := "./public/pics"
				extractedFilePath := filepath.Join(
					targetDir,
					file.Name,
				)
				if file.FileInfo().IsDir() {
					// this one will print the unziped folder location not the image inside it
					os.MkdirAll(extractedFilePath, file.Mode())
				} else {
					currentFileName := file.Name
					_,_,dotDotPath,err := generateDotDotPath(rootDirectory, currentFileName)
					unzipedPhotoDetails.UnzipedPhotoPath = dotDotPath
					PhotoModel.UnzipedPhotos = append(PhotoModel.UnzipedPhotos, unzipedPhotoDetails)
					outputFile, err := os.OpenFile(
						extractedFilePath,
						os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
						file.Mode(),
					)
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
			io.Copy(hash, multipartFile)
			fileName := fmt.Sprintf("%x", hash.Sum(nil)) + "." + ext
			path,_,dotDotPath,err :=generateDotDotPath(rootDirectory, fileName)
			newFile, err := os.Create(path) // if theres already the file, it will truncate the existing one
			if err != nil {
				fmt.Println(err)
			}
			defer newFile.Close()

			// copy
			multipartFile.Seek(0, 0)
			io.Copy(newFile, multipartFile)

			// read the uploaded image in public/pics path(file server)
			reader, err := os.Open(path)
			if err != nil {
				fmt.Println(err)
			}

			// get the image configuration by decoding the reader that contains the file
			imgConfig, _, err :=image.DecodeConfig(reader)
			if err != nil{
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
				generateThumbnail(2,imgConfig, path, detectedFileType, w, rootDirectory, halfSizeThumbnail, quaterSizeThumbnail)

				// quater size
				generateThumbnail(4,imgConfig, path, detectedFileType, w, rootDirectory, halfSizeThumbnail, quaterSizeThumbnail)
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

func generateThumbnail(n int, imgConfig image.Config, path string, detectedFileType string, w http.ResponseWriter,
	rootDirectory string, halfSizeThumbnail HalfSizeThumbnail, quaterSizeThumbnail QuaterSizeThumbnail){

	// load original image
	img, err := imaging.Open(path)
	if err != nil {
		renderError(w, "CANT_OPEN_DIRECTORY", http.StatusInternalServerError)
		return
	}

	// generate thumbnail based on the original img
	thumbnail := imaging.Resize(img, imgConfig.Width/n , 0, imaging.Box)

	// generate random token to name the new thumbnail file
	thumbnailName := randToken(12)
	fileEndings, err := mime.ExtensionsByType(detectedFileType)
	thumbnailFullName := thumbnailName+fileEndings[0]
	if err != nil {
		renderError(w, "CANT_READ_FILE_TYPE", http.StatusInternalServerError)
		return
	}

	thumbnailPath := filepath.Join(rootDirectory,"public","pics", thumbnailFullName)
	// this replace all the \ slash of windows directory to / slash so that html can read
	fowardSlashthumbnailPath := strings.Replace(thumbnailPath, "\\", "/", -1)
	// convert the real path to absolute path(in the root)
	dotDotthumbnailPath := strings.Replace(fowardSlashthumbnailPath, "C:/Users/khair/go/src/khairi-go-image-upload/", "../", -1)

	//save resized image
	err = imaging.Save(thumbnail,thumbnailPath)
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
	thumbnailImgConfig, _, err :=image.DecodeConfig(thumbnailReader)
	if err != nil{
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

func generateDotDotPath(rootDirectory string, fileName string) (path string, fowardSlashPath string, dotDotPath string, err error){
	// this will produce the real path, path from the C: drive
	path = filepath.Join(rootDirectory, "public", "pics", fileName)
	// this replace all the \ slash of windows directory to / slash so that html can read
	fowardSlashPath = strings.Replace(path, "\\", "/", -1)
	// convert the real path to absolute path(in the root)
	dotDotPath = strings.Replace(fowardSlashPath, "C:/Users/khair/go/src/khairi-go-image-upload/", "../", -1)
	return
}
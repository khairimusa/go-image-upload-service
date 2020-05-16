package main

import (
	"crypto/sha1"
	"fmt"
	"html/template"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
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



// PhotoModel struct
var PhotoModel struct{
	Photos []OriginalPhoto
	HalfSizeThumbnails [] HalfSizeThumbnail
	QuaterSizeThumbnails []QuaterSizeThumbnail
}

// html template
var tpl *template.Template

func init() {
	// Must() method must be succeed before the program shuts down
	// parseGlob() just parse the template files in templates/ folder and get it reeady to be use
	tpl = template.Must(template.ParseGlob("templates/*"))
}

// function to initialize all enpoints and the port
func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/about", about)
	http.Handle("/favicon.ico", http.NotFoundHandler())
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


	// process form submission
	if req.Method == http.MethodPost {
		multipartFile, fileHeader, err := req.FormFile("nf")
		if err != nil {
			fmt.Println(err)
		}
		defer multipartFile.Close()

		// create sha for file name
		ext := strings.Split(fileHeader.Filename, ".")[1]
		hash := sha1.New()
		io.Copy(hash, multipartFile)
		fileName := fmt.Sprintf("%x", hash.Sum(nil)) + "." + ext

		// create new file
		rootDirectory, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
		}
		path := filepath.Join(rootDirectory, "public", "pics", fileName) // will produce: 04_upload_pictures/public/pic/{filename}
		newFile, err := os.Create(path) // if theres already the file, it will truncate the existing one
		if err != nil {
			fmt.Println(err)
		}
		defer newFile.Close()

		// copy
		multipartFile.Seek(0, 0)
		io.Copy(newFile, multipartFile)

		// scan file in directory and extract the width and height
		//filesInDirectory, _ := ioutil.ReadDir(path)
		//for _, imgFile := range filesInDirectory {
		//	if reader, err := os.Open(path); err == nil {
		//		defer reader.Close()
		//		imageConfiguration, _, err := image.DecodeConfig(reader)
		//		if err != nil {
		//			fmt.Fprintf(os.Stderr, "%s: %v\n", imgFile.Name(), err)
		//			continue
		//		}
		//		fmt.Printf("%s %d %d\n", imgFile.Name(), imageConfiguration.Width, imageConfiguration.Height)
		//		fmt.Fprintf(w, imgFile.Name(), imageConfiguration.Width, imageConfiguration.Height)
		//	} else {
		//		fmt.Println("Impossible to open the file:", err)
		//	}
		//}

		// read the uploaded image in public/pics path
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

		// get original width of file
		createdPhoto.PhotoWidth = imgConfig.Width

		// get original height of file
		createdPhoto.PhotoHeight = imgConfig.Height

		// generate thumbnail 1/4 of file width and height
		quaterSizeThumbnail.QuaterSizeThumbnailHeight = imgConfig.Height/4
		quaterSizeThumbnail.QuaterSizeThumbnailWidth = imgConfig.Width/4

		// generate thumbnail 1/2 of file width
		halfSizeThumbnail.HalfSizeThumbnailHeight = imgConfig.Height/2
		halfSizeThumbnail.HalfSizeThumbnailWidth = imgConfig.Width/2

		// formula to maintain aspect ratio = (original height / original width) x new width = new height


		// construct the photo struct and append to photoModel struct
		createdPhoto.PhotoPath = path
		createdPhoto.PhotoName = fileName
		PhotoModel.Photos = append(PhotoModel.Photos, createdPhoto)

		// check if the width or height less than 128px

	}

	// this method writes into the template takes in writer, template name, and data
	tpl.ExecuteTemplate(w, "index.gohtml", PhotoModel)
}


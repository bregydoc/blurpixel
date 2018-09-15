package main

import (
	"github.com/kataras/iris"
	"image"
	"net/http"
	"strings"

	"./processor"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	app := iris.New()

	app.Get("/api", func(c iris.Context) {
		urlImageSrc := c.URLParam("src")
		log.Println(urlImageSrc)

		resp, err := http.Get(urlImageSrc)
		if err != nil {
			c.StatusCode(iris.StatusInternalServerError)
			c.JSON(iris.Map{
				"error": err.Error(),
			})
			return
		}
		log.Println(resp.Header)

		contentType := resp.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "image") {
			c.StatusCode(iris.StatusInternalServerError)
			c.JSON(iris.Map{
				"error": "Not valid image, content type: " + contentType,
			})
			return
		}

		imageSrc, filename, err := image.Decode(resp.Body)
		if err != nil {
			c.StatusCode(iris.StatusInternalServerError)
			c.JSON(iris.Map{
				"error": err.Error(),
			})
			return
		}
		log.Println(filename, imageSrc.Bounds().Dx(), imageSrc.Bounds().Dy())

		defaultWidth := uint32(imageSrc.Bounds().Dx())
		defaultHeight := uint32(imageSrc.Bounds().Dy())
		defaultRadius := uint32(20)

		var doneChannel = make(chan struct{}, defaultRadius)

		result := processor.Process(imageSrc, defaultWidth, defaultHeight, defaultRadius, doneChannel)
		<-doneChannel

		var localFile *os.File

		//result = imaging.Resize(result, 128, 128, imaging.Lanczos)

		switch contentType {
		case "image/jpeg":
			localFile, err = ioutil.TempFile(os.TempDir(), "blrpx_")
			if err != nil {
				c.StatusCode(iris.StatusInternalServerError)
				c.JSON(iris.Map{
					"error": err.Error(),
				})
				return
			}

			err = jpeg.Encode(localFile, result, nil)
			if err != nil {
				c.StatusCode(iris.StatusInternalServerError)
				c.JSON(iris.Map{
					"error": err.Error(),
				})
				return
			}
			break
		case "image/png":
			localFile, err = ioutil.TempFile(os.TempDir(), "blrpx_")
			if err != nil {
				c.StatusCode(iris.StatusInternalServerError)
				c.JSON(iris.Map{
					"error": err.Error(),
				})
				return
			}
			defer localFile.Close()

			err = png.Encode(localFile, result)
			if err != nil {
				c.StatusCode(iris.StatusInternalServerError)
				c.JSON(iris.Map{
					"error": err.Error(),
				})
				return
			}
			break
		default:
			break
		}

		defer localFile.Close()

		err = c.SendFile(localFile.Name(), "blrpx_"+filename)
		if err != nil {
			c.StatusCode(iris.StatusInternalServerError)
			c.JSON(iris.Map{
				"error": err.Error(),
			})
			return
		}
		c.Header("Content-Type", "image/png")
		c.StatusCode(iris.StatusOK)
	})

	app.Run(iris.Addr(":4233"))

}

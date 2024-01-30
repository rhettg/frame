package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"log"
	"net/http"
	"text/template"

	"github.com/google/uuid"
)

//go:embed index.html
var content embed.FS

// Define the background color as a global variable

var (
	colors = [4]color.RGBA{
		{0, 128, 0, 255},   // Green
		{128, 0, 128, 255}, // Purple
		{255, 0, 0, 255},   // Red
		{0, 0, 255, 255},   // Blue
	}
	backgroundColor = color.RGBA{255, 255, 255, 255}
)

func main() {
	log.Println("Starting server on :8080")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request: %s %s", r.Method, r.URL.Path)

		// Log request parameters
		if r.Method == http.MethodPost {
			//{"untrustedData":{"fid":244761,"url":"https://rhetts-air.taildefad.ts.net","messageHash":"0xe84754f4668a4d2d779fb694af806f2edd2ccc53","timestamp":1706655303000,"network":1,"buttonIndex":3,"castId":{"fid":244761,"hash":"0x0000000000000000000000000000000000000001"}},"trustedData":{"messageBytes":"0a53080d1099f80e18c7b0ac2e20018201430a2368747470733a2f2f7268657474732d6169722e7461696c64656661642e74732e6e657410031a1a0899f80e121400000000000000000000000000000000000000011214e84754f4668a4d2d779fb694af806f2edd2ccc53180122402a0ecedcf037e335eb72df6716bdaf0420f7ce39bc6f2554fe941a629fb145da7fb69a50c85d2e16f215c46d28b5e0f1c8dda1a2e16dd3b34146efc5a84a3d06280132200ca1713a7d3128291332ee03c8db5e9b21a5c6698d7a44f5ef8bf941e19ea3d1"}}

			type postData struct {
				Fid         int    `json:"fid"`
				Url         string `json:"url"`
				MessageHash string `json:"messageHash"`
				Timestamp   int64  `json:"timestamp"`
				Network     int    `json:"network"`
				ButtonIndex int    `json:"buttonIndex"`
				CastId      struct {
					Fid  int    `json:"fid"`
					Hash string `json:"hash"`
				} `json:"castId"`
			}

			req := struct {
				UntrustedData postData `json:"untrustedData"`
			}{}

			err := json.NewDecoder(r.Body).Decode(&req)
			if err != nil {
				log.Printf("Error decoding untrusted data: %v", err)
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			log.Printf("Untrusted data received: %+v", req.UntrustedData)

			if req.UntrustedData.ButtonIndex > 0 && req.UntrustedData.ButtonIndex <= len(colors) {
				log.Printf("setting color %d", req.UntrustedData.ButtonIndex)
				backgroundColor = colors[req.UntrustedData.ButtonIndex-1]
			}

			queryParams := r.URL.Query()
			for key, values := range queryParams {
				for _, value := range values {
					log.Printf("Query parameter: %s = %s", key, value)
				}
			}

			body, _ := io.ReadAll(r.Body)
			defer r.Body.Close()
			fmt.Println("Body: ", string(body))

		}

		type PageData struct {
			UUID string
		}

		data := PageData{
			UUID: uuid.New().String(),
		}

		tmpl, err := template.ParseFS(content, "index.html")
		if err != nil {
			log.Printf("Error parsing index.html as template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err = tmpl.Execute(w, data)
		if err != nil {
			log.Printf("Error executing template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		log.Println("Served index.html with UUID")
	})

	http.HandleFunc("/image", func(w http.ResponseWriter, r *http.Request) {
		// Create a new 1080P image (1920x1080)
		img := image.NewRGBA(image.Rect(0, 0, 1375, 720))

		// Fill the image with the background color
		for y := 0; y < img.Bounds().Dy(); y++ {
			for x := 0; x < img.Bounds().Dx(); x++ {
				img.Set(x, y, backgroundColor)
			}
		}

		// Set the header and write the image to the response
		w.Header().Set("Content-Type", "image/jpeg")
		err := jpeg.Encode(w, img, &jpeg.Options{Quality: 90})
		if err != nil {
			log.Printf("Error encoding image to PNG: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		log.Println("Served image")
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

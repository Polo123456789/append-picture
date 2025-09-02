package main

import "flag"

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/skip2/go-qrcode"
	"golang.ngrok.com/ngrok/v2"
)

//go:embed index.html
var indexHTML string

func main() {
	ngrokFlag := flag.Bool("ngrok", false, "Exponer el servidor usando ngrok")
	flag.Parse()

	err := os.MkdirAll("pictures", os.ModePerm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "No se pudo crear el directorio pictures: %v\n", err)
		return
	}

	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, indexHTML)
	})

	http.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(10 << 20) // 10MB max
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(os.Stdout, "Error al procesar el formulario: %v\n", err)
			return
		}
		file, handler, err := r.FormFile("foto")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(os.Stdout, "No se recibió archivo: %v\n", err)
			return
		}
		defer file.Close()

		ext := filepath.Ext(handler.Filename)
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		filename := filepath.Join("pictures", timestamp+ext)
		out, err := os.Create(filename)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(os.Stdout, "No se pudo crear la imagen: %v\n", err)
			return
		}
		defer out.Close()
		_, err = io.Copy(out, file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(os.Stdout, "No se pudo escribir la imagen: %v\n", err)
			return
		}
		fmt.Fprintf(os.Stdout, "![](%s)", filename)
	})

	if *ngrokFlag {
		listener, err := ngrok.Listen(context.Background())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error iniciando ngrok: %v\n", err)
			return
		}
		ngrokUrl := listener.URL()
		qr, err := qrcode.New(ngrokUrl.String(), qrcode.Medium)
		if err == nil {
			ascii := qr.ToString(false)
			fmt.Fprintf(os.Stderr, "Escanea este QR para abrir la web y enviar fotos:\n%s\nURL: %s\n", ascii, ngrokUrl)
		} else {
			fmt.Fprintf(os.Stderr, "No se pudo generar el QR: %v\n", err)
		}
		fmt.Fprintf(os.Stderr, "Servidor escuchando en %s ...\n", ngrokUrl)
		http.Serve(listener, nil)
	} else {
		addr := ":8080"
		fmt.Fprintf(os.Stderr, "Servidor local escuchando en http://localhost%s ...\n", addr)
		http.ListenAndServe(addr, nil)
	}
}

package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/skip2/go-qrcode"
)

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return "localhost"
}

func main() {
	err := os.MkdirAll("pictures", os.ModePerm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "No se pudo crear el directorio pictures: %v\n", err)
		return
	}

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

	ip := getLocalIP()
	url := fmt.Sprintf("http://%s:8080/", ip)
	qr, err := qrcode.New(url, qrcode.Medium)
	if err == nil {
		ascii := qr.ToString(false)
		fmt.Fprintf(os.Stderr, "Escanea este QR para acceder al servidor:\n%s\nURL: %s\n", ascii, url)
	} else {
		fmt.Fprintf(os.Stderr, "No se pudo generar el QR: %v\n", err)
	}

	fmt.Fprintf(os.Stderr, "Servidor escuchando en %s ...\n", url)
	http.ListenAndServe(":8080", nil)
}

package main

import (
	"bytes"
	"encoding/base64"
	"image"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	_ "image/gif"
	"image/jpeg"
	"image/png"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/rand"
)

// Fungsi untuk mensimulasikan pengunduhan gambar dari MinIO
func simulateMinioDownload(imageID int) ([]byte, error) {
	// Simulasi waktu download gambar dari MinIO
	time.Sleep(1 * time.Second) // Simulasi waktu download 1 detik per gambar

	// Membuat data acak untuk gambar, ukuran tetap 2 MB
	imageData := make([]byte, 3*1024*1024) // 2 MB (dalam byte)

	// Membuat gambar dengan data acak
	rand.Seed(uint64(time.Now().UnixNano()))
	for i := range imageData {
		// Mengisi data gambar dengan angka acak antara 0 hingga 255
		imageData[i] = byte(rand.Intn(256))
	}

	// Log untuk menandai bahwa gambar telah "didownload"
	log.Printf("Gambar %d berhasil didownload dari MinIO\n", imageID)

	return imageData, nil
}

// Fungsi untuk mengubah gambar menjadi Base64
func convertToBase64(imageData []byte) string {
	// Mengubah gambar menjadi string Base64
	return base64.StdEncoding.EncodeToString(imageData)
}

func downloadImages(c *gin.Context) {
	// Struktur untuk menyimpan hasil
	type ImageResult struct {
		ImageID   int    `json:"imageID"`
		Base64Str string `json:"base64Str"`
	}

	var wg sync.WaitGroup
	var base64Results []ImageResult

	// Mengunduh 10 gambar secara paralel
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(imageID int) {
			defer wg.Done()

			// Simulasi download gambar
			imageData, err := simulateMinioDownload(imageID)
			if err != nil {
				log.Println("Error downloading image:", err)
				return
			}

			// Konversi gambar ke Base64
			base64Str := convertToBase64(imageData)

			// Menyimpan hasil Base64 untuk dicetak nanti
			base64Results = append(base64Results, ImageResult{
				ImageID:   imageID,
				Base64Str: base64Str,
			})

			// Cetak ukuran Base64 dalam KB
			log.Printf("Ukuran Base64 gambar %d: %.2f KB\n", imageID, float64(len(base64Str))/1024)
		}(i)
	}

	// Menunggu semua goroutine selesai
	wg.Wait()

	sort.Slice(base64Results, func(i, j int) bool {
		return base64Results[i].ImageID < base64Results[j].ImageID
	})

	// Kirim hasil Base64 sebagai response dalam format JSON
	c.JSON(http.StatusOK, base64Results)
}

// Fungsi untuk mengompres gambar ke dalam buffer dengan menggunakan kompresi PNG
func compressImageToBuffer(img image.Image, ext string) (*bytes.Buffer, error) {
	// Membuat buffer untuk menampung hasil kompresi
	var buf bytes.Buffer

	// Jika formatnya sudah JPG, kompres langsung
	if ext == ".jpeg" || ext == ".jpg" || ext == ".png" {
		// Kompres gambar dalam format JPG
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80})
		if err != nil {
			log.Println("Error encoding JPG:", err)
			return nil, err
		}
	} else {
		// Jika bukan PNG atau JPG, kompres menggunakan PNG (default)
		encoder := png.Encoder{CompressionLevel: png.DefaultCompression}
		err := encoder.Encode(&buf, img)
		if err != nil {
			log.Println("Error encoding PNG:", err)
			return nil, err
		}
	}

	// Mengubah ukuran file dari byte ke kilobyte
	fileSizeKB := float64(len(buf.Bytes())) / 1024

	// Mencetak ukuran file gambar dalam kilobyte (KB)
	log.Printf("Ukuran file setelah di kompres: %.2f KB\n", fileSizeKB)

	log.Println("Gambar berhasil dikompres ke buffer")
	return &buf, nil
}

// Fungsi untuk mensimulasikan upload ke MinIO
func simulateMinioUpload(buf *bytes.Buffer) error {
	// Simulasi waktu upload ke MinIO
	time.Sleep(2 * time.Second) // Misalnya upload membutuhkan waktu 2 detik

	// Simulasikan sukses upload
	log.Println("Gambar berhasil di-upload ke MinIO")
	return nil
}

func uploadImage(c *gin.Context) {
	// Ambil file gambar yang diupload
	file, _ := c.FormFile("image")
	if file == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	// Mengubah ukuran file dari byte ke kilobyte
	fileSizeKB := float64(file.Size) / 1024

	// Mencetak ukuran file gambar dalam kilobyte (KB)
	log.Printf("Ukuran file yang diupload: %.2f KB\n", fileSizeKB)

	// Buka file gambar
	imgFile, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open image"})
		return
	}
	defer imgFile.Close()

	ext := filepath.Ext(file.Filename)
	ext = strings.ToLower(ext)

	allowedExtensions := []string{".jpg", ".jpeg", ".png"}

	var valid bool

	for _, allowed := range allowedExtensions {
		if ext == allowed {
			valid = true
			break
		}
	}

	if !valid {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error file extension not allowed"})
		return
	}

	maxSize := 2 * 1024 * 1024

	if file.Size > int64(maxSize) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error file too large"})
		return
	}

	// Decode gambar
	img, _, err := image.Decode(imgFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode image"})
		return
	}

	// Mulai pengukuran waktu kompresi
	startTime := time.Now()

	// Kompres gambar ke dalam buffer
	buf, err := compressImageToBuffer(img, ext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to compress image"})
		return
	}

	if err := simulateMinioUpload(buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload minio image"})
		return
	}

	// Menghitung durasi kompresi
	duration := time.Since(startTime)

	// Mencetak waktu kompresi
	log.Printf("Waktu kompresi: %v\n", duration)

	// Konversi buffer ke Base64
	base64Str := base64.StdEncoding.EncodeToString(buf.Bytes())

	// Menghitung ukuran dari string Base64
	base64Size := len(base64Str)

	// Cetak ukuran Base64 dalam KB
	log.Printf("Ukuran Base64 gambar: %.2f KB\n", float64(base64Size)/1024)

	time.Sleep(2 * time.Second)

	// Kirim base64 sebagai response
	c.JSON(http.StatusOK, base64Str)

}

func main() {
	// Setup router
	r := gin.Default()

	// Route untuk upload gambar
	r.POST("/upload", uploadImage)

	// Route untuk download gambar dan mengkonversinya ke Base64
	r.GET("/download", downloadImages)

	// Mulai server
	err := r.Run(":8080")
	if err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

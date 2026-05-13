package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	file, fileHeader, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't fetch thumbnail", err)
		return
	}

	mediaType, _, err := mime.ParseMediaType(fileHeader.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error fetching file header", err)
		return
	}

	typeSlices := strings.Split(mediaType, "/")
	fileExtension := typeSlices[1]
	if typeSlices[0] != "image" {
		respondWithError(w, http.StatusBadRequest, "Media type not supported", err)
		return
	}

	videoData, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "User is not authorized", err)
		return
	}

	key := make([]byte, 32)
	_, err = rand.Read(key)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Cannot make random bytes", err)
		return
	}

	randomStringID := base64.RawURLEncoding.EncodeToString(key)

	newThumbnailFilePath := filepath.Join(cfg.assetsRoot, randomStringID) + "." + fileExtension

	newFilePath, err := os.Create(newThumbnailFilePath)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to create file", err)
		return
	}

	io.Copy(newFilePath, file)

	thumbnailURL := fmt.Sprintf("http://localhost:%s/assets/%s.%s", cfg.port, randomStringID, fileExtension)
	videoData.ThumbnailURL = &thumbnailURL

	cfg.db.UpdateVideo(videoData)

	respondWithJSON(w, http.StatusOK, videoData)
}

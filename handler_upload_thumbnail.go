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

	const maxMemory int = 10 << 20
	err = r.ParseMultipartForm(int64(maxMemory))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't parse form file", err)
		return
	}

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Can't parse thumbnail form file", err)
		return
	}
	defer file.Close()

	mediaType := header.Header.Get("Content-Type")

	videoMetaData, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't read thumbnail file", err)
		return
	}

	if videoMetaData.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Video is not owned by user", err)
		return
	}

	contentType, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not parse content type", err)
		return
	}
	switch contentType {
	case "image/png":
	case "image/jpeg":
	default:
		respondWithError(w, http.StatusBadRequest, "File type not supported", err)
		return
	}
	extension := strings.TrimPrefix(contentType, "image/")

	titleBytes := make([]byte, 32)
	rand.Read(titleBytes)
	title := base64.RawURLEncoding.EncodeToString(titleBytes)

	thumbnailFileName := fmt.Sprintf("%s.%s", title, extension)
	thumbnailPath := filepath.Join(cfg.assetsRoot, thumbnailFileName)
	fileHandler, err := os.Create(thumbnailPath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create file", err)
		return
	}

	_, err = io.Copy(fileHandler, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't write to file", err)
		return
	}

	trimmedAssetPath := strings.TrimPrefix(cfg.assetsRoot, "./")
	thumbnailURL := fmt.Sprintf("http://localhost:%s/%s/%s", cfg.port, trimmedAssetPath, thumbnailFileName)
	videoMetaData.ThumbnailURL = &thumbnailURL

	err = cfg.db.UpdateVideo(videoMetaData)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't write to database", err)
		return
	}

	respondWithJSON(w, http.StatusOK, struct{}{})
}

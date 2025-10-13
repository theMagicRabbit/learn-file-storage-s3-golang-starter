package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
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

	videoMetaData, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't find video meta data", err)
		return
	}

	if videoMetaData.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Video is not owned by user", err)
		return
	}

	maxFileSizeBytes := 1 << 30
	http.MaxBytesReader(w, r.Body, int64(maxFileSizeBytes))
	videoData, videoHeader, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't find video file data in request", err)
		return
	}
	defer videoData.Close()

	contentType := videoHeader.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't parse mime type", err)
		return
	}

	switch mediaType {
	case "video/mp4":
	default:
		respondWithError(w, http.StatusBadRequest, "Invalid mime type", err)
		return
	}

	tempVideoFile, err := os.CreateTemp("", "tubyVideo-")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create temp file", err)
		return
	}
	defer os.Remove(tempVideoFile.Name())
	defer tempVideoFile.Close()

	_, err = io.Copy(tempVideoFile, videoData)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't write to temp file", err)
		return
	}
	
	tempVideoFile.Seek(0, io.SeekStart)

	videoAspectRatio, err := getVideoAspectRatio(tempVideoFile.Name())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to get video metadata", err)
		return
	}

	fastStartFilePath, err := processVideoForFastStart(tempVideoFile.Name())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't reprocess this file", err)
		return
	}
	defer os.Remove(fastStartFilePath)

	fastStartFile, err := os.Open(fastStartFilePath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to open fast start file", err)
		return
	}

	videoHexID := make([]byte, 32)
	rand.Read(videoHexID)
	title := base64.RawURLEncoding.EncodeToString(videoHexID)
	extension := strings.TrimPrefix(mediaType, "video/")
	videoS3Key := fmt.Sprintf("%s/%s.%s", videoAspectRatio, title, extension)
	
	params := s3.PutObjectInput {
		Bucket: &cfg.s3Bucket,
		Key: &videoS3Key,
		Body: fastStartFile,
		ContentType: &mediaType,
	}
	_, err = cfg.s3Client.PutObject(context.Background(), &params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to write to S3", err)
		return
	}

	videoAWSURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, videoS3Key)
	videoMetaData.VideoURL = &videoAWSURL
	err = cfg.db.UpdateVideo(videoMetaData)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to update database", err)
		return
	}
	respondWithJSON(w, http.StatusNoContent, "")
}

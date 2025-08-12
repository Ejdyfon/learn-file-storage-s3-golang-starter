package main

import (
	"fmt"
	"io"
	"net/http"

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

	// "thumbnail" should match the HTML form input name
	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	imgData, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error while reading file", err)
		return
	}

	vid, err := cfg.db.GetVideo(videoID)
	if vid.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Thats not your video", err)
		return
	}
	thm := thumbnail{data: imgData, mediaType: header.Header.Get("Content-Type")}
	videoThumbnails[videoID] = thm
	url := fmt.Sprintf("http://localhost:%v/api/thumbnails/%v", cfg.port, videoID)
	vid.ThumbnailURL = &url

	cfg.db.UpdateVideo(vid)
	respondWithJSON(w, http.StatusOK, vid)
}

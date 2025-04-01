package handler

import (
	"fmt"
	"net/http"
	"simple_api/database"
	"simple_api/model"
	"simple_api/utils"
	"strconv"
)

func GetBeatmaps(w http.ResponseWriter, r *http.Request) {
	jh := utils.NewJsonHandler(w, r)
	db := database.GetDatabase()

	beatmaps := []model.BeatmapSet{}

	tx := db.Find(&beatmaps)
	if err := tx.Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jh.WriteResponse(beatmaps)
}

func InsertBeatmap(w http.ResponseWriter, r *http.Request) {
	type BodyParams struct {
		URL string `json:"url"`
	}

	jh := utils.NewJsonHandler(w, r)
	db := database.GetDatabase()

	params := &BodyParams{}
	err := jh.ParseBody(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if params.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	beatmapSetId := utils.ExtractBeatmapSetIDFromURL(params.URL)
	if beatmapSetId == "" {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	setId, err := strconv.Atoi(beatmapSetId)
	if err != nil {
		http.Error(w, "Invalid beatmap set ID", http.StatusBadRequest)
		return
	}

	// Check if the beatmap already exists in the database
	var existingBeatmap model.BeatmapSet
	tx := db.Where("beatmap_set_id = ?", setId).First(&existingBeatmap)
	if tx.Error == nil {
		jh.WriteMessageWithStatus("Beatmap já existe", http.StatusConflict)
		return
	}

	beatmap_set, error := utils.GetOsuBeatmapSets(setId)
	if error != nil {
		http.Error(w, error.Error(), http.StatusInternalServerError)
		return
	}

	// Create a new Beatmap object
	newBeatmap := model.BeatmapSet{
		BeatmapSetID: setId,
		Artist:       beatmap_set.Artist,
		Title:        beatmap_set.Title,
		Mapper:       beatmap_set.Creator,
		Description:  beatmap_set.Description.Description,
	}

	// Save the new beatmap to the database
	tx = db.Create(&newBeatmap)
	if tx.Error != nil {
		http.Error(w, tx.Error.Error(), http.StatusInternalServerError)
		return
	}
	// Return the created beatmap

	jh.WriteMessageWithStatus(fmt.Sprintf("Beatmap encontrado: %s - %s by %s", beatmap_set.Artist, beatmap_set.Title, beatmap_set.Creator), http.StatusOK)
}

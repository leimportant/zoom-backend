package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
	"zoom-backend/database"
	"zoom-backend/models"
	"zoom-backend/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Zoom Backend OK!"})
}

// GetAllMeetings - List semua meeting dari DB
func GetAllMeetings(c *gin.Context) {
	rows, err := database.DB.Query("SELECT id, zoom_id, topic, agenda, start_time, duration, join_url, start_url FROM zoom_meetings")
	if err != nil {
		log.Println("DB query error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch meetings"})
		return
	}
	defer rows.Close()

	meetings := []models.Meeting{}

	for rows.Next() {
		var m models.Meeting
		err := rows.Scan(&m.ID, &m.ZoomID, &m.Topic, &m.Agenda, &m.StartTime, &m.Duration, &m.JoinURL, &m.StartURL)
		if err != nil {
			log.Println("Row scan error:", err)
			continue
		}
		meetings = append(meetings, m)
	}

	c.JSON(http.StatusOK, meetings)
}

// GetMeetingByID - Ambil meeting by ID dari DB
func GetMeetingByID(c *gin.Context) {
	id := c.Param("id")
	var m models.Meeting

	err := database.DB.QueryRow("SELECT id, zoom_id, topic, agenda, start_time, duration, join_url, start_url FROM zoom_meetings WHERE id = ?", id).
		Scan(&m.ID, &m.ZoomID, &m.Topic, &m.Agenda, &m.StartTime, &m.Duration, &m.JoinURL, &m.StartURL)

	if err != nil {
		log.Println("DB query error:", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Meeting not found"})
		return
	}

	c.JSON(http.StatusOK, m)
}

// CreateMeeting - Buat meeting di Zoom dan simpan ke DB
func CreateMeeting(c *gin.Context) {
	var input struct {
		Topic     string `json:"topic" binding:"required"`
		Agenda    string `json:"agenda"`
		StartTime string `json:"start_time" binding:"required"` // ISO8601 datetime
		Duration  int    `json:"duration" binding:"required"`   // in minutes
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token := services.GetAccessToken()

	meetingReq := map[string]interface{}{
		"topic":      input.Topic,
		"agenda":     input.Agenda,
		"type":       2, // Scheduled meeting
		"start_time": input.StartTime,
		"duration":   input.Duration,
		"settings": map[string]interface{}{
			"host_video":        true,
			"participant_video": true,
			"join_before_host":  false,
		},
	}

	bodyBytes, _ := json.Marshal(meetingReq)

	req, err := http.NewRequest("POST", os.Getenv("ZOOM_API_URL")+"/users/me/meetings", bytes.NewBuffer(bodyBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call Zoom API"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		c.JSON(resp.StatusCode, gin.H{"error": "Zoom API error"})
		return
	}

	var zoomResp struct {
		ID        int64  `json:"id"`
		JoinURL   string `json:"join_url"`
		StartURL  string `json:"start_url"`
		Topic     string `json:"topic"`
		Agenda    string `json:"agenda"`
		StartTime string `json:"start_time"`
		Duration  int    `json:"duration"`
	}

	json.NewDecoder(resp.Body).Decode(&zoomResp)

	id := uuid.New().String()

	_, err = database.DB.Exec(`INSERT INTO zoom_meetings 
        (id, zoom_id, topic, agenda, start_time, duration, join_url, start_url) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id,
		zoomResp.ID,
		zoomResp.Topic,
		zoomResp.Agenda,
		zoomResp.StartTime,
		zoomResp.Duration,
		zoomResp.JoinURL,
		zoomResp.StartURL)

	if err != nil {
		log.Println("DB insert error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save meeting"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id, "zoom_id": zoomResp.ID})
}

// UpdateMeeting - Update Zoom meeting dan DB
func UpdateMeeting(c *gin.Context) {
	id := c.Param("id") // Zoom Meeting ID langsung dari URL, misal "74372059854"

	var zoomMeetingID string
	err := database.DB.QueryRow("SELECT zoom_id FROM zoom_meetings WHERE id = ?", id).Scan(&zoomMeetingID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Meeting not found"})
		return
	}
	// Bind input JSON
	var input struct {
		Topic     string `json:"topic" binding:"required"`
		Agenda    string `json:"agenda"`
		StartTime string `json:"start_time" binding:"required"` // Harus RFC3339 format
		Duration  int    `json:"duration" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token := services.GetAccessToken() // Ambil token JWT

	zoomURL := fmt.Sprintf("https://api.zoom.us/v2/meetings/%s", zoomMeetingID)

	updateBody := map[string]interface{}{
		"topic":      input.Topic,
		"agenda":     input.Agenda,
		"start_time": input.StartTime,
		"duration":   input.Duration,
	}
	jsonBody, err := json.Marshal(updateBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal JSON"})
		return
	}

	req, err := http.NewRequest("PATCH", zoomURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create HTTP request"})
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send request to Zoom API"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("Zoom API error response (status %d): %s\n", resp.StatusCode, string(bodyBytes))
		c.JSON(resp.StatusCode, gin.H{"error": string(bodyBytes)})
		return
	}

	// Kalau perlu, update data meeting di database kamu pakai ZoomMeetingID
	_, err = database.DB.Exec(`UPDATE zoom_meetings SET topic=?, agenda=?, start_time=?, duration=? WHERE id=?`,
		input.Topic, input.Agenda, input.StartTime, input.Duration, zoomMeetingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update meeting in database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Meeting updated successfully"})
}

// DeleteMeeting - Hapus meeting dari Zoom dan DB
func DeleteMeeting(c *gin.Context) {
	id := c.Param("id")

	var zoomID int64
	err := database.DB.QueryRow("SELECT zoom_id FROM zoom_meetings WHERE id = ?", id).Scan(&zoomID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Meeting not found"})
		return
	}

	token := services.GetAccessToken()

	zoomURL := fmt.Sprintf("%s/meetings/%d", os.Getenv("ZOOM_API_URL"), zoomID)
	req, err := http.NewRequest("DELETE", zoomURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call Zoom API"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("Zoom API error response (status %d): %s\n", resp.StatusCode, string(bodyBytes))
		c.JSON(resp.StatusCode, gin.H{"error": "Failed to delete meeting from Zoom API"})
		return
	}

	_, err = database.DB.Exec("DELETE FROM zoom_meetings WHERE id = ?", id)
	if err != nil {
		log.Println("DB delete error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete meeting from database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Meeting deleted successfully"})
}

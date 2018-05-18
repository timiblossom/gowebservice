package webpojo

// UserFile represents file to frontend
type UserFile struct {
	Link     string `json:"link"`
	FileName string `json:"file_name"`
	FileID   string `json:"id"`
}

// FileIDResponse contains fake ID of uploaded file
type FileIDResponse struct {
	FileID string `json:"file_id"`
}

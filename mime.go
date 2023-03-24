package mail

import "path/filepath"

func getMimeType(filename string) string {
	switch ext := filepath.Ext(filename); ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".pdf":
		return "application/pdf"
	case ".doc", ".docx":
		return "application/msword"
	case ".xls", ".xlsx":
		return "application/vnd.ms-excel"
	case ".ppt", ".pptx":
		return "application/vnd.ms-powerpoint"
	case ".zip":
		return "application/zip"
	case ".rar":
		return "application/x-rar-compressed"
	case ".7z":
		return "application/x-7z-compressed"
	case ".txt":
		return "text/plain"
	case ".html", ".htm":
		return "text/html"
	default:
		return "application/octet-stream"
	}
}

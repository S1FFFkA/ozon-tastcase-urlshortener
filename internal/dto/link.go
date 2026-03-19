package dto

type CreateShortLinkRequest struct {
	URL string `json:"url" binding:"required,url"`
}

type CreateShortLinkResponse struct {
	ShortURL string `json:"short_url"`
}

type ResolveShortLinkResponse struct {
	OriginalURL string `json:"original_url"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

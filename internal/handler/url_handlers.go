package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/S1FFFkA/ozon-tastcase-urlshortener/internal/domain"
	"github.com/S1FFFkA/ozon-tastcase-urlshortener/internal/dto"
	"github.com/S1FFFkA/ozon-tastcase-urlshortener/internal/service"
	"github.com/gin-gonic/gin"
)

type URLHandler struct {
	urlService *service.URLService
	baseURL    string
}

func NewURLHandler(urlService *service.URLService, baseURL string) *URLHandler {
	return &URLHandler{
		urlService: urlService,
		baseURL:    strings.TrimRight(baseURL, "/"),
	}
}

func (h *URLHandler) RegisterRoutes(router *gin.Engine) {
	shrt := router.Group("/shrt")
	shrt.POST("/links", h.CreateShortLink)
	shrt.GET("/links/:code", h.GetOriginalURLByShortCode)
	shrt.GET("/:code", h.RedirectByShortCode)
}

func (h *URLHandler) CreateShortLink(c *gin.Context) {
	var request dto.CreateShortLinkRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		status, message := mapDomainError(domain.ErrBadRequest)
		c.JSON(status, dto.ErrorResponse{Error: message})
		return
	}

	link, err := h.urlService.Create(c.Request.Context(), request.URL)
	if err != nil {
		status, message := mapDomainError(err)
		c.JSON(status, dto.ErrorResponse{Error: message})
		return
	}

	c.JSON(http.StatusCreated, dto.CreateShortLinkResponse{
		ShortURL: h.buildPublicShortURL(link.ShortURL),
	})
}

func (h *URLHandler) RedirectByShortCode(c *gin.Context) {
	link, ok := h.resolveByCode(c)
	if !ok {
		return
	}

	c.Redirect(http.StatusFound, link.OriginalURL)
}

func (h *URLHandler) GetOriginalURLByShortCode(c *gin.Context) {
	link, ok := h.resolveByCode(c)
	if !ok {
		return
	}

	c.JSON(http.StatusOK, dto.ResolveShortLinkResponse{
		OriginalURL: link.OriginalURL,
	})
}

func (h *URLHandler) resolveByCode(c *gin.Context) (*domain.Link, bool) {
	code := strings.TrimSpace(c.Param("code"))
	if err := domain.ValidateShortCode(code); err != nil {
		status, message := mapDomainError(err)
		c.JSON(status, dto.ErrorResponse{Error: message})
		return nil, false
	}

	link, err := h.urlService.Resolve(c.Request.Context(), code)
	if err != nil {
		status, message := mapDomainError(err)
		c.JSON(status, dto.ErrorResponse{Error: message})
		return nil, false
	}

	return link, true
}

func (h *URLHandler) buildPublicShortURL(code string) string {
	if h.baseURL == "" {
		return "/shrt/" + code
	}

	return h.baseURL + "/shrt/" + code
}

func mapDomainError(err error) (int, string) {
	switch {
	case errors.Is(err, domain.ErrBadRequest):
		return http.StatusBadRequest, domain.ErrBadRequest.Error()
	case errors.Is(err, domain.ErrInvalidURL):
		return http.StatusBadRequest, domain.ErrInvalidURL.Error()
	case errors.Is(err, domain.ErrInvalidShortCode):
		return http.StatusBadRequest, domain.ErrInvalidShortCode.Error()
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, domain.ErrNotFound.Error()
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict, domain.ErrConflict.Error()
	case errors.Is(err, domain.ErrAlreadyExists):
		return http.StatusConflict, domain.ErrAlreadyExists.Error()
	default:
		return http.StatusInternalServerError, domain.ErrInternal.Error()
	}
}

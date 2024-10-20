package controllers

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"net/http"
	"os"
	"pastebin/internal/models"
	"pastebin/internal/services/clean"
)

type StorageProvider interface {
	SaveText(ctx context.Context, name, hashedUrl string) error
	ProvideHashedUrl(ctx context.Context, name string) (string, error)
	CheckNamePresence(ctx context.Context, name string) (bool, error)
	RemoveRecord(ctx context.Context, name string)
}

type RedisProvider interface {
	CheckHashedUrl(name string) (string, error)
	SaveHashedUrl(name, hashedUrl string) error
}

type ObjectStorageProvider interface {
	AddText(id, text string) error
	Text(textUrl string) (string, error)
}

type HashClientProvider interface {
	HashUrl(ctx context.Context, textUrl string) (string, error)
}

type Handler struct {
	log                   *slog.Logger
	ctx                   context.Context
	storageProvider       StorageProvider
	redisProvider         RedisProvider
	objectStorageProvider ObjectStorageProvider
}

func New(
	ctx context.Context,
	storageProvider StorageProvider,
	redisProvider RedisProvider,
	objectStorageProvider ObjectStorageProvider,
) *Handler {
	return &Handler{
		ctx:                   ctx,
		storageProvider:       storageProvider,
		redisProvider:         redisProvider,
		objectStorageProvider: objectStorageProvider,
	}
}

func (h *Handler) CreateText() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)

		const op = "controllers.CreateText"

		log.With(
			slog.String("op", op),
		)

		log.Info("starting to add a new block of text")

		var req *models.NText

		if err := c.Bind(&req); err != nil {
			c.IndentedJSON(http.StatusBadRequest, fmt.Errorf("%s:%w", op, err))
			return
		}

		// TODO: Checking if provided name already exists already exists

		log.Info("checking if the name already exists")

		ok, err := h.storageProvider.CheckNamePresence(h.ctx, req.Name)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}

		if ok {
			c.IndentedJSON(http.StatusBadRequest, "this name already exists")
			return
		}

		pass := uuid.New().String()[:8]

		log.Info("successfully hashed text url")

		hashedUrl, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
		}
		// TODO: add text url and hashed text url to the database
		if err := h.storageProvider.SaveText(h.ctx, req.Name, clean.CleanHahsedUrl(hashedUrl)); err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}

		log.Info("successfully added new message's credentials to the database")

		log.Info("starting to add the text to the object storage")

		// TODO: add text block to object storage
		if err := h.objectStorageProvider.AddText(clean.CleanHahsedUrl(hashedUrl), req.Msg); err != nil {
			h.storageProvider.RemoveRecord(h.ctx, req.Name)
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}

		log.Info("successfully added the message to the object storage")

		// TODO: return hashed url to the user
		c.IndentedJSON(http.StatusCreated, gin.H{"your url": pass})
	}
}

func (h *Handler) OpenText() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)

		const op = "controllers.OpenText"

		log.With(
			slog.String("op", op),
		)

		log.Info("starting to return text by name and url")

		var rText *models.RText

		if err := c.Bind(&rText); err != nil {
			c.IndentedJSON(http.StatusBadRequest, "try again")
			return
		}

		log.Info("successfully received url")

		hashedUrl, err := h.redisProvider.CheckHashedUrl(rText.Name)
		if err != nil {
			if errors.Is(err, fmt.Errorf("not found")) {
				c.IndentedJSON(http.StatusBadRequest, "wrong credentials")
				return
			}
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}

		log.Info("checking if hashed url exists in cache")

		if hashedUrl != "" {
			log.Info("hashed url exists in cache")
			log.Info("starting to get the body")

			if err := bcrypt.CompareHashAndPassword([]byte(hashedUrl), []byte(rText.Url)); err != nil {
				c.IndentedJSON(http.StatusBadRequest, "wrong credentials")
				return
			}

			text, err := h.objectStorageProvider.Text(hashedUrl)
			if err != nil {
				c.IndentedJSON(http.StatusInternalServerError, err)
				return
			}

			log.Info("successfully read the body")

			c.IndentedJSON(http.StatusOK, text)
			return
		}

		log.Info("hashed url doesn't exist in cache")
		log.Info("trying to get text url from the storage")

		hashedUrl, err = h.storageProvider.ProvideHashedUrl(h.ctx, rText.Name)
		if err := bcrypt.CompareHashAndPassword([]byte(hashedUrl), []byte(rText.Url)); err != nil {
			c.IndentedJSON(http.StatusBadRequest, "wrong credentials")
			return
		}

		if err != nil {
			if hashedUrl == "" {
				c.IndentedJSON(http.StatusBadRequest, "text with provided name does not exist")
				return
			}

			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}

		log.Info("successfully retrieved text Url from the storage")

		log.Info("trying to get the body from the object storage")

		text, err := h.objectStorageProvider.Text(clean.CleanHahsedUrl([]byte(hashedUrl)))
		if err != nil {
			log.Info(hashedUrl, err.Error())
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}

		log.Info("successfully received the body of the text")

		if err := h.redisProvider.SaveHashedUrl(rText.Name, hashedUrl); err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}

		log.Info("saved hashed url in cache")

		c.IndentedJSON(http.StatusOK, text)
	}
}

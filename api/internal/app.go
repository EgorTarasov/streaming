package internal

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/EgorTarasov/streaming/api/internal/config"
	"github.com/EgorTarasov/streaming/api/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func Run(ctx context.Context) error {
	app := fiber.New()

	cfg := config.MustNew()
	log.Info().Interface("cfg", cfg)
	client := service.New(cfg.Service.Host, cfg.Service.Port)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Post("/start/video", func(c *fiber.Ctx) error {
		multiPartFile, err := c.FormFile("video")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		file, err := multiPartFile.Open()
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		defer file.Close()

		buffer, err := io.ReadAll(file)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		jobID, err := client.StartProcessingVideoFile(ctx, buffer)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.JSON(fiber.Map{"job_id": jobID})
	})

	app.Post("/start/stream", func(c *fiber.Ctx) error {
		rtspUrl := c.Query("url")
		if rtspUrl == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "url is required",
			})
		}
		title := c.Query("title")
		if title == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "title is required",
			})
		}
		jobID, err := client.StartProcessingStream(ctx, rtspUrl, title)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.JSON(fiber.Map{"job_id": jobID})
	})

	app.Get("/status/:job_id", func(c *fiber.Ctx) error {
		rawJobID := c.Params("job_id")
		jobId, err := strconv.ParseInt(rawJobID, 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		status, err := client.GetProcessingStatus(ctx, jobId)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.JSON(fiber.Map{"status": status})
	})

	app.Post("/stop/:job_id", func(c *fiber.Ctx) error {
		rawJobID := c.Params("job_id")
		jobId, err := strconv.ParseInt(rawJobID, 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		err = client.StopProcessing(ctx, jobId)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.SendStatus(fiber.StatusOK)
	})

	if err := app.Listen(fmt.Sprintf(":%d", cfg.Port)); err != nil {
		return err
	}

	return nil
}

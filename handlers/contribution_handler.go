package handlers

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gofiber/fiber/v2"
	"github.com/resendlabs/resend-go"
)

type Contribution struct {
	StartStation         string   `form:"startStation"`
	EndStation           string   `form:"endStation"`
	IntermediateStations []string `form:"intermediateStations"`
	Price                float64  `form:"price"`
	Notes                string   `form:"notes"`
}

func HandleContribution(c *fiber.Ctx) error {
	// Get form values
	startStation := c.FormValue("startStation")
	endStation := c.FormValue("endStation")
	price := c.FormValue("price")
	notes := c.FormValue("notes")

	// Check required fields
	if startStation == "" || endStation == "" || price == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Start station, end station, and price are required",
		})
	}

	// Handle image uploads
	var startStationImageURL, endStationImageURL string
	var intermediateImageURLs []string

	// Handle file uploads
	if startStationFile, err := c.FormFile("startStationImage"); err == nil && startStationFile != nil {
		url, err := uploadImage(startStationFile)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to upload start station image: %v", err),
			})
		}
		startStationImageURL = url
	}

	if endStationFile, err := c.FormFile("endStationImage"); err == nil && endStationFile != nil {
		url, err := uploadImage(endStationFile)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to upload end station image: %v", err),
			})
		}
		endStationImageURL = url
	}

	// Handle intermediate station images
	form, err := c.MultipartForm()
	if err == nil {
		for key, files := range form.File {
			if strings.HasPrefix(key, "intermediateStationImage") && len(files) > 0 {
				url, err := uploadImage(files[0])
				if err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"error": fmt.Sprintf("Failed to upload intermediate station image: %v", err),
					})
				}
				intermediateImageURLs = append(intermediateImageURLs, url)
			}
		}
	}

	// Get intermediate stations
	var intermediateStations []string
	form, err = c.MultipartForm()
	if err == nil {
		for key, values := range form.Value {
			if strings.HasPrefix(key, "intermediateStation") && len(values) > 0 {
				intermediateStations = append(intermediateStations, values[0])
			}
		}
	}

	// Create email body
	var emailBody strings.Builder
	emailBody.WriteString("<h2>New Route Contribution</h2>")
	emailBody.WriteString(fmt.Sprintf("<p><strong>Start Station:</strong> %s</p>", startStation))
	emailBody.WriteString(fmt.Sprintf("<p><strong>End Station:</strong> %s</p>", endStation))
	emailBody.WriteString(fmt.Sprintf("<p><strong>Price:</strong> %s Birr</p>", price))

	if len(intermediateStations) > 0 {
		emailBody.WriteString("<h3>Intermediate Stations:</h3>")
		for i, station := range intermediateStations {
			emailBody.WriteString(fmt.Sprintf("<p>%d. %s</p>", i+1, station))
		}
	}

	if notes != "" {
		emailBody.WriteString(fmt.Sprintf("<p><strong>Notes:</strong> %s</p>", notes))
	}

	if startStationImageURL != "" || endStationImageURL != "" || len(intermediateImageURLs) > 0 {
		emailBody.WriteString("<h3>Images:</h3>")
		if startStationImageURL != "" {
			emailBody.WriteString(fmt.Sprintf(`
				<div>
					<p><strong>Start Station Image:</strong></p>
					<img src="%s" alt="Start Station" style="max-width: 500px;" />
					<p><a href="%s" target="_blank">View Start Station Image</a></p>
				</div>
			`, startStationImageURL, startStationImageURL))
		}
		if endStationImageURL != "" {
			emailBody.WriteString(fmt.Sprintf(`
				<div>
					<p><strong>End Station Image:</strong></p>
					<img src="%s" alt="End Station" style="max-width: 500px;" />
					<p><a href="%s" target="_blank">View End Station Image</a></p>
				</div>
			`, endStationImageURL, endStationImageURL))
		}
		for i, url := range intermediateImageURLs {
			emailBody.WriteString(fmt.Sprintf(`
				<div>
					<p><strong>Intermediate Station Image %d:</strong></p>
					<img src="%s" alt="Intermediate Station %d" style="max-width: 500px;" />
					<p><a href="%s" target="_blank">View Intermediate Station Image %d</a></p>
				</div>
			`, i+1, url, i+1, url, i+1))
		}
	}

	// Send email using Resend
	client := resend.NewClient(os.Getenv("RESEND_API_KEY"))

	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Admin email not configured",
		})
	}

	params := &resend.SendEmailRequest{
		From:    "Redat Contributions <onboarding@resend.dev>",
		To:      []string{adminEmail},
		Subject: fmt.Sprintf("New Route Contribution: %s to %s", startStation, endStation),
		Html:    emailBody.String(),
	}

	_, err = client.Emails.Send(params)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to send email: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Contribution received successfully",
	})
}

func uploadImage(file *multipart.FileHeader) (string, error) {
	// Initialize Cloudinary
	cloudinaryURL := os.Getenv("CLOUDINARY_URL")
	if cloudinaryURL == "" {
		return "", fmt.Errorf("CLOUDINARY_URL not configured")
	}

	cld, err := cloudinary.NewFromURL(cloudinaryURL)
	if err != nil {
		return "", fmt.Errorf("failed to initialize Cloudinary: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Open the file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer src.Close()

	// Upload to Cloudinary directly from the file
	uploadResult, err := cld.Upload.Upload(ctx, src, uploader.UploadParams{
		Folder: "redat-contributions",
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to Cloudinary: %v", err)
	}

	fmt.Printf("Image uploaded successfully to: %s\n", uploadResult.SecureURL)
	return uploadResult.SecureURL, nil
}

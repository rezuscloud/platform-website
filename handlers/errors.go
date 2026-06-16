package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// ErrorHandler renders styled error pages for 404/500 responses.
func ErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	c.Status(code)
	c.Set("Content-Type", "text/html; charset=utf-8")
	c.Set("Cache-Control", "no-cache")

	switch code {
	case fiber.StatusNotFound:
		return c.SendString(errorPageHTML(
			"404",
			"Page Not Found",
			"The requested page does not exist. Try heading back to the homepage.",
		))
	default:
		return c.SendString(errorPageHTML(
			"500",
			"Server Error",
			"Something went wrong. Try again in a moment.",
		))
	}
}

func errorPageHTML(code, title, message string) string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8"/>
	<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
	<title>` + code + " " + title + ` | RezusCloud</title>
	<link rel="icon" href="/assets/img/icon.svg" type="image/svg+xml"/>
	<style>
		*{margin:0;padding:0;box-sizing:border-box}
		body{font-family:system-ui,-apple-system,sans-serif;min-height:100vh;display:flex;align-items:center;justify-content:center;background:oklch(99.5% 0.004 85);color:oklch(14% 0.01 85)}
		@media(prefers-color-scheme:dark){body{background:oklch(6% 0.005 270);color:oklch(88% 0.005 270)}}
		.container{text-align:center;padding:2rem}
		.code{font-size:6rem;font-weight:800;letter-spacing:-0.05em;line-height:1}
		.title{font-size:1.25rem;font-weight:600;margin-top:1rem}
		.message{font-size:0.875rem;margin-top:0.5rem;opacity:0.6}
		a{display:inline-block;margin-top:2rem;color:oklch(78% 0.16 75);text-decoration:none;font-weight:600;border-bottom:1px solid currentColor}
		@media(prefers-color-scheme:dark){a{color:oklch(60% 0.08 170)}}
	</style>
</head>
<body>
	<div class="container">
		<div class="code">` + code + `</div>
		<div class="title">` + title + `</div>
		<div class="message">` + message + `</div>
		<a href="/">Return to rezus.cloud</a>
	</div>
</body>
</html>`
}

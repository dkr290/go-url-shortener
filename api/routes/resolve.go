package routes

import (
	"github.com/dkr290/go-url-shortener/database"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
)

func ResolveURL(c *fiber.Ctx) error {
	// resolve the real url from redis
	url := c.Params("url")
	dbClient := database.CreateClient(0)
	defer dbClient.Close()

	// get key value from redis and check if it connects
	v, err := dbClient.Get(database.Ctx, url).Result()
	if err == redis.Nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "short not found in the database"})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "cannot connect to the DB"})
	}
	clientRdr := database.CreateClient(1)
	defer clientRdr.Close()

	_ = clientRdr.Incr(database.Ctx, "counter")

	return c.Redirect(v, 301)

}

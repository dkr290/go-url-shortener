package routes

import (
	"os"
	"strconv"
	"time"

	"github.com/dkr290/go-url-shortener/database"
	"github.com/dkr290/go-url-shortener/helpers"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
)

type Request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}

type Response struct {
	URL             string        `json:"url"`
	CustomShort     string        `json:"short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate_limit"`
	XRateLimitReset time.Duration `json:"rate_limit_reset"`
}

func ShortenURL(ctx *fiber.Ctx) error {

	body := new(Request)

	if err := ctx.BodyParser(&body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse json"})
	}

	//implement rate limiting, if the ip is in the database means user is already there
	//allow the ip for 30 min to call 10 times

	rCl := database.CreateClient(1)
	defer rCl.Close()
	//checking if the IP as a key in redis database exists
	v, err := rCl.Get(database.Ctx, ctx.IP()).Result()
	//if there is not ip set it as key for the url (tracking the client ip woth the quota from env and expiry time for quota 30 min)
	if err == redis.Nil {
		err = rCl.Set(database.Ctx, ctx.IP(), os.Getenv("API_QUOTA"), 30*60*time.Second).Err()

		if err != nil {
			panic(err)
		}

	} else {
		v, _ = rCl.Get(database.Ctx, ctx.IP()).Result()
		vInt, err := strconv.Atoi(v)
		if err != nil {
			panic(err)
		}
		if vInt <= 0 {
			limit, _ := rCl.TTL(database.Ctx, ctx.IP()).Result()
			return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error":           "rate limit is exceeded",
				"rate_limit_rest": limit / time.Nanosecond / time.Minute,
			})
		}

	}

	//check if the input is an actual URL

	if !govalidator.IsURL(body.URL) {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid url"})
	}
	//check for domain error

	if !helpers.RemoveDomainError(body.URL) {
		return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "your account does not exists"})
	}

	//enforce https, SSL

	body.URL = helpers.EnforceHTTPS(body.URL)

	rCl.Decr(database.Ctx, ctx.IP())
}

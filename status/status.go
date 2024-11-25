package status

import "github.com/gofiber/fiber/v2"

func InternalServerError(c *fiber.Ctx, err error) error {
	message := "server error"
	if err != nil {
		message = err.Error()
	}

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"message": message,
	})
}

func BadRequest(c *fiber.Ctx, err error) error {
	message := "bad request"
	if err != nil {
		message = err.Error()
	}

	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"message": message,
	})
}

func Unauthorized(c *fiber.Ctx, err error) error {
	message := "unauthorized"
	if err != nil {
		message = err.Error()
	}

	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"message": message,
	})
}

func Ok(c *fiber.Ctx, content interface{}) error {
	var data interface{} = &fiber.Map{
		"status": "ok",
	}
	if content != nil {
		data = content
	}

	return c.Status(fiber.StatusOK).JSON(data)
}

func Created(c *fiber.Ctx, content interface{}) error {
	var data interface{} = &fiber.Map{}
	if content != nil {
		data = content
	}

	return c.Status(fiber.StatusCreated).JSON(data)
}

package webhook

func HandleWebhook(body Body) error {
	if body.Object != DEPOSIT_OBJ {
		return nil
	}

	// TODO: complete the implementation

	return nil
}

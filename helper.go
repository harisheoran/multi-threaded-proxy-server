package main

func (app *app) handleError(message string, err error) {
	if err != nil {
		app.errorLogger.Printf("%s: %s\n", message, err)
	}
}

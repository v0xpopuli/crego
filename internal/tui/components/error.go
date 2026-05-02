package components

func ErrorPanel(styles Styles, err error) string {
	if err == nil {
		return ""
	}
	return styles.Error.Render("Error: " + err.Error())
}

package components

func Preview(styles Styles, body string) string {
	if body == "" {
		return ""
	}
	return styles.Preview.Render(body)
}

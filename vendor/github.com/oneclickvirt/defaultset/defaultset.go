package defaultset

import "fmt"

func Red(text string) string {
	return fmt.Sprintf("\033[31m\033[01m%s\033[0m", text)
}

func Green(text string) string {
	return fmt.Sprintf("\033[32m\033[01m%s\033[0m", text)
}

func DarkGreen(text string) string {
	return fmt.Sprintf("\033[32m\033[02m%s\033[0m", text)
}

func Yellow(text string) string {
	return fmt.Sprintf("\033[33m\033[01m%s\033[0m", text)
}

func Blue(text string) string {
	return fmt.Sprintf("\033[36m\033[01m%s\033[0m", text)
}

func Purple(text string) string {
	return fmt.Sprintf("\033[35m\033[01m%s\033[0m", text)
}

func Cyan(text string) string {
	return fmt.Sprintf("\033[36m\033[01m%s\033[0m", text)
}

func White(text string) string {
	return fmt.Sprintf("\033[37m\033[01m%s\033[0m", text)
}

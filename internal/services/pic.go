package services

import "fmt"

func GetPictureLink(host, filename string) string {
	return fmt.Sprintf("%s/file/%s", host, filename)
}

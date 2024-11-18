package services

import "fmt"

func GetPictureLink(filename string) string {
	return fmt.Sprintf("/file/%s", filename)
}

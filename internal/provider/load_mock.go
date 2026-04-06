package provider

import "os"

func loadMock(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

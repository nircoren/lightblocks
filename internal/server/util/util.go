package util

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func GetFileData[T any](filePath string) (T, error) {
	file, err := os.Open(filePath)
	var model T

	if err != nil {
		fmt.Println("Error opening file:", err)
		return model, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return model, err
	}

	// Check that format of input.json is valid
	err = json.Unmarshal(data, &model)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return model, err
	}

	return model, nil

}

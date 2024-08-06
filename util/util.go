package util

import (
	"encoding/json"
	"fmt"
	"io"
	messaging "main/message"
	"os"
)

func GetFileData(filePath string, model interface{}) ([]messaging.Message, error) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return nil, err
	}

	// Check that format of input.json is valid
	var messages []messaging.Message
	err = json.Unmarshal(data, &messages)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return nil, err
	}

	return messages, nil

}

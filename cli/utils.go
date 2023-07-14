package cli

import (
	"encoding/json"
	"fmt"
)

func PrintJson(s any) {

	jsonData, err := json.MarshalIndent(s, "", "    ")
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
	}

	fmt.Println(string(jsonData))
}

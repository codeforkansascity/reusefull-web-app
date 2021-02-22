package main

import (
	"fmt"
	"log"
	"regexp"
)

/*
	General utility functions
*/

// Formats phone number strings for display
func FormatPhone(phone string) string {
	/*
		Some phone nums currently contain non-num chars.
		Once cleaning performed on submission, this filter
		probably won't be necessary.
	*/
	reg, err := regexp.Compile("[^0-9]+")
	if err != nil {
		log.Println(err)
	}
	cleanedPhone := reg.ReplaceAllString(phone, "")

	formatted := fmt.Sprintf("(%s) %s-%s", cleanedPhone[0:3], cleanedPhone[3:6], cleanedPhone[6:])
	return formatted
}
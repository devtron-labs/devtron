package util

import "strings"

func ConvertEmailToLowerCase(email string) string {
	if CheckIfAdminOrApiToken(email) {
		return email
	}
	return strings.ToLower(email)
}

func ConvertEmailsToLowerCase(emails []string) []string {
	lowerCaseEmails := make([]string, 0, len(emails))
	for _, email := range emails {
		lowerCaseEmails = append(lowerCaseEmails, ConvertEmailToLowerCase(email))
	}
	return lowerCaseEmails
}

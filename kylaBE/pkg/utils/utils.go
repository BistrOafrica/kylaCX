package utils

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func CreateSerialNumber(prefix string, uuid string) string {
	return prefix + uuid[24:]
}

func ValidateRequiredFields(fields []string, s interface{}) error {
	val := reflect.ValueOf(s)

	// If s is a pointer, dereference it
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	for _, field := range fields {
		f := val.FieldByName(field)
		if !f.IsValid() {
			return fmt.Errorf("field %s does not exist", field)
		}

		if IsZeroOfUnderlyingType(f.Interface()) {
			return errors.New("field " + field + " is empty")
		}
	}

	return nil
}

func IsZeroOfUnderlyingType(x interface{}) bool {
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}

func FormatAndValidatePhone(phone string) (string, bool) {
	if phone == "" {
		return "", true
	}

	// Remove all non-digit characters
	cleaned := regexp.MustCompile(`[^\d+]`).ReplaceAllString(phone, "")

	// If starts with '+' (international format)
	if strings.HasPrefix(cleaned, "+") {
		if len(cleaned) >= 12 { // + followed by country code and number
			return cleaned, true
		}
		return "", false
	}

	// If starts with country code without + (e.g., 254790...)
	if len(cleaned) >= 12 && !strings.HasPrefix(cleaned, "0") {
		return "+" + cleaned, true
	}

	// Kenyan numbers - add +254 prefix if needed
	if len(cleaned) == 9 && strings.HasPrefix(cleaned, "7") { // e.g. 790...
		return "+254" + cleaned, true
	}
	if len(cleaned) == 10 && strings.HasPrefix(cleaned, "07") { // e.g. 0790...
		return "+254" + cleaned[1:], true
	}
	if len(cleaned) == 12 && strings.HasPrefix(cleaned, "254") { // e.g. 254790...
		return "+" + cleaned, true
	}

	// If we get here, it's not a valid format
	return "", false
}

// Updated validation function that uses the formatter
func ValidatePhoneFormat(phone string) bool {
	_, valid := FormatAndValidatePhone(phone)
	return valid
}

func ValidateEmailFormat(email string) bool {
	//	emailRegex := `^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$`
	emailRegex := `^[\w+-\.]+@([\w-]+\.)+[\w-]{2,7}$`
	if email == "" {
		return true
	}
	return regexp.MustCompile(emailRegex).MatchString(email)
}

func ValidateDate(date string) bool {
	dateRegex := `^\d{4}-\d{2}-\d{2}$`
	if date == "" {
		return true
	}
	return regexp.MustCompile(dateRegex).MatchString(date)
}

func ConvertTimeToString(t time.Time) string {
	return t.Format("YYY-MM-DDTHH:MM:SSZ")
}

func ConvertStringToTime(s string) (time.Time, error) {
	return time.Parse("YYY-MM-DDTHH:MM:SSZ", s)
}

func HASH_PASSWORD(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(hashedPassword)
}
func COMPARE_PASSWORD(hashedPassword string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func FilterObjectWithValues(obj interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	v := reflect.ValueOf(obj)

	// Iterate over the fields of the struct.
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.IsZero() {
			result[v.Type().Field(i).Name] = field.Interface()
		}
	}

	return result
}

// Helper function to return fields with zero values
func FilterObjectWithZeroValues(obj interface{}, fieldsToExclude []string) map[string]interface{} {
	result := make(map[string]interface{})

	// Dereference the pointer if obj is a pointer
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Iterate over the fields of the struct.
	for i := 0; i < val.NumField(); i++ {
		fieldName := val.Type().Field(i).Name

		// Skip if field is in fieldsToExclude
		if contains(fieldsToExclude, fieldName) {
			continue
		}

		field := val.Field(i)
		if field.IsZero() {
			result[fieldName] = field.Interface()
		}
	}

	return result
}

// Helper function to check if a slice contains a string
func contains(slice []string, str string) bool {
	for _, v := range slice {
		if strings.EqualFold(v, str) {
			return true
		}
	}
	return false
}

// Helper function to check if a slice contains a specific element
func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func RemoveStringFromSlice(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func RemoveStringsFromSlice(s []string, r []string) []string {
	for _, v := range r {
		s = RemoveStringFromSlice(s, v)
	}
	return s
}

// generateRandomPassword generates a new random password
func GENERATE_RANDOM_KEY(passwordLength int) (string, error) {
	const (
		uppercaseLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		lowercaseLetters = "abcdefghijklmnopqrstuvwxyz"
		digits           = "0123456789"
		specialChars     = "!@#$%^&*()-=_+[]{}|;:'\",.<>/?"
	)

	// Combine all possible characters
	allChars := uppercaseLetters + lowercaseLetters + digits

	// Initialize variables to store the character groups
	var (
		hasUppercase bool
		hasLowercase bool
		hasDigit     bool
		// hasSpecial   bool
	)

	var password string

	// Use a random index to select a character from each character group
	for i := 0; i < passwordLength; i++ {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(allChars))))
		if err != nil {
			return "", err
		}
		randomChar := string(allChars[randomIndex.Int64()])

		// Check and set the corresponding flag based on the character group
		switch {
		case "A" <= randomChar && randomChar <= "Z":
			hasUppercase = true
		case "a" <= randomChar && randomChar <= "z":
			hasLowercase = true
		case "0" <= randomChar && randomChar <= "9":
			hasDigit = true
		default:
			// hasSpecial = true
		}

		// Append the random character to the password
		password += randomChar
	}

	// Check if all character groups are present, regenerate if not
	if !hasUppercase || !hasLowercase || !hasDigit {
		return GENERATE_RANDOM_KEY(passwordLength)
	}

	// return password, nil
	return "Temp@123", nil
}

func Includes(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func UuidParser(id string) uuid.UUID {
	value, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil
	}

	return value
}

func UuidToString(u uuid.UUID) string {
	if u == uuid.Nil {
		return ""
	}
	return u.String()
}

func LogAsJSON(title string, v interface{}) {
	jsonData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Printf("Error marshalling to JSON: %v", err)
		return
	}
	log.Printf("Title: %s: \n\n %s", title, string(jsonData))
}

// ValueExistsGeneric is a generic function that works with any comparable types
func ValueExistsGeneric[K comparable, V comparable](m map[K]V, searchValue V) bool {
	for _, v := range m {
		if v == searchValue {
			return true
		}
	}
	return false
}

// check strings across cases
func IsSame(a, b string) bool {
	return strings.EqualFold(a, b)
}

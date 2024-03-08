package itpg

import (
	"encoding/json"
	"fmt"
	"itpg/responses"
	"net/http"
	"slices"
	"strings"
)

// isEmptyStr checks if any of the provided strings are empty.
func isEmptyStr(w http.ResponseWriter, str ...string) (err error) {
	for _, s := range str {
		if s == "" {
			w.WriteHeader(http.StatusBadRequest)
			responses.ErrEmptyValue.WriteJSON(w)
			return responses.ErrEmptyValue
		}
	}
	return
}

// decodeCredentials decodes JSON data from the request body into a Credentials struct.
func decodeCredentials(w http.ResponseWriter, r *http.Request) (*Credentials, error) {
	var credentials Credentials
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		responses.ErrBadRequest.WriteJSON(w)
		return nil, err
	}
	return &credentials, nil
}

// decodeCredentialsReset decodes JSON data from the request body into a Credentials Reset struct.
func decodeCredentialsReset(w http.ResponseWriter, r *http.Request) (*CredentialsReset, error) {
	var credentialsReset CredentialsReset
	if err := json.NewDecoder(r.Body).Decode(&credentialsReset); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		responses.ErrBadRequest.WriteJSON(w)
		return nil, err
	}
	return &credentialsReset, nil
}

// decodeCredentialsChange decodes JSON data from the request body into a Credentials Change struct.
func decodeCredentialsChange(w http.ResponseWriter, r *http.Request) (*CredentialsChange, error) {
	var credentialsChange CredentialsChange
	if err := json.NewDecoder(r.Body).Decode(&credentialsChange); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		responses.ErrBadRequest.WriteJSON(w)
		return nil, err
	}
	return &credentialsChange, nil
}

// decodeGradeData decodes JSON data from the request body into a Grade Data struct.
func decodeGradeData(w http.ResponseWriter, r *http.Request) (*GradeData, error) {
	var gradeData GradeData
	if err := json.NewDecoder(r.Body).Decode(&gradeData); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		responses.ErrBadRequest.WriteJSON(w)
		return nil, err
	}
	return &gradeData, nil
}

// extractDomain extracts the domain part from an email address.
// It takes an email address string as input and returns the domain part.
// If the email address is in an invalid format (e.g., missing "@" symbol),
// it returns an empty string.
func extractDomain(email string) (domain string, err error) {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return domain, fmt.Errorf("invalid email format")
	}
	return parts[1], err
}

// validAllowedDomains checks if the list of allowed mail domains is empty.
func validAllowedDomains(domains []string) (err error) {
	if len(domains) == 0 {
		return fmt.Errorf("got empty list of allowed mail domains")
	}
	return
}

// checkDomainAllowed checks if the given domain is allowed based on the list of allowed mail domains.
func checkDomainAllowed(domain string) (err error) {
	if AllowedMailDomains[0] == "*" {
		return
	}
	if !slices.Contains(AllowedMailDomains, domain) {
		return responses.ErrEmailDomainNotAllowed
	}
	return
}

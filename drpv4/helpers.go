package drpv4

import "math/rand"

// expandStringList converts a interface{} to a []string
func expandStringList(v interface{}) []string {
	if v == nil {
		return nil
	}
	result := make([]string, len(v.([]interface{})))
	for i, s := range v.([]interface{}) {
		result[i] = s.(string)
	}
	return result
}

// randomString generates a random string of length n
func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

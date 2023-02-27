package drpv4

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

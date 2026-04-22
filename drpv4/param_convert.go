package drpv4

import (
	"encoding/json"
	"fmt"

	"gitlab.com/rackn/provision/v4/models"
)

func getParam(c *Config, name string) (*models.Param, error) {
	var p *models.Param
	if err := c.session.Req().UrlFor("params", name).Do(&p); err != nil {
		return nil, err
	}
	return p, nil
}

func getParamSchemaType(c *Config, name string) string {
	param, err := getParam(c, name)
	if err != nil {
		return ""
	}
	if param.Schema == nil {
		return ""
	}
	sm, ok := param.Schema.(map[string]interface{})
	if !ok || sm["type"] == nil {
		return ""
	}
	s, ok := sm["type"].(string)
	if !ok {
		return ""
	}
	return s
}

func convertParamToType(c *Config, name string, value string) (interface{}, error) {
	paramType := getParamSchemaType(c, name)
	switch paramType {
	case "string":
		return value, nil
	default:
		var out interface{}
		if err := json.Unmarshal([]byte(value), &out); err != nil {
			return value, nil
		}
		return out, nil
	}
}

func convertParamToString(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	default:
		out, err := json.Marshal(v)
		return string(out), err
	}
}

func isParamSecure(c *Config, name string) bool {
	res, err := c.session.GetModel("params", name)
	if err != nil {
		return false
	}
	param := res.(*models.Param)
	return param.Secure
}

func getPublicKey(c *Config, profile string) ([]byte, error) {
	var pubkey []byte
	if err := c.session.Req().UrlFor("profiles", profile, "pubkey").Do(&pubkey); err != nil {
		return nil, err
	}
	return pubkey, nil
}

func defaultParamSchema() map[string]interface{} {
	return map[string]interface{}{"type": "string"}
}

func stringMapToInterfaceMap(m map[string]string) map[string]interface{} {
	if m == nil {
		return nil
	}
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func interfaceMapToStringMap(m map[string]interface{}) (map[string]string, error) {
	if m == nil {
		return nil, nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		switch t := v.(type) {
		case string:
			out[k] = t
		default:
			b, err := json.Marshal(t)
			if err != nil {
				return nil, fmt.Errorf("param schema key %q: %w", k, err)
			}
			out[k] = string(b)
		}
	}
	return out, nil
}

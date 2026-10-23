//go:build !js

package postcode

import "encoding/json"

// UnmarshalJSON handles both the official nested envelope {"error": {"code": ..., "message": ...}}
// and flat payloads {"code": ..., "message": ...}.
func (e *APIError) UnmarshalJSON(data []byte) error {
	var nested struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(data, &nested); err == nil && (nested.Error.Code != "" || nested.Error.Message != "") {
		e.Code = nested.Error.Code
		e.Message = nested.Error.Message
		return nil
	}

	var flat struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(data, &flat); err == nil {
		e.Code = flat.Code
		e.Message = flat.Message
	}
	return nil
}

// UnmarshalJSON supports both a direct slice of units `[...]` (returned by live NIPOST Gateway)
// and an object payload `{"results": [...]}` (returned by simulators/proxies).
func (r *NearbyResponse) UnmarshalJSON(data []byte) error {
	var units []NearbyUnit
	if err := json.Unmarshal(data, &units); err == nil {
		r.Results = units
		return nil
	}

	type alias NearbyResponse
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	r.Results = a.Results
	return nil
}

func (r *referenceStatesResponse) UnmarshalJSON(data []byte) error {
	var list []NamedCode
	if err := json.Unmarshal(data, &list); err == nil {
		r.States = list
		return nil
	}
	type alias referenceStatesResponse
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	r.States = a.States
	return nil
}

func (r *referenceLGAsResponse) UnmarshalJSON(data []byte) error {
	var list []NamedCode
	if err := json.Unmarshal(data, &list); err == nil {
		r.LGAs = list
		return nil
	}
	type alias referenceLGAsResponse
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	r.LGAs = a.LGAs
	return nil
}

func (r *referenceDistrictsResponse) UnmarshalJSON(data []byte) error {
	var list []NamedCode
	if err := json.Unmarshal(data, &list); err == nil {
		r.Districts = list
		return nil
	}
	type alias referenceDistrictsResponse
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	r.Districts = a.Districts
	return nil
}

func (r *referenceAreasResponse) UnmarshalJSON(data []byte) error {
	var list []NamedCode
	if err := json.Unmarshal(data, &list); err == nil {
		r.Areas = list
		return nil
	}
	type alias referenceAreasResponse
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	r.Areas = a.Areas
	return nil
}

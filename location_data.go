package postcode

import (
	"cmp"
	"slices"
	"strings"
	"sync"
)

// StateRecord holds reference coordinates and metadata for a Nigerian state.
type StateRecord struct {
	Code      string
	Name      string
	Capital   string
	Latitude  float64
	Longitude float64
	Zone      string
}

// LGARecord holds reference coordinates and names for an LGA within a state.
type LGARecord struct {
	StateCode string
	LGACode   string
	Name      string
	Latitude  float64
	Longitude float64
}

// BuildingRecord holds high-precision point geometry for a specific building unit.
type BuildingRecord struct {
	Postcode  string // Canonical (e.g. "EK-01-A03-FK-01") or Compact ("EK01A03FK01")
	Latitude  float64
	Longitude float64
	Address   string
	StateCode string
	StateName string
	LGACode   string
	LGAName   string
	Zone      string
}

// NigerianStates contains geographic centroids and administrative capitals for all 36 States and FCT.
var NigerianStates = map[string]StateRecord{
	"AB": {Code: "AB", Name: "Abia", Capital: "Umuahia", Latitude: 5.5320, Longitude: 7.4860, Zone: "SOUTH EAST"},
	"AD": {Code: "AD", Name: "Adamawa", Capital: "Yola", Latitude: 9.2094, Longitude: 12.4818, Zone: "NORTH EAST"},
	"AK": {Code: "AK", Name: "Akwa Ibom", Capital: "Uyo", Latitude: 5.0377, Longitude: 7.9128, Zone: "SOUTH SOUTH"},
	"AN": {Code: "AN", Name: "Anambra", Capital: "Awka", Latitude: 6.2209, Longitude: 7.0670, Zone: "SOUTH EAST"},
	"BA": {Code: "BA", Name: "Bauchi", Capital: "Bauchi", Latitude: 10.3158, Longitude: 9.8442, Zone: "NORTH EAST"},
	"BY": {Code: "BY", Name: "Bayelsa", Capital: "Yenagoa", Latitude: 4.9267, Longitude: 6.2676, Zone: "SOUTH SOUTH"},
	"BN": {Code: "BN", Name: "Benue", Capital: "Makurdi", Latitude: 7.7304, Longitude: 8.5211, Zone: "NORTH CENTRAL"},
	"BO": {Code: "BO", Name: "Borno", Capital: "Maiduguri", Latitude: 11.8333, Longitude: 13.1500, Zone: "NORTH EAST"},
	"CR": {Code: "CR", Name: "Cross River", Capital: "Calabar", Latitude: 4.9757, Longitude: 8.3417, Zone: "SOUTH SOUTH"},
	"DE": {Code: "DE", Name: "Delta", Capital: "Asaba", Latitude: 6.1984, Longitude: 6.7329, Zone: "SOUTH SOUTH"},
	"EB": {Code: "EB", Name: "Ebonyi", Capital: "Abakaliki", Latitude: 6.3249, Longitude: 8.1137, Zone: "SOUTH EAST"},
	"ED": {Code: "ED", Name: "Edo", Capital: "Benin City", Latitude: 6.3350, Longitude: 5.6037, Zone: "SOUTH SOUTH"},
	"EK": {Code: "EK", Name: "Ekiti", Capital: "Ado Ekiti", Latitude: 7.6211, Longitude: 5.2215, Zone: "SOUTH WEST"},
	"EN": {Code: "EN", Name: "Enugu", Capital: "Enugu", Latitude: 6.4584, Longitude: 7.5464, Zone: "SOUTH EAST"},
	"FC": {Code: "FC", Name: "Federal Capital Territory", Capital: "Abuja", Latitude: 9.0579, Longitude: 7.4951, Zone: "NORTH CENTRAL"},
	"GO": {Code: "GO", Name: "Gombe", Capital: "Gombe", Latitude: 10.2897, Longitude: 11.1711, Zone: "NORTH EAST"},
	"IM": {Code: "IM", Name: "Imo", Capital: "Owerri", Latitude: 5.4858, Longitude: 7.0355, Zone: "SOUTH EAST"},
	"JI": {Code: "JI", Name: "Jigawa", Capital: "Dutse", Latitude: 11.7594, Longitude: 9.3389, Zone: "NORTH WEST"},
	"KD": {Code: "KD", Name: "Kaduna", Capital: "Kaduna", Latitude: 10.5105, Longitude: 7.4165, Zone: "NORTH WEST"},
	"KN": {Code: "KN", Name: "Kano", Capital: "Kano", Latitude: 12.0022, Longitude: 8.5920, Zone: "NORTH WEST"},
	"KT": {Code: "KT", Name: "Katsina", Capital: "Katsina", Latitude: 12.9855, Longitude: 7.6171, Zone: "NORTH WEST"},
	"KE": {Code: "KE", Name: "Kebbi", Capital: "Birnin Kebbi", Latitude: 12.4504, Longitude: 4.1999, Zone: "NORTH WEST"},
	"KO": {Code: "KO", Name: "Kogi", Capital: "Lokoja", Latitude: 7.8023, Longitude: 6.7333, Zone: "NORTH CENTRAL"},
	"KW": {Code: "KW", Name: "Kwara", Capital: "Ilorin", Latitude: 8.4966, Longitude: 4.5421, Zone: "NORTH CENTRAL"},
	"LA": {Code: "LA", Name: "Lagos", Capital: "Ikeja", Latitude: 6.6018, Longitude: 3.3515, Zone: "SOUTH WEST"},
	"NA": {Code: "NA", Name: "Nasarawa", Capital: "Lafia", Latitude: 8.4932, Longitude: 8.5153, Zone: "NORTH CENTRAL"},
	"NI": {Code: "NI", Name: "Niger", Capital: "Minna", Latitude: 9.6139, Longitude: 6.5569, Zone: "NORTH CENTRAL"},
	"OG": {Code: "OG", Name: "Ogun", Capital: "Abeokuta", Latitude: 7.1475, Longitude: 3.3619, Zone: "SOUTH WEST"},
	"ON": {Code: "ON", Name: "Ondo", Capital: "Akure", Latitude: 7.2571, Longitude: 5.2058, Zone: "SOUTH WEST"},
	"OS": {Code: "OS", Name: "Osun", Capital: "Osogbo", Latitude: 7.7827, Longitude: 4.5418, Zone: "SOUTH WEST"},
	"OY": {Code: "OY", Name: "Oyo", Capital: "Ibadan", Latitude: 7.3775, Longitude: 3.9470, Zone: "SOUTH WEST"},
	"PL": {Code: "PL", Name: "Plateau", Capital: "Jos", Latitude: 9.9271, Longitude: 8.8921, Zone: "NORTH CENTRAL"},
	"RI": {Code: "RI", Name: "Rivers", Capital: "Port Harcourt", Latitude: 4.8156, Longitude: 7.0498, Zone: "SOUTH SOUTH"},
	"SO": {Code: "SO", Name: "Sokoto", Capital: "Sokoto", Latitude: 13.0609, Longitude: 5.2476, Zone: "NORTH WEST"},
	"TA": {Code: "TA", Name: "Taraba", Capital: "Jalingo", Latitude: 8.8937, Longitude: 11.3600, Zone: "NORTH EAST"},
	"YO": {Code: "YO", Name: "Yobe", Capital: "Damaturu", Latitude: 11.7470, Longitude: 11.9608, Zone: "NORTH EAST"},
	"ZA": {Code: "ZA", Name: "Zamfara", Capital: "Gusau", Latitude: 12.1702, Longitude: 6.6592, Zone: "NORTH WEST"},
}

// knownLGAs maps StateCode + LGACode (e.g. "EK01", "LA11") to LGA metadata.
var knownLGAs = map[string]LGARecord{
	// Abia
	"AB01": {StateCode: "AB", LGACode: "01", Name: "Aba North", Latitude: 5.1156, Longitude: 7.3676},
	"AB02": {StateCode: "AB", LGACode: "02", Name: "Aba South", Latitude: 5.0982, Longitude: 7.3612},
	"AB15": {StateCode: "AB", LGACode: "15", Name: "Umuahia North", Latitude: 5.5320, Longitude: 7.4860},
	// Adamawa
	"AD01": {StateCode: "AD", LGACode: "01", Name: "Yola North", Latitude: 9.2094, Longitude: 12.4818},
	"AD02": {StateCode: "AD", LGACode: "02", Name: "Yola South", Latitude: 9.1932, Longitude: 12.4191},
	// Akwa Ibom
	"AK01": {StateCode: "AK", LGACode: "01", Name: "Eket", Latitude: 4.6423, Longitude: 7.9244},
	"AK11": {StateCode: "AK", LGACode: "11", Name: "Uyo", Latitude: 5.0377, Longitude: 7.9128},
	// Anambra
	"AN01": {StateCode: "AN", LGACode: "01", Name: "Awka South", Latitude: 6.2209, Longitude: 7.0670},
	"AN02": {StateCode: "AN", LGACode: "02", Name: "Onitsha North", Latitude: 6.1498, Longitude: 6.7857},
	"AN03": {StateCode: "AN", LGACode: "03", Name: "Nnewi North", Latitude: 6.0195, Longitude: 6.9148},
	// Bauchi
	"BA02": {StateCode: "BA", LGACode: "02", Name: "Bauchi", Latitude: 10.3158, Longitude: 9.8442},
	// Bayelsa
	"BY01": {StateCode: "BY", LGACode: "01", Name: "Yenagoa", Latitude: 4.9267, Longitude: 6.2676},
	// Benue
	"BN01": {StateCode: "BN", LGACode: "01", Name: "Makurdi", Latitude: 7.7304, Longitude: 8.5211},
	"BN02": {StateCode: "BN", LGACode: "02", Name: "Gboko", Latitude: 7.3250, Longitude: 9.0011},
	"BN03": {StateCode: "BN", LGACode: "03", Name: "Otukpo", Latitude: 7.1931, Longitude: 8.1339},
	// Borno
	"BO01": {StateCode: "BO", LGACode: "01", Name: "Maiduguri", Latitude: 11.8333, Longitude: 13.1500},
	// Cross River
	"CR01": {StateCode: "CR", LGACode: "01", Name: "Calabar Municipal", Latitude: 4.9757, Longitude: 8.3417},
	"CR02": {StateCode: "CR", LGACode: "02", Name: "Calabar South", Latitude: 4.9482, Longitude: 8.3250},
	// Delta
	"DE01": {StateCode: "DE", LGACode: "01", Name: "Oshimili South", Latitude: 6.1984, Longitude: 6.7329},
	"DE02": {StateCode: "DE", LGACode: "02", Name: "Warri South", Latitude: 5.5167, Longitude: 5.7500},
	// Ebonyi
	"EB13": {StateCode: "EB", LGACode: "13", Name: "Abakaliki", Latitude: 6.3249, Longitude: 8.1137},
	// Edo
	"ED01": {StateCode: "ED", LGACode: "01", Name: "Oredo", Latitude: 6.3350, Longitude: 5.6037},
	"ED02": {StateCode: "ED", LGACode: "02", Name: "Egor", Latitude: 6.3688, Longitude: 5.6012},
	"ED03": {StateCode: "ED", LGACode: "03", Name: "Ikpoba-Okha", Latitude: 6.3021, Longitude: 5.6431},
	// Ekiti
	"EK01": {StateCode: "EK", LGACode: "01", Name: "Ado Ekiti", Latitude: 7.6211, Longitude: 5.2215},
	// Enugu
	"EN01": {StateCode: "EN", LGACode: "01", Name: "Enugu East", Latitude: 6.4721, Longitude: 7.5210},
	"EN02": {StateCode: "EN", LGACode: "02", Name: "Enugu South", Latitude: 6.4180, Longitude: 7.4981},
	"EN05": {StateCode: "EN", LGACode: "05", Name: "Enugu North", Latitude: 6.4584, Longitude: 7.5464},
	// Federal Capital Territory (FCT) Area Councils
	"FC01": {StateCode: "FC", LGACode: "01", Name: "Abaji", Latitude: 8.4721, Longitude: 6.9537},
	"FC02": {StateCode: "FC", LGACode: "02", Name: "Bwari", Latitude: 9.1538, Longitude: 7.3220},
	"FC03": {StateCode: "FC", LGACode: "03", Name: "Abuja Municipal", Latitude: 9.0579, Longitude: 7.4951},
	"FC04": {StateCode: "FC", LGACode: "04", Name: "Gwagwalada", Latitude: 8.9431, Longitude: 7.0825},
	"FC05": {StateCode: "FC", LGACode: "05", Name: "Kuje", Latitude: 8.8797, Longitude: 7.2306},
	"FC06": {StateCode: "FC", LGACode: "06", Name: "Kwali", Latitude: 8.8842, Longitude: 7.0142},
	// Gombe
	"GO01": {StateCode: "GO", LGACode: "01", Name: "Gombe", Latitude: 10.2897, Longitude: 11.1711},
	// Imo
	"IM01": {StateCode: "IM", LGACode: "01", Name: "Owerri Municipal", Latitude: 5.4858, Longitude: 7.0355},
	"IM02": {StateCode: "IM", LGACode: "02", Name: "Owerri North", Latitude: 5.5121, Longitude: 7.0580},
	"IM03": {StateCode: "IM", LGACode: "03", Name: "Owerri West", Latitude: 5.4620, Longitude: 6.9920},
	// Jigawa
	"JI24": {StateCode: "JI", LGACode: "24", Name: "Dutse", Latitude: 11.7594, Longitude: 9.3389},
	// Kaduna
	"KD01": {StateCode: "KD", LGACode: "01", Name: "Kaduna North", Latitude: 10.5105, Longitude: 7.4165},
	"KD02": {StateCode: "KD", LGACode: "02", Name: "Kaduna South", Latitude: 10.4611, Longitude: 7.4192},
	"KD03": {StateCode: "KD", LGACode: "03", Name: "Chikun", Latitude: 10.4210, Longitude: 7.3812},
	"KD04": {StateCode: "KD", LGACode: "04", Name: "Zaria", Latitude: 11.0855, Longitude: 7.7199},
	// Kano
	"KN01": {StateCode: "KN", LGACode: "01", Name: "Fagge", Latitude: 12.0080, Longitude: 8.5310},
	"KN02": {StateCode: "KN", LGACode: "02", Name: "Nassarawa", Latitude: 12.0110, Longitude: 8.5620},
	"KN03": {StateCode: "KN", LGACode: "03", Name: "Tarauni", Latitude: 11.9760, Longitude: 8.5490},
	"KN31": {StateCode: "KN", LGACode: "31", Name: "Kano Municipal", Latitude: 12.0022, Longitude: 8.5920},
	// Katsina
	"KT01": {StateCode: "KT", LGACode: "01", Name: "Katsina", Latitude: 12.9855, Longitude: 7.6171},
	// Kebbi
	"KE01": {StateCode: "KE", LGACode: "01", Name: "Birnin Kebbi", Latitude: 12.4504, Longitude: 4.1999},
	// Kogi
	"KO01": {StateCode: "KO", LGACode: "01", Name: "Lokoja", Latitude: 7.8023, Longitude: 6.7333},
	// Kwara
	"KW01": {StateCode: "KW", LGACode: "01", Name: "Ilorin West", Latitude: 8.4966, Longitude: 4.5421},
	"KW02": {StateCode: "KW", LGACode: "02", Name: "Ilorin East", Latitude: 8.5120, Longitude: 4.5810},
	"KW03": {StateCode: "KW", LGACode: "03", Name: "Ilorin South", Latitude: 8.4680, Longitude: 4.5680},
	// All 20 Local Government Areas (LGAs) of Lagos State
	"LA01": {StateCode: "LA", LGACode: "01", Name: "Agege", Latitude: 6.6180, Longitude: 3.3209},
	"LA02": {StateCode: "LA", LGACode: "02", Name: "Ajeromi-Ifelodun", Latitude: 6.4554, Longitude: 3.3342},
	"LA03": {StateCode: "LA", LGACode: "03", Name: "Alimosho", Latitude: 6.6091, Longitude: 3.2561},
	"LA04": {StateCode: "LA", LGACode: "04", Name: "Amuwo-Odofin", Latitude: 6.4312, Longitude: 3.2842},
	"LA05": {StateCode: "LA", LGACode: "05", Name: "Apapa", Latitude: 6.4484, Longitude: 3.3639},
	"LA06": {StateCode: "LA", LGACode: "06", Name: "Badagry", Latitude: 6.4316, Longitude: 2.8876},
	"LA07": {StateCode: "LA", LGACode: "07", Name: "Epe", Latitude: 6.5841, Longitude: 3.9834},
	"LA08": {StateCode: "LA", LGACode: "08", Name: "Eti-Osa", Latitude: 6.4584, Longitude: 3.5684},
	"LA09": {StateCode: "LA", LGACode: "09", Name: "Ibeju-Lekki", Latitude: 6.4862, Longitude: 3.8643},
	"LA10": {StateCode: "LA", LGACode: "10", Name: "Ifako-Ijaiye", Latitude: 6.6713, Longitude: 3.3082},
	"LA11": {StateCode: "LA", LGACode: "11", Name: "Ikeja", Latitude: 6.6018, Longitude: 3.3515},
	"LA12": {StateCode: "LA", LGACode: "12", Name: "Ikorodu", Latitude: 6.6194, Longitude: 3.5105},
	"LA13": {StateCode: "LA", LGACode: "13", Name: "Kosofe", Latitude: 6.5742, Longitude: 3.3941},
	"LA14": {StateCode: "LA", LGACode: "14", Name: "Lagos Island", Latitude: 6.4549, Longitude: 3.4246},
	"LA15": {StateCode: "LA", LGACode: "15", Name: "Lagos Mainland", Latitude: 6.4969, Longitude: 3.3776},
	"LA16": {StateCode: "LA", LGACode: "16", Name: "Mushin", Latitude: 6.5298, Longitude: 3.3515},
	"LA17": {StateCode: "LA", LGACode: "17", Name: "Ojo", Latitude: 6.4684, Longitude: 3.1924},
	"LA18": {StateCode: "LA", LGACode: "18", Name: "Oshodi-Isolo", Latitude: 6.5388, Longitude: 3.3276},
	"LA19": {StateCode: "LA", LGACode: "19", Name: "Shomolu", Latitude: 6.5385, Longitude: 3.3831},
	"LA20": {StateCode: "LA", LGACode: "20", Name: "Surulere", Latitude: 6.4969, Longitude: 3.3565},
	// Nasarawa
	"NA01": {StateCode: "NA", LGACode: "01", Name: "Lafia", Latitude: 8.4932, Longitude: 8.5153},
	"NA02": {StateCode: "NA", LGACode: "02", Name: "Karu", Latitude: 9.0067, Longitude: 7.6481},
	// Niger
	"NI01": {StateCode: "NI", LGACode: "01", Name: "Bida", Latitude: 9.0804, Longitude: 6.0100},
	"NI02": {StateCode: "NI", LGACode: "02", Name: "Suleja", Latitude: 9.1806, Longitude: 7.1794},
	"NI09": {StateCode: "NI", LGACode: "09", Name: "Chanchaga", Latitude: 9.6139, Longitude: 6.5569},
	// Ogun
	"OG01": {StateCode: "OG", LGACode: "01", Name: "Abeokuta North", Latitude: 7.1601, Longitude: 3.3321},
	"OG02": {StateCode: "OG", LGACode: "02", Name: "Ado-Odo/Ota", Latitude: 6.6841, Longitude: 3.2320},
	"OG14": {StateCode: "OG", LGACode: "14", Name: "Abeokuta South", Latitude: 7.1475, Longitude: 3.3619},
	// Ondo
	"ON01": {StateCode: "ON", LGACode: "01", Name: "Akure South", Latitude: 7.2571, Longitude: 5.2058},
	"ON02": {StateCode: "ON", LGACode: "02", Name: "Akure North", Latitude: 7.3020, Longitude: 5.2510},
	// Osun
	"OS01": {StateCode: "OS", LGACode: "01", Name: "Osogbo", Latitude: 7.7827, Longitude: 4.5418},
	"OS02": {StateCode: "OS", LGACode: "02", Name: "Ife Central", Latitude: 7.4905, Longitude: 4.5521},
	// Oyo
	"OY01": {StateCode: "OY", LGACode: "01", Name: "Ibadan North", Latitude: 7.4018, Longitude: 3.9110},
	"OY02": {StateCode: "OY", LGACode: "02", Name: "Ibadan North-East", Latitude: 7.3880, Longitude: 3.9320},
	"OY03": {StateCode: "OY", LGACode: "03", Name: "Ibadan North-West", Latitude: 7.3860, Longitude: 3.8790},
	"OY04": {StateCode: "OY", LGACode: "04", Name: "Ibadan South-East", Latitude: 7.3520, Longitude: 3.9180},
	"OY05": {StateCode: "OY", LGACode: "05", Name: "Ibadan South-West", Latitude: 7.3610, Longitude: 3.8680},
	// Plateau
	"PL01": {StateCode: "PL", LGACode: "01", Name: "Jos North", Latitude: 9.9271, Longitude: 8.8921},
	"PL02": {StateCode: "PL", LGACode: "02", Name: "Jos South", Latitude: 9.8020, Longitude: 8.8650},
	// Rivers
	"RI01": {StateCode: "RI", LGACode: "01", Name: "Port Harcourt", Latitude: 4.8156, Longitude: 7.0498},
	"RI02": {StateCode: "RI", LGACode: "02", Name: "Obio/Akpor", Latitude: 4.8472, Longitude: 7.0134},
	"RI03": {StateCode: "RI", LGACode: "03", Name: "Eleme", Latitude: 4.7912, Longitude: 7.1205},
	// Sokoto
	"SO01": {StateCode: "SO", LGACode: "01", Name: "Sokoto North", Latitude: 13.0609, Longitude: 5.2476},
	"SO02": {StateCode: "SO", LGACode: "02", Name: "Sokoto South", Latitude: 13.0410, Longitude: 5.2390},
	// Taraba
	"TA01": {StateCode: "TA", LGACode: "01", Name: "Jalingo", Latitude: 8.8937, Longitude: 11.3600},
	// Yobe
	"YO01": {StateCode: "YO", LGACode: "01", Name: "Damaturu", Latitude: 11.7470, Longitude: 11.9608},
	// Zamfara
	"ZA01": {StateCode: "ZA", LGACode: "01", Name: "Gusau", Latitude: 12.1702, Longitude: 6.6592},
}

// StateLGAs returns all reference LGAs for a given state code from the built-in reference dataset.
func StateLGAs(state string) []NamedCode {
	s := strings.ToUpper(strings.TrimSpace(state))
	var res []NamedCode
	for _, v := range knownLGAs {
		if v.StateCode == s {
			res = append(res, NamedCode{Code: v.LGACode, Name: v.Name})
		}
	}
	slices.SortFunc(res, func(a, b NamedCode) int {
		return cmp.Compare(a.Code, b.Code)
	})
	return res
}

// stateAliases maps common Nigerian state aliases, ISO codes, and abbreviations to their 2-letter NIPOST code.
var stateAliases = map[string]string{
	"FCT":           "FC",
	"ABUJA":         "FC",
	"ABJ":           "FC",
	"FCT ABUJA":     "FC",
	"F.C.T":         "FC",
	"F.C.T.":        "FC",
	"BE":            "BN", // Common ISO 3166-2 / user confusion for Benue (BN)
	"PH":            "RI",
	"PHC":           "RI",
	"PORT HARCOURT": "RI",
	"AKWA-IBOM":     "AK",
	"CROSS-RIVER":   "CR",
	"NASSARAWA":     "NA",
	"ZAMFERA":       "ZA",
	"PLATUE":        "PL",
}

// LookupState resolves a Nigerian state record by its 2-letter NIPOST code, full state name,
// capital city name, or common alias (e.g. "LA", "Lagos", "Lagos State", "Ikeja", "FCT", "Abuja").
func LookupState(query string) (StateRecord, bool) {
	q := strings.ToUpper(strings.TrimSpace(query))
	q = strings.TrimSuffix(q, " STATE")
	q = strings.TrimSpace(q)
	if q == "" {
		return StateRecord{}, false
	}
	if rec, ok := NigerianStates[q]; ok {
		return rec, true
	}
	if code, ok := stateAliases[q]; ok {
		if rec, ok2 := NigerianStates[code]; ok2 {
			return rec, true
		}
	}
	for _, rec := range NigerianStates {
		if strings.EqualFold(rec.Name, q) || strings.EqualFold(rec.Capital, q) {
			return rec, true
		}
	}
	return StateRecord{}, false
}

// LookupLGA resolves an LGA record within a state by its 2-digit code (e.g. "11", "01", "1")
// or name (e.g. "Ikeja", "Eti-Osa", "Eti Osa", "AMAC"). If state is empty, searches across all states.
func LookupLGA(state, query string) (LGARecord, bool) {
	q := strings.TrimSpace(query)
	if q == "" {
		return LGARecord{}, false
	}
	stCode := ""
	if state != "" {
		if stRec, ok := LookupState(state); ok {
			stCode = stRec.Code
		} else {
			stCode = strings.ToUpper(strings.TrimSpace(state))
		}
	}

	// Check if query is a numeric LGA code
	numQ := q
	if len(numQ) == 1 && numQ[0] >= '1' && numQ[0] <= '9' {
		numQ = "0" + numQ
	}
	if stCode != "" && len(numQ) == 2 {
		if rec, ok := knownLGAs[stCode+numQ]; ok {
			return rec, true
		}
	}

	normQ := normalizePlaceToken(q)
	if normQ == "AMAC" && (stCode == "" || stCode == "FC") {
		if rec, ok := knownLGAs["FC03"]; ok {
			return rec, true
		}
	}

	for _, rec := range knownLGAs {
		if stCode != "" && rec.StateCode != stCode {
			continue
		}
		if strings.EqualFold(rec.Name, q) || normalizePlaceToken(rec.Name) == normQ {
			return rec, true
		}
	}
	return LGARecord{}, false
}

func normalizePlaceToken(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	s = strings.TrimSuffix(s, " LGA")
	s = strings.TrimSuffix(s, " LOCAL GOVERNMENT")
	s = strings.TrimSuffix(s, " AREA COUNCIL")
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			b.WriteByte(c)
		}
	}
	return b.String()
}

// knownBuildings maps compact 11-char postcodes to building-level precision.
var knownBuildings = map[string]BuildingRecord{
	"LA08A86RG01": {
		Postcode:  "LA-08-A86-RG-01",
		Latitude:  6.476111,
		Longitude: 3.633990,
		Address:   "Eti-Osa, Lekki, Lagos",
		StateCode: "LA",
		StateName: "Lagos",
		LGACode:   "08",
		LGAName:   "Eti-Osa",
		Zone:      "SOUTH WEST",
	},
	"EK01A03FK01": {
		Postcode:  "EK-01-A03-FK-01",
		Latitude:  7.6211,
		Longitude: 5.2215,
		Address:   "NTA Road, Back of Fabian Hotel, Ado Ekiti",
		StateCode: "EK",
		StateName: "Ekiti",
		LGACode:   "01",
		LGAName:   "Ado Ekiti",
		Zone:      "SOUTH WEST",
	},
	"AK11I61ZF12": {
		Postcode:  "AK-11-I61-ZF-12",
		Latitude:  5.0377,
		Longitude: 7.9128,
		Address:   "12 Oron Road, Uyo",
		StateCode: "AK",
		StateName: "Akwa Ibom",
		LGACode:   "11",
		LGAName:   "Uyo",
		Zone:      "SOUTH SOUTH",
	},
	"AK11H40WD11": {
		Postcode:  "AK-11-H40-WD-11",
		Latitude:  5.0333,
		Longitude: 7.9266,
		Address:   "11 Wellington Bassey Way, Uyo",
		StateCode: "AK",
		StateName: "Akwa Ibom",
		LGACode:   "11",
		LGAName:   "Uyo",
		Zone:      "SOUTH SOUTH",
	},
	"BA02M67BL69": {
		Postcode:  "BA-02-M67-BL-69",
		Latitude:  10.3158,
		Longitude: 9.8442,
		Address:   "69 Bank Road, GRA, Bauchi",
		StateCode: "BA",
		StateName: "Bauchi",
		LGACode:   "02",
		LGAName:   "Bauchi",
		Zone:      "NORTH EAST",
	},
	"BA02E99NE30": {
		Postcode:  "BA-02-E99-NE-30",
		Latitude:  10.3012,
		Longitude: 9.8234,
		Address:   "30 Ahmadu Bello Way, Bauchi",
		StateCode: "BA",
		StateName: "Bauchi",
		LGACode:   "02",
		LGAName:   "Bauchi",
		Zone:      "NORTH EAST",
	},
	"EB13G95FR90": {
		Postcode:  "EB-13-G95-FR-90",
		Latitude:  6.3249,
		Longitude: 8.1137,
		Address:   "90 Ogoja Road, Abakaliki",
		StateCode: "EB",
		StateName: "Ebonyi",
		LGACode:   "13",
		LGAName:   "Abakaliki",
		Zone:      "SOUTH EAST",
	},
	"EB13I97AB30": {
		Postcode:  "EB-13-I97-AB-30",
		Latitude:  6.3180,
		Longitude: 8.1022,
		Address:   "30 Water Works Road, Abakaliki",
		StateCode: "EB",
		StateName: "Ebonyi",
		LGACode:   "13",
		LGAName:   "Abakaliki",
		Zone:      "SOUTH EAST",
	},
	"EN05V19CD22": {
		Postcode:  "EN-05-V19-CD-22",
		Latitude:  6.4584,
		Longitude: 7.5464,
		Address:   "22 Chime Avenue, New Haven, Enugu",
		StateCode: "EN",
		StateName: "Enugu",
		LGACode:   "05",
		LGAName:   "Enugu North",
		Zone:      "SOUTH EAST",
	},
	"EN05V19FT20": {
		Postcode:  "EN-05-V19-FT-20",
		Latitude:  6.4412,
		Longitude: 7.5023,
		Address:   "20 Ogui Road, Enugu",
		StateCode: "EN",
		StateName: "Enugu",
		LGACode:   "05",
		LGAName:   "Enugu North",
		Zone:      "SOUTH EAST",
	},
	"FC03B06AG12": {
		Postcode:  "FC-03-B06-AG-12",
		Latitude:  9.0579,
		Longitude: 7.4951,
		Address:   "12 Shehu Shagari Way, Garki, Abuja",
		StateCode: "FC",
		StateName: "Federal Capital Territory",
		LGACode:   "03",
		LGAName:   "Abuja Municipal",
		Zone:      "NORTH CENTRAL",
	},
	"FC02B19RT30": {
		Postcode:  "FC-02-B19-RT-30",
		Latitude:  9.1538,
		Longitude: 7.3220,
		Address:   "30 Gado Nasko Way, Phase 4, Kubwa, Abuja",
		StateCode: "FC",
		StateName: "Federal Capital Territory",
		LGACode:   "02",
		LGAName:   "Bwari",
		Zone:      "NORTH CENTRAL",
	},
	"JI24O18JP23": {
		Postcode:  "JI-24-O18-JP-23",
		Latitude:  11.7594,
		Longitude: 9.3389,
		Address:   "23 Sani Abacha Way, Dutse",
		StateCode: "JI",
		StateName: "Jigawa",
		LGACode:   "24",
		LGAName:   "Dutse",
		Zone:      "NORTH WEST",
	},
	"JI24N11VM58": {
		Postcode:  "JI-24-N11-VM-58",
		Latitude:  11.7231,
		Longitude: 9.3102,
		Address:   "58 Kano-Dutse Expressway, Dutse",
		StateCode: "JI",
		StateName: "Jigawa",
		LGACode:   "24",
		LGAName:   "Dutse",
		Zone:      "NORTH WEST",
	},
	"KN31F82WJ80": {
		Postcode:  "KN-31-F82-WJ-80",
		Latitude:  12.0022,
		Longitude: 8.5920,
		Address:   "80 Badu Road, Bompai, Kano",
		StateCode: "KN",
		StateName: "Kano",
		LGACode:   "31",
		LGAName:   "Kano Municipal",
		Zone:      "NORTH WEST",
	},
	"KN31D78IQ38": {
		Postcode:  "KN-31-D78-IQ-38",
		Latitude:  12.0150,
		Longitude: 8.5201,
		Address:   "38 Ibrahim Taiwo Road, Kano",
		StateCode: "KN",
		StateName: "Kano",
		LGACode:   "31",
		LGAName:   "Kano Municipal",
		Zone:      "NORTH WEST",
	},
	"LA11W06TC10": {
		Postcode:  "LA-11-W06-TC-10",
		Latitude:  6.6018,
		Longitude: 3.3515,
		Address:   "10 Obafemi Awolowo Way, Ikeja, Lagos",
		StateCode: "LA",
		StateName: "Lagos",
		LGACode:   "11",
		LGAName:   "Ikeja",
		Zone:      "SOUTH WEST",
	},
	"LA11U34ZR63": {
		Postcode:  "LA-11-U34-ZR-63",
		Latitude:  6.5891,
		Longitude: 3.3590,
		Address:   "63 Isaac John Street, GRA Ikeja, Lagos",
		StateCode: "LA",
		StateName: "Lagos",
		LGACode:   "11",
		LGAName:   "Ikeja",
		Zone:      "SOUTH WEST",
	},
	"NI09J67QC65": {
		Postcode:  "NI-09-J67-QC-65",
		Latitude:  9.6139,
		Longitude: 6.5569,
		Address:   "65 Bosso Road, Minna",
		StateCode: "NI",
		StateName: "Niger",
		LGACode:   "09",
		LGAName:   "Chanchaga",
		Zone:      "NORTH CENTRAL",
	},
	"NI09A75DA10": {
		Postcode:  "NI-09-A75-DA-10",
		Latitude:  9.6288,
		Longitude: 6.5412,
		Address:   "10 Paida Road, Minna",
		StateCode: "NI",
		StateName: "Niger",
		LGACode:   "09",
		LGAName:   "Chanchaga",
		Zone:      "NORTH CENTRAL",
	},
	"OG14T18BN16": {
		Postcode:  "OG-14-T18-BN-16",
		Latitude:  7.1475,
		Longitude: 3.3619,
		Address:   "16 Lalubu Street, Oke-Ilewo, Abeokuta",
		StateCode: "OG",
		StateName: "Ogun",
		LGACode:   "14",
		LGAName:   "Abeokuta South",
		Zone:      "SOUTH WEST",
	},
	"OG14M82QA09": {
		Postcode:  "OG-14-M82-QA-09",
		Latitude:  7.1550,
		Longitude: 3.3480,
		Address:   "9 Quarry Road, Abeokuta",
		StateCode: "OG",
		StateName: "Ogun",
		LGACode:   "14",
		LGAName:   "Abeokuta South",
		Zone:      "SOUTH WEST",
	},
}

var (
	customMu        sync.RWMutex
	customBuildings = make(map[string]BuildingRecord)
)

// RegisterKnownBuilding allows applications to register custom building units into
// the offline geocoding cache.
func RegisterKnownBuilding(rec BuildingRecord) {
	customMu.Lock()
	defer customMu.Unlock()

	parsed, err := Parse(rec.Postcode)
	if err == nil {
		customBuildings[parsed.Raw()] = rec
	} else {
		customBuildings[rec.Postcode] = rec
	}
}

// resolvePostcodeLocation implements hierarchical local geocoding:
// Tier 1: Known Building registry (exact coordinates)
// Tier 2: Known LGA centroid
// Tier 3: State administrative centroid
func resolvePostcodeLocation(p Postcode) Location {
	if p.IsZero() {
		return Location{}
	}

	ensureCacheLoaded()

	compact := p.Raw()
	canonical := p.Formatted()
	stateCode := p.State()
	lgaCode := p.LGA()

	// 1. Check built-in official building registry first
	if bRec, hasBuilding := knownBuildings[compact]; hasBuilding {
		return Location{
			Postcode:  canonical,
			Compact:   compact,
			Latitude:  bRec.Latitude,
			Longitude: bRec.Longitude,
			Precision: PrecisionBuilding,
			Address:   bRec.Address,
			StateCode: bRec.StateCode,
			StateName: bRec.StateName,
			LGACode:   bRec.LGACode,
			LGAName:   bRec.LGAName,
			Zone:      bRec.Zone,
		}
	}

	// 2. Check custom and cached buildings
	customMu.RLock()
	cRec, hasCustom := customBuildings[compact]
	customMu.RUnlock()
	if hasCustom {
		return Location{
			Postcode:  canonical,
			Compact:   compact,
			Latitude:  cRec.Latitude,
			Longitude: cRec.Longitude,
			Precision: PrecisionBuilding,
			Address:   cRec.Address,
			StateCode: cRec.StateCode,
			StateName: cRec.StateName,
			LGACode:   cRec.LGACode,
			LGAName:   cRec.LGAName,
			Zone:      cRec.Zone,
		}
	}

	// 3. Fallback to Known LGA centroid
	lgaKey := stateCode + lgaCode
	if lgaRec, hasLGA := knownLGAs[lgaKey]; hasLGA {
		stRec := NigerianStates[stateCode]
		return Location{
			Postcode:  canonical,
			Compact:   compact,
			Latitude:  lgaRec.Latitude,
			Longitude: lgaRec.Longitude,
			Precision: PrecisionLGA,
			StateCode: stateCode,
			StateName: stRec.Name,
			LGACode:   lgaCode,
			LGAName:   lgaRec.Name,
			Zone:      stRec.Zone,
		}
	}

	// 4. Fallback to State administrative centroid
	stRec, hasState := NigerianStates[stateCode]
	if hasState {
		return Location{
			Postcode:  canonical,
			Compact:   compact,
			Latitude:  stRec.Latitude,
			Longitude: stRec.Longitude,
			Precision: PrecisionState,
			StateCode: stateCode,
			StateName: stRec.Name,
			LGACode:   lgaCode,
			LGAName:   "LGA " + lgaCode,
			Zone:      stRec.Zone,
		}
	}

	// 5. Unknown State fallback
	return Location{
		Postcode:  canonical,
		Compact:   compact,
		StateCode: stateCode,
		LGACode:   lgaCode,
		Precision: PrecisionState,
	}
}

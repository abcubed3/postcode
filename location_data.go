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
// Complete, authoritative registry of all 774 Nigerian Local Government Areas.
var knownLGAs = map[string]LGARecord{
	// --- State AB ---
	"AB01": {StateCode: "AB", LGACode: "01", Name: "Aba North", Latitude: 5.1156, Longitude: 7.3676},
	"AB02": {StateCode: "AB", LGACode: "02", Name: "Aba South", Latitude: 5.0982, Longitude: 7.3612},
	"AB03": {StateCode: "AB", LGACode: "03", Name: "Arochukwu", Latitude: 5.6669, Longitude: 7.5895},
	"AB04": {StateCode: "AB", LGACode: "04", Name: "Bende", Latitude: 5.4955, Longitude: 7.2792},
	"AB05": {StateCode: "AB", LGACode: "05", Name: "Ikwuano", Latitude: 5.3977, Longitude: 7.6968},
	"AB06": {StateCode: "AB", LGACode: "06", Name: "Isiala Ngwa North", Latitude: 5.8121, Longitude: 7.4109},
	"AB07": {StateCode: "AB", LGACode: "07", Name: "Isiala Ngwa South", Latitude: 5.4876, Longitude: 7.4629},
	"AB08": {StateCode: "AB", LGACode: "08", Name: "Isuikwuato", Latitude: 5.5628, Longitude: 7.5706},
	"AB09": {StateCode: "AB", LGACode: "09", Name: "Obi Ngwa", Latitude: 5.5817, Longitude: 7.3659},
	"AB10": {StateCode: "AB", LGACode: "10", Name: "Ohafia", Latitude: 5.3779, Longitude: 7.5578},
	"AB11": {StateCode: "AB", LGACode: "11", Name: "Osisioma Ngwa", Latitude: 5.7323, Longitude: 7.5491},
	"AB12": {StateCode: "AB", LGACode: "12", Name: "Ugwunagbo", Latitude: 5.4070, Longitude: 7.2695},
	"AB13": {StateCode: "AB", LGACode: "13", Name: "Ukwa East", Latitude: 5.4692, Longitude: 7.7691},
	"AB14": {StateCode: "AB", LGACode: "14", Name: "Ukwa West", Latitude: 5.5730, Longitude: 7.4573},
	"AB15": {StateCode: "AB", LGACode: "15", Name: "Umuahia North", Latitude: 5.5320, Longitude: 7.4860},
	"AB16": {StateCode: "AB", LGACode: "16", Name: "Umuahia South", Latitude: 5.6156, Longitude: 7.5856},
	"AB17": {StateCode: "AB", LGACode: "17", Name: "Umu Nneochi", Latitude: 5.5394, Longitude: 7.3162},
	// --- State AD ---
	"AD01": {StateCode: "AD", LGACode: "01", Name: "Demsa", Latitude: 9.2702, Longitude: 12.4154},
	"AD02": {StateCode: "AD", LGACode: "02", Name: "Fufore", Latitude: 9.0799, Longitude: 12.4931},
	"AD03": {StateCode: "AD", LGACode: "03", Name: "Ganye", Latitude: 9.3443, Longitude: 12.5853},
	"AD04": {StateCode: "AD", LGACode: "04", Name: "Girei", Latitude: 9.1729, Longitude: 12.2750},
	"AD05": {StateCode: "AD", LGACode: "05", Name: "Gombi", Latitude: 9.0751, Longitude: 12.6926},
	"AD06": {StateCode: "AD", LGACode: "06", Name: "Guyuk", Latitude: 9.4895, Longitude: 12.4067},
	"AD07": {StateCode: "AD", LGACode: "07", Name: "Hong", Latitude: 9.1650, Longitude: 12.4587},
	"AD08": {StateCode: "AD", LGACode: "08", Name: "Jada", Latitude: 9.2402, Longitude: 12.5664},
	"AD09": {StateCode: "AD", LGACode: "09", Name: "Lamurde", Latitude: 9.2591, Longitude: 12.3617},
	"AD10": {StateCode: "AD", LGACode: "10", Name: "Madagali", Latitude: 9.0553, Longitude: 12.5536},
	"AD11": {StateCode: "AD", LGACode: "11", Name: "Maiha", Latitude: 9.4097, Longitude: 12.5449},
	"AD12": {StateCode: "AD", LGACode: "12", Name: "Mayo-Belwa", Latitude: 9.0844, Longitude: 12.2653},
	"AD13": {StateCode: "AD", LGACode: "13", Name: "Michika", Latitude: 9.1466, Longitude: 12.7649},
	"AD14": {StateCode: "AD", LGACode: "14", Name: "Mubi North", Latitude: 9.2504, Longitude: 12.4531},
	"AD15": {StateCode: "AD", LGACode: "15", Name: "Mubi South", Latitude: 9.1202, Longitude: 12.4701},
	"AD16": {StateCode: "AD", LGACode: "16", Name: "Numan", Latitude: 9.2930, Longitude: 12.5814},
	"AD17": {StateCode: "AD", LGACode: "17", Name: "Shelleng", Latitude: 9.2168, Longitude: 12.3120},
	"AD18": {StateCode: "AD", LGACode: "18", Name: "Song", Latitude: 9.0609, Longitude: 12.6303},
	"AD19": {StateCode: "AD", LGACode: "19", Name: "Toungo", Latitude: 9.4592, Longitude: 12.4709},
	"AD20": {StateCode: "AD", LGACode: "20", Name: "Yola North", Latitude: 9.2094, Longitude: 12.4818},
	"AD21": {StateCode: "AD", LGACode: "21", Name: "Yola South", Latitude: 9.1932, Longitude: 12.4191},
	// --- State AK ---
	"AK01": {StateCode: "AK", LGACode: "01", Name: "Abak", Latitude: 5.0985, Longitude: 7.8464},
	"AK02": {StateCode: "AK", LGACode: "02", Name: "Eastern Obolo", Latitude: 4.9082, Longitude: 7.9241},
	"AK03": {StateCode: "AK", LGACode: "03", Name: "Eket", Latitude: 4.6423, Longitude: 7.9244},
	"AK04": {StateCode: "AK", LGACode: "04", Name: "Esit Eket", Latitude: 5.0012, Longitude: 7.7060},
	"AK05": {StateCode: "AK", LGACode: "05", Name: "Essien Udim", Latitude: 4.9034, Longitude: 8.1236},
	"AK06": {StateCode: "AK", LGACode: "06", Name: "Etim Ekpo", Latitude: 5.3178, Longitude: 7.8377},
	"AK07": {StateCode: "AK", LGACode: "07", Name: "Etinan", Latitude: 4.9933, Longitude: 7.8897},
	"AK08": {StateCode: "AK", LGACode: "08", Name: "Ibeno", Latitude: 5.0685, Longitude: 7.9974},
	"AK09": {StateCode: "AK", LGACode: "09", Name: "Ibesikpo Asutan", Latitude: 5.0874, Longitude: 7.7927},
	"AK10": {StateCode: "AK", LGACode: "10", Name: "Ibiono-Ibom", Latitude: 4.8836, Longitude: 7.9846},
	"AK11": {StateCode: "AK", LGACode: "11", Name: "Ika", Latitude: 5.2380, Longitude: 7.9759},
	"AK12": {StateCode: "AK", LGACode: "12", Name: "Ikono", Latitude: 4.9127, Longitude: 7.6963},
	"AK13": {StateCode: "AK", LGACode: "13", Name: "Ikot Abasi", Latitude: 4.9749, Longitude: 8.1959},
	"AK14": {StateCode: "AK", LGACode: "14", Name: "Ikot Ekpene", Latitude: 5.0787, Longitude: 7.8841},
	"AK15": {StateCode: "AK", LGACode: "15", Name: "Ini", Latitude: 4.9485, Longitude: 7.9011},
	"AK16": {StateCode: "AK", LGACode: "16", Name: "Itu", Latitude: 5.1213, Longitude: 8.0124},
	"AK17": {StateCode: "AK", LGACode: "17", Name: "Mbo", Latitude: 5.0451, Longitude: 7.7430},
	"AK18": {StateCode: "AK", LGACode: "18", Name: "Mkpat-Enin", Latitude: 4.8892, Longitude: 8.0613},
	"AK19": {StateCode: "AK", LGACode: "19", Name: "Nsit-Atai", Latitude: 5.2875, Longitude: 7.9019},
	"AK20": {StateCode: "AK", LGACode: "20", Name: "Nsit-Ibom", Latitude: 4.8155, Longitude: 7.7264},
	"AK21": {StateCode: "AK", LGACode: "21", Name: "Nsit-Ubium", Latitude: 5.0442, Longitude: 7.9624},
	"AK22": {StateCode: "AK", LGACode: "22", Name: "Obot Akara", Latitude: 5.0893, Longitude: 7.8391},
	"AK23": {StateCode: "AK", LGACode: "23", Name: "Okobo", Latitude: 4.9108, Longitude: 7.9409},
	"AK24": {StateCode: "AK", LGACode: "24", Name: "Onna", Latitude: 5.1849, Longitude: 7.9978},
	"AK25": {StateCode: "AK", LGACode: "25", Name: "Oron", Latitude: 4.9746, Longitude: 7.7125},
	"AK26": {StateCode: "AK", LGACode: "26", Name: "Oruk Anam", Latitude: 4.9320, Longitude: 8.1394},
	"AK27": {StateCode: "AK", LGACode: "27", Name: "Udung-Uko", Latitude: 5.3056, Longitude: 7.8018},
	"AK28": {StateCode: "AK", LGACode: "28", Name: "Ukanafun", Latitude: 4.9907, Longitude: 7.8957},
	"AK29": {StateCode: "AK", LGACode: "29", Name: "Uruan", Latitude: 5.0793, Longitude: 7.9926},
	"AK30": {StateCode: "AK", LGACode: "30", Name: "Urue-Offong/Oruko", Latitude: 5.0713, Longitude: 7.7872},
	"AK31": {StateCode: "AK", LGACode: "31", Name: "Uyo", Latitude: 5.0377, Longitude: 7.9128},
	// --- State AN ---
	"AN01": {StateCode: "AN", LGACode: "01", Name: "Aguata", Latitude: 6.2817, Longitude: 7.0006},
	"AN02": {StateCode: "AN", LGACode: "02", Name: "Anambra East", Latitude: 6.0914, Longitude: 7.0783},
	"AN03": {StateCode: "AN", LGACode: "03", Name: "Anambra West", Latitude: 6.3558, Longitude: 7.1705},
	"AN04": {StateCode: "AN", LGACode: "04", Name: "Anaocha", Latitude: 6.1844, Longitude: 6.8602},
	"AN05": {StateCode: "AN", LGACode: "05", Name: "Awka North", Latitude: 6.0866, Longitude: 7.2778},
	"AN06": {StateCode: "AN", LGACode: "06", Name: "Awka South", Latitude: 6.2209, Longitude: 7.0670},
	"AN07": {StateCode: "AN", LGACode: "07", Name: "Ayamelum", Latitude: 6.1765, Longitude: 7.0439},
	"AN08": {StateCode: "AN", LGACode: "08", Name: "Dunukofia", Latitude: 6.2517, Longitude: 7.1516},
	"AN09": {StateCode: "AN", LGACode: "09", Name: "Ekwusigo", Latitude: 6.2706, Longitude: 6.9469},
	"AN10": {StateCode: "AN", LGACode: "10", Name: "Idemili North", Latitude: 6.0668, Longitude: 7.1388},
	"AN11": {StateCode: "AN", LGACode: "11", Name: "Idemili South", Latitude: 6.4212, Longitude: 7.1301},
	"AN12": {StateCode: "AN", LGACode: "12", Name: "Ihiala", Latitude: 6.0959, Longitude: 6.8505},
	"AN13": {StateCode: "AN", LGACode: "13", Name: "Njikoka", Latitude: 6.1581, Longitude: 7.3501},
	"AN14": {StateCode: "AN", LGACode: "14", Name: "Nnewi North", Latitude: 6.0195, Longitude: 6.9148},
	"AN15": {StateCode: "AN", LGACode: "15", Name: "Nnewi South", Latitude: 6.1317, Longitude: 7.0553},
	"AN16": {StateCode: "AN", LGACode: "16", Name: "Ogbaru", Latitude: 6.3045, Longitude: 7.1666},
	"AN17": {StateCode: "AN", LGACode: "17", Name: "Onitsha North", Latitude: 6.1498, Longitude: 6.7857},
	"AN18": {StateCode: "AN", LGACode: "18", Name: "Onitsha South", Latitude: 6.0724, Longitude: 7.2155},
	"AN19": {StateCode: "AN", LGACode: "19", Name: "Orumba North", Latitude: 6.4707, Longitude: 7.0561},
	"AN20": {StateCode: "AN", LGACode: "20", Name: "Orumba South", Latitude: 5.9987, Longitude: 6.8806},
	"AN21": {StateCode: "AN", LGACode: "21", Name: "Oyi", Latitude: 6.2274, Longitude: 7.1166},
	// --- State BA ---
	"BA01": {StateCode: "BA", LGACode: "01", Name: "Alkaleri", Latitude: 10.3766, Longitude: 9.7778},
	"BA02": {StateCode: "BA", LGACode: "02", Name: "Bauchi", Latitude: 10.3158, Longitude: 9.8442},
	"BA03": {StateCode: "BA", LGACode: "03", Name: "Bogoro", Latitude: 10.4507, Longitude: 9.9477},
	"BA04": {StateCode: "BA", LGACode: "04", Name: "Damban", Latitude: 10.2793, Longitude: 9.6374},
	"BA05": {StateCode: "BA", LGACode: "05", Name: "Darazo", Latitude: 10.1815, Longitude: 10.0550},
	"BA06": {StateCode: "BA", LGACode: "06", Name: "Dass", Latitude: 10.5959, Longitude: 9.7691},
	"BA07": {StateCode: "BA", LGACode: "07", Name: "Gamawa", Latitude: 10.2714, Longitude: 9.8211},
	"BA08": {StateCode: "BA", LGACode: "08", Name: "Ganjuwa", Latitude: 10.3466, Longitude: 9.9288},
	"BA09": {StateCode: "BA", LGACode: "09", Name: "Giade", Latitude: 10.3655, Longitude: 9.7241},
	"BA10": {StateCode: "BA", LGACode: "10", Name: "Itas/Gadau", Latitude: 10.1617, Longitude: 9.9160},
	"BA11": {StateCode: "BA", LGACode: "11", Name: "Jama'are", Latitude: 10.5161, Longitude: 9.9073},
	"BA12": {StateCode: "BA", LGACode: "12", Name: "Katagum", Latitude: 10.1908, Longitude: 9.6277},
	"BA13": {StateCode: "BA", LGACode: "13", Name: "Kirfi", Latitude: 10.2530, Longitude: 10.1273},
	"BA14": {StateCode: "BA", LGACode: "14", Name: "Misau", Latitude: 10.3568, Longitude: 9.8155},
	"BA15": {StateCode: "BA", LGACode: "15", Name: "Ningi", Latitude: 10.2266, Longitude: 9.8325},
	"BA16": {StateCode: "BA", LGACode: "16", Name: "Shira", Latitude: 10.3994, Longitude: 9.9438},
	"BA17": {StateCode: "BA", LGACode: "17", Name: "Tafawa Balewa", Latitude: 10.3232, Longitude: 9.6744},
	"BA18": {StateCode: "BA", LGACode: "18", Name: "Toro", Latitude: 10.1673, Longitude: 9.9927},
	"BA19": {StateCode: "BA", LGACode: "19", Name: "Warji", Latitude: 10.5656, Longitude: 9.8333},
	"BA20": {StateCode: "BA", LGACode: "20", Name: "Zaki", Latitude: 10.0936, Longitude: 9.6578},
	// --- State BY ---
	"BY01": {StateCode: "BY", LGACode: "01", Name: "Brass", Latitude: 4.9875, Longitude: 6.2012},
	"BY02": {StateCode: "BY", LGACode: "02", Name: "Ekeremor", Latitude: 4.7972, Longitude: 6.2789},
	"BY03": {StateCode: "BY", LGACode: "03", Name: "Kolokuma/Opokuma", Latitude: 5.0616, Longitude: 6.3711},
	"BY04": {StateCode: "BY", LGACode: "04", Name: "Nembe", Latitude: 4.8902, Longitude: 6.0608},
	"BY05": {StateCode: "BY", LGACode: "05", Name: "Ogbia", Latitude: 4.7924, Longitude: 6.4784},
	"BY06": {StateCode: "BY", LGACode: "06", Name: "Sagbama", Latitude: 5.2068, Longitude: 6.1925},
	"BY07": {StateCode: "BY", LGACode: "07", Name: "Southern Ijaw", Latitude: 4.8823, Longitude: 6.2445},
	"BY08": {StateCode: "BY", LGACode: "08", Name: "Yenagoa", Latitude: 4.9267, Longitude: 6.2676},
	// --- State BN ---
	"BN01": {StateCode: "BN", LGACode: "01", Name: "Ado", Latitude: 7.7912, Longitude: 8.4547},
	"BN02": {StateCode: "BN", LGACode: "02", Name: "Agatu", Latitude: 7.6009, Longitude: 8.5324},
	"BN03": {StateCode: "BN", LGACode: "03", Name: "Apa", Latitude: 7.8653, Longitude: 8.6246},
	"BN04": {StateCode: "BN", LGACode: "04", Name: "Buruku", Latitude: 7.6939, Longitude: 8.3143},
	"BN05": {StateCode: "BN", LGACode: "05", Name: "Gboko", Latitude: 7.3250, Longitude: 9.0011},
	"BN06": {StateCode: "BN", LGACode: "06", Name: "Guma", Latitude: 8.0105, Longitude: 8.4460},
	"BN07": {StateCode: "BN", LGACode: "07", Name: "Gwer East", Latitude: 7.6860, Longitude: 8.4980},
	"BN08": {StateCode: "BN", LGACode: "08", Name: "Gwer West", Latitude: 7.7612, Longitude: 8.6057},
	"BN09": {StateCode: "BN", LGACode: "09", Name: "Katsina-Ala", Latitude: 7.7801, Longitude: 8.4010},
	"BN10": {StateCode: "BN", LGACode: "10", Name: "Konshisha", Latitude: 7.5763, Longitude: 8.5929},
	"BN11": {StateCode: "BN", LGACode: "11", Name: "Kwande", Latitude: 7.9307, Longitude: 8.5842},
	"BN12": {StateCode: "BN", LGACode: "12", Name: "Logo", Latitude: 7.6054, Longitude: 8.3046},
	"BN13": {StateCode: "BN", LGACode: "13", Name: "Makurdi", Latitude: 7.7304, Longitude: 8.5211},
	"BN14": {StateCode: "BN", LGACode: "14", Name: "Obi", Latitude: 7.7714, Longitude: 8.4924},
	"BN15": {StateCode: "BN", LGACode: "15", Name: "Ogbadibo", Latitude: 7.6412, Longitude: 8.5094},
	"BN16": {StateCode: "BN", LGACode: "16", Name: "Ohimini", Latitude: 7.8140, Longitude: 8.6207},
	"BN17": {StateCode: "BN", LGACode: "17", Name: "Oju", Latitude: 7.7378, Longitude: 8.3513},
	"BN18": {StateCode: "BN", LGACode: "18", Name: "Okpokwu", Latitude: 7.5819, Longitude: 8.6696},
	"BN19": {StateCode: "BN", LGACode: "19", Name: "Otukpo", Latitude: 7.1931, Longitude: 8.1339},
	"BN20": {StateCode: "BN", LGACode: "20", Name: "Tarka", Latitude: 7.5082, Longitude: 8.3347},
	"BN21": {StateCode: "BN", LGACode: "21", Name: "Ukum", Latitude: 7.7369, Longitude: 8.5707},
	"BN22": {StateCode: "BN", LGACode: "22", Name: "Ushongo", Latitude: 7.7820, Longitude: 8.4474},
	"BN23": {StateCode: "BN", LGACode: "23", Name: "Vandeikya", Latitude: 7.6035, Longitude: 8.5492},
	// --- State BO ---
	"BO01": {StateCode: "BO", LGACode: "01", Name: "Abadam", Latitude: 11.8941, Longitude: 13.0836},
	"BO02": {StateCode: "BO", LGACode: "02", Name: "Askira/Uba", Latitude: 11.7038, Longitude: 13.1613},
	"BO03": {StateCode: "BO", LGACode: "03", Name: "Bama", Latitude: 11.9682, Longitude: 13.2535},
	"BO04": {StateCode: "BO", LGACode: "04", Name: "Bayo", Latitude: 11.7968, Longitude: 12.9432},
	"BO05": {StateCode: "BO", LGACode: "05", Name: "Biu", Latitude: 11.6990, Longitude: 13.3608},
	"BO06": {StateCode: "BO", LGACode: "06", Name: "Chibok", Latitude: 12.1134, Longitude: 13.0749},
	"BO07": {StateCode: "BO", LGACode: "07", Name: "Damboa", Latitude: 11.7889, Longitude: 13.1269},
	"BO08": {StateCode: "BO", LGACode: "08", Name: "Dikwa", Latitude: 11.8641, Longitude: 13.2346},
	"BO09": {StateCode: "BO", LGACode: "09", Name: "Gubio", Latitude: 11.8830, Longitude: 13.0299},
	"BO10": {StateCode: "BO", LGACode: "10", Name: "Guzamala", Latitude: 11.6792, Longitude: 13.2218},
	"BO11": {StateCode: "BO", LGACode: "11", Name: "Gwoza", Latitude: 12.0336, Longitude: 13.2131},
	"BO12": {StateCode: "BO", LGACode: "12", Name: "Hawul", Latitude: 11.7083, Longitude: 12.9335},
	"BO13": {StateCode: "BO", LGACode: "13", Name: "Jere", Latitude: 11.7705, Longitude: 13.4331},
	"BO14": {StateCode: "BO", LGACode: "14", Name: "Kaga", Latitude: 11.8743, Longitude: 13.1213},
	"BO15": {StateCode: "BO", LGACode: "15", Name: "Kala/Balge", Latitude: 11.7441, Longitude: 13.1383},
	"BO16": {StateCode: "BO", LGACode: "16", Name: "Konduga", Latitude: 11.9169, Longitude: 13.2496},
	"BO17": {StateCode: "BO", LGACode: "17", Name: "Kukawa", Latitude: 11.8407, Longitude: 12.9802},
	"BO18": {StateCode: "BO", LGACode: "18", Name: "Kwaya Kusar", Latitude: 11.6848, Longitude: 13.2985},
	"BO19": {StateCode: "BO", LGACode: "19", Name: "Mafa", Latitude: 12.0831, Longitude: 13.1391},
	"BO20": {StateCode: "BO", LGACode: "20", Name: "Magumeri", Latitude: 11.6111, Longitude: 12.9636},
	"BO21": {StateCode: "BO", LGACode: "21", Name: "Maiduguri", Latitude: 11.8333, Longitude: 13.1500},
	"BO22": {StateCode: "BO", LGACode: "22", Name: "Marte", Latitude: 11.8849, Longitude: 13.0763},
	"BO23": {StateCode: "BO", LGACode: "23", Name: "Mobbar", Latitude: 11.7064, Longitude: 13.1781},
	"BO24": {StateCode: "BO", LGACode: "24", Name: "Monguno", Latitude: 11.9805, Longitude: 13.2350},
	"BO25": {StateCode: "BO", LGACode: "25", Name: "Ngala", Latitude: 11.7702, Longitude: 12.9497},
	"BO26": {StateCode: "BO", LGACode: "26", Name: "Nganzai", Latitude: 11.7276, Longitude: 13.3766},
	"BO27": {StateCode: "BO", LGACode: "27", Name: "Shani", Latitude: 12.1012, Longitude: 13.0390},
	// --- State CR ---
	"CR01": {StateCode: "CR", LGACode: "01", Name: "Abi", Latitude: 5.0365, Longitude: 8.2753},
	"CR02": {StateCode: "CR", LGACode: "02", Name: "Akamkpa", Latitude: 4.8462, Longitude: 8.3530},
	"CR03": {StateCode: "CR", LGACode: "03", Name: "Akpabuyo", Latitude: 5.1106, Longitude: 8.4452},
	"CR04": {StateCode: "CR", LGACode: "04", Name: "Bakassi", Latitude: 4.9392, Longitude: 8.1349},
	"CR05": {StateCode: "CR", LGACode: "05", Name: "Bekwarra", Latitude: 4.8414, Longitude: 8.5525},
	"CR06": {StateCode: "CR", LGACode: "06", Name: "Biase", Latitude: 5.2558, Longitude: 8.2666},
	"CR07": {StateCode: "CR", LGACode: "07", Name: "Boki", Latitude: 4.9313, Longitude: 8.3186},
	"CR08": {StateCode: "CR", LGACode: "08", Name: "Calabar Municipal", Latitude: 4.9757, Longitude: 8.3417},
	"CR09": {StateCode: "CR", LGACode: "09", Name: "Calabar South", Latitude: 4.9482, Longitude: 8.3250},
	"CR10": {StateCode: "CR", LGACode: "10", Name: "Etung", Latitude: 4.8216, Longitude: 8.4135},
	"CR11": {StateCode: "CR", LGACode: "11", Name: "Ikom", Latitude: 5.1760, Longitude: 8.4048},
	"CR12": {StateCode: "CR", LGACode: "12", Name: "Obanliku", Latitude: 4.8507, Longitude: 8.1252},
	"CR13": {StateCode: "CR", LGACode: "13", Name: "Obubra", Latitude: 4.9129, Longitude: 8.6248},
	"CR14": {StateCode: "CR", LGACode: "14", Name: "Obudu", Latitude: 5.0167, Longitude: 8.3130},
	"CR15": {StateCode: "CR", LGACode: "15", Name: "Odukpani", Latitude: 4.8865, Longitude: 8.3300},
	"CR16": {StateCode: "CR", LGACode: "16", Name: "Ogoja", Latitude: 5.0593, Longitude: 8.4413},
	"CR17": {StateCode: "CR", LGACode: "17", Name: "Yakuur", Latitude: 4.9831, Longitude: 8.1719},
	"CR18": {StateCode: "CR", LGACode: "18", Name: "Yala", Latitude: 4.8272, Longitude: 8.4902},
	// --- State DE ---
	"DE01": {StateCode: "DE", LGACode: "01", Name: "Aniocha North", Latitude: 6.2592, Longitude: 6.6665},
	"DE02": {StateCode: "DE", LGACode: "02", Name: "Aniocha South", Latitude: 6.0689, Longitude: 6.7442},
	"DE03": {StateCode: "DE", LGACode: "03", Name: "Bomadi", Latitude: 6.3333, Longitude: 6.8364},
	"DE04": {StateCode: "DE", LGACode: "04", Name: "Burutu", Latitude: 6.1619, Longitude: 6.5261},
	"DE05": {StateCode: "DE", LGACode: "05", Name: "Ethiope East", Latitude: 6.0641, Longitude: 6.9437},
	"DE06": {StateCode: "DE", LGACode: "06", Name: "Ethiope West", Latitude: 6.4785, Longitude: 6.6578},
	"DE07": {StateCode: "DE", LGACode: "07", Name: "Ika North East", Latitude: 6.1540, Longitude: 6.7098},
	"DE08": {StateCode: "DE", LGACode: "08", Name: "Ika South", Latitude: 6.2292, Longitude: 6.8175},
	"DE09": {StateCode: "DE", LGACode: "09", Name: "Isoko North", Latitude: 6.2481, Longitude: 6.6128},
	"DE10": {StateCode: "DE", LGACode: "10", Name: "Isoko South", Latitude: 6.0443, Longitude: 6.8047},
	"DE11": {StateCode: "DE", LGACode: "11", Name: "Ndokwa East", Latitude: 6.3987, Longitude: 6.7960},
	"DE12": {StateCode: "DE", LGACode: "12", Name: "Ndokwa West", Latitude: 6.0734, Longitude: 6.5164},
	"DE13": {StateCode: "DE", LGACode: "13", Name: "Okpe", Latitude: 6.1356, Longitude: 7.0160},
	"DE14": {StateCode: "DE", LGACode: "14", Name: "Oshimili North", Latitude: 6.2394, Longitude: 6.7042},
	"DE15": {StateCode: "DE", LGACode: "15", Name: "Oshimili South", Latitude: 6.1984, Longitude: 6.7329},
	"DE16": {StateCode: "DE", LGACode: "16", Name: "Patani", Latitude: 6.2820, Longitude: 6.8325},
	"DE17": {StateCode: "DE", LGACode: "17", Name: "Sapele", Latitude: 6.2058, Longitude: 6.5631},
	"DE18": {StateCode: "DE", LGACode: "18", Name: "Udu", Latitude: 6.0499, Longitude: 6.8814},
	"DE19": {StateCode: "DE", LGACode: "19", Name: "Ughelli North", Latitude: 6.4482, Longitude: 6.7220},
	"DE20": {StateCode: "DE", LGACode: "20", Name: "Ughelli South", Latitude: 5.9762, Longitude: 6.5465},
	"DE21": {StateCode: "DE", LGACode: "21", Name: "Ukwuani", Latitude: 6.2049, Longitude: 6.7825},
	"DE22": {StateCode: "DE", LGACode: "22", Name: "Uvwie", Latitude: 6.2500, Longitude: 6.6592},
	"DE23": {StateCode: "DE", LGACode: "23", Name: "Warri North", Latitude: 6.0715, Longitude: 6.7610},
	"DE24": {StateCode: "DE", LGACode: "24", Name: "Warri South", Latitude: 5.5167, Longitude: 5.7500},
	"DE25": {StateCode: "DE", LGACode: "25", Name: "Warri South West", Latitude: 6.1353, Longitude: 6.5326},
	// --- State EB ---
	"EB01": {StateCode: "EB", LGACode: "01", Name: "Abakaliki", Latitude: 6.3249, Longitude: 8.1137},
	"EB02": {StateCode: "EB", LGACode: "02", Name: "Afikpo North", Latitude: 6.1954, Longitude: 8.1250},
	"EB03": {StateCode: "EB", LGACode: "03", Name: "Afikpo South (Edda)", Latitude: 6.4598, Longitude: 8.2172},
	"EB04": {StateCode: "EB", LGACode: "04", Name: "Ebonyi", Latitude: 6.2884, Longitude: 7.9069},
	"EB05": {StateCode: "EB", LGACode: "05", Name: "Ezza North", Latitude: 6.1906, Longitude: 8.3245},
	"EB06": {StateCode: "EB", LGACode: "06", Name: "Ezza South", Latitude: 6.6050, Longitude: 8.0386},
	"EB07": {StateCode: "EB", LGACode: "07", Name: "Ikwo", Latitude: 6.2805, Longitude: 8.0906},
	"EB08": {StateCode: "EB", LGACode: "08", Name: "Ishielu", Latitude: 6.3557, Longitude: 8.1983},
	"EB09": {StateCode: "EB", LGACode: "09", Name: "Ivo", Latitude: 6.3746, Longitude: 7.9936},
	"EB10": {StateCode: "EB", LGACode: "10", Name: "Izzi", Latitude: 6.1708, Longitude: 8.1855},
	"EB11": {StateCode: "EB", LGACode: "11", Name: "Ohaozara", Latitude: 6.5252, Longitude: 8.1768},
	"EB12": {StateCode: "EB", LGACode: "12", Name: "Ohaukwu", Latitude: 6.1999, Longitude: 7.8972},
	"EB13": {StateCode: "EB", LGACode: "13", Name: "Onicha", Latitude: 6.2621, Longitude: 8.3968},
	// --- State ED ---
	"ED01": {StateCode: "ED", LGACode: "01", Name: "Akoko-Edo", Latitude: 6.3958, Longitude: 5.5373},
	"ED02": {StateCode: "ED", LGACode: "02", Name: "Egor", Latitude: 6.3688, Longitude: 5.6012},
	"ED03": {StateCode: "ED", LGACode: "03", Name: "Esan Central", Latitude: 6.4699, Longitude: 5.7072},
	"ED04": {StateCode: "ED", LGACode: "04", Name: "Esan North-East", Latitude: 6.2985, Longitude: 5.3969},
	"ED05": {StateCode: "ED", LGACode: "05", Name: "Esan South-East", Latitude: 6.2007, Longitude: 5.8145},
	"ED06": {StateCode: "ED", LGACode: "06", Name: "Esan West", Latitude: 6.6151, Longitude: 5.5286},
	"ED07": {StateCode: "ED", LGACode: "07", Name: "Etsako Central", Latitude: 6.2906, Longitude: 5.5806},
	"ED08": {StateCode: "ED", LGACode: "08", Name: "Etsako East", Latitude: 6.3658, Longitude: 5.6883},
	"ED09": {StateCode: "ED", LGACode: "09", Name: "Etsako West", Latitude: 6.3847, Longitude: 5.4836},
	"ED10": {StateCode: "ED", LGACode: "10", Name: "Igueben", Latitude: 6.1809, Longitude: 5.6755},
	"ED11": {StateCode: "ED", LGACode: "11", Name: "Ikpoba-Okha", Latitude: 6.3021, Longitude: 5.6431},
	"ED12": {StateCode: "ED", LGACode: "12", Name: "Oredo", Latitude: 6.3350, Longitude: 5.6037},
	"ED13": {StateCode: "ED", LGACode: "13", Name: "Orhionmwon", Latitude: 6.2722, Longitude: 5.8868},
	"ED14": {StateCode: "ED", LGACode: "14", Name: "Ovia North-East", Latitude: 6.3760, Longitude: 5.5750},
	"ED15": {StateCode: "ED", LGACode: "15", Name: "Ovia South-West", Latitude: 6.2458, Longitude: 5.5920},
	"ED16": {StateCode: "ED", LGACode: "16", Name: "Owan East", Latitude: 6.4186, Longitude: 5.7033},
	"ED17": {StateCode: "ED", LGACode: "17", Name: "Owan West", Latitude: 6.3424, Longitude: 5.4339},
	"ED18": {StateCode: "ED", LGACode: "18", Name: "Uhunmwonde", Latitude: 6.1865, Longitude: 5.7522},
	// --- State EK ---
	"EK01": {StateCode: "EK", LGACode: "01", Name: "Ado Ekiti", Latitude: 7.6211, Longitude: 5.2215},
	"EK02": {StateCode: "EK", LGACode: "02", Name: "Efon", Latitude: 7.4916, Longitude: 5.2328},
	"EK03": {StateCode: "EK", LGACode: "03", Name: "Ekiti East", Latitude: 7.7560, Longitude: 5.3250},
	"EK04": {StateCode: "EK", LGACode: "04", Name: "Ekiti South-West", Latitude: 7.5846, Longitude: 5.0147},
	"EK05": {StateCode: "EK", LGACode: "05", Name: "Ekiti West", Latitude: 7.4868, Longitude: 5.4323},
	"EK06": {StateCode: "EK", LGACode: "06", Name: "Emure", Latitude: 7.9012, Longitude: 5.1464},
	"EK07": {StateCode: "EK", LGACode: "07", Name: "Gbonyin", Latitude: 7.5767, Longitude: 5.1984},
	"EK08": {StateCode: "EK", LGACode: "08", Name: "Ido-Osi", Latitude: 7.6519, Longitude: 5.3061},
	"EK09": {StateCode: "EK", LGACode: "09", Name: "Ijero", Latitude: 7.6708, Longitude: 5.1014},
	"EK10": {StateCode: "EK", LGACode: "10", Name: "Ikere", Latitude: 7.4670, Longitude: 5.2933},
	"EK11": {StateCode: "EK", LGACode: "11", Name: "Ikole", Latitude: 7.8214, Longitude: 5.2846},
	"EK12": {StateCode: "EK", LGACode: "12", Name: "Ilejemeje", Latitude: 7.4961, Longitude: 5.0050},
	"EK13": {StateCode: "EK", LGACode: "13", Name: "Irepodun/Ifelodun", Latitude: 7.5583, Longitude: 5.5046},
	"EK14": {StateCode: "EK", LGACode: "14", Name: "Ise/Orun", Latitude: 7.6621, Longitude: 5.1928},
	"EK15": {StateCode: "EK", LGACode: "15", Name: "Moba", Latitude: 7.5319, Longitude: 5.2098},
	"EK16": {StateCode: "EK", LGACode: "16", Name: "Oye", Latitude: 7.7047, Longitude: 5.3211},
	// --- State EN ---
	"EN01": {StateCode: "EN", LGACode: "01", Name: "Aninri", Latitude: 6.5192, Longitude: 7.4800},
	"EN02": {StateCode: "EN", LGACode: "02", Name: "Awgu", Latitude: 6.3289, Longitude: 7.5577},
	"EN03": {StateCode: "EN", LGACode: "03", Name: "Enugu East", Latitude: 6.4721, Longitude: 7.5210},
	"EN04": {StateCode: "EN", LGACode: "04", Name: "Enugu North", Latitude: 6.4584, Longitude: 7.5464},
	"EN05": {StateCode: "EN", LGACode: "05", Name: "Enugu South", Latitude: 6.4180, Longitude: 7.4981},
	"EN06": {StateCode: "EN", LGACode: "06", Name: "Ezeagu", Latitude: 6.7385, Longitude: 7.4713},
	"EN07": {StateCode: "EN", LGACode: "07", Name: "Igbo Etiti", Latitude: 6.4140, Longitude: 7.5233},
	"EN08": {StateCode: "EN", LGACode: "08", Name: "Igbo Eze North", Latitude: 6.4892, Longitude: 7.6310},
	"EN09": {StateCode: "EN", LGACode: "09", Name: "Igbo Eze South", Latitude: 6.5081, Longitude: 7.4263},
	"EN10": {StateCode: "EN", LGACode: "10", Name: "Isi Uzo", Latitude: 6.3043, Longitude: 7.6182},
	"EN11": {StateCode: "EN", LGACode: "11", Name: "Nkanu East", Latitude: 6.6587, Longitude: 7.6095},
	"EN12": {StateCode: "EN", LGACode: "12", Name: "Nkanu West", Latitude: 6.3334, Longitude: 7.3299},
	"EN13": {StateCode: "EN", LGACode: "13", Name: "Nsukka", Latitude: 6.3956, Longitude: 7.8295},
	"EN14": {StateCode: "EN", LGACode: "14", Name: "Oji River", Latitude: 6.4994, Longitude: 7.5177},
	"EN15": {StateCode: "EN", LGACode: "15", Name: "Udenu", Latitude: 6.3692, Longitude: 7.5347},
	"EN16": {StateCode: "EN", LGACode: "16", Name: "Udi", Latitude: 6.5420, Longitude: 7.6460},
	"EN17": {StateCode: "EN", LGACode: "17", Name: "Uzo-Uwani", Latitude: 6.4658, Longitude: 7.3766},
	// --- State FC ---
	"FC01": {StateCode: "FC", LGACode: "01", Name: "Abaji", Latitude: 8.4721, Longitude: 6.9537},
	"FC02": {StateCode: "FC", LGACode: "02", Name: "Abuja Municipal", Latitude: 9.0579, Longitude: 7.4951},
	"FC03": {StateCode: "FC", LGACode: "03", Name: "Bwari", Latitude: 9.1538, Longitude: 7.3220},
	"FC04": {StateCode: "FC", LGACode: "04", Name: "Gwagwalada", Latitude: 8.9431, Longitude: 7.0825},
	"FC05": {StateCode: "FC", LGACode: "05", Name: "Kuje", Latitude: 8.8797, Longitude: 7.2306},
	"FC06": {StateCode: "FC", LGACode: "06", Name: "Kwali", Latitude: 8.8842, Longitude: 7.0142},
	// --- State GO ---
	"GO01": {StateCode: "GO", LGACode: "01", Name: "Akko", Latitude: 10.3505, Longitude: 11.1047},
	"GO02": {StateCode: "GO", LGACode: "02", Name: "Balanga", Latitude: 10.1602, Longitude: 11.1824},
	"GO03": {StateCode: "GO", LGACode: "03", Name: "Billiri", Latitude: 10.4246, Longitude: 11.2746},
	"GO04": {StateCode: "GO", LGACode: "04", Name: "Dukku", Latitude: 10.2532, Longitude: 10.9643},
	"GO05": {StateCode: "GO", LGACode: "05", Name: "Funakaye", Latitude: 10.1554, Longitude: 11.3819},
	"GO06": {StateCode: "GO", LGACode: "06", Name: "Gombe", Latitude: 10.2897, Longitude: 11.1711},
	"GO07": {StateCode: "GO", LGACode: "07", Name: "Kaltungo", Latitude: 10.2453, Longitude: 11.1480},
	"GO08": {StateCode: "GO", LGACode: "08", Name: "Kwami", Latitude: 10.3205, Longitude: 11.2557},
	"GO09": {StateCode: "GO", LGACode: "09", Name: "Nafada", Latitude: 10.3394, Longitude: 11.0510},
	"GO10": {StateCode: "GO", LGACode: "10", Name: "Shongom", Latitude: 10.1356, Longitude: 11.2429},
	"GO11": {StateCode: "GO", LGACode: "11", Name: "Yamaltu/Deba", Latitude: 10.4900, Longitude: 11.2342},
	// --- State IM ---
	"IM01": {StateCode: "IM", LGACode: "01", Name: "Aboh Mbaise", Latitude: 5.5466, Longitude: 6.9691},
	"IM02": {StateCode: "IM", LGACode: "02", Name: "Ahiazu Mbaise", Latitude: 5.3563, Longitude: 7.0468},
	"IM03": {StateCode: "IM", LGACode: "03", Name: "Ehime Mbano", Latitude: 5.6207, Longitude: 7.1390},
	"IM04": {StateCode: "IM", LGACode: "04", Name: "Ezinihitte", Latitude: 5.4493, Longitude: 6.8287},
	"IM05": {StateCode: "IM", LGACode: "05", Name: "Ideato North", Latitude: 5.3515, Longitude: 7.2463},
	"IM06": {StateCode: "IM", LGACode: "06", Name: "Ideato South", Latitude: 5.7659, Longitude: 6.9604},
	"IM07": {StateCode: "IM", LGACode: "07", Name: "Ihitte/Uboma", Latitude: 5.4414, Longitude: 7.0124},
	"IM08": {StateCode: "IM", LGACode: "08", Name: "Ikeduru", Latitude: 5.5166, Longitude: 7.1201},
	"IM09": {StateCode: "IM", LGACode: "09", Name: "Isiala Mbano", Latitude: 5.5355, Longitude: 6.9154},
	"IM10": {StateCode: "IM", LGACode: "10", Name: "Isu", Latitude: 5.3317, Longitude: 7.1073},
	"IM11": {StateCode: "IM", LGACode: "11", Name: "Mbaitoli", Latitude: 5.6861, Longitude: 7.0986},
	"IM12": {StateCode: "IM", LGACode: "12", Name: "Ngor Okpala", Latitude: 5.3608, Longitude: 6.8190},
	"IM13": {StateCode: "IM", LGACode: "13", Name: "Njaba", Latitude: 5.4230, Longitude: 7.3186},
	"IM14": {StateCode: "IM", LGACode: "14", Name: "Nkwerre", Latitude: 5.5268, Longitude: 7.0068},
	"IM15": {StateCode: "IM", LGACode: "15", Name: "Nwangele", Latitude: 5.3966, Longitude: 7.0238},
	"IM16": {StateCode: "IM", LGACode: "16", Name: "Obowo", Latitude: 5.5694, Longitude: 7.1351},
	"IM17": {StateCode: "IM", LGACode: "17", Name: "Oguta", Latitude: 5.4932, Longitude: 6.8657},
	"IM18": {StateCode: "IM", LGACode: "18", Name: "Ohaji/Egbema", Latitude: 5.3373, Longitude: 7.1840},
	"IM19": {StateCode: "IM", LGACode: "19", Name: "Okigwe", Latitude: 5.7356, Longitude: 7.0246},
	"IM20": {StateCode: "IM", LGACode: "20", Name: "Onuimo", Latitude: 5.2636, Longitude: 6.8491},
	"IM21": {StateCode: "IM", LGACode: "21", Name: "Orlu", Latitude: 5.4923, Longitude: 7.0851},
	"IM22": {StateCode: "IM", LGACode: "22", Name: "Orsu", Latitude: 5.5374, Longitude: 6.9618},
	"IM23": {StateCode: "IM", LGACode: "23", Name: "Oru East", Latitude: 5.3589, Longitude: 7.0636},
	"IM24": {StateCode: "IM", LGACode: "24", Name: "Oru West", Latitude: 5.6330, Longitude: 7.1205},
	"IM25": {StateCode: "IM", LGACode: "25", Name: "Owerri Municipal", Latitude: 5.4858, Longitude: 7.0355},
	"IM26": {StateCode: "IM", LGACode: "26", Name: "Owerri North", Latitude: 5.5121, Longitude: 7.0580},
	"IM27": {StateCode: "IM", LGACode: "27", Name: "Owerri West", Latitude: 5.4620, Longitude: 6.9920},
	// --- State JI ---
	"JI01": {StateCode: "JI", LGACode: "01", Name: "Auyo", Latitude: 11.8202, Longitude: 9.2725},
	"JI02": {StateCode: "JI", LGACode: "02", Name: "Babura", Latitude: 11.6299, Longitude: 9.3502},
	"JI03": {StateCode: "JI", LGACode: "03", Name: "Biriniwa", Latitude: 11.8943, Longitude: 9.4424},
	"JI04": {StateCode: "JI", LGACode: "04", Name: "Birnin Kudu", Latitude: 11.7229, Longitude: 9.1321},
	"JI05": {StateCode: "JI", LGACode: "05", Name: "Buji", Latitude: 11.6251, Longitude: 9.5497},
	"JI06": {StateCode: "JI", LGACode: "06", Name: "Dutse", Latitude: 11.7594, Longitude: 9.3389},
	"JI07": {StateCode: "JI", LGACode: "07", Name: "Gagarawa", Latitude: 11.7150, Longitude: 9.3158},
	"JI08": {StateCode: "JI", LGACode: "08", Name: "Garki", Latitude: 11.7902, Longitude: 9.4235},
	"JI09": {StateCode: "JI", LGACode: "09", Name: "Gumel", Latitude: 11.8091, Longitude: 9.2188},
	"JI10": {StateCode: "JI", LGACode: "10", Name: "Guri", Latitude: 11.6053, Longitude: 9.4107},
	"JI11": {StateCode: "JI", LGACode: "11", Name: "Gwaram", Latitude: 11.9597, Longitude: 9.4020},
	"JI12": {StateCode: "JI", LGACode: "12", Name: "Gwiwa", Latitude: 11.6344, Longitude: 9.1224},
	"JI13": {StateCode: "JI", LGACode: "13", Name: "Hadejia", Latitude: 11.6966, Longitude: 9.6220},
	"JI14": {StateCode: "JI", LGACode: "14", Name: "Jahun", Latitude: 11.8004, Longitude: 9.3102},
	"JI15": {StateCode: "JI", LGACode: "15", Name: "Kafin Hausa", Latitude: 11.6702, Longitude: 9.3272},
	"JI16": {StateCode: "JI", LGACode: "16", Name: "Kaugama", Latitude: 11.8430, Longitude: 9.4385},
	"JI17": {StateCode: "JI", LGACode: "17", Name: "Kazaure", Latitude: 11.7668, Longitude: 9.1691},
	"JI18": {StateCode: "JI", LGACode: "18", Name: "Kiri Kasama", Latitude: 11.6109, Longitude: 9.4874},
	"JI19": {StateCode: "JI", LGACode: "19", Name: "Kiyawa", Latitude: 12.0092, Longitude: 9.3280},
	"JI20": {StateCode: "JI", LGACode: "20", Name: "Maigatari", Latitude: 11.5372, Longitude: 9.1525},
	"JI21": {StateCode: "JI", LGACode: "21", Name: "Malam Madori", Latitude: 11.7659, Longitude: 9.3885},
	"JI22": {StateCode: "JI", LGACode: "22", Name: "Miga", Latitude: 11.8110, Longitude: 9.2652},
	"JI23": {StateCode: "JI", LGACode: "23", Name: "Ringim", Latitude: 11.6325, Longitude: 9.3670},
	"JI24": {StateCode: "JI", LGACode: "24", Name: "Roni", Latitude: 11.9066, Longitude: 9.4239},
	"JI25": {StateCode: "JI", LGACode: "25", Name: "Sule Tankarkar", Latitude: 11.6963, Longitude: 9.1386},
	"JI26": {StateCode: "JI", LGACode: "26", Name: "Taura", Latitude: 11.6537, Longitude: 9.5655},
	"JI27": {StateCode: "JI", LGACode: "27", Name: "Yankwashi", Latitude: 12.0273, Longitude: 9.2279},
	// --- State KD ---
	"KD01": {StateCode: "KD", LGACode: "01", Name: "Birnin Gwari", Latitude: 10.5713, Longitude: 7.3501},
	"KD02": {StateCode: "KD", LGACode: "02", Name: "Chikun", Latitude: 10.4210, Longitude: 7.3812},
	"KD03": {StateCode: "KD", LGACode: "03", Name: "Giwa", Latitude: 10.6454, Longitude: 7.5200},
	"KD04": {StateCode: "KD", LGACode: "04", Name: "Igabi", Latitude: 10.4740, Longitude: 7.2097},
	"KD05": {StateCode: "KD", LGACode: "05", Name: "Ikara", Latitude: 10.3762, Longitude: 7.6273},
	"KD06": {StateCode: "KD", LGACode: "06", Name: "Jaba", Latitude: 10.7906, Longitude: 7.3414},
	"KD07": {StateCode: "KD", LGACode: "07", Name: "Jema'a", Latitude: 10.4661, Longitude: 7.3934},
	"KD08": {StateCode: "KD", LGACode: "08", Name: "Kachia", Latitude: 10.5413, Longitude: 7.5011},
	"KD09": {StateCode: "KD", LGACode: "09", Name: "Kaduna North", Latitude: 10.5105, Longitude: 7.4165},
	"KD10": {StateCode: "KD", LGACode: "10", Name: "Kaduna South", Latitude: 10.4611, Longitude: 7.4192},
	"KD11": {StateCode: "KD", LGACode: "11", Name: "Kagarko", Latitude: 10.7108, Longitude: 7.4796},
	"KD12": {StateCode: "KD", LGACode: "12", Name: "Kajuru", Latitude: 10.3855, Longitude: 7.2000},
	"KD13": {StateCode: "KD", LGACode: "13", Name: "Kaura", Latitude: 10.4477, Longitude: 7.6996},
	"KD14": {StateCode: "KD", LGACode: "14", Name: "Kauru", Latitude: 10.5515, Longitude: 7.3878},
	"KD15": {StateCode: "KD", LGACode: "15", Name: "Kubau", Latitude: 10.4213, Longitude: 7.4048},
	"KD16": {StateCode: "KD", LGACode: "16", Name: "Kudan", Latitude: 10.5941, Longitude: 7.5161},
	"KD17": {StateCode: "KD", LGACode: "17", Name: "Lere", Latitude: 10.5179, Longitude: 7.2467},
	"KD18": {StateCode: "KD", LGACode: "18", Name: "Makarfi", Latitude: 10.3620, Longitude: 7.5650},
	"KD19": {StateCode: "KD", LGACode: "19", Name: "Sabon Gari", Latitude: 10.7603, Longitude: 7.4056},
	"KD20": {StateCode: "KD", LGACode: "20", Name: "Sanga", Latitude: 10.2883, Longitude: 7.2301},
	"KD21": {StateCode: "KD", LGACode: "21", Name: "Soba", Latitude: 10.5170, Longitude: 7.4661},
	"KD22": {StateCode: "KD", LGACode: "22", Name: "Zangon Kataf", Latitude: 10.5621, Longitude: 7.3428},
	"KD23": {StateCode: "KD", LGACode: "23", Name: "Zaria", Latitude: 11.0855, Longitude: 7.7199},
	// --- State KN ---
	"KN01": {StateCode: "KN", LGACode: "01", Name: "Ajingi", Latitude: 12.0630, Longitude: 8.5256},
	"KN02": {StateCode: "KN", LGACode: "02", Name: "Albasu", Latitude: 11.8727, Longitude: 8.6033},
	"KN03": {StateCode: "KN", LGACode: "03", Name: "Bagwai", Latitude: 12.1371, Longitude: 8.6955},
	"KN04": {StateCode: "KN", LGACode: "04", Name: "Bebeji", Latitude: 11.9657, Longitude: 8.3852},
	"KN05": {StateCode: "KN", LGACode: "05", Name: "Bichi", Latitude: 11.8679, Longitude: 8.8028},
	"KN06": {StateCode: "KN", LGACode: "06", Name: "Bunkure", Latitude: 12.2823, Longitude: 8.5169},
	"KN07": {StateCode: "KN", LGACode: "07", Name: "Dala", Latitude: 11.9578, Longitude: 8.5689},
	"KN08": {StateCode: "KN", LGACode: "08", Name: "Dambatta", Latitude: 12.0330, Longitude: 8.6766},
	"KN09": {StateCode: "KN", LGACode: "09", Name: "Dawakin Kudu", Latitude: 12.0519, Longitude: 8.4719},
	"KN10": {StateCode: "KN", LGACode: "10", Name: "Dawakin Tofa", Latitude: 11.8481, Longitude: 8.6638},
	"KN11": {StateCode: "KN", LGACode: "11", Name: "Doguwa", Latitude: 12.2025, Longitude: 8.6551},
	"KN12": {StateCode: "KN", LGACode: "12", Name: "Fagge", Latitude: 12.0080, Longitude: 8.5310},
	"KN13": {StateCode: "KN", LGACode: "13", Name: "Gabasawa", Latitude: 11.9394, Longitude: 8.8751},
	"KN14": {StateCode: "KN", LGACode: "14", Name: "Garko", Latitude: 12.0432, Longitude: 8.5633},
	"KN15": {StateCode: "KN", LGACode: "15", Name: "Garun Mallam", Latitude: 11.9130, Longitude: 8.5803},
	"KN16": {StateCode: "KN", LGACode: "16", Name: "Gaya", Latitude: 12.0858, Longitude: 8.6916},
	"KN17": {StateCode: "KN", LGACode: "17", Name: "Gezawa", Latitude: 12.0096, Longitude: 8.4222},
	"KN18": {StateCode: "KN", LGACode: "18", Name: "Gwale", Latitude: 11.8537, Longitude: 8.7405},
	"KN19": {StateCode: "KN", LGACode: "19", Name: "Gwarzo", Latitude: 12.2520, Longitude: 8.5811},
	"KN20": {StateCode: "KN", LGACode: "20", Name: "Kabo", Latitude: 11.7800, Longitude: 8.4056},
	"KN21": {StateCode: "KN", LGACode: "21", Name: "Kano Municipal", Latitude: 12.0022, Longitude: 8.5920},
	"KN22": {StateCode: "KN", LGACode: "22", Name: "Karaye", Latitude: 12.0538, Longitude: 8.5183},
	"KN23": {StateCode: "KN", LGACode: "23", Name: "Kibiya", Latitude: 11.8753, Longitude: 8.6201},
	"KN24": {StateCode: "KN", LGACode: "24", Name: "Kiru", Latitude: 12.1494, Longitude: 8.6770},
	"KN25": {StateCode: "KN", LGACode: "25", Name: "Kumbotso", Latitude: 11.9391, Longitude: 8.3917},
	"KN26": {StateCode: "KN", LGACode: "26", Name: "Kunchi", Latitude: 11.8965, Longitude: 8.8186},
	"KN27": {StateCode: "KN", LGACode: "27", Name: "Kura", Latitude: 12.2701, Longitude: 8.4810},
	"KN28": {StateCode: "KN", LGACode: "28", Name: "Madobi", Latitude: 11.9552, Longitude: 8.5749},
	"KN29": {StateCode: "KN", LGACode: "29", Name: "Makoda", Latitude: 12.0438, Longitude: 8.6718},
	"KN30": {StateCode: "KN", LGACode: "30", Name: "Minjibir", Latitude: 12.0358, Longitude: 8.4664},
	"KN31": {StateCode: "KN", LGACode: "31", Name: "Nassarawa", Latitude: 12.0110, Longitude: 8.5620},
	"KN32": {StateCode: "KN", LGACode: "32", Name: "Rano", Latitude: 12.2090, Longitude: 8.6285},
	"KN33": {StateCode: "KN", LGACode: "33", Name: "Rimin Gado", Latitude: 11.8500, Longitude: 8.3937},
	"KN34": {StateCode: "KN", LGACode: "34", Name: "Rogo", Latitude: 11.9769, Longitude: 8.8809},
	"KN35": {StateCode: "KN", LGACode: "35", Name: "Shanono", Latitude: 12.0391, Longitude: 8.5582},
	"KN36": {StateCode: "KN", LGACode: "36", Name: "Sumaila", Latitude: 11.9122, Longitude: 8.5920},
	"KN37": {StateCode: "KN", LGACode: "37", Name: "Takai", Latitude: 12.0980, Longitude: 8.6798},
	"KN38": {StateCode: "KN", LGACode: "38", Name: "Tarauni", Latitude: 11.9760, Longitude: 8.5490},
	"KN39": {StateCode: "KN", LGACode: "39", Name: "Tofa", Latitude: 11.8744, Longitude: 8.7586},
	"KN40": {StateCode: "KN", LGACode: "40", Name: "Tsanyawa", Latitude: 12.2484, Longitude: 8.5486},
	"KN41": {StateCode: "KN", LGACode: "41", Name: "Tudun Wada", Latitude: 11.7576, Longitude: 8.4362},
	"KN42": {StateCode: "KN", LGACode: "42", Name: "Ungogo", Latitude: 12.0151, Longitude: 8.6403},
	"KN43": {StateCode: "KN", LGACode: "43", Name: "Warawa", Latitude: 12.0438, Longitude: 8.5122},
	"KN44": {StateCode: "KN", LGACode: "44", Name: "Wudil", Latitude: 11.8800, Longitude: 8.6365},
	// --- State KT ---
	"KT01": {StateCode: "KT", LGACode: "01", Name: "Bakori", Latitude: 13.0463, Longitude: 7.5507},
	"KT02": {StateCode: "KT", LGACode: "02", Name: "Batagarawa", Latitude: 12.8560, Longitude: 7.6284},
	"KT03": {StateCode: "KT", LGACode: "03", Name: "Batsari", Latitude: 13.1204, Longitude: 7.7206},
	"KT04": {StateCode: "KT", LGACode: "04", Name: "Baure", Latitude: 12.9490, Longitude: 7.4103},
	"KT05": {StateCode: "KT", LGACode: "05", Name: "Bindawa", Latitude: 12.8512, Longitude: 7.8279},
	"KT06": {StateCode: "KT", LGACode: "06", Name: "Charanchi", Latitude: 13.2656, Longitude: 7.5420},
	"KT07": {StateCode: "KT", LGACode: "07", Name: "Dandume", Latitude: 12.9411, Longitude: 7.5940},
	"KT08": {StateCode: "KT", LGACode: "08", Name: "Danja", Latitude: 13.0163, Longitude: 7.7017},
	"KT09": {StateCode: "KT", LGACode: "09", Name: "Dan Musa", Latitude: 13.0352, Longitude: 7.4970},
	"KT10": {StateCode: "KT", LGACode: "10", Name: "Daura", Latitude: 12.8314, Longitude: 7.6889},
	"KT11": {StateCode: "KT", LGACode: "11", Name: "Dutsi", Latitude: 13.1858, Longitude: 7.6802},
	"KT12": {StateCode: "KT", LGACode: "12", Name: "Dutsin Ma", Latitude: 12.8605, Longitude: 7.4006},
	"KT13": {StateCode: "KT", LGACode: "13", Name: "Faskari", Latitude: 12.9227, Longitude: 7.9002},
	"KT14": {StateCode: "KT", LGACode: "14", Name: "Funtua", Latitude: 13.0265, Longitude: 7.5884},
	"KT15": {StateCode: "KT", LGACode: "15", Name: "Ingawa", Latitude: 12.8963, Longitude: 7.6054},
	"KT16": {StateCode: "KT", LGACode: "16", Name: "Jibia", Latitude: 13.0691, Longitude: 7.7167},
	"KT17": {StateCode: "KT", LGACode: "17", Name: "Kafur", Latitude: 12.9929, Longitude: 7.4473},
	"KT18": {StateCode: "KT", LGACode: "18", Name: "Kaita", Latitude: 12.8370, Longitude: 7.7656},
	"KT19": {StateCode: "KT", LGACode: "19", Name: "Kankara", Latitude: 13.2353, Longitude: 7.6062},
	"KT20": {StateCode: "KT", LGACode: "20", Name: "Kankia", Latitude: 12.7633, Longitude: 7.4307},
	"KT21": {StateCode: "KT", LGACode: "21", Name: "Katsina", Latitude: 12.9855, Longitude: 7.6171},
	"KT22": {StateCode: "KT", LGACode: "22", Name: "Kurfi", Latitude: 13.0371, Longitude: 7.5434},
	"KT23": {StateCode: "KT", LGACode: "23", Name: "Kusada", Latitude: 12.8586, Longitude: 7.6452},
	"KT24": {StateCode: "KT", LGACode: "24", Name: "Mai'Adua", Latitude: 13.1327, Longitude: 7.7021},
	"KT25": {StateCode: "KT", LGACode: "25", Name: "Malumfashi", Latitude: 12.9224, Longitude: 7.4168},
	"KT26": {StateCode: "KT", LGACode: "26", Name: "Mani", Latitude: 12.8798, Longitude: 7.8437},
	"KT27": {StateCode: "KT", LGACode: "27", Name: "Mashi", Latitude: 13.2534, Longitude: 7.5061},
	"KT28": {StateCode: "KT", LGACode: "28", Name: "Matazu", Latitude: 12.9385, Longitude: 7.6000},
	"KT29": {StateCode: "KT", LGACode: "29", Name: "Musawa", Latitude: 13.0271, Longitude: 7.6969},
	"KT30": {StateCode: "KT", LGACode: "30", Name: "Rimi", Latitude: 13.0191, Longitude: 7.4915},
	"KT31": {StateCode: "KT", LGACode: "31", Name: "Sabuwa", Latitude: 12.8421, Longitude: 7.7084},
	"KT32": {StateCode: "KT", LGACode: "32", Name: "Safana", Latitude: 13.1923, Longitude: 7.6536},
	"KT33": {StateCode: "KT", LGACode: "33", Name: "Sandamu", Latitude: 12.8333, Longitude: 7.4188},
	"KT34": {StateCode: "KT", LGACode: "34", Name: "Zango", Latitude: 12.9602, Longitude: 7.9060},
	// --- State KE ---
	"KE01": {StateCode: "KE", LGACode: "01", Name: "Aleiro", Latitude: 12.5112, Longitude: 4.1335},
	"KE02": {StateCode: "KE", LGACode: "02", Name: "Arewa Dandi", Latitude: 12.3209, Longitude: 4.2112},
	"KE03": {StateCode: "KE", LGACode: "03", Name: "Argungu", Latitude: 12.5853, Longitude: 4.3034},
	"KE04": {StateCode: "KE", LGACode: "04", Name: "Augie", Latitude: 12.4139, Longitude: 3.9931},
	"KE05": {StateCode: "KE", LGACode: "05", Name: "Bagudo", Latitude: 12.3161, Longitude: 4.4107},
	"KE06": {StateCode: "KE", LGACode: "06", Name: "Birnin Kebbi", Latitude: 12.4504, Longitude: 4.1999},
	"KE07": {StateCode: "KE", LGACode: "07", Name: "Bunza", Latitude: 12.4060, Longitude: 4.1768},
	"KE08": {StateCode: "KE", LGACode: "08", Name: "Dandi", Latitude: 12.4812, Longitude: 4.2845},
	"KE09": {StateCode: "KE", LGACode: "09", Name: "Fakai", Latitude: 12.5001, Longitude: 4.0798},
	"KE10": {StateCode: "KE", LGACode: "10", Name: "Gwandu", Latitude: 12.2963, Longitude: 4.2717},
	"KE11": {StateCode: "KE", LGACode: "11", Name: "Jega", Latitude: 12.6507, Longitude: 4.2630},
	"KE12": {StateCode: "KE", LGACode: "12", Name: "Kalgo", Latitude: 12.3254, Longitude: 3.9834},
	"KE13": {StateCode: "KE", LGACode: "13", Name: "Koko/Besse", Latitude: 12.3876, Longitude: 4.4830},
	"KE14": {StateCode: "KE", LGACode: "14", Name: "Maiyama", Latitude: 12.4914, Longitude: 4.1712},
	"KE15": {StateCode: "KE", LGACode: "15", Name: "Ngaski", Latitude: 12.3612, Longitude: 4.1882},
	"KE16": {StateCode: "KE", LGACode: "16", Name: "Sakaba", Latitude: 12.5340, Longitude: 4.2995},
	"KE17": {StateCode: "KE", LGACode: "17", Name: "Shanga", Latitude: 12.4578, Longitude: 4.0301},
	"KE18": {StateCode: "KE", LGACode: "18", Name: "Suru", Latitude: 12.3019, Longitude: 4.3484},
	"KE19": {StateCode: "KE", LGACode: "19", Name: "Danko/Wasagu", Latitude: 12.7002, Longitude: 4.1890},
	"KE20": {StateCode: "KE", LGACode: "20", Name: "Yauri", Latitude: 12.2282, Longitude: 4.0135},
	"KE21": {StateCode: "KE", LGACode: "21", Name: "Zuru", Latitude: 12.4569, Longitude: 4.2495},
	// --- State KO ---
	"KO01": {StateCode: "KO", LGACode: "01", Name: "Adavi", Latitude: 7.8631, Longitude: 6.6669},
	"KO02": {StateCode: "KO", LGACode: "02", Name: "Ajaokuta", Latitude: 7.6728, Longitude: 6.7446},
	"KO03": {StateCode: "KO", LGACode: "03", Name: "Ankpa", Latitude: 7.9372, Longitude: 6.8368},
	"KO04": {StateCode: "KO", LGACode: "04", Name: "Bassa", Latitude: 7.7658, Longitude: 6.5265},
	"KO05": {StateCode: "KO", LGACode: "05", Name: "Dekina", Latitude: 7.6680, Longitude: 6.9441},
	"KO06": {StateCode: "KO", LGACode: "06", Name: "Ibaji", Latitude: 8.0824, Longitude: 6.6582},
	"KO07": {StateCode: "KO", LGACode: "07", Name: "Idah", Latitude: 7.7579, Longitude: 6.7102},
	"KO08": {StateCode: "KO", LGACode: "08", Name: "Igalamela-Odolu", Latitude: 7.8331, Longitude: 6.8179},
	"KO09": {StateCode: "KO", LGACode: "09", Name: "Ijumu", Latitude: 7.8520, Longitude: 6.6132},
	"KO10": {StateCode: "KO", LGACode: "10", Name: "Kabba/Bunu", Latitude: 7.6482, Longitude: 6.8051},
	"KO11": {StateCode: "KO", LGACode: "11", Name: "Kogi", Latitude: 8.0026, Longitude: 6.7964},
	"KO12": {StateCode: "KO", LGACode: "12", Name: "Lokoja", Latitude: 7.8023, Longitude: 6.7333},
	"KO13": {StateCode: "KO", LGACode: "13", Name: "Mopa-Muro", Latitude: 7.7395, Longitude: 7.0164},
	"KO14": {StateCode: "KO", LGACode: "14", Name: "Ofu", Latitude: 7.8433, Longitude: 6.7046},
	"KO15": {StateCode: "KO", LGACode: "15", Name: "Ogori/Magongo", Latitude: 7.7131, Longitude: 6.7216},
	"KO16": {StateCode: "KO", LGACode: "16", Name: "Okehi", Latitude: 7.8859, Longitude: 6.8329},
	"KO17": {StateCode: "KO", LGACode: "17", Name: "Okene", Latitude: 7.8097, Longitude: 6.5635},
	"KO18": {StateCode: "KO", LGACode: "18", Name: "Olamaboro", Latitude: 7.6538, Longitude: 6.8818},
	"KO19": {StateCode: "KO", LGACode: "19", Name: "Omala", Latitude: 8.0521, Longitude: 6.7224},
	"KO20": {StateCode: "KO", LGACode: "20", Name: "Yagba East", Latitude: 7.5801, Longitude: 6.5469},
	"KO21": {StateCode: "KO", LGACode: "21", Name: "Yagba West", Latitude: 7.8088, Longitude: 6.7829},
	// --- State KW ---
	"KW01": {StateCode: "KW", LGACode: "01", Name: "Asa", Latitude: 8.5574, Longitude: 4.4757},
	"KW02": {StateCode: "KW", LGACode: "02", Name: "Baruten", Latitude: 8.3671, Longitude: 4.5534},
	"KW03": {StateCode: "KW", LGACode: "03", Name: "Edu", Latitude: 8.6315, Longitude: 4.6456},
	"KW04": {StateCode: "KW", LGACode: "04", Name: "Ekiti", Latitude: 8.4601, Longitude: 4.3353},
	"KW05": {StateCode: "KW", LGACode: "05", Name: "Ifelodun", Latitude: 8.3623, Longitude: 4.7529},
	"KW06": {StateCode: "KW", LGACode: "06", Name: "Ilorin East", Latitude: 8.5120, Longitude: 4.5810},
	"KW07": {StateCode: "KW", LGACode: "07", Name: "Ilorin South", Latitude: 8.4680, Longitude: 4.5680},
	"KW08": {StateCode: "KW", LGACode: "08", Name: "Ilorin West", Latitude: 8.4966, Longitude: 4.5421},
	"KW09": {StateCode: "KW", LGACode: "09", Name: "Irepodun", Latitude: 8.5463, Longitude: 4.4220},
	"KW10": {StateCode: "KW", LGACode: "10", Name: "Isin", Latitude: 8.3425, Longitude: 4.6139},
	"KW11": {StateCode: "KW", LGACode: "11", Name: "Kaiama", Latitude: 8.6969, Longitude: 4.6052},
	"KW12": {StateCode: "KW", LGACode: "12", Name: "Moro", Latitude: 8.3716, Longitude: 4.3256},
	"KW13": {StateCode: "KW", LGACode: "13", Name: "Offa", Latitude: 8.4338, Longitude: 4.8252},
	"KW14": {StateCode: "KW", LGACode: "14", Name: "Oke Ero", Latitude: 8.5376, Longitude: 4.5134},
	"KW15": {StateCode: "KW", LGACode: "15", Name: "Oyun", Latitude: 8.4074, Longitude: 4.5304},
	"KW16": {StateCode: "KW", LGACode: "16", Name: "Pategi", Latitude: 8.5802, Longitude: 4.6417},
	// --- State LA ---
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
	// --- State NA ---
	"NA01": {StateCode: "NA", LGACode: "01", Name: "Akwanga", Latitude: 8.5540, Longitude: 8.4489},
	"NA02": {StateCode: "NA", LGACode: "02", Name: "Awe", Latitude: 8.3637, Longitude: 8.5266},
	"NA03": {StateCode: "NA", LGACode: "03", Name: "Doma", Latitude: 8.6281, Longitude: 8.6188},
	"NA04": {StateCode: "NA", LGACode: "04", Name: "Karu", Latitude: 9.0067, Longitude: 7.6481},
	"NA05": {StateCode: "NA", LGACode: "05", Name: "Keana", Latitude: 8.3589, Longitude: 8.7261},
	"NA06": {StateCode: "NA", LGACode: "06", Name: "Keffi", Latitude: 8.7733, Longitude: 8.4402},
	"NA07": {StateCode: "NA", LGACode: "07", Name: "Kokona", Latitude: 8.4488, Longitude: 8.4922},
	"NA08": {StateCode: "NA", LGACode: "08", Name: "Lafia", Latitude: 8.4932, Longitude: 8.5153},
	"NA09": {StateCode: "NA", LGACode: "09", Name: "Nasarawa", Latitude: 8.5429, Longitude: 8.3952},
	"NA10": {StateCode: "NA", LGACode: "10", Name: "Nasarawa Egon", Latitude: 8.3391, Longitude: 8.5871},
	"NA11": {StateCode: "NA", LGACode: "11", Name: "Obi", Latitude: 8.6935, Longitude: 8.5784},
	"NA12": {StateCode: "NA", LGACode: "12", Name: "Toto", Latitude: 8.3682, Longitude: 8.2988},
	"NA13": {StateCode: "NA", LGACode: "13", Name: "Wamba", Latitude: 8.4304, Longitude: 8.7984},
	// --- State NI ---
	"NI01": {StateCode: "NI", LGACode: "01", Name: "Agaie", Latitude: 9.6747, Longitude: 6.4905},
	"NI02": {StateCode: "NI", LGACode: "02", Name: "Agwara", Latitude: 9.4844, Longitude: 6.5682},
	"NI03": {StateCode: "NI", LGACode: "03", Name: "Bida", Latitude: 9.0804, Longitude: 6.0100},
	"NI04": {StateCode: "NI", LGACode: "04", Name: "Borgu", Latitude: 9.5774, Longitude: 6.3501},
	"NI05": {StateCode: "NI", LGACode: "05", Name: "Bosso", Latitude: 9.4796, Longitude: 6.7677},
	"NI06": {StateCode: "NI", LGACode: "06", Name: "Chanchaga", Latitude: 9.6139, Longitude: 6.5569},
	"NI07": {StateCode: "NI", LGACode: "07", Name: "Edati", Latitude: 9.5695, Longitude: 6.5338},
	"NI08": {StateCode: "NI", LGACode: "08", Name: "Gbako", Latitude: 9.6447, Longitude: 6.6415},
	"NI09": {StateCode: "NI", LGACode: "09", Name: "Gurara", Latitude: 9.6636, Longitude: 6.4368},
	"NI10": {StateCode: "NI", LGACode: "10", Name: "Katcha", Latitude: 9.4598, Longitude: 6.6287},
	"NI11": {StateCode: "NI", LGACode: "11", Name: "Kontagora", Latitude: 9.8142, Longitude: 6.6200},
	"NI12": {StateCode: "NI", LGACode: "12", Name: "Lapai", Latitude: 9.4889, Longitude: 6.3404},
	"NI13": {StateCode: "NI", LGACode: "13", Name: "Lavun", Latitude: 9.5511, Longitude: 6.8400},
	"NI14": {StateCode: "NI", LGACode: "14", Name: "Magama", Latitude: 9.6549, Longitude: 6.5282},
	"NI15": {StateCode: "NI", LGACode: "15", Name: "Mariga", Latitude: 9.5247, Longitude: 6.5452},
	"NI16": {StateCode: "NI", LGACode: "16", Name: "Mashegu", Latitude: 9.6975, Longitude: 6.6565},
	"NI17": {StateCode: "NI", LGACode: "17", Name: "Mokwa", Latitude: 9.6213, Longitude: 6.3871},
	"NI18": {StateCode: "NI", LGACode: "18", Name: "Moya", Latitude: 9.4654, Longitude: 6.7054},
	"NI19": {StateCode: "NI", LGACode: "19", Name: "Paikoro", Latitude: 9.8637, Longitude: 6.5460},
	"NI20": {StateCode: "NI", LGACode: "20", Name: "Rafi", Latitude: 9.3917, Longitude: 6.3705},
	"NI21": {StateCode: "NI", LGACode: "21", Name: "Rijau", Latitude: 9.6204, Longitude: 6.6065},
	"NI22": {StateCode: "NI", LGACode: "22", Name: "Shiroro", Latitude: 9.6655, Longitude: 6.4832},
	"NI23": {StateCode: "NI", LGACode: "23", Name: "Suleja", Latitude: 9.1806, Longitude: 7.1794},
	"NI24": {StateCode: "NI", LGACode: "24", Name: "Tafa", Latitude: 9.7611, Longitude: 6.6419},
	"NI25": {StateCode: "NI", LGACode: "25", Name: "Wushishi", Latitude: 9.5508, Longitude: 6.3566},
	// --- State OG ---
	"OG01": {StateCode: "OG", LGACode: "01", Name: "Abeokuta North", Latitude: 7.1601, Longitude: 3.3321},
	"OG02": {StateCode: "OG", LGACode: "02", Name: "Abeokuta South", Latitude: 7.1475, Longitude: 3.3619},
	"OG03": {StateCode: "OG", LGACode: "03", Name: "Ado-Odo/Ota", Latitude: 6.6841, Longitude: 3.2320},
	"OG04": {StateCode: "OG", LGACode: "04", Name: "Ewekoro", Latitude: 7.1110, Longitude: 3.1551},
	"OG05": {StateCode: "OG", LGACode: "05", Name: "Ifo", Latitude: 7.0132, Longitude: 3.5727},
	"OG06": {StateCode: "OG", LGACode: "06", Name: "Ijebu East", Latitude: 7.4276, Longitude: 3.2868},
	"OG07": {StateCode: "OG", LGACode: "07", Name: "Ijebu North", Latitude: 7.1031, Longitude: 3.3388},
	"OG08": {StateCode: "OG", LGACode: "08", Name: "Ijebu North East", Latitude: 7.1783, Longitude: 3.4465},
	"OG09": {StateCode: "OG", LGACode: "09", Name: "Ijebu Ode", Latitude: 7.1972, Longitude: 3.2418},
	"OG10": {StateCode: "OG", LGACode: "10", Name: "Ikenne", Latitude: 6.9934, Longitude: 3.4337},
	"OG11": {StateCode: "OG", LGACode: "11", Name: "Imeko Afon", Latitude: 7.3478, Longitude: 3.4250},
	"OG12": {StateCode: "OG", LGACode: "12", Name: "Ipokia", Latitude: 7.0225, Longitude: 3.1454},
	"OG13": {StateCode: "OG", LGACode: "13", Name: "Obafemi Owode", Latitude: 7.0847, Longitude: 3.6450},
	"OG14": {StateCode: "OG", LGACode: "14", Name: "Odeda", Latitude: 7.1885, Longitude: 3.3332},
	"OG15": {StateCode: "OG", LGACode: "15", Name: "Odogbolu", Latitude: 7.0583, Longitude: 3.3502},
	"OG16": {StateCode: "OG", LGACode: "16", Name: "Ogun Waterside", Latitude: 7.2311, Longitude: 3.4615},
	"OG17": {StateCode: "OG", LGACode: "17", Name: "Remo North", Latitude: 7.1549, Longitude: 3.1921},
	"OG18": {StateCode: "OG", LGACode: "18", Name: "Sagamu", Latitude: 6.9990, Longitude: 3.5104},
	"OG19": {StateCode: "OG", LGACode: "19", Name: "Yewa North", Latitude: 7.3973, Longitude: 3.3510},
	"OG20": {StateCode: "OG", LGACode: "20", Name: "Yewa South", Latitude: 6.9253, Longitude: 3.1755},
	// --- State ON ---
	"ON01": {StateCode: "ON", LGACode: "01", Name: "Akoko North-East", Latitude: 7.3179, Longitude: 5.1394},
	"ON02": {StateCode: "ON", LGACode: "02", Name: "Akoko North-West", Latitude: 7.1276, Longitude: 5.2171},
	"ON03": {StateCode: "ON", LGACode: "03", Name: "Akoko South-East", Latitude: 7.3920, Longitude: 5.3093},
	"ON04": {StateCode: "ON", LGACode: "04", Name: "Akoko South-West", Latitude: 7.2206, Longitude: 4.9990},
	"ON05": {StateCode: "ON", LGACode: "05", Name: "Akure North", Latitude: 7.3020, Longitude: 5.2510},
	"ON06": {StateCode: "ON", LGACode: "06", Name: "Akure South", Latitude: 7.2571, Longitude: 5.2058},
	"ON07": {StateCode: "ON", LGACode: "07", Name: "Ese Odo", Latitude: 7.2127, Longitude: 5.1827},
	"ON08": {StateCode: "ON", LGACode: "08", Name: "Idanre", Latitude: 7.2879, Longitude: 5.2904},
	"ON09": {StateCode: "ON", LGACode: "09", Name: "Ifedore", Latitude: 7.3068, Longitude: 5.0857},
	"ON10": {StateCode: "ON", LGACode: "10", Name: "Ilaje", Latitude: 7.1030, Longitude: 5.2776},
	"ON11": {StateCode: "ON", LGACode: "11", Name: "Ile Oluji/Okeigbo", Latitude: 7.4574, Longitude: 5.2689},
	"ON12": {StateCode: "ON", LGACode: "12", Name: "Irele", Latitude: 7.1321, Longitude: 4.9893},
	"ON13": {StateCode: "ON", LGACode: "13", Name: "Odigbo", Latitude: 7.1943, Longitude: 5.4889},
	"ON14": {StateCode: "ON", LGACode: "14", Name: "Okitipupa", Latitude: 7.2981, Longitude: 5.1771},
	"ON15": {StateCode: "ON", LGACode: "15", Name: "Ondo East", Latitude: 7.1679, Longitude: 5.1941},
	"ON16": {StateCode: "ON", LGACode: "16", Name: "Ondo West", Latitude: 7.3407, Longitude: 5.3054},
	"ON17": {StateCode: "ON", LGACode: "17", Name: "Ose", Latitude: 7.2645, Longitude: 5.0360},
	"ON18": {StateCode: "ON", LGACode: "18", Name: "Owo", Latitude: 7.1086, Longitude: 5.3543},
	// --- State OS ---
	"OS01": {StateCode: "OS", LGACode: "01", Name: "Aiyedaade", Latitude: 7.8435, Longitude: 4.4754},
	"OS02": {StateCode: "OS", LGACode: "02", Name: "Aiyedire", Latitude: 7.6532, Longitude: 4.5531},
	"OS03": {StateCode: "OS", LGACode: "03", Name: "Atakunmosa East", Latitude: 7.9176, Longitude: 4.6453},
	"OS04": {StateCode: "OS", LGACode: "04", Name: "Atakunmosa West", Latitude: 7.7462, Longitude: 4.3350},
	"OS05": {StateCode: "OS", LGACode: "05", Name: "Boluwaduro", Latitude: 7.6484, Longitude: 4.7526},
	"OS06": {StateCode: "OS", LGACode: "06", Name: "Boripe", Latitude: 8.0628, Longitude: 4.4667},
	"OS07": {StateCode: "OS", LGACode: "07", Name: "Ede North", Latitude: 7.7383, Longitude: 4.5187},
	"OS08": {StateCode: "OS", LGACode: "08", Name: "Ede South", Latitude: 7.8135, Longitude: 4.6264},
	"OS09": {StateCode: "OS", LGACode: "09", Name: "Egbedore", Latitude: 7.8324, Longitude: 4.4217},
	"OS10": {StateCode: "OS", LGACode: "10", Name: "Ejigbo", Latitude: 7.6286, Longitude: 4.6136},
	"OS11": {StateCode: "OS", LGACode: "11", Name: "Ife Central", Latitude: 7.4905, Longitude: 4.5521},
	"OS12": {StateCode: "OS", LGACode: "12", Name: "Ife East", Latitude: 7.6577, Longitude: 4.3253},
	"OS13": {StateCode: "OS", LGACode: "13", Name: "Ife North", Latitude: 7.7199, Longitude: 4.8249},
	"OS14": {StateCode: "OS", LGACode: "14", Name: "Ife South", Latitude: 7.8237, Longitude: 4.5131},
	"OS15": {StateCode: "OS", LGACode: "15", Name: "Ifedayo", Latitude: 7.6935, Longitude: 4.5301},
	"OS16": {StateCode: "OS", LGACode: "16", Name: "Ifelodun", Latitude: 7.8663, Longitude: 4.6414},
	"OS17": {StateCode: "OS", LGACode: "17", Name: "Ila", Latitude: 7.7901, Longitude: 4.3720},
	"OS18": {StateCode: "OS", LGACode: "18", Name: "Ilesa East", Latitude: 7.6342, Longitude: 4.6903},
	"OS19": {StateCode: "OS", LGACode: "19", Name: "Ilesa West", Latitude: 8.0325, Longitude: 4.5309},
	"OS20": {StateCode: "OS", LGACode: "20", Name: "Irepodun", Latitude: 7.5605, Longitude: 4.3554},
	"OS21": {StateCode: "OS", LGACode: "21", Name: "Irewole", Latitude: 7.7892, Longitude: 4.5914},
	"OS22": {StateCode: "OS", LGACode: "22", Name: "Isokan", Latitude: 7.8343, Longitude: 4.4681},
	"OS23": {StateCode: "OS", LGACode: "23", Name: "Iwo", Latitude: 7.6558, Longitude: 4.5699},
	"OS24": {StateCode: "OS", LGACode: "24", Name: "Obokun", Latitude: 7.9299, Longitude: 4.6268},
	"OS25": {StateCode: "OS", LGACode: "25", Name: "Odo Otin", Latitude: 7.7196, Longitude: 4.3415},
	"OS26": {StateCode: "OS", LGACode: "26", Name: "Ola Oluwa", Latitude: 7.6770, Longitude: 4.7684},
	"OS27": {StateCode: "OS", LGACode: "27", Name: "Olorunda", Latitude: 8.0506, Longitude: 4.4308},
	"OS28": {StateCode: "OS", LGACode: "28", Name: "Oriade", Latitude: 7.7357, Longitude: 4.5247},
	"OS29": {StateCode: "OS", LGACode: "29", Name: "Orolu", Latitude: 7.8243, Longitude: 4.6216},
	"OS30": {StateCode: "OS", LGACode: "30", Name: "Osogbo", Latitude: 7.7827, Longitude: 4.5418},
	// --- State OY ---
	"OY01": {StateCode: "OY", LGACode: "01", Name: "Afijio", Latitude: 7.4488, Longitude: 3.8656},
	"OY02": {StateCode: "OY", LGACode: "02", Name: "Akinyele", Latitude: 7.2585, Longitude: 3.9433},
	"OY03": {StateCode: "OY", LGACode: "03", Name: "Atiba", Latitude: 7.5229, Longitude: 4.0355},
	"OY04": {StateCode: "OY", LGACode: "04", Name: "Atisbo", Latitude: 7.3515, Longitude: 3.7252},
	"OY05": {StateCode: "OY", LGACode: "05", Name: "Egbeda", Latitude: 7.2537, Longitude: 4.1428},
	"OY06": {StateCode: "OY", LGACode: "06", Name: "Ibadan North", Latitude: 7.4018, Longitude: 3.9110},
	"OY07": {StateCode: "OY", LGACode: "07", Name: "Ibadan North-East", Latitude: 7.3880, Longitude: 3.9320},
	"OY08": {StateCode: "OY", LGACode: "08", Name: "Ibadan North-West", Latitude: 7.3860, Longitude: 3.8790},
	"OY09": {StateCode: "OY", LGACode: "09", Name: "Ibadan South-East", Latitude: 7.3520, Longitude: 3.9180},
	"OY10": {StateCode: "OY", LGACode: "10", Name: "Ibadan South-West", Latitude: 7.3610, Longitude: 3.8680},
	"OY11": {StateCode: "OY", LGACode: "11", Name: "Ibarapa Central", Latitude: 7.5883, Longitude: 3.9951},
	"OY12": {StateCode: "OY", LGACode: "12", Name: "Ibarapa East", Latitude: 7.2630, Longitude: 3.7155},
	"OY13": {StateCode: "OY", LGACode: "13", Name: "Ibarapa North", Latitude: 7.3252, Longitude: 4.2151},
	"OY14": {StateCode: "OY", LGACode: "14", Name: "Ido", Latitude: 7.4290, Longitude: 3.9033},
	"OY15": {StateCode: "OY", LGACode: "15", Name: "Irepo", Latitude: 7.2988, Longitude: 3.9203},
	"OY16": {StateCode: "OY", LGACode: "16", Name: "Iseyin", Latitude: 7.4716, Longitude: 4.0316},
	"OY17": {StateCode: "OY", LGACode: "17", Name: "Itesiwaju", Latitude: 7.3954, Longitude: 3.7622},
	"OY18": {StateCode: "OY", LGACode: "18", Name: "Iwajowa", Latitude: 7.2395, Longitude: 4.0805},
	"OY19": {StateCode: "OY", LGACode: "19", Name: "Kajola", Latitude: 7.6378, Longitude: 3.9211},
	"OY20": {StateCode: "OY", LGACode: "20", Name: "Lagelu", Latitude: 7.1658, Longitude: 3.7456},
	"OY21": {StateCode: "OY", LGACode: "21", Name: "Ogbomosho North", Latitude: 7.3945, Longitude: 3.9816},
	"OY22": {StateCode: "OY", LGACode: "22", Name: "Ogbomosho South", Latitude: 7.4396, Longitude: 3.8583},
	"OY23": {StateCode: "OY", LGACode: "23", Name: "Ogo Oluwa", Latitude: 7.2611, Longitude: 3.9601},
	"OY24": {StateCode: "OY", LGACode: "24", Name: "Olorunsogo", Latitude: 7.5352, Longitude: 4.0170},
	"OY25": {StateCode: "OY", LGACode: "25", Name: "Oluyole", Latitude: 7.3249, Longitude: 3.7317},
	"OY26": {StateCode: "OY", LGACode: "26", Name: "Ona Ara", Latitude: 7.2823, Longitude: 4.1586},
	"OY27": {StateCode: "OY", LGACode: "27", Name: "Orelope", Latitude: 7.6559, Longitude: 3.8210},
	"OY28": {StateCode: "OY", LGACode: "28", Name: "Ori Ire", Latitude: 7.3410, Longitude: 3.9149},
	"OY29": {StateCode: "OY", LGACode: "29", Name: "Oyo East", Latitude: 7.4296, Longitude: 4.0118},
	"OY30": {StateCode: "OY", LGACode: "30", Name: "Oyo West", Latitude: 7.4216, Longitude: 3.8064},
	"OY31": {StateCode: "OY", LGACode: "31", Name: "Saki East", Latitude: 7.2446, Longitude: 4.0233},
	"OY32": {StateCode: "OY", LGACode: "32", Name: "Saki West", Latitude: 7.5948, Longitude: 3.9685},
	"OY33": {StateCode: "OY", LGACode: "33", Name: "Surulere", Latitude: 7.2358, Longitude: 3.7337},
	// --- State PL ---
	"PL01": {StateCode: "PL", LGACode: "01", Name: "Barkin Ladi", Latitude: 9.9879, Longitude: 8.8257},
	"PL02": {StateCode: "PL", LGACode: "02", Name: "Bassa", Latitude: 9.7976, Longitude: 8.9034},
	"PL03": {StateCode: "PL", LGACode: "03", Name: "Bokkos", Latitude: 10.0620, Longitude: 8.9956},
	"PL04": {StateCode: "PL", LGACode: "04", Name: "Jos East", Latitude: 9.8906, Longitude: 8.6853},
	"PL05": {StateCode: "PL", LGACode: "05", Name: "Jos North", Latitude: 9.9271, Longitude: 8.8921},
	"PL06": {StateCode: "PL", LGACode: "06", Name: "Jos South", Latitude: 9.8020, Longitude: 8.8650},
	"PL07": {StateCode: "PL", LGACode: "07", Name: "Kanam", Latitude: 9.8827, Longitude: 8.8690},
	"PL08": {StateCode: "PL", LGACode: "08", Name: "Kanke", Latitude: 9.9579, Longitude: 8.9767},
	"PL09": {StateCode: "PL", LGACode: "09", Name: "Langtang North", Latitude: 9.9768, Longitude: 8.7720},
	"PL10": {StateCode: "PL", LGACode: "10", Name: "Langtang South", Latitude: 9.7730, Longitude: 8.9639},
	"PL11": {StateCode: "PL", LGACode: "11", Name: "Mangu", Latitude: 10.1274, Longitude: 8.9552},
	"PL12": {StateCode: "PL", LGACode: "12", Name: "Mikang", Latitude: 9.8021, Longitude: 8.6756},
	"PL13": {StateCode: "PL", LGACode: "13", Name: "Pankshin", Latitude: 9.8643, Longitude: 9.1752},
	"PL14": {StateCode: "PL", LGACode: "14", Name: "Qua'an Pan", Latitude: 9.9681, Longitude: 8.8634},
	"PL15": {StateCode: "PL", LGACode: "15", Name: "Riyom", Latitude: 9.8379, Longitude: 8.8804},
	"PL16": {StateCode: "PL", LGACode: "16", Name: "Shendam", Latitude: 10.0107, Longitude: 8.9917},
	"PL17": {StateCode: "PL", LGACode: "17", Name: "Wase", Latitude: 9.9345, Longitude: 8.7223},
	// --- State RI ---
	"RI01": {StateCode: "RI", LGACode: "01", Name: "Abua/Odual", Latitude: 4.8764, Longitude: 6.9834},
	"RI02": {StateCode: "RI", LGACode: "02", Name: "Ahoada East", Latitude: 4.6861, Longitude: 7.0611},
	"RI03": {StateCode: "RI", LGACode: "03", Name: "Ahoada West", Latitude: 4.9505, Longitude: 7.1533},
	"RI04": {StateCode: "RI", LGACode: "04", Name: "Akuku-Toru", Latitude: 4.7791, Longitude: 6.8430},
	"RI05": {StateCode: "RI", LGACode: "05", Name: "Andoni", Latitude: 4.6813, Longitude: 7.2606},
	"RI06": {StateCode: "RI", LGACode: "06", Name: "Asari-Toru", Latitude: 5.0957, Longitude: 6.9747},
	"RI07": {StateCode: "RI", LGACode: "07", Name: "Bonny", Latitude: 4.7712, Longitude: 7.0267},
	"RI08": {StateCode: "RI", LGACode: "08", Name: "Degema", Latitude: 4.8464, Longitude: 7.1344},
	"RI09": {StateCode: "RI", LGACode: "09", Name: "Eleme", Latitude: 4.7912, Longitude: 7.1205},
	"RI10": {StateCode: "RI", LGACode: "10", Name: "Emuoha", Latitude: 4.6615, Longitude: 7.1216},
	"RI11": {StateCode: "RI", LGACode: "11", Name: "Etche", Latitude: 5.0159, Longitude: 7.1129},
	"RI12": {StateCode: "RI", LGACode: "12", Name: "Gokana", Latitude: 4.6906, Longitude: 6.8333},
	"RI13": {StateCode: "RI", LGACode: "13", Name: "Ikwerre", Latitude: 4.7528, Longitude: 7.3329},
	"RI14": {StateCode: "RI", LGACode: "14", Name: "Khana", Latitude: 4.8566, Longitude: 7.0211},
	"RI15": {StateCode: "RI", LGACode: "15", Name: "Obio/Akpor", Latitude: 4.8472, Longitude: 7.0134},
	"RI16": {StateCode: "RI", LGACode: "16", Name: "Ogba/Egbema/Ndoni", Latitude: 4.8992, Longitude: 7.1494},
	"RI17": {StateCode: "RI", LGACode: "17", Name: "Ogu/Bolo", Latitude: 4.8230, Longitude: 6.8800},
	"RI18": {StateCode: "RI", LGACode: "18", Name: "Okrika", Latitude: 4.6671, Longitude: 7.1983},
	"RI19": {StateCode: "RI", LGACode: "19", Name: "Omuma", Latitude: 5.0654, Longitude: 7.0389},
	"RI20": {StateCode: "RI", LGACode: "20", Name: "Opobo/Nkoro", Latitude: 4.5934, Longitude: 6.8634},
	"RI21": {StateCode: "RI", LGACode: "21", Name: "Oyigbo", Latitude: 4.8221, Longitude: 7.0994},
	"RI22": {StateCode: "RI", LGACode: "22", Name: "Port Harcourt", Latitude: 4.8156, Longitude: 7.0498},
	"RI23": {StateCode: "RI", LGACode: "23", Name: "Tai", Latitude: 4.6887, Longitude: 7.0779},
	// --- State SO ---
	"SO01": {StateCode: "SO", LGACode: "01", Name: "Binji", Latitude: 13.1217, Longitude: 5.1812},
	"SO02": {StateCode: "SO", LGACode: "02", Name: "Bodinga", Latitude: 12.9314, Longitude: 5.2589},
	"SO03": {StateCode: "SO", LGACode: "03", Name: "Dange Shuni", Latitude: 13.1958, Longitude: 5.3511},
	"SO04": {StateCode: "SO", LGACode: "04", Name: "Gada", Latitude: 13.0244, Longitude: 5.0408},
	"SO05": {StateCode: "SO", LGACode: "05", Name: "Goronyo", Latitude: 12.9266, Longitude: 5.4584},
	"SO06": {StateCode: "SO", LGACode: "06", Name: "Gudu", Latitude: 13.3410, Longitude: 5.1725},
	"SO07": {StateCode: "SO", LGACode: "07", Name: "Gawabawa", Latitude: 13.0165, Longitude: 5.2245},
	"SO08": {StateCode: "SO", LGACode: "08", Name: "Illela", Latitude: 13.0917, Longitude: 5.3322},
	"SO09": {StateCode: "SO", LGACode: "09", Name: "Isa", Latitude: 13.1106, Longitude: 5.1275},
	"SO10": {StateCode: "SO", LGACode: "10", Name: "Kebbe", Latitude: 12.9068, Longitude: 5.3194},
	"SO11": {StateCode: "SO", LGACode: "11", Name: "Kware", Latitude: 13.2612, Longitude: 5.3107},
	"SO12": {StateCode: "SO", LGACode: "12", Name: "Rabah", Latitude: 12.9359, Longitude: 5.0311},
	"SO13": {StateCode: "SO", LGACode: "13", Name: "Sabon Birni", Latitude: 12.9981, Longitude: 5.5307},
	"SO14": {StateCode: "SO", LGACode: "14", Name: "Shagari", Latitude: 13.1019, Longitude: 5.2189},
	"SO15": {StateCode: "SO", LGACode: "15", Name: "Silame", Latitude: 12.9717, Longitude: 5.2359},
	"SO16": {StateCode: "SO", LGACode: "16", Name: "Sokoto North", Latitude: 13.0609, Longitude: 5.2476},
	"SO17": {StateCode: "SO", LGACode: "17", Name: "Sokoto South", Latitude: 13.0410, Longitude: 5.2390},
	"SO18": {StateCode: "SO", LGACode: "18", Name: "Tambuwal", Latitude: 12.9124, Longitude: 5.3961},
	"SO19": {StateCode: "SO", LGACode: "19", Name: "Tangaza", Latitude: 13.3107, Longitude: 5.2367},
	"SO20": {StateCode: "SO", LGACode: "20", Name: "Tureta", Latitude: 12.8387, Longitude: 5.0612},
	"SO21": {StateCode: "SO", LGACode: "21", Name: "Wamako", Latitude: 13.0674, Longitude: 5.2972},
	"SO22": {StateCode: "SO", LGACode: "22", Name: "Wurno", Latitude: 13.1125, Longitude: 5.1739},
	"SO23": {StateCode: "SO", LGACode: "23", Name: "Yabo", Latitude: 12.9340, Longitude: 5.2757},
	// --- State TA ---
	"TA01": {StateCode: "TA", LGACode: "01", Name: "Ardo Kola", Latitude: 8.9545, Longitude: 11.2936},
	"TA02": {StateCode: "TA", LGACode: "02", Name: "Bali", Latitude: 8.7642, Longitude: 11.3713},
	"TA03": {StateCode: "TA", LGACode: "03", Name: "Donga", Latitude: 9.0286, Longitude: 11.4635},
	"TA04": {StateCode: "TA", LGACode: "04", Name: "Gashaka", Latitude: 8.8572, Longitude: 11.1532},
	"TA05": {StateCode: "TA", LGACode: "05", Name: "Gassol", Latitude: 8.7594, Longitude: 11.5708},
	"TA06": {StateCode: "TA", LGACode: "06", Name: "Ibi", Latitude: 9.1738, Longitude: 11.2849},
	"TA07": {StateCode: "TA", LGACode: "07", Name: "Jalingo", Latitude: 8.8937, Longitude: 11.3600},
	"TA08": {StateCode: "TA", LGACode: "08", Name: "Karim Lamido", Latitude: 8.9245, Longitude: 11.4446},
	"TA09": {StateCode: "TA", LGACode: "09", Name: "Kurmi", Latitude: 8.9434, Longitude: 11.2399},
	"TA10": {StateCode: "TA", LGACode: "10", Name: "Lau", Latitude: 8.7396, Longitude: 11.4318},
	"TA11": {StateCode: "TA", LGACode: "11", Name: "Sardauna", Latitude: 9.0940, Longitude: 11.4231},
	"TA12": {StateCode: "TA", LGACode: "12", Name: "Takum", Latitude: 8.7687, Longitude: 11.1435},
	"TA13": {StateCode: "TA", LGACode: "13", Name: "Ussa", Latitude: 8.8309, Longitude: 11.6431},
	"TA14": {StateCode: "TA", LGACode: "14", Name: "Wukari", Latitude: 8.9347, Longitude: 11.3313},
	"TA15": {StateCode: "TA", LGACode: "15", Name: "Yorro", Latitude: 8.8045, Longitude: 11.3483},
	"TA16": {StateCode: "TA", LGACode: "16", Name: "Zing", Latitude: 8.9773, Longitude: 11.4596},
	// --- State YO ---
	"YO01": {StateCode: "YO", LGACode: "01", Name: "Bade", Latitude: 11.8078, Longitude: 11.8944},
	"YO02": {StateCode: "YO", LGACode: "02", Name: "Bursari", Latitude: 11.6175, Longitude: 11.9721},
	"YO03": {StateCode: "YO", LGACode: "03", Name: "Damaturu", Latitude: 11.7470, Longitude: 11.9608},
	"YO04": {StateCode: "YO", LGACode: "04", Name: "Fika", Latitude: 11.7105, Longitude: 11.7540},
	"YO05": {StateCode: "YO", LGACode: "05", Name: "Fune", Latitude: 11.6127, Longitude: 12.1716},
	"YO06": {StateCode: "YO", LGACode: "06", Name: "Geidam", Latitude: 12.0271, Longitude: 11.8857},
	"YO07": {StateCode: "YO", LGACode: "07", Name: "Gujba", Latitude: 11.7026, Longitude: 11.9377},
	"YO08": {StateCode: "YO", LGACode: "08", Name: "Gulani", Latitude: 11.7778, Longitude: 12.0454},
	"YO09": {StateCode: "YO", LGACode: "09", Name: "Jakusko", Latitude: 11.7967, Longitude: 11.8407},
	"YO10": {StateCode: "YO", LGACode: "10", Name: "Karasuwa", Latitude: 11.5929, Longitude: 12.0326},
	"YO11": {StateCode: "YO", LGACode: "11", Name: "Machina", Latitude: 11.9473, Longitude: 12.0239},
	"YO12": {StateCode: "YO", LGACode: "12", Name: "Nangere", Latitude: 11.6220, Longitude: 11.7443},
	"YO13": {StateCode: "YO", LGACode: "13", Name: "Nguru", Latitude: 11.6842, Longitude: 12.2439},
	"YO14": {StateCode: "YO", LGACode: "14", Name: "Potiskum", Latitude: 11.7880, Longitude: 11.9321},
	"YO15": {StateCode: "YO", LGACode: "15", Name: "Tarmuwa", Latitude: 11.6578, Longitude: 11.9491},
	"YO16": {StateCode: "YO", LGACode: "16", Name: "Yunusari", Latitude: 11.8306, Longitude: 12.0604},
	"YO17": {StateCode: "YO", LGACode: "17", Name: "Yusufari", Latitude: 11.7544, Longitude: 11.7910},
	// --- State ZA ---
	"ZA01": {StateCode: "ZA", LGACode: "01", Name: "Anka", Latitude: 12.2310, Longitude: 6.5928},
	"ZA02": {StateCode: "ZA", LGACode: "02", Name: "Bakura", Latitude: 12.0407, Longitude: 6.6705},
	"ZA03": {StateCode: "ZA", LGACode: "03", Name: "Birnin Magaji/Kiyaw", Latitude: 12.3051, Longitude: 6.7627},
	"ZA04": {StateCode: "ZA", LGACode: "04", Name: "Bukkuyum", Latitude: 12.1337, Longitude: 6.4524},
	"ZA05": {StateCode: "ZA", LGACode: "05", Name: "Bungudu", Latitude: 12.0359, Longitude: 6.8700},
	"ZA06": {StateCode: "ZA", LGACode: "06", Name: "Gummi", Latitude: 12.4503, Longitude: 6.5841},
	"ZA07": {StateCode: "ZA", LGACode: "07", Name: "Gusau", Latitude: 12.1702, Longitude: 6.6592},
	"ZA08": {StateCode: "ZA", LGACode: "08", Name: "Kaura Namoda", Latitude: 12.2010, Longitude: 6.7438},
	"ZA09": {StateCode: "ZA", LGACode: "09", Name: "Maradun", Latitude: 12.2199, Longitude: 6.5391},
	"ZA10": {StateCode: "ZA", LGACode: "10", Name: "Maru", Latitude: 12.0161, Longitude: 6.7310},
	"ZA11": {StateCode: "ZA", LGACode: "11", Name: "Shinkafi", Latitude: 12.3705, Longitude: 6.7223},
	"ZA12": {StateCode: "ZA", LGACode: "12", Name: "Talata Mafara", Latitude: 12.0452, Longitude: 6.4427},
	"ZA13": {StateCode: "ZA", LGACode: "13", Name: "Tsafe", Latitude: 12.1074, Longitude: 6.9423},
	"ZA14": {StateCode: "ZA", LGACode: "14", Name: "Zurmi", Latitude: 12.2112, Longitude: 6.6305},
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

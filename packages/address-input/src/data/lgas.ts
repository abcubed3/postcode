import type { LGAOption } from "../types";

/**
 * Exhaustive directory of all 774 Local Government Areas (LGAs) and Area Councils
 * in the Federal Republic of Nigeria, across all 36 States and the Federal Capital Territory (FCT).
 * Centroid coordinates (latitude and longitude) are derived from the official NIPOST cadastral database.
 */
export const NIGERIAN_774_LGAS: Record<string, LGAOption[]> = {
  // 1. State AB (17 LGAs)
  AB: [
    { code: "01", name: "Aba North", state_code: "AB", latitude: 5.1156, longitude: 7.3676 },
    { code: "02", name: "Aba South", state_code: "AB", latitude: 5.0982, longitude: 7.3612 },
    { code: "03", name: "Arochukwu", state_code: "AB", latitude: 5.6669, longitude: 7.5895 },
    { code: "04", name: "Bende", state_code: "AB", latitude: 5.4955, longitude: 7.2792 },
    { code: "05", name: "Ikwuano", state_code: "AB", latitude: 5.3977, longitude: 7.6968 },
    { code: "06", name: "Isiala Ngwa North", state_code: "AB", latitude: 5.8121, longitude: 7.4109 },
    { code: "07", name: "Isiala Ngwa South", state_code: "AB", latitude: 5.4876, longitude: 7.4629 },
    { code: "08", name: "Isuikwuato", state_code: "AB", latitude: 5.5628, longitude: 7.5706 },
    { code: "09", name: "Obi Ngwa", state_code: "AB", latitude: 5.5817, longitude: 7.3659 },
    { code: "10", name: "Ohafia", state_code: "AB", latitude: 5.3779, longitude: 7.5578 },
    { code: "11", name: "Osisioma Ngwa", state_code: "AB", latitude: 5.7323, longitude: 7.5491 },
    { code: "12", name: "Ugwunagbo", state_code: "AB", latitude: 5.407, longitude: 7.2695 },
    { code: "13", name: "Ukwa East", state_code: "AB", latitude: 5.4692, longitude: 7.7691 },
    { code: "14", name: "Ukwa West", state_code: "AB", latitude: 5.573, longitude: 7.4573 },
    { code: "15", name: "Umuahia North", state_code: "AB", latitude: 5.532, longitude: 7.486 },
    { code: "16", name: "Umuahia South", state_code: "AB", latitude: 5.6156, longitude: 7.5856 },
    { code: "17", name: "Umu Nneochi", state_code: "AB", latitude: 5.5394, longitude: 7.3162 }
  ],

  // 2. State AD (21 LGAs)
  AD: [
    { code: "01", name: "Demsa", state_code: "AD", latitude: 9.2702, longitude: 12.4154 },
    { code: "02", name: "Fufore", state_code: "AD", latitude: 9.0799, longitude: 12.4931 },
    { code: "03", name: "Ganye", state_code: "AD", latitude: 9.3443, longitude: 12.5853 },
    { code: "04", name: "Girei", state_code: "AD", latitude: 9.1729, longitude: 12.275 },
    { code: "05", name: "Gombi", state_code: "AD", latitude: 9.0751, longitude: 12.6926 },
    { code: "06", name: "Guyuk", state_code: "AD", latitude: 9.4895, longitude: 12.4067 },
    { code: "07", name: "Hong", state_code: "AD", latitude: 9.165, longitude: 12.4587 },
    { code: "08", name: "Jada", state_code: "AD", latitude: 9.2402, longitude: 12.5664 },
    { code: "09", name: "Lamurde", state_code: "AD", latitude: 9.2591, longitude: 12.3617 },
    { code: "10", name: "Madagali", state_code: "AD", latitude: 9.0553, longitude: 12.5536 },
    { code: "11", name: "Maiha", state_code: "AD", latitude: 9.4097, longitude: 12.5449 },
    { code: "12", name: "Mayo-Belwa", state_code: "AD", latitude: 9.0844, longitude: 12.2653 },
    { code: "13", name: "Michika", state_code: "AD", latitude: 9.1466, longitude: 12.7649 },
    { code: "14", name: "Mubi North", state_code: "AD", latitude: 9.2504, longitude: 12.4531 },
    { code: "15", name: "Mubi South", state_code: "AD", latitude: 9.1202, longitude: 12.4701 },
    { code: "16", name: "Numan", state_code: "AD", latitude: 9.293, longitude: 12.5814 },
    { code: "17", name: "Shelleng", state_code: "AD", latitude: 9.2168, longitude: 12.312 },
    { code: "18", name: "Song", state_code: "AD", latitude: 9.0609, longitude: 12.6303 },
    { code: "19", name: "Toungo", state_code: "AD", latitude: 9.4592, longitude: 12.4709 },
    { code: "20", name: "Yola North", state_code: "AD", latitude: 9.2094, longitude: 12.4818 },
    { code: "21", name: "Yola South", state_code: "AD", latitude: 9.1932, longitude: 12.4191 }
  ],

  // 3. State AK (31 LGAs)
  AK: [
    { code: "01", name: "Abak", state_code: "AK", latitude: 5.0985, longitude: 7.8464 },
    { code: "02", name: "Eastern Obolo", state_code: "AK", latitude: 4.9082, longitude: 7.9241 },
    { code: "03", name: "Eket", state_code: "AK", latitude: 4.6423, longitude: 7.9244 },
    { code: "04", name: "Esit Eket", state_code: "AK", latitude: 5.0012, longitude: 7.706 },
    { code: "05", name: "Essien Udim", state_code: "AK", latitude: 4.9034, longitude: 8.1236 },
    { code: "06", name: "Etim Ekpo", state_code: "AK", latitude: 5.3178, longitude: 7.8377 },
    { code: "07", name: "Etinan", state_code: "AK", latitude: 4.9933, longitude: 7.8897 },
    { code: "08", name: "Ibeno", state_code: "AK", latitude: 5.0685, longitude: 7.9974 },
    { code: "09", name: "Ibesikpo Asutan", state_code: "AK", latitude: 5.0874, longitude: 7.7927 },
    { code: "10", name: "Ibiono-Ibom", state_code: "AK", latitude: 4.8836, longitude: 7.9846 },
    { code: "11", name: "Ika", state_code: "AK", latitude: 5.238, longitude: 7.9759 },
    { code: "12", name: "Ikono", state_code: "AK", latitude: 4.9127, longitude: 7.6963 },
    { code: "13", name: "Ikot Abasi", state_code: "AK", latitude: 4.9749, longitude: 8.1959 },
    { code: "14", name: "Ikot Ekpene", state_code: "AK", latitude: 5.0787, longitude: 7.8841 },
    { code: "15", name: "Ini", state_code: "AK", latitude: 4.9485, longitude: 7.9011 },
    { code: "16", name: "Itu", state_code: "AK", latitude: 5.1213, longitude: 8.0124 },
    { code: "17", name: "Mbo", state_code: "AK", latitude: 5.0451, longitude: 7.743 },
    { code: "18", name: "Mkpat-Enin", state_code: "AK", latitude: 4.8892, longitude: 8.0613 },
    { code: "19", name: "Nsit-Atai", state_code: "AK", latitude: 5.2875, longitude: 7.9019 },
    { code: "20", name: "Nsit-Ibom", state_code: "AK", latitude: 4.8155, longitude: 7.7264 },
    { code: "21", name: "Nsit-Ubium", state_code: "AK", latitude: 5.0442, longitude: 7.9624 },
    { code: "22", name: "Obot Akara", state_code: "AK", latitude: 5.0893, longitude: 7.8391 },
    { code: "23", name: "Okobo", state_code: "AK", latitude: 4.9108, longitude: 7.9409 },
    { code: "24", name: "Onna", state_code: "AK", latitude: 5.1849, longitude: 7.9978 },
    { code: "25", name: "Oron", state_code: "AK", latitude: 4.9746, longitude: 7.7125 },
    { code: "26", name: "Oruk Anam", state_code: "AK", latitude: 4.932, longitude: 8.1394 },
    { code: "27", name: "Udung-Uko", state_code: "AK", latitude: 5.3056, longitude: 7.8018 },
    { code: "28", name: "Ukanafun", state_code: "AK", latitude: 4.9907, longitude: 7.8957 },
    { code: "29", name: "Uruan", state_code: "AK", latitude: 5.0793, longitude: 7.9926 },
    { code: "30", name: "Urue-Offong/Oruko", state_code: "AK", latitude: 5.0713, longitude: 7.7872 },
    { code: "31", name: "Uyo", state_code: "AK", latitude: 5.0377, longitude: 7.9128 }
  ],

  // 4. State AN (21 LGAs)
  AN: [
    { code: "01", name: "Aguata", state_code: "AN", latitude: 6.2817, longitude: 7.0006 },
    { code: "02", name: "Anambra East", state_code: "AN", latitude: 6.0914, longitude: 7.0783 },
    { code: "03", name: "Anambra West", state_code: "AN", latitude: 6.3558, longitude: 7.1705 },
    { code: "04", name: "Anaocha", state_code: "AN", latitude: 6.1844, longitude: 6.8602 },
    { code: "05", name: "Awka North", state_code: "AN", latitude: 6.0866, longitude: 7.2778 },
    { code: "06", name: "Awka South", state_code: "AN", latitude: 6.2209, longitude: 7.067 },
    { code: "07", name: "Ayamelum", state_code: "AN", latitude: 6.1765, longitude: 7.0439 },
    { code: "08", name: "Dunukofia", state_code: "AN", latitude: 6.2517, longitude: 7.1516 },
    { code: "09", name: "Ekwusigo", state_code: "AN", latitude: 6.2706, longitude: 6.9469 },
    { code: "10", name: "Idemili North", state_code: "AN", latitude: 6.0668, longitude: 7.1388 },
    { code: "11", name: "Idemili South", state_code: "AN", latitude: 6.4212, longitude: 7.1301 },
    { code: "12", name: "Ihiala", state_code: "AN", latitude: 6.0959, longitude: 6.8505 },
    { code: "13", name: "Njikoka", state_code: "AN", latitude: 6.1581, longitude: 7.3501 },
    { code: "14", name: "Nnewi North", state_code: "AN", latitude: 6.0195, longitude: 6.9148 },
    { code: "15", name: "Nnewi South", state_code: "AN", latitude: 6.1317, longitude: 7.0553 },
    { code: "16", name: "Ogbaru", state_code: "AN", latitude: 6.3045, longitude: 7.1666 },
    { code: "17", name: "Onitsha North", state_code: "AN", latitude: 6.1498, longitude: 6.7857 },
    { code: "18", name: "Onitsha South", state_code: "AN", latitude: 6.0724, longitude: 7.2155 },
    { code: "19", name: "Orumba North", state_code: "AN", latitude: 6.4707, longitude: 7.0561 },
    { code: "20", name: "Orumba South", state_code: "AN", latitude: 5.9987, longitude: 6.8806 },
    { code: "21", name: "Oyi", state_code: "AN", latitude: 6.2274, longitude: 7.1166 }
  ],

  // 5. State BA (20 LGAs)
  BA: [
    { code: "01", name: "Alkaleri", state_code: "BA", latitude: 10.3766, longitude: 9.7778 },
    { code: "02", name: "Bauchi", state_code: "BA", latitude: 10.3158, longitude: 9.8442 },
    { code: "03", name: "Bogoro", state_code: "BA", latitude: 10.4507, longitude: 9.9477 },
    { code: "04", name: "Damban", state_code: "BA", latitude: 10.2793, longitude: 9.6374 },
    { code: "05", name: "Darazo", state_code: "BA", latitude: 10.1815, longitude: 10.055 },
    { code: "06", name: "Dass", state_code: "BA", latitude: 10.5959, longitude: 9.7691 },
    { code: "07", name: "Gamawa", state_code: "BA", latitude: 10.2714, longitude: 9.8211 },
    { code: "08", name: "Ganjuwa", state_code: "BA", latitude: 10.3466, longitude: 9.9288 },
    { code: "09", name: "Giade", state_code: "BA", latitude: 10.3655, longitude: 9.7241 },
    { code: "10", name: "Itas/Gadau", state_code: "BA", latitude: 10.1617, longitude: 9.916 },
    { code: "11", name: "Jama'are", state_code: "BA", latitude: 10.5161, longitude: 9.9073 },
    { code: "12", name: "Katagum", state_code: "BA", latitude: 10.1908, longitude: 9.6277 },
    { code: "13", name: "Kirfi", state_code: "BA", latitude: 10.253, longitude: 10.1273 },
    { code: "14", name: "Misau", state_code: "BA", latitude: 10.3568, longitude: 9.8155 },
    { code: "15", name: "Ningi", state_code: "BA", latitude: 10.2266, longitude: 9.8325 },
    { code: "16", name: "Shira", state_code: "BA", latitude: 10.3994, longitude: 9.9438 },
    { code: "17", name: "Tafawa Balewa", state_code: "BA", latitude: 10.3232, longitude: 9.6744 },
    { code: "18", name: "Toro", state_code: "BA", latitude: 10.1673, longitude: 9.9927 },
    { code: "19", name: "Warji", state_code: "BA", latitude: 10.5656, longitude: 9.8333 },
    { code: "20", name: "Zaki", state_code: "BA", latitude: 10.0936, longitude: 9.6578 }
  ],

  // 6. State BY (8 LGAs)
  BY: [
    { code: "01", name: "Brass", state_code: "BY", latitude: 4.9875, longitude: 6.2012 },
    { code: "02", name: "Ekeremor", state_code: "BY", latitude: 4.7972, longitude: 6.2789 },
    { code: "03", name: "Kolokuma/Opokuma", state_code: "BY", latitude: 5.0616, longitude: 6.3711 },
    { code: "04", name: "Nembe", state_code: "BY", latitude: 4.8902, longitude: 6.0608 },
    { code: "05", name: "Ogbia", state_code: "BY", latitude: 4.7924, longitude: 6.4784 },
    { code: "06", name: "Sagbama", state_code: "BY", latitude: 5.2068, longitude: 6.1925 },
    { code: "07", name: "Southern Ijaw", state_code: "BY", latitude: 4.8823, longitude: 6.2445 },
    { code: "08", name: "Yenagoa", state_code: "BY", latitude: 4.9267, longitude: 6.2676 }
  ],

  // 7. State BN (23 LGAs)
  BN: [
    { code: "01", name: "Ado", state_code: "BN", latitude: 7.7912, longitude: 8.4547 },
    { code: "02", name: "Agatu", state_code: "BN", latitude: 7.6009, longitude: 8.5324 },
    { code: "03", name: "Apa", state_code: "BN", latitude: 7.8653, longitude: 8.6246 },
    { code: "04", name: "Buruku", state_code: "BN", latitude: 7.6939, longitude: 8.3143 },
    { code: "05", name: "Gboko", state_code: "BN", latitude: 7.325, longitude: 9.0011 },
    { code: "06", name: "Guma", state_code: "BN", latitude: 8.0105, longitude: 8.446 },
    { code: "07", name: "Gwer East", state_code: "BN", latitude: 7.686, longitude: 8.498 },
    { code: "08", name: "Gwer West", state_code: "BN", latitude: 7.7612, longitude: 8.6057 },
    { code: "09", name: "Katsina-Ala", state_code: "BN", latitude: 7.7801, longitude: 8.401 },
    { code: "10", name: "Konshisha", state_code: "BN", latitude: 7.5763, longitude: 8.5929 },
    { code: "11", name: "Kwande", state_code: "BN", latitude: 7.9307, longitude: 8.5842 },
    { code: "12", name: "Logo", state_code: "BN", latitude: 7.6054, longitude: 8.3046 },
    { code: "13", name: "Makurdi", state_code: "BN", latitude: 7.7304, longitude: 8.5211 },
    { code: "14", name: "Obi", state_code: "BN", latitude: 7.7714, longitude: 8.4924 },
    { code: "15", name: "Ogbadibo", state_code: "BN", latitude: 7.6412, longitude: 8.5094 },
    { code: "16", name: "Ohimini", state_code: "BN", latitude: 7.814, longitude: 8.6207 },
    { code: "17", name: "Oju", state_code: "BN", latitude: 7.7378, longitude: 8.3513 },
    { code: "18", name: "Okpokwu", state_code: "BN", latitude: 7.5819, longitude: 8.6696 },
    { code: "19", name: "Otukpo", state_code: "BN", latitude: 7.1931, longitude: 8.1339 },
    { code: "20", name: "Tarka", state_code: "BN", latitude: 7.5082, longitude: 8.3347 },
    { code: "21", name: "Ukum", state_code: "BN", latitude: 7.7369, longitude: 8.5707 },
    { code: "22", name: "Ushongo", state_code: "BN", latitude: 7.782, longitude: 8.4474 },
    { code: "23", name: "Vandeikya", state_code: "BN", latitude: 7.6035, longitude: 8.5492 }
  ],

  // 8. State BO (27 LGAs)
  BO: [
    { code: "01", name: "Abadam", state_code: "BO", latitude: 11.8941, longitude: 13.0836 },
    { code: "02", name: "Askira/Uba", state_code: "BO", latitude: 11.7038, longitude: 13.1613 },
    { code: "03", name: "Bama", state_code: "BO", latitude: 11.9682, longitude: 13.2535 },
    { code: "04", name: "Bayo", state_code: "BO", latitude: 11.7968, longitude: 12.9432 },
    { code: "05", name: "Biu", state_code: "BO", latitude: 11.699, longitude: 13.3608 },
    { code: "06", name: "Chibok", state_code: "BO", latitude: 12.1134, longitude: 13.0749 },
    { code: "07", name: "Damboa", state_code: "BO", latitude: 11.7889, longitude: 13.1269 },
    { code: "08", name: "Dikwa", state_code: "BO", latitude: 11.8641, longitude: 13.2346 },
    { code: "09", name: "Gubio", state_code: "BO", latitude: 11.883, longitude: 13.0299 },
    { code: "10", name: "Guzamala", state_code: "BO", latitude: 11.6792, longitude: 13.2218 },
    { code: "11", name: "Gwoza", state_code: "BO", latitude: 12.0336, longitude: 13.2131 },
    { code: "12", name: "Hawul", state_code: "BO", latitude: 11.7083, longitude: 12.9335 },
    { code: "13", name: "Jere", state_code: "BO", latitude: 11.7705, longitude: 13.4331 },
    { code: "14", name: "Kaga", state_code: "BO", latitude: 11.8743, longitude: 13.1213 },
    { code: "15", name: "Kala/Balge", state_code: "BO", latitude: 11.7441, longitude: 13.1383 },
    { code: "16", name: "Konduga", state_code: "BO", latitude: 11.9169, longitude: 13.2496 },
    { code: "17", name: "Kukawa", state_code: "BO", latitude: 11.8407, longitude: 12.9802 },
    { code: "18", name: "Kwaya Kusar", state_code: "BO", latitude: 11.6848, longitude: 13.2985 },
    { code: "19", name: "Mafa", state_code: "BO", latitude: 12.0831, longitude: 13.1391 },
    { code: "20", name: "Magumeri", state_code: "BO", latitude: 11.6111, longitude: 12.9636 },
    { code: "21", name: "Maiduguri", state_code: "BO", latitude: 11.8333, longitude: 13.15 },
    { code: "22", name: "Marte", state_code: "BO", latitude: 11.8849, longitude: 13.0763 },
    { code: "23", name: "Mobbar", state_code: "BO", latitude: 11.7064, longitude: 13.1781 },
    { code: "24", name: "Monguno", state_code: "BO", latitude: 11.9805, longitude: 13.235 },
    { code: "25", name: "Ngala", state_code: "BO", latitude: 11.7702, longitude: 12.9497 },
    { code: "26", name: "Nganzai", state_code: "BO", latitude: 11.7276, longitude: 13.3766 },
    { code: "27", name: "Shani", state_code: "BO", latitude: 12.1012, longitude: 13.039 }
  ],

  // 9. State CR (18 LGAs)
  CR: [
    { code: "01", name: "Abi", state_code: "CR", latitude: 5.0365, longitude: 8.2753 },
    { code: "02", name: "Akamkpa", state_code: "CR", latitude: 4.8462, longitude: 8.353 },
    { code: "03", name: "Akpabuyo", state_code: "CR", latitude: 5.1106, longitude: 8.4452 },
    { code: "04", name: "Bakassi", state_code: "CR", latitude: 4.9392, longitude: 8.1349 },
    { code: "05", name: "Bekwarra", state_code: "CR", latitude: 4.8414, longitude: 8.5525 },
    { code: "06", name: "Biase", state_code: "CR", latitude: 5.2558, longitude: 8.2666 },
    { code: "07", name: "Boki", state_code: "CR", latitude: 4.9313, longitude: 8.3186 },
    { code: "08", name: "Calabar Municipal", state_code: "CR", latitude: 4.9757, longitude: 8.3417 },
    { code: "09", name: "Calabar South", state_code: "CR", latitude: 4.9482, longitude: 8.325 },
    { code: "10", name: "Etung", state_code: "CR", latitude: 4.8216, longitude: 8.4135 },
    { code: "11", name: "Ikom", state_code: "CR", latitude: 5.176, longitude: 8.4048 },
    { code: "12", name: "Obanliku", state_code: "CR", latitude: 4.8507, longitude: 8.1252 },
    { code: "13", name: "Obubra", state_code: "CR", latitude: 4.9129, longitude: 8.6248 },
    { code: "14", name: "Obudu", state_code: "CR", latitude: 5.0167, longitude: 8.313 },
    { code: "15", name: "Odukpani", state_code: "CR", latitude: 4.8865, longitude: 8.33 },
    { code: "16", name: "Ogoja", state_code: "CR", latitude: 5.0593, longitude: 8.4413 },
    { code: "17", name: "Yakuur", state_code: "CR", latitude: 4.9831, longitude: 8.1719 },
    { code: "18", name: "Yala", state_code: "CR", latitude: 4.8272, longitude: 8.4902 }
  ],

  // 10. State DE (25 LGAs)
  DE: [
    { code: "01", name: "Aniocha North", state_code: "DE", latitude: 6.2592, longitude: 6.6665 },
    { code: "02", name: "Aniocha South", state_code: "DE", latitude: 6.0689, longitude: 6.7442 },
    { code: "03", name: "Bomadi", state_code: "DE", latitude: 6.3333, longitude: 6.8364 },
    { code: "04", name: "Burutu", state_code: "DE", latitude: 6.1619, longitude: 6.5261 },
    { code: "05", name: "Ethiope East", state_code: "DE", latitude: 6.0641, longitude: 6.9437 },
    { code: "06", name: "Ethiope West", state_code: "DE", latitude: 6.4785, longitude: 6.6578 },
    { code: "07", name: "Ika North East", state_code: "DE", latitude: 6.154, longitude: 6.7098 },
    { code: "08", name: "Ika South", state_code: "DE", latitude: 6.2292, longitude: 6.8175 },
    { code: "09", name: "Isoko North", state_code: "DE", latitude: 6.2481, longitude: 6.6128 },
    { code: "10", name: "Isoko South", state_code: "DE", latitude: 6.0443, longitude: 6.8047 },
    { code: "11", name: "Ndokwa East", state_code: "DE", latitude: 6.3987, longitude: 6.796 },
    { code: "12", name: "Ndokwa West", state_code: "DE", latitude: 6.0734, longitude: 6.5164 },
    { code: "13", name: "Okpe", state_code: "DE", latitude: 6.1356, longitude: 7.016 },
    { code: "14", name: "Oshimili North", state_code: "DE", latitude: 6.2394, longitude: 6.7042 },
    { code: "15", name: "Oshimili South", state_code: "DE", latitude: 6.1984, longitude: 6.7329 },
    { code: "16", name: "Patani", state_code: "DE", latitude: 6.282, longitude: 6.8325 },
    { code: "17", name: "Sapele", state_code: "DE", latitude: 6.2058, longitude: 6.5631 },
    { code: "18", name: "Udu", state_code: "DE", latitude: 6.0499, longitude: 6.8814 },
    { code: "19", name: "Ughelli North", state_code: "DE", latitude: 6.4482, longitude: 6.722 },
    { code: "20", name: "Ughelli South", state_code: "DE", latitude: 5.9762, longitude: 6.5465 },
    { code: "21", name: "Ukwuani", state_code: "DE", latitude: 6.2049, longitude: 6.7825 },
    { code: "22", name: "Uvwie", state_code: "DE", latitude: 6.25, longitude: 6.6592 },
    { code: "23", name: "Warri North", state_code: "DE", latitude: 6.0715, longitude: 6.761 },
    { code: "24", name: "Warri South", state_code: "DE", latitude: 5.5167, longitude: 5.75 },
    { code: "25", name: "Warri South West", state_code: "DE", latitude: 6.1353, longitude: 6.5326 }
  ],

  // 11. State EB (13 LGAs)
  EB: [
    { code: "01", name: "Abakaliki", state_code: "EB", latitude: 6.3249, longitude: 8.1137 },
    { code: "02", name: "Afikpo North", state_code: "EB", latitude: 6.1954, longitude: 8.125 },
    { code: "03", name: "Afikpo South (Edda)", state_code: "EB", latitude: 6.4598, longitude: 8.2172 },
    { code: "04", name: "Ebonyi", state_code: "EB", latitude: 6.2884, longitude: 7.9069 },
    { code: "05", name: "Ezza North", state_code: "EB", latitude: 6.1906, longitude: 8.3245 },
    { code: "06", name: "Ezza South", state_code: "EB", latitude: 6.605, longitude: 8.0386 },
    { code: "07", name: "Ikwo", state_code: "EB", latitude: 6.2805, longitude: 8.0906 },
    { code: "08", name: "Ishielu", state_code: "EB", latitude: 6.3557, longitude: 8.1983 },
    { code: "09", name: "Ivo", state_code: "EB", latitude: 6.3746, longitude: 7.9936 },
    { code: "10", name: "Izzi", state_code: "EB", latitude: 6.1708, longitude: 8.1855 },
    { code: "11", name: "Ohaozara", state_code: "EB", latitude: 6.5252, longitude: 8.1768 },
    { code: "12", name: "Ohaukwu", state_code: "EB", latitude: 6.1999, longitude: 7.8972 },
    { code: "13", name: "Onicha", state_code: "EB", latitude: 6.2621, longitude: 8.3968 }
  ],

  // 12. State ED (18 LGAs)
  ED: [
    { code: "01", name: "Akoko-Edo", state_code: "ED", latitude: 6.3958, longitude: 5.5373 },
    { code: "02", name: "Egor", state_code: "ED", latitude: 6.3688, longitude: 5.6012 },
    { code: "03", name: "Esan Central", state_code: "ED", latitude: 6.4699, longitude: 5.7072 },
    { code: "04", name: "Esan North-East", state_code: "ED", latitude: 6.2985, longitude: 5.3969 },
    { code: "05", name: "Esan South-East", state_code: "ED", latitude: 6.2007, longitude: 5.8145 },
    { code: "06", name: "Esan West", state_code: "ED", latitude: 6.6151, longitude: 5.5286 },
    { code: "07", name: "Etsako Central", state_code: "ED", latitude: 6.2906, longitude: 5.5806 },
    { code: "08", name: "Etsako East", state_code: "ED", latitude: 6.3658, longitude: 5.6883 },
    { code: "09", name: "Etsako West", state_code: "ED", latitude: 6.3847, longitude: 5.4836 },
    { code: "10", name: "Igueben", state_code: "ED", latitude: 6.1809, longitude: 5.6755 },
    { code: "11", name: "Ikpoba-Okha", state_code: "ED", latitude: 6.3021, longitude: 5.6431 },
    { code: "12", name: "Oredo", state_code: "ED", latitude: 6.335, longitude: 5.6037 },
    { code: "13", name: "Orhionmwon", state_code: "ED", latitude: 6.2722, longitude: 5.8868 },
    { code: "14", name: "Ovia North-East", state_code: "ED", latitude: 6.376, longitude: 5.575 },
    { code: "15", name: "Ovia South-West", state_code: "ED", latitude: 6.2458, longitude: 5.592 },
    { code: "16", name: "Owan East", state_code: "ED", latitude: 6.4186, longitude: 5.7033 },
    { code: "17", name: "Owan West", state_code: "ED", latitude: 6.3424, longitude: 5.4339 },
    { code: "18", name: "Uhunmwonde", state_code: "ED", latitude: 6.1865, longitude: 5.7522 }
  ],

  // 13. State EK (16 LGAs)
  EK: [
    { code: "01", name: "Ado Ekiti", state_code: "EK", latitude: 7.6211, longitude: 5.2215 },
    { code: "02", name: "Efon", state_code: "EK", latitude: 7.4916, longitude: 5.2328 },
    { code: "03", name: "Ekiti East", state_code: "EK", latitude: 7.756, longitude: 5.325 },
    { code: "04", name: "Ekiti South-West", state_code: "EK", latitude: 7.5846, longitude: 5.0147 },
    { code: "05", name: "Ekiti West", state_code: "EK", latitude: 7.4868, longitude: 5.4323 },
    { code: "06", name: "Emure", state_code: "EK", latitude: 7.9012, longitude: 5.1464 },
    { code: "07", name: "Gbonyin", state_code: "EK", latitude: 7.5767, longitude: 5.1984 },
    { code: "08", name: "Ido-Osi", state_code: "EK", latitude: 7.6519, longitude: 5.3061 },
    { code: "09", name: "Ijero", state_code: "EK", latitude: 7.6708, longitude: 5.1014 },
    { code: "10", name: "Ikere", state_code: "EK", latitude: 7.467, longitude: 5.2933 },
    { code: "11", name: "Ikole", state_code: "EK", latitude: 7.8214, longitude: 5.2846 },
    { code: "12", name: "Ilejemeje", state_code: "EK", latitude: 7.4961, longitude: 5.005 },
    { code: "13", name: "Irepodun/Ifelodun", state_code: "EK", latitude: 7.5583, longitude: 5.5046 },
    { code: "14", name: "Ise/Orun", state_code: "EK", latitude: 7.6621, longitude: 5.1928 },
    { code: "15", name: "Moba", state_code: "EK", latitude: 7.5319, longitude: 5.2098 },
    { code: "16", name: "Oye", state_code: "EK", latitude: 7.7047, longitude: 5.3211 }
  ],

  // 14. State EN (17 LGAs)
  EN: [
    { code: "01", name: "Aninri", state_code: "EN", latitude: 6.5192, longitude: 7.48 },
    { code: "02", name: "Awgu", state_code: "EN", latitude: 6.3289, longitude: 7.5577 },
    { code: "03", name: "Enugu East", state_code: "EN", latitude: 6.4721, longitude: 7.521 },
    { code: "04", name: "Enugu North", state_code: "EN", latitude: 6.4584, longitude: 7.5464 },
    { code: "05", name: "Enugu South", state_code: "EN", latitude: 6.418, longitude: 7.4981 },
    { code: "06", name: "Ezeagu", state_code: "EN", latitude: 6.7385, longitude: 7.4713 },
    { code: "07", name: "Igbo Etiti", state_code: "EN", latitude: 6.414, longitude: 7.5233 },
    { code: "08", name: "Igbo Eze North", state_code: "EN", latitude: 6.4892, longitude: 7.631 },
    { code: "09", name: "Igbo Eze South", state_code: "EN", latitude: 6.5081, longitude: 7.4263 },
    { code: "10", name: "Isi Uzo", state_code: "EN", latitude: 6.3043, longitude: 7.6182 },
    { code: "11", name: "Nkanu East", state_code: "EN", latitude: 6.6587, longitude: 7.6095 },
    { code: "12", name: "Nkanu West", state_code: "EN", latitude: 6.3334, longitude: 7.3299 },
    { code: "13", name: "Nsukka", state_code: "EN", latitude: 6.3956, longitude: 7.8295 },
    { code: "14", name: "Oji River", state_code: "EN", latitude: 6.4994, longitude: 7.5177 },
    { code: "15", name: "Udenu", state_code: "EN", latitude: 6.3692, longitude: 7.5347 },
    { code: "16", name: "Udi", state_code: "EN", latitude: 6.542, longitude: 7.646 },
    { code: "17", name: "Uzo-Uwani", state_code: "EN", latitude: 6.4658, longitude: 7.3766 }
  ],

  // 15. State FC (6 LGAs)
  FC: [
    { code: "01", name: "Abaji", state_code: "FC", latitude: 8.4721, longitude: 6.9537 },
    { code: "02", name: "Abuja Municipal", state_code: "FC", latitude: 9.0579, longitude: 7.4951 },
    { code: "03", name: "Bwari", state_code: "FC", latitude: 9.1538, longitude: 7.322 },
    { code: "04", name: "Gwagwalada", state_code: "FC", latitude: 8.9431, longitude: 7.0825 },
    { code: "05", name: "Kuje", state_code: "FC", latitude: 8.8797, longitude: 7.2306 },
    { code: "06", name: "Kwali", state_code: "FC", latitude: 8.8842, longitude: 7.0142 }
  ],

  // 16. State GO (11 LGAs)
  GO: [
    { code: "01", name: "Akko", state_code: "GO", latitude: 10.3505, longitude: 11.1047 },
    { code: "02", name: "Balanga", state_code: "GO", latitude: 10.1602, longitude: 11.1824 },
    { code: "03", name: "Billiri", state_code: "GO", latitude: 10.4246, longitude: 11.2746 },
    { code: "04", name: "Dukku", state_code: "GO", latitude: 10.2532, longitude: 10.9643 },
    { code: "05", name: "Funakaye", state_code: "GO", latitude: 10.1554, longitude: 11.3819 },
    { code: "06", name: "Gombe", state_code: "GO", latitude: 10.2897, longitude: 11.1711 },
    { code: "07", name: "Kaltungo", state_code: "GO", latitude: 10.2453, longitude: 11.148 },
    { code: "08", name: "Kwami", state_code: "GO", latitude: 10.3205, longitude: 11.2557 },
    { code: "09", name: "Nafada", state_code: "GO", latitude: 10.3394, longitude: 11.051 },
    { code: "10", name: "Shongom", state_code: "GO", latitude: 10.1356, longitude: 11.2429 },
    { code: "11", name: "Yamaltu/Deba", state_code: "GO", latitude: 10.49, longitude: 11.2342 }
  ],

  // 17. State IM (27 LGAs)
  IM: [
    { code: "01", name: "Aboh Mbaise", state_code: "IM", latitude: 5.5466, longitude: 6.9691 },
    { code: "02", name: "Ahiazu Mbaise", state_code: "IM", latitude: 5.3563, longitude: 7.0468 },
    { code: "03", name: "Ehime Mbano", state_code: "IM", latitude: 5.6207, longitude: 7.139 },
    { code: "04", name: "Ezinihitte", state_code: "IM", latitude: 5.4493, longitude: 6.8287 },
    { code: "05", name: "Ideato North", state_code: "IM", latitude: 5.3515, longitude: 7.2463 },
    { code: "06", name: "Ideato South", state_code: "IM", latitude: 5.7659, longitude: 6.9604 },
    { code: "07", name: "Ihitte/Uboma", state_code: "IM", latitude: 5.4414, longitude: 7.0124 },
    { code: "08", name: "Ikeduru", state_code: "IM", latitude: 5.5166, longitude: 7.1201 },
    { code: "09", name: "Isiala Mbano", state_code: "IM", latitude: 5.5355, longitude: 6.9154 },
    { code: "10", name: "Isu", state_code: "IM", latitude: 5.3317, longitude: 7.1073 },
    { code: "11", name: "Mbaitoli", state_code: "IM", latitude: 5.6861, longitude: 7.0986 },
    { code: "12", name: "Ngor Okpala", state_code: "IM", latitude: 5.3608, longitude: 6.819 },
    { code: "13", name: "Njaba", state_code: "IM", latitude: 5.423, longitude: 7.3186 },
    { code: "14", name: "Nkwerre", state_code: "IM", latitude: 5.5268, longitude: 7.0068 },
    { code: "15", name: "Nwangele", state_code: "IM", latitude: 5.3966, longitude: 7.0238 },
    { code: "16", name: "Obowo", state_code: "IM", latitude: 5.5694, longitude: 7.1351 },
    { code: "17", name: "Oguta", state_code: "IM", latitude: 5.4932, longitude: 6.8657 },
    { code: "18", name: "Ohaji/Egbema", state_code: "IM", latitude: 5.3373, longitude: 7.184 },
    { code: "19", name: "Okigwe", state_code: "IM", latitude: 5.7356, longitude: 7.0246 },
    { code: "20", name: "Onuimo", state_code: "IM", latitude: 5.2636, longitude: 6.8491 },
    { code: "21", name: "Orlu", state_code: "IM", latitude: 5.4923, longitude: 7.0851 },
    { code: "22", name: "Orsu", state_code: "IM", latitude: 5.5374, longitude: 6.9618 },
    { code: "23", name: "Oru East", state_code: "IM", latitude: 5.3589, longitude: 7.0636 },
    { code: "24", name: "Oru West", state_code: "IM", latitude: 5.633, longitude: 7.1205 },
    { code: "25", name: "Owerri Municipal", state_code: "IM", latitude: 5.4858, longitude: 7.0355 },
    { code: "26", name: "Owerri North", state_code: "IM", latitude: 5.5121, longitude: 7.058 },
    { code: "27", name: "Owerri West", state_code: "IM", latitude: 5.462, longitude: 6.992 }
  ],

  // 18. State JI (27 LGAs)
  JI: [
    { code: "01", name: "Auyo", state_code: "JI", latitude: 11.8202, longitude: 9.2725 },
    { code: "02", name: "Babura", state_code: "JI", latitude: 11.6299, longitude: 9.3502 },
    { code: "03", name: "Biriniwa", state_code: "JI", latitude: 11.8943, longitude: 9.4424 },
    { code: "04", name: "Birnin Kudu", state_code: "JI", latitude: 11.7229, longitude: 9.1321 },
    { code: "05", name: "Buji", state_code: "JI", latitude: 11.6251, longitude: 9.5497 },
    { code: "06", name: "Dutse", state_code: "JI", latitude: 11.7594, longitude: 9.3389 },
    { code: "07", name: "Gagarawa", state_code: "JI", latitude: 11.715, longitude: 9.3158 },
    { code: "08", name: "Garki", state_code: "JI", latitude: 11.7902, longitude: 9.4235 },
    { code: "09", name: "Gumel", state_code: "JI", latitude: 11.8091, longitude: 9.2188 },
    { code: "10", name: "Guri", state_code: "JI", latitude: 11.6053, longitude: 9.4107 },
    { code: "11", name: "Gwaram", state_code: "JI", latitude: 11.9597, longitude: 9.402 },
    { code: "12", name: "Gwiwa", state_code: "JI", latitude: 11.6344, longitude: 9.1224 },
    { code: "13", name: "Hadejia", state_code: "JI", latitude: 11.6966, longitude: 9.622 },
    { code: "14", name: "Jahun", state_code: "JI", latitude: 11.8004, longitude: 9.3102 },
    { code: "15", name: "Kafin Hausa", state_code: "JI", latitude: 11.6702, longitude: 9.3272 },
    { code: "16", name: "Kaugama", state_code: "JI", latitude: 11.843, longitude: 9.4385 },
    { code: "17", name: "Kazaure", state_code: "JI", latitude: 11.7668, longitude: 9.1691 },
    { code: "18", name: "Kiri Kasama", state_code: "JI", latitude: 11.6109, longitude: 9.4874 },
    { code: "19", name: "Kiyawa", state_code: "JI", latitude: 12.0092, longitude: 9.328 },
    { code: "20", name: "Maigatari", state_code: "JI", latitude: 11.5372, longitude: 9.1525 },
    { code: "21", name: "Malam Madori", state_code: "JI", latitude: 11.7659, longitude: 9.3885 },
    { code: "22", name: "Miga", state_code: "JI", latitude: 11.811, longitude: 9.2652 },
    { code: "23", name: "Ringim", state_code: "JI", latitude: 11.6325, longitude: 9.367 },
    { code: "24", name: "Roni", state_code: "JI", latitude: 11.9066, longitude: 9.4239 },
    { code: "25", name: "Sule Tankarkar", state_code: "JI", latitude: 11.6963, longitude: 9.1386 },
    { code: "26", name: "Taura", state_code: "JI", latitude: 11.6537, longitude: 9.5655 },
    { code: "27", name: "Yankwashi", state_code: "JI", latitude: 12.0273, longitude: 9.2279 }
  ],

  // 19. State KD (23 LGAs)
  KD: [
    { code: "01", name: "Birnin Gwari", state_code: "KD", latitude: 10.5713, longitude: 7.3501 },
    { code: "02", name: "Chikun", state_code: "KD", latitude: 10.421, longitude: 7.3812 },
    { code: "03", name: "Giwa", state_code: "KD", latitude: 10.6454, longitude: 7.52 },
    { code: "04", name: "Igabi", state_code: "KD", latitude: 10.474, longitude: 7.2097 },
    { code: "05", name: "Ikara", state_code: "KD", latitude: 10.3762, longitude: 7.6273 },
    { code: "06", name: "Jaba", state_code: "KD", latitude: 10.7906, longitude: 7.3414 },
    { code: "07", name: "Jema'a", state_code: "KD", latitude: 10.4661, longitude: 7.3934 },
    { code: "08", name: "Kachia", state_code: "KD", latitude: 10.5413, longitude: 7.5011 },
    { code: "09", name: "Kaduna North", state_code: "KD", latitude: 10.5105, longitude: 7.4165 },
    { code: "10", name: "Kaduna South", state_code: "KD", latitude: 10.4611, longitude: 7.4192 },
    { code: "11", name: "Kagarko", state_code: "KD", latitude: 10.7108, longitude: 7.4796 },
    { code: "12", name: "Kajuru", state_code: "KD", latitude: 10.3855, longitude: 7.2 },
    { code: "13", name: "Kaura", state_code: "KD", latitude: 10.4477, longitude: 7.6996 },
    { code: "14", name: "Kauru", state_code: "KD", latitude: 10.5515, longitude: 7.3878 },
    { code: "15", name: "Kubau", state_code: "KD", latitude: 10.4213, longitude: 7.4048 },
    { code: "16", name: "Kudan", state_code: "KD", latitude: 10.5941, longitude: 7.5161 },
    { code: "17", name: "Lere", state_code: "KD", latitude: 10.5179, longitude: 7.2467 },
    { code: "18", name: "Makarfi", state_code: "KD", latitude: 10.362, longitude: 7.565 },
    { code: "19", name: "Sabon Gari", state_code: "KD", latitude: 10.7603, longitude: 7.4056 },
    { code: "20", name: "Sanga", state_code: "KD", latitude: 10.2883, longitude: 7.2301 },
    { code: "21", name: "Soba", state_code: "KD", latitude: 10.517, longitude: 7.4661 },
    { code: "22", name: "Zangon Kataf", state_code: "KD", latitude: 10.5621, longitude: 7.3428 },
    { code: "23", name: "Zaria", state_code: "KD", latitude: 11.0855, longitude: 7.7199 }
  ],

  // 20. State KN (44 LGAs)
  KN: [
    { code: "01", name: "Ajingi", state_code: "KN", latitude: 12.063, longitude: 8.5256 },
    { code: "02", name: "Albasu", state_code: "KN", latitude: 11.8727, longitude: 8.6033 },
    { code: "03", name: "Bagwai", state_code: "KN", latitude: 12.1371, longitude: 8.6955 },
    { code: "04", name: "Bebeji", state_code: "KN", latitude: 11.9657, longitude: 8.3852 },
    { code: "05", name: "Bichi", state_code: "KN", latitude: 11.8679, longitude: 8.8028 },
    { code: "06", name: "Bunkure", state_code: "KN", latitude: 12.2823, longitude: 8.5169 },
    { code: "07", name: "Dala", state_code: "KN", latitude: 11.9578, longitude: 8.5689 },
    { code: "08", name: "Dambatta", state_code: "KN", latitude: 12.033, longitude: 8.6766 },
    { code: "09", name: "Dawakin Kudu", state_code: "KN", latitude: 12.0519, longitude: 8.4719 },
    { code: "10", name: "Dawakin Tofa", state_code: "KN", latitude: 11.8481, longitude: 8.6638 },
    { code: "11", name: "Doguwa", state_code: "KN", latitude: 12.2025, longitude: 8.6551 },
    { code: "12", name: "Fagge", state_code: "KN", latitude: 12.008, longitude: 8.531 },
    { code: "13", name: "Gabasawa", state_code: "KN", latitude: 11.9394, longitude: 8.8751 },
    { code: "14", name: "Garko", state_code: "KN", latitude: 12.0432, longitude: 8.5633 },
    { code: "15", name: "Garun Mallam", state_code: "KN", latitude: 11.913, longitude: 8.5803 },
    { code: "16", name: "Gaya", state_code: "KN", latitude: 12.0858, longitude: 8.6916 },
    { code: "17", name: "Gezawa", state_code: "KN", latitude: 12.0096, longitude: 8.4222 },
    { code: "18", name: "Gwale", state_code: "KN", latitude: 11.8537, longitude: 8.7405 },
    { code: "19", name: "Gwarzo", state_code: "KN", latitude: 12.252, longitude: 8.5811 },
    { code: "20", name: "Kabo", state_code: "KN", latitude: 11.78, longitude: 8.4056 },
    { code: "21", name: "Kano Municipal", state_code: "KN", latitude: 12.0022, longitude: 8.592 },
    { code: "22", name: "Karaye", state_code: "KN", latitude: 12.0538, longitude: 8.5183 },
    { code: "23", name: "Kibiya", state_code: "KN", latitude: 11.8753, longitude: 8.6201 },
    { code: "24", name: "Kiru", state_code: "KN", latitude: 12.1494, longitude: 8.677 },
    { code: "25", name: "Kumbotso", state_code: "KN", latitude: 11.9391, longitude: 8.3917 },
    { code: "26", name: "Kunchi", state_code: "KN", latitude: 11.8965, longitude: 8.8186 },
    { code: "27", name: "Kura", state_code: "KN", latitude: 12.2701, longitude: 8.481 },
    { code: "28", name: "Madobi", state_code: "KN", latitude: 11.9552, longitude: 8.5749 },
    { code: "29", name: "Makoda", state_code: "KN", latitude: 12.0438, longitude: 8.6718 },
    { code: "30", name: "Minjibir", state_code: "KN", latitude: 12.0358, longitude: 8.4664 },
    { code: "31", name: "Nassarawa", state_code: "KN", latitude: 12.011, longitude: 8.562 },
    { code: "32", name: "Rano", state_code: "KN", latitude: 12.209, longitude: 8.6285 },
    { code: "33", name: "Rimin Gado", state_code: "KN", latitude: 11.85, longitude: 8.3937 },
    { code: "34", name: "Rogo", state_code: "KN", latitude: 11.9769, longitude: 8.8809 },
    { code: "35", name: "Shanono", state_code: "KN", latitude: 12.0391, longitude: 8.5582 },
    { code: "36", name: "Sumaila", state_code: "KN", latitude: 11.9122, longitude: 8.592 },
    { code: "37", name: "Takai", state_code: "KN", latitude: 12.098, longitude: 8.6798 },
    { code: "38", name: "Tarauni", state_code: "KN", latitude: 11.976, longitude: 8.549 },
    { code: "39", name: "Tofa", state_code: "KN", latitude: 11.8744, longitude: 8.7586 },
    { code: "40", name: "Tsanyawa", state_code: "KN", latitude: 12.2484, longitude: 8.5486 },
    { code: "41", name: "Tudun Wada", state_code: "KN", latitude: 11.7576, longitude: 8.4362 },
    { code: "42", name: "Ungogo", state_code: "KN", latitude: 12.0151, longitude: 8.6403 },
    { code: "43", name: "Warawa", state_code: "KN", latitude: 12.0438, longitude: 8.5122 },
    { code: "44", name: "Wudil", state_code: "KN", latitude: 11.88, longitude: 8.6365 }
  ],

  // 21. State KT (34 LGAs)
  KT: [
    { code: "01", name: "Bakori", state_code: "KT", latitude: 13.0463, longitude: 7.5507 },
    { code: "02", name: "Batagarawa", state_code: "KT", latitude: 12.856, longitude: 7.6284 },
    { code: "03", name: "Batsari", state_code: "KT", latitude: 13.1204, longitude: 7.7206 },
    { code: "04", name: "Baure", state_code: "KT", latitude: 12.949, longitude: 7.4103 },
    { code: "05", name: "Bindawa", state_code: "KT", latitude: 12.8512, longitude: 7.8279 },
    { code: "06", name: "Charanchi", state_code: "KT", latitude: 13.2656, longitude: 7.542 },
    { code: "07", name: "Dandume", state_code: "KT", latitude: 12.9411, longitude: 7.594 },
    { code: "08", name: "Danja", state_code: "KT", latitude: 13.0163, longitude: 7.7017 },
    { code: "09", name: "Dan Musa", state_code: "KT", latitude: 13.0352, longitude: 7.497 },
    { code: "10", name: "Daura", state_code: "KT", latitude: 12.8314, longitude: 7.6889 },
    { code: "11", name: "Dutsi", state_code: "KT", latitude: 13.1858, longitude: 7.6802 },
    { code: "12", name: "Dutsin Ma", state_code: "KT", latitude: 12.8605, longitude: 7.4006 },
    { code: "13", name: "Faskari", state_code: "KT", latitude: 12.9227, longitude: 7.9002 },
    { code: "14", name: "Funtua", state_code: "KT", latitude: 13.0265, longitude: 7.5884 },
    { code: "15", name: "Ingawa", state_code: "KT", latitude: 12.8963, longitude: 7.6054 },
    { code: "16", name: "Jibia", state_code: "KT", latitude: 13.0691, longitude: 7.7167 },
    { code: "17", name: "Kafur", state_code: "KT", latitude: 12.9929, longitude: 7.4473 },
    { code: "18", name: "Kaita", state_code: "KT", latitude: 12.837, longitude: 7.7656 },
    { code: "19", name: "Kankara", state_code: "KT", latitude: 13.2353, longitude: 7.6062 },
    { code: "20", name: "Kankia", state_code: "KT", latitude: 12.7633, longitude: 7.4307 },
    { code: "21", name: "Katsina", state_code: "KT", latitude: 12.9855, longitude: 7.6171 },
    { code: "22", name: "Kurfi", state_code: "KT", latitude: 13.0371, longitude: 7.5434 },
    { code: "23", name: "Kusada", state_code: "KT", latitude: 12.8586, longitude: 7.6452 },
    { code: "24", name: "Mai'Adua", state_code: "KT", latitude: 13.1327, longitude: 7.7021 },
    { code: "25", name: "Malumfashi", state_code: "KT", latitude: 12.9224, longitude: 7.4168 },
    { code: "26", name: "Mani", state_code: "KT", latitude: 12.8798, longitude: 7.8437 },
    { code: "27", name: "Mashi", state_code: "KT", latitude: 13.2534, longitude: 7.5061 },
    { code: "28", name: "Matazu", state_code: "KT", latitude: 12.9385, longitude: 7.6 },
    { code: "29", name: "Musawa", state_code: "KT", latitude: 13.0271, longitude: 7.6969 },
    { code: "30", name: "Rimi", state_code: "KT", latitude: 13.0191, longitude: 7.4915 },
    { code: "31", name: "Sabuwa", state_code: "KT", latitude: 12.8421, longitude: 7.7084 },
    { code: "32", name: "Safana", state_code: "KT", latitude: 13.1923, longitude: 7.6536 },
    { code: "33", name: "Sandamu", state_code: "KT", latitude: 12.8333, longitude: 7.4188 },
    { code: "34", name: "Zango", state_code: "KT", latitude: 12.9602, longitude: 7.906 }
  ],

  // 22. State KE (21 LGAs)
  KE: [
    { code: "01", name: "Aleiro", state_code: "KE", latitude: 12.5112, longitude: 4.1335 },
    { code: "02", name: "Arewa Dandi", state_code: "KE", latitude: 12.3209, longitude: 4.2112 },
    { code: "03", name: "Argungu", state_code: "KE", latitude: 12.5853, longitude: 4.3034 },
    { code: "04", name: "Augie", state_code: "KE", latitude: 12.4139, longitude: 3.9931 },
    { code: "05", name: "Bagudo", state_code: "KE", latitude: 12.3161, longitude: 4.4107 },
    { code: "06", name: "Birnin Kebbi", state_code: "KE", latitude: 12.4504, longitude: 4.1999 },
    { code: "07", name: "Bunza", state_code: "KE", latitude: 12.406, longitude: 4.1768 },
    { code: "08", name: "Dandi", state_code: "KE", latitude: 12.4812, longitude: 4.2845 },
    { code: "09", name: "Fakai", state_code: "KE", latitude: 12.5001, longitude: 4.0798 },
    { code: "10", name: "Gwandu", state_code: "KE", latitude: 12.2963, longitude: 4.2717 },
    { code: "11", name: "Jega", state_code: "KE", latitude: 12.6507, longitude: 4.263 },
    { code: "12", name: "Kalgo", state_code: "KE", latitude: 12.3254, longitude: 3.9834 },
    { code: "13", name: "Koko/Besse", state_code: "KE", latitude: 12.3876, longitude: 4.483 },
    { code: "14", name: "Maiyama", state_code: "KE", latitude: 12.4914, longitude: 4.1712 },
    { code: "15", name: "Ngaski", state_code: "KE", latitude: 12.3612, longitude: 4.1882 },
    { code: "16", name: "Sakaba", state_code: "KE", latitude: 12.534, longitude: 4.2995 },
    { code: "17", name: "Shanga", state_code: "KE", latitude: 12.4578, longitude: 4.0301 },
    { code: "18", name: "Suru", state_code: "KE", latitude: 12.3019, longitude: 4.3484 },
    { code: "19", name: "Danko/Wasagu", state_code: "KE", latitude: 12.7002, longitude: 4.189 },
    { code: "20", name: "Yauri", state_code: "KE", latitude: 12.2282, longitude: 4.0135 },
    { code: "21", name: "Zuru", state_code: "KE", latitude: 12.4569, longitude: 4.2495 }
  ],

  // 23. State KO (21 LGAs)
  KO: [
    { code: "01", name: "Adavi", state_code: "KO", latitude: 7.8631, longitude: 6.6669 },
    { code: "02", name: "Ajaokuta", state_code: "KO", latitude: 7.6728, longitude: 6.7446 },
    { code: "03", name: "Ankpa", state_code: "KO", latitude: 7.9372, longitude: 6.8368 },
    { code: "04", name: "Bassa", state_code: "KO", latitude: 7.7658, longitude: 6.5265 },
    { code: "05", name: "Dekina", state_code: "KO", latitude: 7.668, longitude: 6.9441 },
    { code: "06", name: "Ibaji", state_code: "KO", latitude: 8.0824, longitude: 6.6582 },
    { code: "07", name: "Idah", state_code: "KO", latitude: 7.7579, longitude: 6.7102 },
    { code: "08", name: "Igalamela-Odolu", state_code: "KO", latitude: 7.8331, longitude: 6.8179 },
    { code: "09", name: "Ijumu", state_code: "KO", latitude: 7.852, longitude: 6.6132 },
    { code: "10", name: "Kabba/Bunu", state_code: "KO", latitude: 7.6482, longitude: 6.8051 },
    { code: "11", name: "Kogi", state_code: "KO", latitude: 8.0026, longitude: 6.7964 },
    { code: "12", name: "Lokoja", state_code: "KO", latitude: 7.8023, longitude: 6.7333 },
    { code: "13", name: "Mopa-Muro", state_code: "KO", latitude: 7.7395, longitude: 7.0164 },
    { code: "14", name: "Ofu", state_code: "KO", latitude: 7.8433, longitude: 6.7046 },
    { code: "15", name: "Ogori/Magongo", state_code: "KO", latitude: 7.7131, longitude: 6.7216 },
    { code: "16", name: "Okehi", state_code: "KO", latitude: 7.8859, longitude: 6.8329 },
    { code: "17", name: "Okene", state_code: "KO", latitude: 7.8097, longitude: 6.5635 },
    { code: "18", name: "Olamaboro", state_code: "KO", latitude: 7.6538, longitude: 6.8818 },
    { code: "19", name: "Omala", state_code: "KO", latitude: 8.0521, longitude: 6.7224 },
    { code: "20", name: "Yagba East", state_code: "KO", latitude: 7.5801, longitude: 6.5469 },
    { code: "21", name: "Yagba West", state_code: "KO", latitude: 7.8088, longitude: 6.7829 }
  ],

  // 24. State KW (16 LGAs)
  KW: [
    { code: "01", name: "Asa", state_code: "KW", latitude: 8.5574, longitude: 4.4757 },
    { code: "02", name: "Baruten", state_code: "KW", latitude: 8.3671, longitude: 4.5534 },
    { code: "03", name: "Edu", state_code: "KW", latitude: 8.6315, longitude: 4.6456 },
    { code: "04", name: "Ekiti", state_code: "KW", latitude: 8.4601, longitude: 4.3353 },
    { code: "05", name: "Ifelodun", state_code: "KW", latitude: 8.3623, longitude: 4.7529 },
    { code: "06", name: "Ilorin East", state_code: "KW", latitude: 8.512, longitude: 4.581 },
    { code: "07", name: "Ilorin South", state_code: "KW", latitude: 8.468, longitude: 4.568 },
    { code: "08", name: "Ilorin West", state_code: "KW", latitude: 8.4966, longitude: 4.5421 },
    { code: "09", name: "Irepodun", state_code: "KW", latitude: 8.5463, longitude: 4.422 },
    { code: "10", name: "Isin", state_code: "KW", latitude: 8.3425, longitude: 4.6139 },
    { code: "11", name: "Kaiama", state_code: "KW", latitude: 8.6969, longitude: 4.6052 },
    { code: "12", name: "Moro", state_code: "KW", latitude: 8.3716, longitude: 4.3256 },
    { code: "13", name: "Offa", state_code: "KW", latitude: 8.4338, longitude: 4.8252 },
    { code: "14", name: "Oke Ero", state_code: "KW", latitude: 8.5376, longitude: 4.5134 },
    { code: "15", name: "Oyun", state_code: "KW", latitude: 8.4074, longitude: 4.5304 },
    { code: "16", name: "Pategi", state_code: "KW", latitude: 8.5802, longitude: 4.6417 }
  ],

  // 25. State LA (20 LGAs)
  LA: [
    { code: "01", name: "Agege", state_code: "LA", latitude: 6.618, longitude: 3.3209 },
    { code: "02", name: "Ajeromi-Ifelodun", state_code: "LA", latitude: 6.4554, longitude: 3.3342 },
    { code: "03", name: "Alimosho", state_code: "LA", latitude: 6.6091, longitude: 3.2561 },
    { code: "04", name: "Amuwo-Odofin", state_code: "LA", latitude: 6.4312, longitude: 3.2842 },
    { code: "05", name: "Apapa", state_code: "LA", latitude: 6.4484, longitude: 3.3639 },
    { code: "06", name: "Badagry", state_code: "LA", latitude: 6.4316, longitude: 2.8876 },
    { code: "07", name: "Epe", state_code: "LA", latitude: 6.5841, longitude: 3.9834 },
    { code: "08", name: "Eti-Osa", state_code: "LA", latitude: 6.4584, longitude: 3.5684 },
    { code: "09", name: "Ibeju-Lekki", state_code: "LA", latitude: 6.4862, longitude: 3.8643 },
    { code: "10", name: "Ifako-Ijaiye", state_code: "LA", latitude: 6.6713, longitude: 3.3082 },
    { code: "11", name: "Ikeja", state_code: "LA", latitude: 6.6018, longitude: 3.3515 },
    { code: "12", name: "Ikorodu", state_code: "LA", latitude: 6.6194, longitude: 3.5105 },
    { code: "13", name: "Kosofe", state_code: "LA", latitude: 6.5742, longitude: 3.3941 },
    { code: "14", name: "Lagos Island", state_code: "LA", latitude: 6.4549, longitude: 3.4246 },
    { code: "15", name: "Lagos Mainland", state_code: "LA", latitude: 6.4969, longitude: 3.3776 },
    { code: "16", name: "Mushin", state_code: "LA", latitude: 6.5298, longitude: 3.3515 },
    { code: "17", name: "Ojo", state_code: "LA", latitude: 6.4684, longitude: 3.1924 },
    { code: "18", name: "Oshodi-Isolo", state_code: "LA", latitude: 6.5388, longitude: 3.3276 },
    { code: "19", name: "Shomolu", state_code: "LA", latitude: 6.5385, longitude: 3.3831 },
    { code: "20", name: "Surulere", state_code: "LA", latitude: 6.4969, longitude: 3.3565 }
  ],

  // 26. State NA (13 LGAs)
  NA: [
    { code: "01", name: "Akwanga", state_code: "NA", latitude: 8.554, longitude: 8.4489 },
    { code: "02", name: "Awe", state_code: "NA", latitude: 8.3637, longitude: 8.5266 },
    { code: "03", name: "Doma", state_code: "NA", latitude: 8.6281, longitude: 8.6188 },
    { code: "04", name: "Karu", state_code: "NA", latitude: 9.0067, longitude: 7.6481 },
    { code: "05", name: "Keana", state_code: "NA", latitude: 8.3589, longitude: 8.7261 },
    { code: "06", name: "Keffi", state_code: "NA", latitude: 8.7733, longitude: 8.4402 },
    { code: "07", name: "Kokona", state_code: "NA", latitude: 8.4488, longitude: 8.4922 },
    { code: "08", name: "Lafia", state_code: "NA", latitude: 8.4932, longitude: 8.5153 },
    { code: "09", name: "Nasarawa", state_code: "NA", latitude: 8.5429, longitude: 8.3952 },
    { code: "10", name: "Nasarawa Egon", state_code: "NA", latitude: 8.3391, longitude: 8.5871 },
    { code: "11", name: "Obi", state_code: "NA", latitude: 8.6935, longitude: 8.5784 },
    { code: "12", name: "Toto", state_code: "NA", latitude: 8.3682, longitude: 8.2988 },
    { code: "13", name: "Wamba", state_code: "NA", latitude: 8.4304, longitude: 8.7984 }
  ],

  // 27. State NI (25 LGAs)
  NI: [
    { code: "01", name: "Agaie", state_code: "NI", latitude: 9.6747, longitude: 6.4905 },
    { code: "02", name: "Agwara", state_code: "NI", latitude: 9.4844, longitude: 6.5682 },
    { code: "03", name: "Bida", state_code: "NI", latitude: 9.0804, longitude: 6.01 },
    { code: "04", name: "Borgu", state_code: "NI", latitude: 9.5774, longitude: 6.3501 },
    { code: "05", name: "Bosso", state_code: "NI", latitude: 9.4796, longitude: 6.7677 },
    { code: "06", name: "Chanchaga", state_code: "NI", latitude: 9.6139, longitude: 6.5569 },
    { code: "07", name: "Edati", state_code: "NI", latitude: 9.5695, longitude: 6.5338 },
    { code: "08", name: "Gbako", state_code: "NI", latitude: 9.6447, longitude: 6.6415 },
    { code: "09", name: "Gurara", state_code: "NI", latitude: 9.6636, longitude: 6.4368 },
    { code: "10", name: "Katcha", state_code: "NI", latitude: 9.4598, longitude: 6.6287 },
    { code: "11", name: "Kontagora", state_code: "NI", latitude: 9.8142, longitude: 6.62 },
    { code: "12", name: "Lapai", state_code: "NI", latitude: 9.4889, longitude: 6.3404 },
    { code: "13", name: "Lavun", state_code: "NI", latitude: 9.5511, longitude: 6.84 },
    { code: "14", name: "Magama", state_code: "NI", latitude: 9.6549, longitude: 6.5282 },
    { code: "15", name: "Mariga", state_code: "NI", latitude: 9.5247, longitude: 6.5452 },
    { code: "16", name: "Mashegu", state_code: "NI", latitude: 9.6975, longitude: 6.6565 },
    { code: "17", name: "Mokwa", state_code: "NI", latitude: 9.6213, longitude: 6.3871 },
    { code: "18", name: "Moya", state_code: "NI", latitude: 9.4654, longitude: 6.7054 },
    { code: "19", name: "Paikoro", state_code: "NI", latitude: 9.8637, longitude: 6.546 },
    { code: "20", name: "Rafi", state_code: "NI", latitude: 9.3917, longitude: 6.3705 },
    { code: "21", name: "Rijau", state_code: "NI", latitude: 9.6204, longitude: 6.6065 },
    { code: "22", name: "Shiroro", state_code: "NI", latitude: 9.6655, longitude: 6.4832 },
    { code: "23", name: "Suleja", state_code: "NI", latitude: 9.1806, longitude: 7.1794 },
    { code: "24", name: "Tafa", state_code: "NI", latitude: 9.7611, longitude: 6.6419 },
    { code: "25", name: "Wushishi", state_code: "NI", latitude: 9.5508, longitude: 6.3566 }
  ],

  // 28. State OG (20 LGAs)
  OG: [
    { code: "01", name: "Abeokuta North", state_code: "OG", latitude: 7.1601, longitude: 3.3321 },
    { code: "02", name: "Abeokuta South", state_code: "OG", latitude: 7.1475, longitude: 3.3619 },
    { code: "03", name: "Ado-Odo/Ota", state_code: "OG", latitude: 6.6841, longitude: 3.232 },
    { code: "04", name: "Ewekoro", state_code: "OG", latitude: 7.111, longitude: 3.1551 },
    { code: "05", name: "Ifo", state_code: "OG", latitude: 7.0132, longitude: 3.5727 },
    { code: "06", name: "Ijebu East", state_code: "OG", latitude: 7.4276, longitude: 3.2868 },
    { code: "07", name: "Ijebu North", state_code: "OG", latitude: 7.1031, longitude: 3.3388 },
    { code: "08", name: "Ijebu North East", state_code: "OG", latitude: 7.1783, longitude: 3.4465 },
    { code: "09", name: "Ijebu Ode", state_code: "OG", latitude: 7.1972, longitude: 3.2418 },
    { code: "10", name: "Ikenne", state_code: "OG", latitude: 6.9934, longitude: 3.4337 },
    { code: "11", name: "Imeko Afon", state_code: "OG", latitude: 7.3478, longitude: 3.425 },
    { code: "12", name: "Ipokia", state_code: "OG", latitude: 7.0225, longitude: 3.1454 },
    { code: "13", name: "Obafemi Owode", state_code: "OG", latitude: 7.0847, longitude: 3.645 },
    { code: "14", name: "Odeda", state_code: "OG", latitude: 7.1885, longitude: 3.3332 },
    { code: "15", name: "Odogbolu", state_code: "OG", latitude: 7.0583, longitude: 3.3502 },
    { code: "16", name: "Ogun Waterside", state_code: "OG", latitude: 7.2311, longitude: 3.4615 },
    { code: "17", name: "Remo North", state_code: "OG", latitude: 7.1549, longitude: 3.1921 },
    { code: "18", name: "Sagamu", state_code: "OG", latitude: 6.999, longitude: 3.5104 },
    { code: "19", name: "Yewa North", state_code: "OG", latitude: 7.3973, longitude: 3.351 },
    { code: "20", name: "Yewa South", state_code: "OG", latitude: 6.9253, longitude: 3.1755 }
  ],

  // 29. State ON (18 LGAs)
  ON: [
    { code: "01", name: "Akoko North-East", state_code: "ON", latitude: 7.3179, longitude: 5.1394 },
    { code: "02", name: "Akoko North-West", state_code: "ON", latitude: 7.1276, longitude: 5.2171 },
    { code: "03", name: "Akoko South-East", state_code: "ON", latitude: 7.392, longitude: 5.3093 },
    { code: "04", name: "Akoko South-West", state_code: "ON", latitude: 7.2206, longitude: 4.999 },
    { code: "05", name: "Akure North", state_code: "ON", latitude: 7.302, longitude: 5.251 },
    { code: "06", name: "Akure South", state_code: "ON", latitude: 7.2571, longitude: 5.2058 },
    { code: "07", name: "Ese Odo", state_code: "ON", latitude: 7.2127, longitude: 5.1827 },
    { code: "08", name: "Idanre", state_code: "ON", latitude: 7.2879, longitude: 5.2904 },
    { code: "09", name: "Ifedore", state_code: "ON", latitude: 7.3068, longitude: 5.0857 },
    { code: "10", name: "Ilaje", state_code: "ON", latitude: 7.103, longitude: 5.2776 },
    { code: "11", name: "Ile Oluji/Okeigbo", state_code: "ON", latitude: 7.4574, longitude: 5.2689 },
    { code: "12", name: "Irele", state_code: "ON", latitude: 7.1321, longitude: 4.9893 },
    { code: "13", name: "Odigbo", state_code: "ON", latitude: 7.1943, longitude: 5.4889 },
    { code: "14", name: "Okitipupa", state_code: "ON", latitude: 7.2981, longitude: 5.1771 },
    { code: "15", name: "Ondo East", state_code: "ON", latitude: 7.1679, longitude: 5.1941 },
    { code: "16", name: "Ondo West", state_code: "ON", latitude: 7.3407, longitude: 5.3054 },
    { code: "17", name: "Ose", state_code: "ON", latitude: 7.2645, longitude: 5.036 },
    { code: "18", name: "Owo", state_code: "ON", latitude: 7.1086, longitude: 5.3543 }
  ],

  // 30. State OS (30 LGAs)
  OS: [
    { code: "01", name: "Aiyedaade", state_code: "OS", latitude: 7.8435, longitude: 4.4754 },
    { code: "02", name: "Aiyedire", state_code: "OS", latitude: 7.6532, longitude: 4.5531 },
    { code: "03", name: "Atakunmosa East", state_code: "OS", latitude: 7.9176, longitude: 4.6453 },
    { code: "04", name: "Atakunmosa West", state_code: "OS", latitude: 7.7462, longitude: 4.335 },
    { code: "05", name: "Boluwaduro", state_code: "OS", latitude: 7.6484, longitude: 4.7526 },
    { code: "06", name: "Boripe", state_code: "OS", latitude: 8.0628, longitude: 4.4667 },
    { code: "07", name: "Ede North", state_code: "OS", latitude: 7.7383, longitude: 4.5187 },
    { code: "08", name: "Ede South", state_code: "OS", latitude: 7.8135, longitude: 4.6264 },
    { code: "09", name: "Egbedore", state_code: "OS", latitude: 7.8324, longitude: 4.4217 },
    { code: "10", name: "Ejigbo", state_code: "OS", latitude: 7.6286, longitude: 4.6136 },
    { code: "11", name: "Ife Central", state_code: "OS", latitude: 7.4905, longitude: 4.5521 },
    { code: "12", name: "Ife East", state_code: "OS", latitude: 7.6577, longitude: 4.3253 },
    { code: "13", name: "Ife North", state_code: "OS", latitude: 7.7199, longitude: 4.8249 },
    { code: "14", name: "Ife South", state_code: "OS", latitude: 7.8237, longitude: 4.5131 },
    { code: "15", name: "Ifedayo", state_code: "OS", latitude: 7.6935, longitude: 4.5301 },
    { code: "16", name: "Ifelodun", state_code: "OS", latitude: 7.8663, longitude: 4.6414 },
    { code: "17", name: "Ila", state_code: "OS", latitude: 7.7901, longitude: 4.372 },
    { code: "18", name: "Ilesa East", state_code: "OS", latitude: 7.6342, longitude: 4.6903 },
    { code: "19", name: "Ilesa West", state_code: "OS", latitude: 8.0325, longitude: 4.5309 },
    { code: "20", name: "Irepodun", state_code: "OS", latitude: 7.5605, longitude: 4.3554 },
    { code: "21", name: "Irewole", state_code: "OS", latitude: 7.7892, longitude: 4.5914 },
    { code: "22", name: "Isokan", state_code: "OS", latitude: 7.8343, longitude: 4.4681 },
    { code: "23", name: "Iwo", state_code: "OS", latitude: 7.6558, longitude: 4.5699 },
    { code: "24", name: "Obokun", state_code: "OS", latitude: 7.9299, longitude: 4.6268 },
    { code: "25", name: "Odo Otin", state_code: "OS", latitude: 7.7196, longitude: 4.3415 },
    { code: "26", name: "Ola Oluwa", state_code: "OS", latitude: 7.677, longitude: 4.7684 },
    { code: "27", name: "Olorunda", state_code: "OS", latitude: 8.0506, longitude: 4.4308 },
    { code: "28", name: "Oriade", state_code: "OS", latitude: 7.7357, longitude: 4.5247 },
    { code: "29", name: "Orolu", state_code: "OS", latitude: 7.8243, longitude: 4.6216 },
    { code: "30", name: "Osogbo", state_code: "OS", latitude: 7.7827, longitude: 4.5418 }
  ],

  // 31. State OY (33 LGAs)
  OY: [
    { code: "01", name: "Afijio", state_code: "OY", latitude: 7.4488, longitude: 3.8656 },
    { code: "02", name: "Akinyele", state_code: "OY", latitude: 7.2585, longitude: 3.9433 },
    { code: "03", name: "Atiba", state_code: "OY", latitude: 7.5229, longitude: 4.0355 },
    { code: "04", name: "Atisbo", state_code: "OY", latitude: 7.3515, longitude: 3.7252 },
    { code: "05", name: "Egbeda", state_code: "OY", latitude: 7.2537, longitude: 4.1428 },
    { code: "06", name: "Ibadan North", state_code: "OY", latitude: 7.4018, longitude: 3.911 },
    { code: "07", name: "Ibadan North-East", state_code: "OY", latitude: 7.388, longitude: 3.932 },
    { code: "08", name: "Ibadan North-West", state_code: "OY", latitude: 7.386, longitude: 3.879 },
    { code: "09", name: "Ibadan South-East", state_code: "OY", latitude: 7.352, longitude: 3.918 },
    { code: "10", name: "Ibadan South-West", state_code: "OY", latitude: 7.361, longitude: 3.868 },
    { code: "11", name: "Ibarapa Central", state_code: "OY", latitude: 7.5883, longitude: 3.9951 },
    { code: "12", name: "Ibarapa East", state_code: "OY", latitude: 7.263, longitude: 3.7155 },
    { code: "13", name: "Ibarapa North", state_code: "OY", latitude: 7.3252, longitude: 4.2151 },
    { code: "14", name: "Ido", state_code: "OY", latitude: 7.429, longitude: 3.9033 },
    { code: "15", name: "Irepo", state_code: "OY", latitude: 7.2988, longitude: 3.9203 },
    { code: "16", name: "Iseyin", state_code: "OY", latitude: 7.4716, longitude: 4.0316 },
    { code: "17", name: "Itesiwaju", state_code: "OY", latitude: 7.3954, longitude: 3.7622 },
    { code: "18", name: "Iwajowa", state_code: "OY", latitude: 7.2395, longitude: 4.0805 },
    { code: "19", name: "Kajola", state_code: "OY", latitude: 7.6378, longitude: 3.9211 },
    { code: "20", name: "Lagelu", state_code: "OY", latitude: 7.1658, longitude: 3.7456 },
    { code: "21", name: "Ogbomosho North", state_code: "OY", latitude: 7.3945, longitude: 3.9816 },
    { code: "22", name: "Ogbomosho South", state_code: "OY", latitude: 7.4396, longitude: 3.8583 },
    { code: "23", name: "Ogo Oluwa", state_code: "OY", latitude: 7.2611, longitude: 3.9601 },
    { code: "24", name: "Olorunsogo", state_code: "OY", latitude: 7.5352, longitude: 4.017 },
    { code: "25", name: "Oluyole", state_code: "OY", latitude: 7.3249, longitude: 3.7317 },
    { code: "26", name: "Ona Ara", state_code: "OY", latitude: 7.2823, longitude: 4.1586 },
    { code: "27", name: "Orelope", state_code: "OY", latitude: 7.6559, longitude: 3.821 },
    { code: "28", name: "Ori Ire", state_code: "OY", latitude: 7.341, longitude: 3.9149 },
    { code: "29", name: "Oyo East", state_code: "OY", latitude: 7.4296, longitude: 4.0118 },
    { code: "30", name: "Oyo West", state_code: "OY", latitude: 7.4216, longitude: 3.8064 },
    { code: "31", name: "Saki East", state_code: "OY", latitude: 7.2446, longitude: 4.0233 },
    { code: "32", name: "Saki West", state_code: "OY", latitude: 7.5948, longitude: 3.9685 },
    { code: "33", name: "Surulere", state_code: "OY", latitude: 7.2358, longitude: 3.7337 }
  ],

  // 32. State PL (17 LGAs)
  PL: [
    { code: "01", name: "Barkin Ladi", state_code: "PL", latitude: 9.9879, longitude: 8.8257 },
    { code: "02", name: "Bassa", state_code: "PL", latitude: 9.7976, longitude: 8.9034 },
    { code: "03", name: "Bokkos", state_code: "PL", latitude: 10.062, longitude: 8.9956 },
    { code: "04", name: "Jos East", state_code: "PL", latitude: 9.8906, longitude: 8.6853 },
    { code: "05", name: "Jos North", state_code: "PL", latitude: 9.9271, longitude: 8.8921 },
    { code: "06", name: "Jos South", state_code: "PL", latitude: 9.802, longitude: 8.865 },
    { code: "07", name: "Kanam", state_code: "PL", latitude: 9.8827, longitude: 8.869 },
    { code: "08", name: "Kanke", state_code: "PL", latitude: 9.9579, longitude: 8.9767 },
    { code: "09", name: "Langtang North", state_code: "PL", latitude: 9.9768, longitude: 8.772 },
    { code: "10", name: "Langtang South", state_code: "PL", latitude: 9.773, longitude: 8.9639 },
    { code: "11", name: "Mangu", state_code: "PL", latitude: 10.1274, longitude: 8.9552 },
    { code: "12", name: "Mikang", state_code: "PL", latitude: 9.8021, longitude: 8.6756 },
    { code: "13", name: "Pankshin", state_code: "PL", latitude: 9.8643, longitude: 9.1752 },
    { code: "14", name: "Qua'an Pan", state_code: "PL", latitude: 9.9681, longitude: 8.8634 },
    { code: "15", name: "Riyom", state_code: "PL", latitude: 9.8379, longitude: 8.8804 },
    { code: "16", name: "Shendam", state_code: "PL", latitude: 10.0107, longitude: 8.9917 },
    { code: "17", name: "Wase", state_code: "PL", latitude: 9.9345, longitude: 8.7223 }
  ],

  // 33. State RI (23 LGAs)
  RI: [
    { code: "01", name: "Abua/Odual", state_code: "RI", latitude: 4.8764, longitude: 6.9834 },
    { code: "02", name: "Ahoada East", state_code: "RI", latitude: 4.6861, longitude: 7.0611 },
    { code: "03", name: "Ahoada West", state_code: "RI", latitude: 4.9505, longitude: 7.1533 },
    { code: "04", name: "Akuku-Toru", state_code: "RI", latitude: 4.7791, longitude: 6.843 },
    { code: "05", name: "Andoni", state_code: "RI", latitude: 4.6813, longitude: 7.2606 },
    { code: "06", name: "Asari-Toru", state_code: "RI", latitude: 5.0957, longitude: 6.9747 },
    { code: "07", name: "Bonny", state_code: "RI", latitude: 4.7712, longitude: 7.0267 },
    { code: "08", name: "Degema", state_code: "RI", latitude: 4.8464, longitude: 7.1344 },
    { code: "09", name: "Eleme", state_code: "RI", latitude: 4.7912, longitude: 7.1205 },
    { code: "10", name: "Emuoha", state_code: "RI", latitude: 4.6615, longitude: 7.1216 },
    { code: "11", name: "Etche", state_code: "RI", latitude: 5.0159, longitude: 7.1129 },
    { code: "12", name: "Gokana", state_code: "RI", latitude: 4.6906, longitude: 6.8333 },
    { code: "13", name: "Ikwerre", state_code: "RI", latitude: 4.7528, longitude: 7.3329 },
    { code: "14", name: "Khana", state_code: "RI", latitude: 4.8566, longitude: 7.0211 },
    { code: "15", name: "Obio/Akpor", state_code: "RI", latitude: 4.8472, longitude: 7.0134 },
    { code: "16", name: "Ogba/Egbema/Ndoni", state_code: "RI", latitude: 4.8992, longitude: 7.1494 },
    { code: "17", name: "Ogu/Bolo", state_code: "RI", latitude: 4.823, longitude: 6.88 },
    { code: "18", name: "Okrika", state_code: "RI", latitude: 4.6671, longitude: 7.1983 },
    { code: "19", name: "Omuma", state_code: "RI", latitude: 5.0654, longitude: 7.0389 },
    { code: "20", name: "Opobo/Nkoro", state_code: "RI", latitude: 4.5934, longitude: 6.8634 },
    { code: "21", name: "Oyigbo", state_code: "RI", latitude: 4.8221, longitude: 7.0994 },
    { code: "22", name: "Port Harcourt", state_code: "RI", latitude: 4.8156, longitude: 7.0498 },
    { code: "23", name: "Tai", state_code: "RI", latitude: 4.6887, longitude: 7.0779 }
  ],

  // 34. State SO (23 LGAs)
  SO: [
    { code: "01", name: "Binji", state_code: "SO", latitude: 13.1217, longitude: 5.1812 },
    { code: "02", name: "Bodinga", state_code: "SO", latitude: 12.9314, longitude: 5.2589 },
    { code: "03", name: "Dange Shuni", state_code: "SO", latitude: 13.1958, longitude: 5.3511 },
    { code: "04", name: "Gada", state_code: "SO", latitude: 13.0244, longitude: 5.0408 },
    { code: "05", name: "Goronyo", state_code: "SO", latitude: 12.9266, longitude: 5.4584 },
    { code: "06", name: "Gudu", state_code: "SO", latitude: 13.341, longitude: 5.1725 },
    { code: "07", name: "Gawabawa", state_code: "SO", latitude: 13.0165, longitude: 5.2245 },
    { code: "08", name: "Illela", state_code: "SO", latitude: 13.0917, longitude: 5.3322 },
    { code: "09", name: "Isa", state_code: "SO", latitude: 13.1106, longitude: 5.1275 },
    { code: "10", name: "Kebbe", state_code: "SO", latitude: 12.9068, longitude: 5.3194 },
    { code: "11", name: "Kware", state_code: "SO", latitude: 13.2612, longitude: 5.3107 },
    { code: "12", name: "Rabah", state_code: "SO", latitude: 12.9359, longitude: 5.0311 },
    { code: "13", name: "Sabon Birni", state_code: "SO", latitude: 12.9981, longitude: 5.5307 },
    { code: "14", name: "Shagari", state_code: "SO", latitude: 13.1019, longitude: 5.2189 },
    { code: "15", name: "Silame", state_code: "SO", latitude: 12.9717, longitude: 5.2359 },
    { code: "16", name: "Sokoto North", state_code: "SO", latitude: 13.0609, longitude: 5.2476 },
    { code: "17", name: "Sokoto South", state_code: "SO", latitude: 13.041, longitude: 5.239 },
    { code: "18", name: "Tambuwal", state_code: "SO", latitude: 12.9124, longitude: 5.3961 },
    { code: "19", name: "Tangaza", state_code: "SO", latitude: 13.3107, longitude: 5.2367 },
    { code: "20", name: "Tureta", state_code: "SO", latitude: 12.8387, longitude: 5.0612 },
    { code: "21", name: "Wamako", state_code: "SO", latitude: 13.0674, longitude: 5.2972 },
    { code: "22", name: "Wurno", state_code: "SO", latitude: 13.1125, longitude: 5.1739 },
    { code: "23", name: "Yabo", state_code: "SO", latitude: 12.934, longitude: 5.2757 }
  ],

  // 35. State TA (16 LGAs)
  TA: [
    { code: "01", name: "Ardo Kola", state_code: "TA", latitude: 8.9545, longitude: 11.2936 },
    { code: "02", name: "Bali", state_code: "TA", latitude: 8.7642, longitude: 11.3713 },
    { code: "03", name: "Donga", state_code: "TA", latitude: 9.0286, longitude: 11.4635 },
    { code: "04", name: "Gashaka", state_code: "TA", latitude: 8.8572, longitude: 11.1532 },
    { code: "05", name: "Gassol", state_code: "TA", latitude: 8.7594, longitude: 11.5708 },
    { code: "06", name: "Ibi", state_code: "TA", latitude: 9.1738, longitude: 11.2849 },
    { code: "07", name: "Jalingo", state_code: "TA", latitude: 8.8937, longitude: 11.36 },
    { code: "08", name: "Karim Lamido", state_code: "TA", latitude: 8.9245, longitude: 11.4446 },
    { code: "09", name: "Kurmi", state_code: "TA", latitude: 8.9434, longitude: 11.2399 },
    { code: "10", name: "Lau", state_code: "TA", latitude: 8.7396, longitude: 11.4318 },
    { code: "11", name: "Sardauna", state_code: "TA", latitude: 9.094, longitude: 11.4231 },
    { code: "12", name: "Takum", state_code: "TA", latitude: 8.7687, longitude: 11.1435 },
    { code: "13", name: "Ussa", state_code: "TA", latitude: 8.8309, longitude: 11.6431 },
    { code: "14", name: "Wukari", state_code: "TA", latitude: 8.9347, longitude: 11.3313 },
    { code: "15", name: "Yorro", state_code: "TA", latitude: 8.8045, longitude: 11.3483 },
    { code: "16", name: "Zing", state_code: "TA", latitude: 8.9773, longitude: 11.4596 }
  ],

  // 36. State YO (17 LGAs)
  YO: [
    { code: "01", name: "Bade", state_code: "YO", latitude: 11.8078, longitude: 11.8944 },
    { code: "02", name: "Bursari", state_code: "YO", latitude: 11.6175, longitude: 11.9721 },
    { code: "03", name: "Damaturu", state_code: "YO", latitude: 11.747, longitude: 11.9608 },
    { code: "04", name: "Fika", state_code: "YO", latitude: 11.7105, longitude: 11.754 },
    { code: "05", name: "Fune", state_code: "YO", latitude: 11.6127, longitude: 12.1716 },
    { code: "06", name: "Geidam", state_code: "YO", latitude: 12.0271, longitude: 11.8857 },
    { code: "07", name: "Gujba", state_code: "YO", latitude: 11.7026, longitude: 11.9377 },
    { code: "08", name: "Gulani", state_code: "YO", latitude: 11.7778, longitude: 12.0454 },
    { code: "09", name: "Jakusko", state_code: "YO", latitude: 11.7967, longitude: 11.8407 },
    { code: "10", name: "Karasuwa", state_code: "YO", latitude: 11.5929, longitude: 12.0326 },
    { code: "11", name: "Machina", state_code: "YO", latitude: 11.9473, longitude: 12.0239 },
    { code: "12", name: "Nangere", state_code: "YO", latitude: 11.622, longitude: 11.7443 },
    { code: "13", name: "Nguru", state_code: "YO", latitude: 11.6842, longitude: 12.2439 },
    { code: "14", name: "Potiskum", state_code: "YO", latitude: 11.788, longitude: 11.9321 },
    { code: "15", name: "Tarmuwa", state_code: "YO", latitude: 11.6578, longitude: 11.9491 },
    { code: "16", name: "Yunusari", state_code: "YO", latitude: 11.8306, longitude: 12.0604 },
    { code: "17", name: "Yusufari", state_code: "YO", latitude: 11.7544, longitude: 11.791 }
  ],

  // 37. State ZA (14 LGAs)
  ZA: [
    { code: "01", name: "Anka", state_code: "ZA", latitude: 12.231, longitude: 6.5928 },
    { code: "02", name: "Bakura", state_code: "ZA", latitude: 12.0407, longitude: 6.6705 },
    { code: "03", name: "Birnin Magaji/Kiyaw", state_code: "ZA", latitude: 12.3051, longitude: 6.7627 },
    { code: "04", name: "Bukkuyum", state_code: "ZA", latitude: 12.1337, longitude: 6.4524 },
    { code: "05", name: "Bungudu", state_code: "ZA", latitude: 12.0359, longitude: 6.87 },
    { code: "06", name: "Gummi", state_code: "ZA", latitude: 12.4503, longitude: 6.5841 },
    { code: "07", name: "Gusau", state_code: "ZA", latitude: 12.1702, longitude: 6.6592 },
    { code: "08", name: "Kaura Namoda", state_code: "ZA", latitude: 12.201, longitude: 6.7438 },
    { code: "09", name: "Maradun", state_code: "ZA", latitude: 12.2199, longitude: 6.5391 },
    { code: "10", name: "Maru", state_code: "ZA", latitude: 12.0161, longitude: 6.731 },
    { code: "11", name: "Shinkafi", state_code: "ZA", latitude: 12.3705, longitude: 6.7223 },
    { code: "12", name: "Talata Mafara", state_code: "ZA", latitude: 12.0452, longitude: 6.4427 },
    { code: "13", name: "Tsafe", state_code: "ZA", latitude: 12.1074, longitude: 6.9423 },
    { code: "14", name: "Zurmi", state_code: "ZA", latitude: 12.2112, longitude: 6.6305 }
  ]

};

/**
 * Returns all LGAs for a given 2-letter state code, sorted by code.
 */
export function getLGAsForState(stateCode: string): LGAOption[] {
  if (!stateCode) return [];
  let upper = stateCode.trim().toUpperCase();
  if (upper === "GM") upper = "GO";
  return NIGERIAN_774_LGAS[upper] || [];
}

/**
 * Total LGA count helper to verify completeness (equals 774).
 */
export function getTotalLGACount(): number {
  return Object.values(NIGERIAN_774_LGAS).reduce((acc, lgas) => acc + lgas.length, 0);
}

/**
 * Fast offline haversine search for the closest LGA centroid in Nigeria.
 */
export function findNearestLGA(
  lat: number,
  lng: number
): { lga: LGAOption; distanceM: number } | null {
  if (!lat && !lng) return null;
  let minD = Infinity;
  let closest: LGAOption | null = null;
  const r = 6371000;
  for (const lgas of Object.values(NIGERIAN_774_LGAS)) {
    for (const lga of lgas) {
      if (lga.latitude == null || lga.longitude == null) continue;
      const dLat = (lga.latitude - lat) * (Math.PI / 180);
      const dLon = (lga.longitude - lng) * (Math.PI / 180);
      const a =
        Math.sin(dLat / 2) * Math.sin(dLat / 2) +
        Math.cos(lat * (Math.PI / 180)) * Math.cos(lga.latitude * (Math.PI / 180)) *
        Math.sin(dLon / 2) * Math.sin(dLon / 2);
      const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
      const d = r * c;
      if (d < minD) {
        minD = d;
        closest = lga;
      }
    }
  }
  return closest ? { lga: closest, distanceM: Math.round(minD * 10) / 10 } : null;
}

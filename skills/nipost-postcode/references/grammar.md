# Nigerian Postcode Grammar & State Code Reference

## 1. Grammatical Specification

Nigeria's National Digital Alphanumeric Postcode system represents physical delivery points with an immutable 11-character identifier.

### Segment Decomposition

```text
 Canonical:   L A - 1 1 - W 0 6 - T C - 1 0
 Compact:     L A 1 1 W 0 6 T C 1 0
 Components: [State][LGA][District][Area][Unit]
 Indices:     0..1   2..3  4..6     7..8  9..10
```

- **Segment 1: State Code (`SS`)**
  - Length: 2 characters.
  - Characters: Uppercase ASCII alphabetic (`[A-Z]{2}`).
  - Semantics: Identifies one of the 36 states or the Federal Capital Territory (`FC`).
- **Segment 2: Local Government Area (`LL`)**
  - Length: 2 characters.
  - Characters: Decimal digits (`[0-9]{2}`).
  - Valid Range: `01` through `99`. **`00` is strictly illegal.**
- **Segment 3: Postal District (`DDD`)**
  - Length: 3 characters.
  - Characters: Alphanumeric uppercase (`[A-Z0-9]{3}`).
  - Semantics: Cadastral delivery district within the LGA.
- **Segment 4: Delivery Area (`AA`)**
  - Length: 2 characters.
  - Characters: Uppercase ASCII alphabetic (`[A-Z]{2}`).
  - Semantics: Sub-district neighborhood or postal zone.
- **Segment 5: Building Unit (`UU`)**
  - Length: 2 characters.
  - Characters: Decimal digits (`[0-9]{2}`).
  - Valid Range: `01` through `99`. **`00` is strictly illegal.**

---

## 2. Recognized State Codes (37 Entities)

| Code | State Name | Geopolitical Zone | Capital |
| :--- | :--- | :--- | :--- |
| **AB** | Abia | South East | Umuahia |
| **AD** | Adamawa | North East | Yola |
| **AK** | Akwa Ibom | South South | Uyo |
| **AN** | Anambra | South East | Awka |
| **BA** | Bauchi | North East | Bauchi |
| **BY** | Bayelsa | South South | Yenagoa |
| **BE** | Benue | North Central | Makurdi |
| **BO** | Borno | North East | Maiduguri |
| **CR** | Cross River | South South | Calabar |
| **DE** | Delta | South South | Asaba |
| **EB** | Ebonyi | South East | Abakaliki |
| **ED** | Edo | South South | Benin City |
| **EK** | Ekiti | South West | Ado-Ekiti |
| **EN** | Enugu | South East | Enugu |
| **FC** | Federal Capital Territory (FCT) | North Central | Abuja |
| **GO** | Gombe | North East | Gombe |
| **IM** | Imo | South East | Owerri |
| **JI** | Jigawa | North West | Dutse |
| **KD** | Kaduna | North West | Kaduna |
| **KN** | Kano | North West | Kano |
| **KT** | Katsina | North West | Katsina |
| **KE** | Kebbi | North West | Birnin Kebbi |
| **KO** | Kogi | North Central | Lokoja |
| **KW** | Kwara | North Central | Ilorin |
| **LA** | Lagos | South West | Ikeja |
| **NA** | Nasarawa | North Central | Lafia |
| **NI** | Niger | North Central | Minna |
| **OG** | Ogun | South West | Abeokuta |
| **ON** | Ondo | South West | Akure |
| **OS** | Osun | South West | Osogbo |
| **OY** | Oyo | South West | Ibadan |
| **PL** | Plateau | North Central | Jos |
| **RI** | Rivers | South South | Port Harcourt |
| **SO** | Sokoto | North West | Sokoto |
| **TA** | Taraba | North East | Jalingo |
| **YO** | Yobe | North East | Damaturu |
| **ZA** | Zamfara | North West | Gusau |

---

## 3. Regular Expression Patterns

- **Canonical Hyphenated Pattern**:
  ```regex
  ^[A-Z]{2}-[0-9]{2}-[A-Z0-9]{3}-[A-Z]{2}-[0-9]{2}$
  ```
- **Compact Pattern**:
  ```regex
  ^[A-Z]{2}[0-9]{2}[A-Z0-9]{3}[A-Z]{2}[0-9]{2}$
  ```
- **Flexible Normalization**:
  Any input matching `[A-Za-z0-9]{11}` after stripping spaces, hyphens, and whitespace can be tested for validity.

---

## 4. Legacy vs Modern Postal Codes

| Feature | Legacy System (Pre-2023) | Modern Digital System (Current) |
| :--- | :--- | :--- |
| **Format** | 6 numeric digits (e.g., `100001`) | 11 alphanumeric characters (`LA-11-W06-TC-10`) |
| **Granularity** | Broad general post office | Specific building / delivery point unit |
| **Cadastral Link** | None (paper routes) | Exact GPS centroid and polygon boundary |
| **Status** | Phased out by NIPOST | Official National Standard |

#!/usr/bin/env node
const fs = require('fs');
const path = require('path');

const goLocPath = path.resolve(__dirname, '../location_data.go');
const outPath = path.resolve(__dirname, '../packages/address-input/src/data/lgas.ts');

if (!fs.existsSync(goLocPath)) {
  console.error(`Error: location_data.go not found at ${goLocPath}`);
  process.exit(1);
}

const goLoc = fs.readFileSync(goLocPath, 'utf8');
const goRegex = /"([A-Z]{2})(\d{2})":\s*\{StateCode:\s*"([A-Z]{2})",\s*LGACode:\s*"(\d{2})",\s*Name:\s*"([^"]+)",\s*Latitude:\s*([\d\.-]+),\s*Longitude:\s*([\d\.-]+)\}/g;

const lgasByState = {};
let match;
let count = 0;

while ((match = goRegex.exec(goLoc)) !== null) {
  const st = match[1];
  lgasByState[st] = lgasByState[st] || [];
  lgasByState[st].push({
    code: match[2],
    name: match[5],
    state_code: st,
    latitude: parseFloat(match[6]),
    longitude: parseFloat(match[7]),
  });
  count++;
}

if (count !== 774) {
  console.error(`Warning: Expected 774 LGAs from location_data.go, but extracted ${count}`);
}

const stateOrder = [
  'AB', 'AD', 'AK', 'AN', 'BA', 'BY', 'BN', 'BO', 'CR', 'DE',
  'EB', 'ED', 'EK', 'EN', 'FC', 'GO', 'IM', 'JI', 'KD', 'KN',
  'KT', 'KE', 'KO', 'KW', 'LA', 'NA', 'NI', 'OG', 'ON', 'OS',
  'OY', 'PL', 'RI', 'SO', 'TA', 'YO', 'ZA',
];

let out = `import type { LGAOption } from "../types";\n\n`;
out += `/**\n * Exhaustive directory of all 774 Local Government Areas (LGAs) and Area Councils\n`;
out += ` * in the Federal Republic of Nigeria, across all 36 States and the Federal Capital Territory (FCT).\n`;
out += ` * Centroid coordinates (latitude and longitude) are derived from the official NIPOST cadastral database.\n */\n`;
out += `export const NIGERIAN_774_LGAS: Record<string, LGAOption[]> = {\n`;

for (let i = 0; i < stateOrder.length; i++) {
  const st = stateOrder[i];
  const list = lgasByState[st] || [];
  list.sort((a, b) => a.code.localeCompare(b.code));
  out += `  // ${i + 1}. State ${st} (${list.length} LGAs)\n`;
  out += `  ${st}: [\n`;
  for (let j = 0; j < list.length; j++) {
    const l = list[j];
    const comma = j < list.length - 1 ? ',' : '';
    out += `    { code: "${l.code}", name: ${JSON.stringify(l.name)}, state_code: "${l.state_code}", latitude: ${l.latitude}, longitude: ${l.longitude} }${comma}\n`;
  }
  const stComma = i < stateOrder.length - 1 ? ',' : '';
  out += `  ]${stComma}\n\n`;
}

out += `};\n\n`;
out += `/**\n * Returns all LGAs for a given 2-letter state code, sorted by code.\n */\n`;
out += `export function getLGAsForState(stateCode: string): LGAOption[] {\n`;
out += `  if (!stateCode) return [];\n`;
out += `  let upper = stateCode.trim().toUpperCase();\n`;
out += `  if (upper === "GM") upper = "GO";\n`;
out += `  return NIGERIAN_774_LGAS[upper] || [];\n`;
out += `}\n\n`;
out += `/**\n * Total LGA count helper to verify completeness (equals 774).\n */\n`;
out += `export function getTotalLGACount(): number {\n`;
out += `  return Object.values(NIGERIAN_774_LGAS).reduce((acc, lgas) => acc + lgas.length, 0);\n`;
out += `}\n\n`;
out += `/**\n * Fast offline haversine search for the closest LGA centroid in Nigeria.\n */\n`;
out += `export function findNearestLGA(\n`;
out += `  lat: number,\n`;
out += `  lng: number\n`;
out += `): { lga: LGAOption; distanceM: number } | null {\n`;
out += `  if (!lat && !lng) return null;\n`;
out += `  let minD = Infinity;\n`;
out += `  let closest: LGAOption | null = null;\n`;
out += `  const r = 6371000;\n`;
out += `  for (const lgas of Object.values(NIGERIAN_774_LGAS)) {\n`;
out += `    for (const lga of lgas) {\n`;
out += `      if (lga.latitude == null || lga.longitude == null) continue;\n`;
out += `      const dLat = (lga.latitude - lat) * (Math.PI / 180);\n`;
out += `      const dLon = (lga.longitude - lng) * (Math.PI / 180);\n`;
out += `      const a =\n`;
out += `        Math.sin(dLat / 2) * Math.sin(dLat / 2) +\n`;
out += `        Math.cos(lat * (Math.PI / 180)) * Math.cos(lga.latitude * (Math.PI / 180)) *\n`;
out += `        Math.sin(dLon / 2) * Math.sin(dLon / 2);\n`;
out += `      const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));\n`;
out += `      const d = r * c;\n`;
out += `      if (d < minD) {\n`;
out += `        minD = d;\n`;
out += `        closest = lga;\n`;
out += `      }\n`;
out += `    }\n`;
out += `  }\n`;
out += `  return closest ? { lga: closest, distanceM: Math.round(minD * 10) / 10 } : null;\n`;
out += `}\n`;

fs.writeFileSync(outPath, out);
console.log(`Successfully generated ${outPath} from location_data.go (${count} LGAs).`);

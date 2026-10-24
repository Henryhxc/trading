"""Generate transactional metadata-only SQL from the HKEX securities list.

Usage: python3 tools/import_hk_directory.py path.xlsx | mysql database
The source date and checksum are printed to stderr for the evidence ledger.
"""
import argparse
import datetime
import hashlib
import json
import re
import sys
import xml.etree.ElementTree as ET
import zipfile
from pathlib import Path

SOURCE = "https://www.hkex.com.hk/eng/services/trading/securities/securitieslists/ListOfSecurities.xlsx"
NS = {"x": "http://schemas.openxmlformats.org/spreadsheetml/2006/main"}


def parse_workbook(path):
    with zipfile.ZipFile(path) as archive:
        strings = []
        if "xl/sharedStrings.xml" in archive.namelist():
            root = ET.fromstring(archive.read("xl/sharedStrings.xml"))
            strings = ["".join(item.itertext()) for item in root.findall("x:si", NS)]
        sheet = ET.fromstring(archive.read("xl/worksheets/sheet1.xml"))
    rows = []
    for row in sheet.findall("x:sheetData/x:row", NS):
        cells = {}
        for cell in row.findall("x:c", NS):
            column = re.match(r"[A-Z]+", cell.attrib["r"]).group()
            value = cell.findtext("x:v", "", NS)
            if cell.get("t") == "s":
                value = strings[int(value)]
            elif cell.get("t") == "inlineStr":
                value = "".join(cell.find("x:is", NS).itertext())
            cells[column] = value.strip()
        rows.append(cells)
    date_match = re.search(r"\b(\d{2}/\d{2}/\d{4})\b", " ".join(value for row in rows[:3] for value in row.values()))
    if not date_match:
        raise ValueError("The source declaration date is missing")
    declared = datetime.datetime.strptime(date_match[1], "%d/%m/%Y").date().isoformat()
    header_index = next(index for index, row in enumerate(rows) if "Stock Code" in row.values())
    headers = {value: column for column, value in rows[header_index].items()}
    required = ("Stock Code", "Name of Securities", "Category", "Board Lot", "Trading Currency")
    if not all(key in headers for key in required):
        raise ValueError("Unexpected source columns")
    items = {}
    for row in rows[header_index + 1:]:
        values = {key: row.get(headers[key], "") for key in required}
        if values["Category"] != "Equity" or values["Trading Currency"] != "HKD":
            continue
        raw_symbol = values["Stock Code"]
        if not re.fullmatch(r"\d{1,5}", raw_symbol):
            raise ValueError("Invalid equity symbol")
        symbol = raw_symbol.zfill(5)
        lot_size = int(values["Board Lot"].replace(",", ""))
        name = values["Name of Securities"]
        if not name or len(name) > 128 or lot_size < 1 or symbol in items:
            raise ValueError("Invalid or duplicate equity metadata")
        items[symbol] = {"symbol": symbol, "name": name, "lot_size": lot_size}
    if not items:
        raise ValueError("The source contains no HKD equities")
    return declared, [items[key] for key in sorted(items)]


def metadata_sql(rows):
    statements = ["SET NAMES utf8mb4;", "START TRANSACTION;"]
    for row in rows:
        name_hex = row["name"].encode("utf-8").hex()
        statements.append(
            "INSERT INTO stocks (symbol,name,lot_size,data_status,updated_at) "
            f"VALUES ('{row['symbol']}',CONVERT(0x{name_hex} USING utf8mb4),{row['lot_size']},'unavailable',UTC_TIMESTAMP(3)) "
            "ON DUPLICATE KEY UPDATE name=VALUES(name),lot_size=VALUES(lot_size);"
        )
    statements.append("COMMIT;")
    return "\n".join(statements)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("workbook", type=Path)
    args = parser.parse_args()
    declared, rows = parse_workbook(args.workbook)
    report = {"source_url": SOURCE, "source_declared_date": declared,
              "source_sha256": hashlib.sha256(args.workbook.read_bytes()).hexdigest(),
              "hkd_equities": len(rows), "metadata_only": True,
              "future_dated_source": declared > datetime.date.today().isoformat()}
    print(json.dumps(report, ensure_ascii=False), file=sys.stderr)
    print(metadata_sql(rows))


if __name__ == "__main__":
    main()

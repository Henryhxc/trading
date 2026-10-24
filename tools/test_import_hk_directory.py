import io
import unittest
import zipfile
from import_hk_directory import parse_workbook


class DirectoryImportTests(unittest.TestCase):
    def test_sparse_cells_and_shared_names_filter_only_hkd_equities(self):
        source = io.BytesIO()
        with zipfile.ZipFile(source, "w") as archive:
            archive.writestr("xl/sharedStrings.xml", '<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><si><t>TEST HOLDINGS</t></si></sst>')
            archive.writestr("xl/worksheets/sheet1.xml", '''<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>
            <row r="2"><c r="A2" t="inlineStr"><is><t>Updated as at 26/10/2026</t></is></c></row>
            <row r="3"><c r="A3" t="inlineStr"><is><t>Stock Code</t></is></c><c r="B3" t="inlineStr"><is><t>Name of Securities</t></is></c><c r="C3" t="inlineStr"><is><t>Category</t></is></c><c r="E3" t="inlineStr"><is><t>Board Lot</t></is></c><c r="Q3" t="inlineStr"><is><t>Trading Currency</t></is></c></row>
            <row r="4"><c r="A4"><v>1</v></c><c r="B4" t="s"><v>0</v></c><c r="C4" t="inlineStr"><is><t>Equity</t></is></c><c r="E4"><v>500</v></c><c r="Q4" t="inlineStr"><is><t>HKD</t></is></c></row>
            <row r="5"><c r="A5"><v>80001</v></c><c r="C5" t="inlineStr"><is><t>Equity</t></is></c><c r="Q5" t="inlineStr"><is><t>RMB</t></is></c></row>
            <row r="6"><c r="A6"><v>10001</v></c><c r="C6" t="inlineStr"><is><t>Derivative Warrants</t></is></c><c r="Q6" t="inlineStr"><is><t>HKD</t></is></c></row>
            </sheetData></worksheet>''')
        source.seek(0)
        declared_date, rows = parse_workbook(source)
        self.assertEqual(declared_date, "2026-10-26")
        self.assertEqual(rows, [{"symbol": "00001", "name": "TEST HOLDINGS", "lot_size": 500}])


if __name__ == "__main__":
    unittest.main()

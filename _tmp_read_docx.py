import zipfile
import re
import sys


def docx_text(path: str) -> str:
    with zipfile.ZipFile(path) as z:
        xml = z.read("word/document.xml").decode("utf-8")
    t = re.sub(r"<w:tab[^>]*/>", "\t", xml)
    t = re.sub(r"</w:p>", "\n", t)
    t = re.sub(r"<[^>]+>", "", t)
    for a, b in [("&lt;", "<"), ("&gt;", ">"), ("&amp;", "&")]:
        t = t.replace(a, b)
    return t


def main():
    for p in sys.argv[1:]:
        print("===", p, "===")
        print(docx_text(p))


if __name__ == "__main__":
    main()

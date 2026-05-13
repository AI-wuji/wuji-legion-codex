import json, os, re, sys, datetime

ERRORS_DIR = ".wuji-errors"
ERRORS_FILE = os.path.join(ERRORS_DIR, "ERRORS.md")

def ensure_db():
    os.makedirs(ERRORS_DIR, exist_ok=True)
    if not os.path.exists(ERRORS_FILE):
        with open(ERRORS_FILE, "w", encoding="utf-8") as f:
            f.write("# Wuji Error DNA Database\n\n")

def load_entries():
    ensure_db()
    with open(ERRORS_FILE, "r", encoding="utf-8") as f:
        content = f.read()
    entries = re.findall(r"## (\d+)\. \[(.+?)\] (.+?)\n- \*\*涉及文件\*\*: (.+?)\n- \*\*现象\*\*: (.+?)\n- \*\*根因\*\*: (.+?)\n- \*\*修复方法\*\*: (.+?)\n- \*\*修复验证\*\*: (.+?)\n- \*\*相关文件\*\*: (.+?)\n- \*\*错误签名\*\*: (.+?)\n- \*\*日期\*\*: (.+?)\n", content)
    return entries, content

def add_entry(category, summary, files, symptom, root_cause, fix, verification, related_files, signature):
    entries, content = load_entries()
    next_num = len(entries) + 1
    today = datetime.date.today().strftime("%Y-%m-%d")
    new_entry = "\n## %d. [%s] %s\n- **涉及文件**: %s\n- **现象**: %s\n- **根因**: %s\n- **修复方法**: %s\n- **修复验证**: %s\n- **相关文件**: %s\n- **错误签名**: %s\n- **日期**: %s\n" % (
        next_num, category, summary, files, symptom, root_cause, fix, verification, related_files, signature, today)
    with open(ERRORS_FILE, "a", encoding="utf-8") as f:
        f.write(new_entry)
    print("[OK] Error DNA recorded (## %d)" % next_num)

def check_file(target_file):
    entries, _ = load_entries()
    target_lower = target_file.lower()
    matches = []
    for entry in entries:
        if target_lower in entry[3].lower() or target_lower in entry[8].lower():
            matches.append(entry)
    return matches

def search_by_pattern(keywords):
    entries, _ = load_entries()
    kw_list = [k.lower() for k in keywords]
    matches = []
    for entry in entries:
        entry_text = " ".join(str(f) for f in entry).lower()
        if any(kw in entry_text for kw in kw_list):
            matches.append(entry)
    return matches

def similarity_check(new_signature):
    entries, _ = load_entries()
    new_kws = set(new_signature.lower().split())
    results = []
    for entry in entries:
        num, cat, summary, files, _, _, _, _, _, old_sig, date = entry
        old_kws = set(old_sig.lower().split())
        if len(new_kws) > 0 and len(old_kws) > 0:
            overlap = len(new_kws & old_kws)
            union = len(new_kws | old_kws)
            sim = overlap / union if union > 0 else 0
            if sim >= 0.3:
                results.append((sim, num, cat, summary, date))
    return sorted(results, key=lambda x: -x[0])

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python errors_db.py <add|check|search|list|dedup> [args]")
        sys.exit(1)
    cmd = sys.argv[1]
    if cmd == "add":
        if len(sys.argv) < 10:
            print("Usage: add <category> <summary> <files> <symptom> <root_cause> <fix> <verification> <related_files> <signature>")
            sys.exit(1)
        add_entry(sys.argv[2], sys.argv[3], sys.argv[4], sys.argv[5], sys.argv[6], sys.argv[7], sys.argv[8], sys.argv[9], sys.argv[10] if len(sys.argv) > 10 else "")
    elif cmd == "check":
        if len(sys.argv) < 3:
            print("Usage: check <filename>")
            sys.exit(1)
        matches = check_file(sys.argv[2])
        if matches:
            print("[WARN] Found %d related history entries:" % len(matches))
            for m in matches:
                print("  ## %s. [%s] %s | File: %s | Date: %s" % (m[0], m[1], m[2], m[3], m[10]))
        else:
            print("[OK] No history for '%s'" % sys.argv[2])
    elif cmd == "search":
        if len(sys.argv) < 3:
            print("Usage: search <keyword1> [keyword2 ...]")
            sys.exit(1)
        matches = search_by_pattern(sys.argv[2:])
        if matches:
            print("Found %d matches:" % len(matches))
            for m in matches:
                print("  ## %s. [%s] %s | %s" % (m[0], m[1], m[2], m[10]))
        else:
            print("No matches")
    elif cmd == "dedup":
        if len(sys.argv) < 3:
            print("Need new error signature")
            sys.exit(1)
        results = similarity_check(sys.argv[2])
        if results:
            print("Found %d potential duplicates:" % len(results))
            for sim, num, cat, summary, date in results:
                warn = " [HIGH RISK]" if sim >= 0.7 else ""
                print("  [%d%%] ## %s. [%s] %s (%s)%s" % (sim*100, num, cat, summary, date, warn))
        else:
            print("[OK] No similar history")
    elif cmd == "list":
        entries, _ = load_entries()
        if entries:
            print("Error DNA DB: %d entries" % len(entries))
            for e in entries:
                print("  ## %s. [%s] %s (%s)" % (e[0], e[1], e[2], e[10]))
        else:
            print("DB is empty")
